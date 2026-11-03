package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	ApolloAPIKey     string   `mapstructure:"APOLLO_API_KEY"`
	ClayAPIKey       string   `mapstructure:"CLAY_API_KEY"`
	OutputDir        string   `mapstructure:"OUTPUT_DIR"`
	OutputFormat     string   `mapstructure:"OUTPUT_FORMAT"`
	TargetRoles      []string `mapstructure:"TARGET_ROLES"`
	TargetSeries     []string `mapstructure:"TARGET_SERIES"`
	TargetIndustries []string `mapstructure:"TARGET_INDUSTRIES"`
	MinEmployees     int      `mapstructure:"MIN_EMPLOYEES"`
	MaxEmployees     int      `mapstructure:"MAX_EMPLOYEES"`
	Locations        []string `mapstructure:"LOCATIONS"`
	RequestsPerMin   int      `mapstructure:"REQUESTS_PER_MINUTE"`
}

func Load() (*Config, error) {
	// Load .env file if exists
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load(".env.example"); err != nil {
			// No env file, continue with env vars only
		}
	}

	v := viper.New()
	v.SetEnvPrefix("OUTREACH")
	v.AutomaticEnv()

	// Bind env vars
	v.BindEnv("APOLLO_API_KEY")
	v.BindEnv("CLAY_API_KEY")
	v.BindEnv("OUTPUT_DIR")
	v.BindEnv("OUTPUT_FORMAT")
	v.BindEnv("TARGET_ROLES")
	v.BindEnv("TARGET_SERIES")
	v.BindEnv("TARGET_INDUSTRIES")
	v.BindEnv("MIN_EMPLOYEES")
	v.BindEnv("MAX_EMPLOYEES")
	v.BindEnv("LOCATIONS")
	v.BindEnv("REQUESTS_PER_MINUTE")

	// Set defaults
	v.SetDefault("OUTPUT_DIR", "./outreach_results")
	v.SetDefault("OUTPUT_FORMAT", "csv")
	v.SetDefault("TARGET_ROLES", []string{"CFO", "VP Finance", "Head of Finance", "Controller"})
	v.SetDefault("TARGET_SERIES", []string{"Series B", "Series C", "Series D", "Series E"})
	v.SetDefault("TARGET_INDUSTRIES", []string{"SaaS", "Software", "FinTech", "HealthTech", "DevTools"})
	v.SetDefault("MIN_EMPLOYEES", 50)
	v.SetDefault("MAX_EMPLOYEES", 2000)
	v.SetDefault("LOCATIONS", []string{"United States", "Canada", "United Kingdom", "Germany", "France"})
	v.SetDefault("REQUESTS_PER_MINUTE", 80)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, err
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(cfg.OutputDir)
	if err == nil {
		cfg.OutputDir = absPath
	}

	return &cfg, nil
}