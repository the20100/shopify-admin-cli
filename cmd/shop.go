package cmd

import (
	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var shopCmd = &cobra.Command{
	Use:   "shop",
	Short: "Manage Shopify store properties",
}

var shopInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show store information",
	Long: `Show store name, email, domain, currency, plan, and other details.

Examples:
  shopify-admin shop info
  shopify-admin shop info --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shop, err := client.GetShop()
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(shop, output.IsPretty(cmd))
		}
		planPlus := ""
		if shop.Plan.ShopifyPlus {
			planPlus = " (Shopify Plus)"
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(shop.ID)},
			{"Name", shop.Name},
			{"Email", shop.Email},
			{"Domain", shop.MyshopifyDomain},
			{"Primary URL", shop.PrimaryDomain.URL},
			{"Currency", shop.CurrencyCode},
			{"Country", shop.CountryCode},
			{"Timezone", shop.TimezoneAbbreviation},
			{"Plan", shop.Plan.DisplayName + planPlus},
			{"Created", output.FormatTime(shop.CreatedAt)},
		})
		return nil
	},
}

func init() {
	shopCmd.AddCommand(shopInfoCmd)
	rootCmd.AddCommand(shopCmd)
}
