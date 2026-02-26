package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Shopify Admin authentication",
}

var authSetupCmd = &cobra.Command{
	Use:   "setup <shop> <access-token>",
	Short: "Save Shopify shop domain and access token to the config file",
	Long: `Save your Shopify shop domain and Admin API access token to the local config file.

<shop> can be the full domain (mystore.myshopify.com) or just the store name (mystore).

To get an access token:
  1. Go to your Shopify Admin → Settings → Apps and sales channels
  2. Click "Develop apps" → Create an app
  3. Configure Admin API scopes and install the app
  4. Copy the Admin API access token

The credentials are stored at:
  macOS:   ~/Library/Application Support/shopify-admin/config.json
  Linux:   ~/.config/shopify-admin/config.json
  Windows: %AppData%\shopify-admin\config.json

You can also set SHOPIFY_SHOP and SHOPIFY_ACCESS_TOKEN env vars instead.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		shop := args[0]
		token := args[1]

		// Normalize shop domain
		if !strings.Contains(shop, ".") {
			shop = shop + ".myshopify.com"
		}
		if len(token) < 16 {
			return fmt.Errorf("access token looks too short")
		}
		if err := config.Save(&config.Config{Shop: shop, AccessToken: token}); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Credentials saved to %s\n", config.Path())
		fmt.Printf("Shop:  %s\n", shop)
		fmt.Printf("Token: %s\n", maskOrEmpty(token))
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Printf("Config: %s\n\n", config.Path())

		envShop := os.Getenv("SHOPIFY_SHOP")
		envToken := os.Getenv("SHOPIFY_ACCESS_TOKEN")

		if envShop != "" && envToken != "" {
			fmt.Println("Source: env vars (take priority over config)")
			fmt.Printf("Shop:   %s\n", envShop)
			fmt.Printf("Token:  %s\n", maskOrEmpty(envToken))
		} else if c.Shop != "" && c.AccessToken != "" {
			fmt.Println("Source: config file")
			fmt.Printf("Shop:   %s\n", c.Shop)
			fmt.Printf("Token:  %s\n", maskOrEmpty(c.AccessToken))
		} else {
			fmt.Println("Status: not authenticated")
			fmt.Printf("\nRun: shopify-admin auth setup <shop> <access-token>\n")
			fmt.Printf("Or:  export SHOPIFY_SHOP=mystore.myshopify.com\n")
			fmt.Printf("     export SHOPIFY_ACCESS_TOKEN=shpat_...\n")
		}
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the saved credentials from the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Println("Credentials removed from config.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authSetupCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
