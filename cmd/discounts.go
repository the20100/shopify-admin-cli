package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var discountsCmd = &cobra.Command{
	Use:   "discounts",
	Short: "Manage Shopify discounts",
}

var (
	discountsListFirst int
	discountsListAfter string
	discountsListQuery string
)

var discountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discounts (code and automatic)",
	Long: `List all discount nodes including code discounts and automatic discounts.

Examples:
  shopify-admin discounts list
  shopify-admin discounts list --query "status:active"
  shopify-admin discounts list --first 20 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListDiscounts(discountsListFirst, discountsListAfter, discountsListQuery)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			items := make([]any, len(conn.Edges))
			for i, e := range conn.Edges {
				items[i] = e.Node
			}
			return output.PrintJSON(items, output.IsPretty(cmd))
		}
		if len(conn.Edges) == 0 {
			fmt.Println("No discounts found.")
			return nil
		}
		headers := []string{"ID", "TITLE", "TYPE", "STATUS", "STARTS", "ENDS"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			d := e.Node
			typeName := strings.TrimPrefix(d.Discount.TypeName, "Discount")
			rows[i] = []string{
				shortID(d.ID),
				output.Truncate(d.Discount.Title, 36),
				typeName,
				strings.ToLower(d.Discount.Status),
				output.FormatTime(d.Discount.StartsAt),
				output.FormatTime(d.Discount.EndsAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

var discountsDeactivateCmd = &cobra.Command{
	Use:   "deactivate <id>",
	Short: "Deactivate a discount",
	Long: `Deactivate a discount by its node ID.

For code discounts, use the DiscountCodeNode ID.
For automatic discounts, use the DiscountAutomaticNode ID.

The CLI attempts to deactivate as a code discount first, then falls back to automatic.

Examples:
  shopify-admin discounts deactivate 1234567890`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try code discount first; if that fails, try automatic
		err := client.DeactivateDiscountCode(args[0])
		if err != nil {
			err2 := client.DeactivateAutomaticDiscount(args[0])
			if err2 != nil {
				return fmt.Errorf("code discount: %v; automatic discount: %v", err, err2)
			}
		}
		fmt.Printf("Discount %s deactivated.\n", args[0])
		return nil
	},
}

func init() {
	discountsListCmd.Flags().IntVar(&discountsListFirst, "first", 50, "Number of discounts to return")
	discountsListCmd.Flags().StringVar(&discountsListAfter, "after", "", "Pagination cursor")
	discountsListCmd.Flags().StringVar(&discountsListQuery, "query", "", "Search query")

	discountsCmd.AddCommand(discountsListCmd, discountsDeactivateCmd)
	rootCmd.AddCommand(discountsCmd)
}
