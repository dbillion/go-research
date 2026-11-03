package templates

import (
	"strings"
	"testing"
)

func baseProspect() ProspectData {
	return ProspectData{
		FirstName:      "Sarah",
		LastName:       "Chen",
		Title:          "CFO",
		Company:        "Acme SaaS",
		Industry:       "FinTech",
		Series:         "Series B",
		Employees:      250,
		CompanyCity:    "Austin",
		CompanyCountry: "United States",
		Description:    "subscription billing platform",
		Technologies:   []string{"Stripe", "NetSuite", "Salesforce"},
	}
}

// Regression: an empty Technologies slice used to produce an empty/"<nil>"
// stack in the email body instead of a sensible fallback.
func TestGenerate_NoTechnologies(t *testing.T) {
	g := NewGenerator()
	p := baseProspect()
	p.Technologies = nil

	tmpl, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !strings.Contains(tmpl.Email1Body, "modern SaaS stack") {
		t.Errorf("expected fallback tech stack in email body, got:\n%s", tmpl.Email1Body)
	}
	if strings.Contains(tmpl.Email1Body, "<nil>") {
		t.Errorf("email body leaked a nil slice rendering:\n%s", tmpl.Email1Body)
	}
}

func TestGenerate_EmptyTechnologiesSlice(t *testing.T) {
	g := NewGenerator()
	p := baseProspect()
	p.Technologies = []string{}

	tmpl, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !strings.Contains(tmpl.Email1Body, "modern SaaS stack") {
		t.Errorf("expected fallback tech stack, got:\n%s", tmpl.Email1Body)
	}
}

func TestGenerate_TechStackTruncatedToThree(t *testing.T) {
	g := NewGenerator()
	p := baseProspect()
	p.Technologies = []string{"Stripe", "NetSuite", "Salesforce", "HubSpot", "Snowflake"}

	tmpl, err := g.Generate(p)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !strings.Contains(tmpl.Email1Body, "Stripe, NetSuite, Salesforce stack") {
		t.Errorf("expected first three technologies in body, got:\n%s", tmpl.Email1Body)
	}
	if strings.Contains(tmpl.Email1Body, "HubSpot") {
		t.Errorf("expected 4th technology to be truncated, got:\n%s", tmpl.Email1Body)
	}
}

func TestGenerate_AllFieldsPopulated(t *testing.T) {
	g := NewGenerator()

	tmpl, err := g.Generate(baseProspect())
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	blanks := []string{}
	if strings.TrimSpace(tmpl.Email1Subject) == "" {
		blanks = append(blanks, "email1 subject")
	}
	if strings.TrimSpace(tmpl.Email1Body) == "" {
		blanks = append(blanks, "email1 body")
	}
	if strings.TrimSpace(tmpl.Email2Subject) == "" {
		blanks = append(blanks, "email2 subject")
	}
	if strings.TrimSpace(tmpl.Email2Body) == "" {
		blanks = append(blanks, "email2 body")
	}
	if strings.TrimSpace(tmpl.Email3Subject) == "" {
		blanks = append(blanks, "email3 subject")
	}
	if strings.TrimSpace(tmpl.Email3Body) == "" {
		blanks = append(blanks, "email3 body")
	}
	if strings.TrimSpace(tmpl.LinkedInConnection) == "" {
		blanks = append(blanks, "linkedin connection note")
	}
	if strings.TrimSpace(tmpl.LinkedInFollowUp) == "" {
		blanks = append(blanks, "linkedin follow-up note")
	}
	if strings.TrimSpace(tmpl.PainPoints) == "" {
		blanks = append(blanks, "pain points")
	}
	if strings.TrimSpace(tmpl.PersonalizationHooks) == "" {
		blanks = append(blanks, "personalization hooks")
	}
	if len(blanks) > 0 {
		t.Errorf("expected every template field populated, but these are empty: %v", blanks)
	}

	if !strings.Contains(tmpl.Email1Subject, "Sarah") || !strings.Contains(tmpl.Email1Subject, "Acme SaaS") {
		t.Errorf("subject missing personalization: %q", tmpl.Email1Subject)
	}
}

func TestIdentifyPainPoints(t *testing.T) {
	g := NewGenerator()

	expected := []string{
		"Complex billing/subscription management at scale",      // description mentions billing
		"Scaling finance ops without proportional headcount",    // Series B
		"Disconnected systems",                                  // Stripe/NetSuite/Salesforce
		"Revenue recognition compliance",                        // FinTech
		"Manual revenue operations processes breaking at scale", // 250 employees
	}

	pains := g.identifyPainPoints(baseProspect())
	for _, want := range expected {
		if !strings.Contains(pains, want) {
			t.Errorf("expected %q in pain points, got: %s", want, pains)
		}
	}
}

func TestIdentifyPainPoints_FallbackWhenNoSignals(t *testing.T) {
	g := NewGenerator()
	p := ProspectData{
		FirstName: "Alex",
		Company:   "Some Corp",
		Industry:  "Retail",
		Series:    "Series D",
		Employees: 2000,
	}

	pains := g.identifyPainPoints(p)
	if !strings.Contains(pains, "Forrester") {
		t.Errorf("expected fallback pain point, got: %s", pains)
	}
}

func TestGenerateHooks_OmitsUnknownFields(t *testing.T) {
	g := NewGenerator()
	p := ProspectData{
		FirstName: "Alex",
		LastName:  "Doe",
		Title:     "VP Finance",
		Company:   "Minimal Inc",
	}

	hooks := g.generateHooks(p)
	if !strings.Contains(hooks, "Alex Doe, VP Finance at Minimal Inc") {
		t.Errorf("expected name hook, got: %s", hooks)
	}
	// No Series / Employees / Tech / City set -> those segments must not appear.
	for _, unwanted := range []string{"-backed", "employees", "Stack:", "Based in"} {
		if strings.Contains(hooks, unwanted) {
			t.Errorf("hook %q should not be present for sparse prospect: %s", unwanted, hooks)
		}
	}
}

func TestGenerateForBatch(t *testing.T) {
	g := NewGenerator()
	prospects := []ProspectData{
		baseProspect(),
		{FirstName: "Bob", Company: "Beta LLC", Industry: "SaaS", Series: "Series C", Employees: 120},
	}

	results, err := g.GenerateForBatch(prospects)
	if err != nil {
		t.Fatalf("GenerateForBatch returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Prospect.Company != prospects[i].Company {
			t.Errorf("result %d prospect mismatch: got %s", i, r.Prospect.Company)
		}
		if strings.TrimSpace(r.Email1Body) == "" {
			t.Errorf("result %d has empty email body", i)
		}
	}
}

func TestGenerateForBatch_Empty(t *testing.T) {
	g := NewGenerator()
	results, err := g.GenerateForBatch(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}
