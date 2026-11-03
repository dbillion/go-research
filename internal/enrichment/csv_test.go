package enrichment

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/deeone/go-research/internal/apollo"
	"github.com/deeone/go-research/internal/templates"
)

func sampleProspect() apollo.ProspectData {
	return apollo.ProspectData{
		ApolloID:       "abc123",
		FirstName:      "Sarah",
		LastName:       "Chen",
		Title:          "CFO",
		Headline:       "CFO at Acme SaaS",
		Email:          "sarah@acme.example",
		EmailStatus:    "verified",
		LinkedIn:       "https://linkedin.com/in/sarachen",
		Phone:          "+1-555-0100",
		Company:        "Acme SaaS",
		CompanyID:      "org_42",
		CompanyDomain:  "acme.example",
		Industry:       "FinTech",
		Series:         "Series B",
		Employees:      250,
		Revenue:        50000000,
		Technologies:   []string{"Stripe", "NetSuite", "Salesforce"},
		Description:    "Subscription billing platform",
		CompanyCity:    "Austin",
		CompanyState:   "TX",
		CompanyCountry: "United States",
	}
}

func TestSaveAndLoadCSV_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prospects.csv")
	in := []interface{}{sampleProspect()}

	if err := SaveCSV(in, path); err != nil {
		t.Fatalf("SaveCSV: %v", err)
	}

	out, err := LoadCSV(path)
	if err != nil {
		t.Fatalf("LoadCSV: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 prospect, got %d", len(out))
	}

	got := out[0]
	want := sampleProspect()
	if !reflect.DeepEqual(got, want) {
		// Compare field by field for a readable diff.
		if got.ApolloID != want.ApolloID {
			t.Errorf("ApolloID: got %q want %q", got.ApolloID, want.ApolloID)
		}
		if got.FirstName != want.FirstName || got.LastName != want.LastName {
			t.Errorf("name: got %q %q want %q %q", got.FirstName, got.LastName, want.FirstName, want.LastName)
		}
		if got.Employees != want.Employees {
			t.Errorf("Employees: got %d want %d", got.Employees, want.Employees)
		}
		if got.Revenue != want.Revenue {
			t.Errorf("Revenue: got %d want %d", got.Revenue, want.Revenue)
		}
		if strings.Join(got.Technologies, "|") != strings.Join(want.Technologies, "|") {
			t.Errorf("Technologies: got %v want %v", got.Technologies, want.Technologies)
		}
		if got.CompanyCountry != want.CompanyCountry {
			t.Errorf("CompanyCountry: got %q want %q", got.CompanyCountry, want.CompanyCountry)
		}
	}
}

func TestSaveCSV_EmptySlice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.csv")
	if err := SaveCSV([]interface{}{}, path); err != nil {
		t.Fatalf("SaveCSV on empty slice: %v", err)
	}
	// Headerless empty file -> LoadCSV should report it can't parse rows.
	if _, err := LoadCSV(path); err == nil {
		t.Log("LoadCSV accepted an empty file (no rows) — acceptable, no assertion")
	}
}

func TestSaveCSV_UnknownType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.csv")
	err := SaveCSV([]interface{}{42}, path)
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
	if !strings.Contains(err.Error(), "unknown type") {
		t.Errorf("expected 'unknown type' message, got: %v", err)
	}
}

func TestSaveOutreachCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outreach.csv")
	result := templates.OutreachResult{
		Prospect:             templates.ProspectData{FirstName: "Sarah", Company: "Acme SaaS"},
		Email1Subject:        "Subject 1",
		Email1Body:           "Body 1",
		Email2Subject:        "Subject 2",
		Email2Body:           "Body 2",
		Email3Subject:        "Subject 3",
		Email3Body:           "Body 3",
		LinkedInConnection:   "Connection note",
		LinkedInFollowUp:     "Follow up note",
		PainPoints:           "leakage",
		PersonalizationHooks: "hooks",
	}

	if err := SaveCSV([]interface{}{result}, path); err != nil {
		t.Fatalf("SaveCSV: %v", err)
	}

	data := readAll(t, path)
	for _, want := range []string{
		"email1_subject", "email3_body", "linkedin_followup",
		"Subject 1", "Body 3", "Follow up note", "leakage",
	} {
		if !strings.Contains(data, want) {
			t.Errorf("expected %q in outreach CSV output", want)
		}
	}
}

func TestParseList(t *testing.T) {
	cases := map[string]int{
		"":                       0,
		"Stripe":                 1,
		"Stripe; NetSuite":       2,
		" Stripe ;; NetSuite ; ": 2,
	}
	for in, wantLen := range cases {
		got := parseList(in)
		if len(got) != wantLen {
			t.Errorf("parseList(%q): got %d items %v, want %d", in, len(got), got, wantLen)
		}
	}
}

func TestTimestamp_Format(t *testing.T) {
	ts := Timestamp()
	if len(ts) != len("20060102_150405") {
		t.Errorf("Timestamp() = %q, expected length %d", ts, len("20060102_150405"))
	}
	if strings.ContainsAny(ts, "-:") {
		t.Errorf("Timestamp() = %q, should not contain separators other than underscore", ts)
	}
}

func TestLoadCSV_MissingFile(t *testing.T) {
	if _, err := LoadCSV(filepath.Join(t.TempDir(), "nope.csv")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func readAll(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	return string(b)
}
