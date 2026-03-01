package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/api"
	"github.com/the20100/shopify-admin-cli/internal/config"
)

var (
	jsonFlag   bool
	prettyFlag bool
	client     *api.Client
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "shopify-admin",
	Short: "Shopify Admin CLI — manage your Shopify store via the Admin GraphQL API",
	Long: `shopify-admin is a CLI tool for the Shopify Admin GraphQL API.

It outputs JSON when piped (for agent use) and human-readable tables in a terminal.

Credential resolution order:
  1. SHOPIFY_ACCESS_TOKEN + SHOPIFY_SHOP env vars
  2. Config file  (~/.config/shopify-admin/config.json  via: shopify-admin auth setup)

Examples:
  shopify-admin auth setup mystore.myshopify.com <access-token>
  shopify-admin shop info
  shopify-admin products list
  shopify-admin orders list --query "financial_status:paid"
  shopify-admin customers get <id>`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "Force pretty-printed JSON output (implies --json)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		noAuthCommands := map[string]bool{"info": true, "search-syntax": true, "completion": true, "help": true}
		if isAuthCommand(cmd) || noAuthCommands[cmd.Name()] {
			return nil
		}
		shop, token, err := resolveCredentials()
		if err != nil {
			return err
		}
		client = api.NewClient(shop, token)
		return nil
	}

	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show tool info: config path, auth status, and environment",
	Run: func(cmd *cobra.Command, args []string) {
		printInfo()
	},
}

func printInfo() {
	fmt.Printf("shopify-admin — Shopify Admin CLI\n\n")
	exe, _ := os.Executable()
	fmt.Printf("  binary:  %s\n", exe)
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Println("  config paths by OS:")
	fmt.Printf("    macOS:    ~/Library/Application Support/shopify-admin/config.json\n")
	fmt.Printf("    Linux:    ~/.config/shopify-admin/config.json\n")
	fmt.Printf("    Windows:  %%AppData%%\\shopify-admin\\config.json\n")
	fmt.Printf("  config:   %s\n", config.Path())
	fmt.Println()
	fmt.Println("  env vars:")
	fmt.Printf("    SHOPIFY_SHOP         = %s\n", maskOrEmpty(os.Getenv("SHOPIFY_SHOP")))
	fmt.Printf("    SHOPIFY_ACCESS_TOKEN = %s\n", maskOrEmpty(os.Getenv("SHOPIFY_ACCESS_TOKEN")))
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

// resolveEnv returns the value of the first non-empty environment variable from the given names.
func resolveEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

// resolveCredentials returns the shop domain and access token from env or config.
func resolveCredentials() (string, string, error) {
	// 1. Env vars (try all aliases)
	envToken := resolveEnv(
		"SHOPIFY_ACCESS_TOKEN", "SHOPIFY_TOKEN", "SHOPIFY_API_TOKEN", "SHOPIFY_API_KEY",
		"SHOPIFY_KEY", "SHOPIFY_API", "API_KEY_SHOPIFY", "API_SHOPIFY",
		"SHOPIFY_SECRET_KEY", "SHOPIFY_API_SECRET", "SHOPIFY_SECRET", "SHOPIFY_SK", "SK_SHOPIFY",
	)
	envShop := resolveEnv(
		"SHOPIFY_SHOP", "SHOPIFY_STORE", "SHOPIFY_DOMAIN", "SHOPIFY_SHOP_URL", "SHOPIFY_STORE_URL", "SHOP_DOMAIN",
	)
	if envShop != "" && envToken != "" {
		return envShop, envToken, nil
	}

	// 2. Config file
	var err error
	cfg, err = config.Load()
	if err != nil {
		return "", "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.Shop != "" && cfg.AccessToken != "" {
		return cfg.Shop, cfg.AccessToken, nil
	}
	return "", "", fmt.Errorf("not authenticated — run: shopify-admin auth setup <shop> <access-token>\nor set SHOPIFY_SHOP and SHOPIFY_ACCESS_TOKEN env vars")
}

func isAuthCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "auth" {
		return true
	}
	p := cmd.Parent()
	for p != nil {
		if p.Name() == "auth" {
			return true
		}
		p = p.Parent()
	}
	return false
}
