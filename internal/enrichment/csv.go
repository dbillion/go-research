package enrichment

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/deeone/go-research/internal/apollo"
	"github.com/deeone/go-research/internal/templates"
)

func Timestamp() string {
	return time.Now().Format("20060102_150405")
}

func SaveCSV(prospects []interface{}, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if len(prospects) == 0 {
		return nil
	}

	// Write header based on first element type
	switch prospects[0].(type) {
	case apollo.ProspectData:
		return saveProspectCSV(writer, prospects)
	case templates.OutreachResult:
		return saveOutreachCSV(writer, prospects)
	default:
		return fmt.Errorf("unknown type for CSV: %T", prospects[0])
	}
}

func saveProspectCSV(writer *csv.Writer, prospects []interface{}) error {
	header := []string{
		"apollo_id", "first_name", "last_name", "title", "headline",
		"email", "email_status", "linkedin", "phone",
		"company", "company_id", "company_domain", "industry", "series",
		"employees", "revenue", "technologies", "description",
		"company_city", "company_state", "company_country",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, p := range prospects {
		prospect := p.(apollo.ProspectData)
		record := []string{
			prospect.ApolloID,
			prospect.FirstName,
			prospect.LastName,
			prospect.Title,
			prospect.Headline,
			prospect.Email,
			prospect.EmailStatus,
			prospect.LinkedIn,
			prospect.Phone,
			prospect.Company,
			prospect.CompanyID,
			prospect.CompanyDomain,
			prospect.Industry,
			prospect.Series,
			fmt.Sprintf("%d", prospect.Employees),
			fmt.Sprintf("%d", prospect.Revenue),
			strings.Join(prospect.Technologies, "; "),
			prospect.Description,
			prospect.CompanyCity,
			prospect.CompanyState,
			prospect.CompanyCountry,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func saveOutreachCSV(writer *csv.Writer, prospects []interface{}) error {
	header := []string{
		"apollo_id", "first_name", "last_name", "title", "company",
		"email", "linkedin", "phone", "industry", "series", "employees",
		"email1_subject", "email1_body", "email2_subject", "email2_body",
		"email3_subject", "email3_body", "linkedin_connection", "linkedin_followup",
		"pain_points", "personalization_hooks",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, p := range prospects {
		result := p.(templates.OutreachResult)
		record := []string{
			result.Prospect.ApolloID,
			result.Prospect.FirstName,
			result.Prospect.LastName,
			result.Prospect.Title,
			result.Prospect.Company,
			result.Prospect.Email,
			result.Prospect.LinkedIn,
			result.Prospect.Phone,
			result.Prospect.Industry,
			result.Prospect.Series,
			fmt.Sprintf("%d", result.Prospect.Employees),
			result.Email1Subject,
			result.Email1Body,
			result.Email2Subject,
			result.Email2Body,
			result.Email3Subject,
			result.Email3Body,
			result.LinkedInConnection,
			result.LinkedInFollowUp,
			result.PainPoints,
			result.PersonalizationHooks,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func LoadCSV(filepath string) ([]apollo.ProspectData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have header and at least one row")
	}

	// Map header to column index
	header := records[0]
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	var prospects []apollo.ProspectData
	for _, row := range records[1:] {
		if len(row) == 0 {
			continue
		}

		get := func(col string) string {
			if idx, ok := colMap[strings.ToLower(col)]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		prospect := apollo.ProspectData{
			ApolloID:       get("apollo_id"),
			FirstName:      get("first_name"),
			LastName:       get("last_name"),
			Title:          get("title"),
			Headline:       get("headline"),
			Email:          get("email"),
			EmailStatus:    get("email_status"),
			LinkedIn:       get("linkedin"),
			Phone:          get("phone"),
			Company:        get("company"),
			CompanyID:      get("company_id"),
			CompanyDomain:  get("company_domain"),
			Industry:       get("industry"),
			Series:         get("series"),
			Employees:      parseInt(get("employees")),
			Revenue:        parseInt64(get("revenue")),
			Technologies:   parseList(get("technologies")),
			Description:    get("description"),
			CompanyCity:    get("company_city"),
			CompanyState:   get("company_state"),
			CompanyCountry: get("company_country"),
		}
		prospects = append(prospects, prospect)
	}

	return prospects, nil
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseInt64(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ";")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}