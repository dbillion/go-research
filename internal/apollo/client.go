package apollo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	baseURL = "https://api.apollo.io/v1"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	limiter    *rate.Limiter
	mu         sync.Mutex
}

func NewClient(apiKey string, requestsPerMin int) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(requestsPerMin)), requestsPerMin),
	}
}

type SearchParams struct {
	Roles       []string
	Series      []string
	Industries  []string
	MinEmployees int
	MaxEmployees int
	Locations   []string
	Limit       int
}

type Person struct {
	ID              string   `json:"id"`
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Name            string   `json:"name"`
	Title           string   `json:"title"`
	Headline        string   `json:"headline"`
	Email           string   `json:"email"`
	EmailStatus     string   `json:"email_status"`
	LinkedInURL     string   `json:"linkedin_url"`
	PhoneNumber     string   `json:"phone_number"`
	OrganizationID  string   `json:"organization_id"`
	OrganizationName string  `json:"organization_name"`
	Department      string   `json:"department"`
	Seniority       string   `json:"seniority"`
	Departments     []string `json:"departments"`
	Functions       []string `json:"functions"`
	Country         string   `json:"country"`
	State           string   `json:"state"`
	City            string   `json:"city"`
}

type Organization struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	WebsiteURL      string   `json:"website_url"`
	Industry        string   `json:"industry"`
	IndustryTag     string   `json:"industry_tag"`
	EmployeeCount   int      `json:"employee_count"`
	AnnualRevenue   int64    `json:"annual_revenue"`
	FoundedYear     int      `json:"founded_year"`
	FundingStage    string   `json:"funding_stage"`
	FundingTotal    int64    `json:"funding_total"`
	LatestFundingAt string   `json:"latest_funding_at"`
	Technologies    []string `json:"technologies"`
	SEODescription  string   `json:"seo_description"`
	ShortDescription string  `json:"short_description"`
	LinkedInURL     string   `json:"linkedin_url"`
	TwitterURL      string   `json:"twitter_url"`
	FacebookURL     string   `json:"facebook_url"`
	PrimaryDomain   string   `json:"primary_domain"`
	Country         string   `json:"country"`
	State           string   `json:"state"`
	City            string   `json:"city"`
	Address         string   `json:"address"`
	PostalCode      string   `json:"postal_code"`
	Phone           string   `json:"phone"`
}

type SearchResponse struct {
	People        []Person      `json:"people"`
	Organizations []Organization `json:"organizations"`
	Pagination    Pagination    `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
}

func (c *Client) SearchProspects(params SearchParams) ([]ProspectData, error) {
	ctx := context.Background()

	// Apollo mixed_companies/search endpoint
	endpoint := "/mixed_companies/search"
	
	reqBody := map[string]interface{}{
		"q_organization_funding_stage": params.Series,
		"q_organization_industry":      params.Industries,
		"q_organization_locations":     params.Locations,
		"organization_num_employees_ranges": fmt.Sprintf("%d,%d", params.MinEmployees, params.MaxEmployees),
		"person_titles":                params.Roles,
		"person_locations":             params.Locations,
		"page":                         1,
		"per_page":                     100,
	}

	var allProspects []ProspectData
	page := 1

	for len(allProspects) < params.Limit {
		reqBody["page"] = page
		
		resp, err := c.doRequest(ctx, "POST", endpoint, reqBody)
		if err != nil {
			return nil, err
		}

		var searchResp SearchResponse
		if err := json.Unmarshal(resp, &searchResp); err != nil {
			return nil, fmt.Errorf("parse response: %w", err)
		}

		if len(searchResp.People) == 0 {
			break
		}

		// Map people to organizations
		orgMap := make(map[string]Organization)
		for _, org := range searchResp.Organizations {
			orgMap[org.ID] = org
		}

		for _, person := range searchResp.People {
			org := orgMap[person.OrganizationID]
			prospect := ProspectData{
				ApolloID:      person.ID,
				FirstName:     person.FirstName,
				LastName:      person.LastName,
				Title:         person.Title,
				Headline:      person.Headline,
				Email:         person.Email,
				EmailStatus:   person.EmailStatus,
				LinkedIn:      person.LinkedInURL,
				Phone:         person.PhoneNumber,
				Company:       person.OrganizationName,
				CompanyID:     person.OrganizationID,
				CompanyDomain: org.PrimaryDomain,
				Industry:      org.Industry,
				Series:        org.FundingStage,
				Employees:     org.EmployeeCount,
				Revenue:       org.AnnualRevenue,
				Technologies:  org.Technologies,
				Description:   org.ShortDescription,
				CompanyCity:   org.City,
				CompanyState:  org.State,
				CompanyCountry: org.Country,
			}
			allProspects = append(allProspects, prospect)
			
			if len(allProspects) >= params.Limit {
				break
			}
		}

		if page >= searchResp.Pagination.TotalPages {
			break
		}
		page++
	}

	return allProspects[:min(len(allProspects), params.Limit)], nil
}

func (c *Client) EnrichPerson(personID string) (*Person, error) {
	ctx := context.Background()
	endpoint := fmt.Sprintf("/people/%s", personID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var person Person
	if err := json.Unmarshal(resp, &person); err != nil {
		return nil, err
	}

	return &person, nil
}

func (c *Client) EnrichOrganization(orgID string) (*Organization, error) {
	ctx := context.Background()
	endpoint := fmt.Sprintf("/organizations/%s", orgID)
	
	resp, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	// Rate limit
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func buildSearchQuery(params SearchParams) string {
	// Not used directly but useful for logging
	values := url.Values{}
	if len(params.Roles) > 0 {
		values.Set("roles", join(params.Roles))
	}
	if len(params.Series) > 0 {
		values.Set("series", join(params.Series))
	}
	return values.Encode()
}

func join(s []string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ","
		}
		result += v
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ProspectData is the normalized prospect structure used across packages
type ProspectData struct {
	ApolloID       string
	FirstName      string
	LastName       string
	Title          string
	Headline       string
	Email          string
	EmailStatus    string
	LinkedIn       string
	Phone          string
	Company        string
	CompanyID      string
	CompanyDomain  string
	Industry       string
	Series         string
	Employees      int
	Revenue        int64
	Technologies   []string
	Description    string
	CompanyCity    string
	CompanyState   string
	CompanyCountry string
}