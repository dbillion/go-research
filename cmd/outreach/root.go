package outreach

import (
	"fmt"
	"os"

	"github.com/deeone/go-research/internal/apollo"
	"github.com/deeone/go-research/internal/config"
	"github.com/deeone/go-research/internal/enrichment"
	"github.com/deeone/go-research/internal/templates"
	"github.com/spf13/cobra"
)

var (
	cfg         *config.Config
	outputFile  string
	dryRun      bool
	limit       int
)

var rootCmd = &cobra.Command{
	Use:   "outreach",
	Short: "Find and enrich CFOs at Series B+ SaaS companies for outreach",
	Long: `Outreach CLI - Primary Research Tool

This tool finds CFOs and finance leaders at Series B+ SaaS companies
using Apollo.io API, enriches their data, and generates personalized
outreach templates based on revenue leakage pain points.`,
	RunE: runOutreach,
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for prospects without enrichment",
	RunE:  runSearch,
}

var enrichCmd = &cobra.Command{
	Use:   "enrich [csv-file]",
	Short: "Enrich existing prospect CSV with additional data",
	Args:  cobra.ExactArgs(1),
	RunE:  runEnrich,
}

var templateCmd = &cobra.Command{
	Use:   "template [csv-file]",
	Short: "Generate outreach templates from enriched prospect data",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplate,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "", "Output file path (default: ./outreach_results/outreach_<timestamp>.csv)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Run without making API calls")
	rootCmd.PersistentFlags().IntVar(&limit, "limit", 50, "Maximum number of prospects to find")

	rootCmd.AddCommand(searchCmd, enrichCmd, templateCmd)

	searchCmd.Flags().StringSlice("roles", nil, "Target roles (overrides config)")
	searchCmd.Flags().StringSlice("series", nil, "Target funding series (overrides config)")
	searchCmd.Flags().StringSlice("industries", nil, "Target industries (overrides config)")
	searchCmd.Flags().Int("min-employees", 0, "Minimum company employees")
	searchCmd.Flags().Int("max-employees", 0, "Maximum company employees")
	searchCmd.Flags().StringSlice("locations", nil, "Target locations (overrides config)")
}

func initConfig() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	// Override from flags if provided
	if roles, _ := rootCmd.Flags().GetStringSlice("roles"); len(roles) > 0 {
		cfg.TargetRoles = roles
	}
	if series, _ := rootCmd.Flags().GetStringSlice("series"); len(series) > 0 {
		cfg.TargetSeries = series
	}
	if industries, _ := rootCmd.Flags().GetStringSlice("industries"); len(industries) > 0 {
		cfg.TargetIndustries = industries
	}
	if minEmp, _ := rootCmd.Flags().GetInt("min-employees"); minEmp > 0 {
		cfg.MinEmployees = minEmp
	}
	if maxEmp, _ := rootCmd.Flags().GetInt("max-employees"); maxEmp > 0 {
		cfg.MaxEmployees = maxEmp
	}
	if locations, _ := rootCmd.Flags().GetStringSlice("locations"); len(locations) > 0 {
		cfg.Locations = locations
	}
}

func runOutreach(cmd *cobra.Command, args []string) error {
	if !dryRun && cfg.ApolloAPIKey == "" {
		return fmt.Errorf("APOLLO_API_KEY not set. Copy .env.example to .env and add your key")
	}

	fmt.Println("🔍 Starting outreach prospecting...")
	fmt.Printf("   Target: %s at %s companies\n", cfg.TargetRoles, cfg.TargetSeries)
	fmt.Printf("   Industries: %v\n", cfg.TargetIndustries)
	fmt.Printf("   Employee range: %d-%d\n", cfg.MinEmployees, cfg.MaxEmployees)
	fmt.Printf("   Locations: %v\n", cfg.Locations)
	fmt.Printf("   Limit: %d prospects\n", limit)

	if dryRun {
		fmt.Println("🧪 DRY RUN - No API calls will be made")
		return runDryRun()
	}

	// Step 1: Search Apollo for prospects
	client := apollo.NewClient(cfg.ApolloAPIKey, cfg.RequestsPerMin)
	prospects, err := client.SearchProspects(apollo.SearchParams{
		Roles:       cfg.TargetRoles,
		Series:      cfg.TargetSeries,
		Industries:  cfg.TargetIndustries,
		MinEmployees: cfg.MinEmployees,
		MaxEmployees: cfg.MaxEmployees,
		Locations:   cfg.Locations,
		Limit:       limit,
	})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Printf("✅ Found %d prospects\n", len(prospects))

	// Step 2: Enrich prospects
	enricher := enrichment.NewEnricher(client)
	enriched, err := enricher.EnrichBatch(prospects)
	if err != nil {
		return fmt.Errorf("enrichment failed: %w", err)
	}

	fmt.Printf("✅ Enriched %d prospects\n", len(enriched))

	// Convert to templates.ProspectData for template generation
	templateProspects := make([]templates.ProspectData, len(enriched))
	for i, p := range enriched {
		templateProspects[i] = templates.ProspectData(p)
	}

	// Step 3: Generate outreach templates
	generator := templates.NewGenerator()
	results, err := generator.GenerateForBatch(templateProspects)
	if err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}

	// Step 4: Save results
	outPath := outputFile
	if outPath == "" {
		outPath = fmt.Sprintf("%s/outreach_%s.csv", cfg.OutputDir, enrichment.Timestamp())
	}

	// Convert to []interface{} for SaveCSV
	resultsInterface := make([]interface{}, len(results))
	for i, r := range results {
		resultsInterface[i] = r
	}

	if err := enrichment.SaveCSV(resultsInterface, outPath); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	fmt.Printf("💾 Results saved to: %s\n", outPath)
	printSummary(results)

	return nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	if cfg.ApolloAPIKey == "" {
		return fmt.Errorf("APOLLO_API_KEY not set")
	}

	client := apollo.NewClient(cfg.ApolloAPIKey, cfg.RequestsPerMin)
	prospects, err := client.SearchProspects(apollo.SearchParams{
		Roles:       cfg.TargetRoles,
		Series:      cfg.TargetSeries,
		Industries:  cfg.TargetIndustries,
		MinEmployees: cfg.MinEmployees,
		MaxEmployees: cfg.MaxEmployees,
		Locations:   cfg.Locations,
		Limit:       limit,
	})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	outPath := outputFile
	if outPath == "" {
		outPath = fmt.Sprintf("%s/search_%s.csv", cfg.OutputDir, enrichment.Timestamp())
	}

	// Convert to []interface{} for SaveCSV
	prospectsInterface := make([]interface{}, len(prospects))
	for i, p := range prospects {
		prospectsInterface[i] = p
	}

	if err := enrichment.SaveCSV(prospectsInterface, outPath); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	fmt.Printf("✅ Found %d prospects, saved to %s\n", len(prospects), outPath)
	return nil
}

func runEnrich(cmd *cobra.Command, args []string) error {
	if cfg.ApolloAPIKey == "" {
		return fmt.Errorf("APOLLO_API_KEY not set")
	}

	inputFile := args[0]
	prospects, err := enrichment.LoadCSV(inputFile)
	if err != nil {
		return fmt.Errorf("load failed: %w", err)
	}

	client := apollo.NewClient(cfg.ApolloAPIKey, cfg.RequestsPerMin)
	enricher := enrichment.NewEnricher(client)
	enriched, err := enricher.EnrichBatch(prospects)
	if err != nil {
		return fmt.Errorf("enrichment failed: %w", err)
	}

	outPath := outputFile
	if outPath == "" {
		outPath = fmt.Sprintf("%s/enriched_%s.csv", cfg.OutputDir, enrichment.Timestamp())
	}

	// Convert to []interface{} for SaveCSV
	enrichedInterface := make([]interface{}, len(enriched))
	for i, p := range enriched {
		enrichedInterface[i] = p
	}

	if err := enrichment.SaveCSV(enrichedInterface, outPath); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	fmt.Printf("✅ Enriched %d prospects, saved to %s\n", len(enriched), outPath)
	return nil
}

func runTemplate(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	prospects, err := enrichment.LoadCSV(inputFile)
	if err != nil {
		return fmt.Errorf("load failed: %w", err)
	}

	// Convert to templates.ProspectData for template generation
	templateProspects := make([]templates.ProspectData, len(prospects))
	for i, p := range prospects {
		templateProspects[i] = templates.ProspectData(p)
	}

	generator := templates.NewGenerator()
	results, err := generator.GenerateForBatch(templateProspects)
	if err != nil {
		return fmt.Errorf("template generation failed: %w", err)
	}

	outPath := outputFile
	if outPath == "" {
		outPath = fmt.Sprintf("%s/templates_%s.csv", cfg.OutputDir, enrichment.Timestamp())
	}

	// Convert to []interface{} for SaveCSV
	resultsInterface := make([]interface{}, len(results))
	for i, r := range results {
		resultsInterface[i] = r
	}

	if err := enrichment.SaveCSV(resultsInterface, outPath); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	fmt.Printf("✅ Generated templates for %d prospects, saved to %s\n", len(results), outPath)
	return nil
}

func runDryRun() error {
	// Show what would be searched
	params := apollo.SearchParams{
		Roles:       cfg.TargetRoles,
		Series:      cfg.TargetSeries,
		Industries:  cfg.TargetIndustries,
		MinEmployees: cfg.MinEmployees,
		MaxEmployees: cfg.MaxEmployees,
		Locations:   cfg.Locations,
		Limit:       limit,
	}

	fmt.Println("\n📋 Search parameters that would be used:")
	fmt.Printf("   Roles: %v\n", params.Roles)
	fmt.Printf("   Series: %v\n", params.Series)
	fmt.Printf("   Industries: %v\n", params.Industries)
	fmt.Printf("   Employee range: %d-%d\n", params.MinEmployees, params.MaxEmployees)
	fmt.Printf("   Locations: %v\n", params.Locations)
	fmt.Printf("   Limit: %d\n", params.Limit)

	// Show sample templates
	fmt.Println("\n📝 Sample outreach templates that would be generated:")
	sample := templates.ProspectData{
		FirstName: "Sarah",
		LastName:  "Chen",
		Title:     "CFO",
		Company:   "Acme SaaS",
		Industry:  "FinTech",
		Series:    "Series B",
		Employees: 120,
		Email:     "sarah.chen@acme.io",
		LinkedIn:  "linkedin.com/in/sarahchen",
	}
	gen := templates.NewGenerator()
	tmpl, _ := gen.Generate(sample)
	fmt.Printf("\n--- EMAIL 1 (Problem-First) ---\nSubject: %s\n%s\n", tmpl.Email1Subject, tmpl.Email1Body)
	fmt.Printf("\n--- EMAIL 2 (Value-Add) ---\nSubject: %s\n%s\n", tmpl.Email2Subject, tmpl.Email2Body)
	fmt.Printf("\n--- LINKEDIN CONNECTION ---\n%s\n", tmpl.LinkedInConnection)

	return nil
}

func printSummary(results []templates.OutreachResult) {
	fmt.Println("\n📊 Summary:")
	total := len(results)
	withEmail := 0
	withLinkedIn := 0
	withPhone := 0

	for _, r := range results {
		if r.Prospect.Email != "" {
			withEmail++
		}
		if r.Prospect.LinkedIn != "" {
			withLinkedIn++
		}
		if r.Prospect.Phone != "" {
			withPhone++
		}
	}

	fmt.Printf("   Total prospects: %d\n", total)
	fmt.Printf("   With email: %d (%.1f%%)\n", withEmail, float64(withEmail)/float64(total)*100)
	fmt.Printf("   With LinkedIn: %d (%.1f%%)\n", withLinkedIn, float64(withLinkedIn)/float64(total)*100)
	fmt.Printf("   With phone: %d (%.1f%%)\n", withPhone, float64(withPhone)/float64(total)*100)
}