# Outreach CLI - Primary Research Tool

Find and enrich CFOs at Series B+ SaaS companies using Apollo.io API, then generate personalized outreach templates based on revenue leakage pain points.

## Installation

```bash
cd /home/deeone/go-research
go build -o outreach
```

## Configuration

Copy `.env.example` to `.env` and add your Apollo.io API key:

```bash
cp .env.example .env
# Edit .env and add your APOLLO_API_KEY
```

Get your API key from: https://app.apollo.io/#/settings/integrations/api

## Usage

### Dry Run (No API Key Required)

```bash
# Show search parameters and sample templates
./outreach --dry-run --limit 10

# Custom filters
./outreach --dry-run --roles "CFO,VP Finance" --series "Series B,Series C" --industries "FinTech,HealthTech" --min-employees 100 --max-employees 500
```

### Search Prospects (Requires API Key)

```bash
# Search and save to CSV
./outreach search --limit 50 --output ./results/search.csv

# With custom filters
./outreach search --roles "CFO" --series "Series B" --industries "FinTech" --limit 30
```

### Enrich Existing Prospects

```bash
# Enrich a CSV file with additional Apollo data
./outreach enrich ./results/search.csv --output ./results/enriched.csv
```

### Generate Outreach Templates

```bash
# Generate templates from enriched prospects
./outreach template ./results/enriched.csv --output ./results/templates.csv
```

### Full Pipeline (Search → Enrich → Template)

```bash
./outreach --limit 50 --output ./results/full_pipeline.csv
```

## Output Format

The tool outputs CSV files with the following columns:

**Search/Enrich Output:**
- apollo_id, first_name, last_name, title, headline
- email, email_status, linkedin, phone
- company, company_id, company_domain, industry, series
- employees, revenue, technologies, description
- company_city, company_state, company_country

**Template Output (includes all above plus):**
- email1_subject, email1_body
- email2_subject, email2_body
- email3_subject, email3_body
- linkedin_connection, linkedin_followup
- pain_points, personalization_hooks

## Example Output

See `./outreach --dry-run` for sample templates.

## Rate Limits

Apollo.io free tier: 100 requests/minute
- Configured default: 80 requests/minute (safe margin)
- Adjust via `REQUESTS_PER_MINUTE` in `.env`

## Filters

| Flag | Description | Default |
|------|-------------|---------|
| `--roles` | Target job titles | CFO, VP Finance, Head of Finance, Controller |
| `--series` | Funding stage | Series B, C, D, E |
| `--industries` | Target industries | SaaS, Software, FinTech, HealthTech, DevTools |
| `--min-employees` | Minimum company size | 50 |
| `--max-employees` | Maximum company size | 2000 |
| `--locations` | Geographic locations | US, Canada, UK, Germany, France |
| `--limit` | Max prospects | 50 |

## Architecture

```
outreach/
├── cmd/outreach/     # CLI commands (Cobra)
├── internal/
│   ├── apollo/       # Apollo.io API client
│   ├── config/       # Configuration (Viper + godotenv)
│   ├── enrichment/   # Data enrichment & CSV I/O
│   └── templates/    # Outreach template generation
└── main.go           # Entry point
```

## Pain Points Targeted

Templates are personalized based on detected pain points:
- Complex billing/subscription management at scale
- Manual revenue operations processes breaking at scale
- Scaling finance ops without proportional headcount
- Disconnected systems (Stripe/NetSuite/Salesforce) causing revenue leakage
- Revenue recognition compliance (ASC 606) + audit pressure
- 1-5% revenue leakage from manual billing processes (Forrester: 46% of RevOps teams)