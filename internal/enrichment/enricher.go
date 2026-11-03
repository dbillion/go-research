package enrichment

import (
	"fmt"
	"sync"

	"github.com/deeone/go-research/internal/apollo"
)

type Enricher struct {
	client *apollo.Client
}

func NewEnricher(client *apollo.Client) *Enricher {
	return &Enricher{client: client}
}

func (e *Enricher) EnrichBatch(prospects []apollo.ProspectData) ([]apollo.ProspectData, error) {
	var (
		wg       sync.WaitGroup
		sem      = make(chan struct{}, 5) // Max 5 concurrent
		results  = make([]apollo.ProspectData, len(prospects))
		errors   = make([]error, len(prospects))
	)

	for i, prospect := range prospects {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, p apollo.ProspectData) {
			defer wg.Done()
			defer func() { <-sem }()

			enriched, err := e.enrichProspect(p)
			if err != nil {
				errors[idx] = err
				results[idx] = p // Keep original on error
				return
			}
			results[idx] = enriched
		}(i, prospect)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, fmt.Errorf("enrichment error: %w", err)
		}
	}

	return results, nil
}

func (e *Enricher) enrichProspect(p apollo.ProspectData) (apollo.ProspectData, error) {

	// Enrich person if we have Apollo ID
	if p.ApolloID != "" {
		person, err := e.client.EnrichPerson(p.ApolloID)
		if err == nil && person != nil {
			p.Email = person.Email
			p.EmailStatus = person.EmailStatus
			p.LinkedIn = person.LinkedInURL
			p.Phone = person.PhoneNumber
			p.Headline = person.Headline
		}
	}

	// Enrich organization if we have Company ID
	if p.CompanyID != "" {
		org, err := e.client.EnrichOrganization(p.CompanyID)
		if err == nil && org != nil {
			p.Company = org.Name
			p.CompanyDomain = org.PrimaryDomain
			p.Industry = org.Industry
			p.Series = org.FundingStage
			p.Employees = org.EmployeeCount
			p.Revenue = org.AnnualRevenue
			p.Technologies = org.Technologies
			p.Description = org.ShortDescription
			p.CompanyCity = org.City
			p.CompanyState = org.State
			p.CompanyCountry = org.Country
		}
	}

	return p, nil
}