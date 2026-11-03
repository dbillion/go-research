package templates

import (
	"fmt"
	"strings"
)

// ProspectData mirrors apollo.ProspectData for template generation
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

type OutreachTemplates struct {
	Email1Subject    string
	Email1Body       string
	Email2Subject    string
	Email2Body       string
	Email3Subject    string
	Email3Body       string
	LinkedInConnection string
	LinkedInFollowUp string
	PainPoints       string
	PersonalizationHooks string
}

type OutreachResult struct {
	Prospect      ProspectData
	Email1Subject string
	Email1Body    string
	Email2Subject string
	Email2Body    string
	Email3Subject string
	Email3Body    string
	LinkedInConnection string
	LinkedInFollowUp string
	PainPoints    string
	PersonalizationHooks string
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(prospect ProspectData) (*OutreachTemplates, error) {
	painPoints := g.identifyPainPoints(prospect)
	hooks := g.generateHooks(prospect)

	return &OutreachTemplates{
		Email1Subject:       g.email1Subject(prospect),
		Email1Body:          g.email1Body(prospect, painPoints, hooks),
		Email2Subject:       g.email2Subject(prospect),
		Email2Body:          g.email2Body(prospect, painPoints, hooks),
		Email3Subject:       g.email3Subject(prospect),
		Email3Body:          g.email3Body(prospect, painPoints, hooks),
		LinkedInConnection:  g.linkedInConnection(prospect, hooks),
		LinkedInFollowUp:    g.linkedInFollowUp(prospect, painPoints),
		PainPoints:          painPoints,
		PersonalizationHooks: hooks,
	}, nil
}

func (g *Generator) GenerateForBatch(prospects []ProspectData) ([]OutreachResult, error) {
	results := make([]OutreachResult, len(prospects))
	for i, p := range prospects {
		tmpl, err := g.Generate(p)
		if err != nil {
			return nil, err
		}
		results[i] = OutreachResult{
			Prospect:            p,
			Email1Subject:       tmpl.Email1Subject,
			Email1Body:          tmpl.Email1Body,
			Email2Subject:       tmpl.Email2Subject,
			Email2Body:          tmpl.Email2Body,
			Email3Subject:       tmpl.Email3Subject,
			Email3Body:          tmpl.Email3Body,
			LinkedInConnection:  tmpl.LinkedInConnection,
			LinkedInFollowUp:    tmpl.LinkedInFollowUp,
			PainPoints:          tmpl.PainPoints,
			PersonalizationHooks: tmpl.PersonalizationHooks,
		}
	}
	return results, nil
}

func (g *Generator) identifyPainPoints(p ProspectData) string {
	var pains []string

	// Revenue leakage indicators
	if strings.Contains(strings.ToLower(p.Description), "billing") ||
	   strings.Contains(strings.ToLower(p.Description), "subscription") ||
	   strings.Contains(strings.ToLower(p.Description), "usage") {
		pains = append(pains, "Complex billing/subscription management at scale")
	}

	// High growth = process breaking
	if p.Employees > 100 && p.Employees < 500 {
		pains = append(pains, "Manual revenue operations processes breaking at scale")
	}

	// Series B/C = scaling challenges
	if p.Series == "Series B" || p.Series == "Series C" {
		pains = append(pains, "Scaling finance ops without proportional headcount")
	}

	// Tech stack complexity
	techStr := strings.ToLower(strings.Join(p.Technologies, " "))
	if strings.Contains(techStr, "stripe") || strings.Contains(techStr, "netsuite") || 
	   strings.Contains(techStr, "salesforce") || strings.Contains(techStr, "hubspot") {
		pains = append(pains, "Disconnected systems (Stripe/NetSuite/Salesforce) causing revenue leakage")
	}

	// FinTech/HealthTech = compliance burden
	if p.Industry == "FinTech" || p.Industry == "HealthTech" {
		pains = append(pains, "Revenue recognition compliance (ASC 606) + audit pressure")
	}

	if len(pains) == 0 {
		pains = append(pains, "1-5% revenue leakage from manual billing processes (Forrester: 46% of RevOps teams)")
	}

	return strings.Join(pains, "; ")
}

func (g *Generator) generateHooks(p ProspectData) string {
	var hooks []string

	// Name + role
	hooks = append(hooks, fmt.Sprintf("%s %s, %s at %s", p.FirstName, p.LastName, p.Title, p.Company))

	// Company context
	if p.Series != "" {
		hooks = append(hooks, fmt.Sprintf("%s-backed (%s)", p.Company, p.Series))
	}
	if p.Employees > 0 {
		hooks = append(hooks, fmt.Sprintf("%d employees", p.Employees))
	}

	// Tech stack
	if len(p.Technologies) > 0 {
		topTech := p.Technologies
		if len(topTech) > 3 {
			topTech = topTech[:3]
		}
		hooks = append(hooks, fmt.Sprintf("Stack: %s", strings.Join(topTech, ", ")))
	}

	// Location
	if p.CompanyCity != "" && p.CompanyCountry != "" {
		hooks = append(hooks, fmt.Sprintf("Based in %s, %s", p.CompanyCity, p.CompanyCountry))
	}

	return strings.Join(hooks, " | ")
}

func (g *Generator) email1Subject(p ProspectData) string {
	return fmt.Sprintf("%s - %s at %s - revenue leakage?", p.FirstName, p.Title, p.Company)
}

func (g *Generator) email1Body(p ProspectData, painPoints, hooks string) string {
	techStack := "modern SaaS"
	if len(p.Technologies) > 0 {
		topTech := p.Technologies
		if len(topTech) > 3 {
			topTech = topTech[:3]
		}
		techStack = strings.Join(topTech, ", ")
	}

	return fmt.Sprintf(`Hi %s,

Saw %s just hit %s — congrats on the momentum.

Quick observation: Forrester found 46%% of RevOps teams at %s-stage companies run mostly manual billing processes. That typically means 1-5%% of top-line revenue leaking out the back door (contract mismatches, unbilled usage, pricing exceptions in spreadsheets).

At %s (%d employees, %s stack), you're likely hitting the inflection point where:
• Sales closes deals faster than Finance can bill them
• Usage data sits in Stripe/event streams, not NetSuite
• Discount approvals happen in Slack/email, not in the billing system

We're helping %s-stage %s companies plug those leaks continuously — not quarterly.

Worth a 10-min chat to see if the math works for %s?

Best,
[Your Name]

P.S. Happy to share the "Revenue Leakage Calculator" we built — no strings attached.`, p.FirstName, p.Company, p.Series, p.Series, p.Company, p.Employees, techStack, p.Series, strings.ToLower(p.Industry), p.Company)
}

func (g *Generator) email2Subject(p ProspectData) string {
	return fmt.Sprintf("Re: %s - revenue leakage calculator (no pitch)", p.Company)
}

func (g *Generator) email2Body(p ProspectData, painPoints, hooks string) string {
	return fmt.Sprintf(`Hi %s,

Following up — built a quick calculator that shows exactly where %s-stage %s companies lose revenue.

It takes 3 inputs:
1. Monthly invoice volume
2. %% of contracts with custom terms
3. Manual touchpoints per invoice

Outputs: $ leaked/month + hours wasted.

No email gate, no sales pitch. Just the math.

Want me to send the link? Runs in-browser, takes 60 seconds.

[Your Name]`, p.FirstName, p.Series, strings.ToLower(p.Industry))
}

func (g *Generator) email3Subject(p ProspectData) string {
	return fmt.Sprintf("Closing the loop - %s", p.Company)
}

func (g *Generator) email3Body(p ProspectData, painPoints, hooks string) string {
	return fmt.Sprintf(`Hi %s,

I'll keep this short.

If revenue leakage isn't a priority right now, no worries — timing matters.

But if it becomes one (board asks about bookings vs. realized margin, audit finds gaps, or you're prepping for next raise), the calculator and a few case studies are here: [link]

Happy to stay in touch via LinkedIn too.

Best,
[Your Name]`, p.FirstName)
}

func (g *Generator) linkedInConnection(p ProspectData, hooks string) string {
	return fmt.Sprintf(`Hi %s, noticed you're leading finance at %s (%s, %d people). 

Connecting because I work with %s-stage %s CFOs on revenue leakage — the 1-5%% that slips between Stripe/NetSuite/Salesforce. 

No pitch, just peer context. Happy to share the leakage calculator we built if useful.

%s`, p.FirstName, p.Company, p.Series, p.Employees, p.Series, strings.ToLower(p.Industry), hooks)
}

func (g *Generator) linkedInFollowUp(p ProspectData, painPoints string) string {
	return fmt.Sprintf(`Thanks for connecting %s.

The leakage calculator: [link] — takes 60 seconds, shows $/month at risk.

Also putting together a small group of %s-stage %s finance leaders for a 30-min roundtable next month: "Scaling RevOps Without Scaling Headcount." 

Would you want an invite?

[Your Name]`, p.FirstName, p.Series, strings.ToLower(p.Industry))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}