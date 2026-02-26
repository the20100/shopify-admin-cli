package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var analyticsCmd = &cobra.Command{
	Use:   "analytics",
	Short: "Run Shopify analytics queries",
}

var analyticsQueryCmd = &cobra.Command{
	Use:   "query <shopifyql>",
	Short: "Run a ShopifyQL analytics query",
	Long: `Execute a ShopifyQL query against your store's analytics data.

ShopifyQL syntax examples:
  FROM sales SHOW SUM(net_sales) SINCE -30d UNTIL today
  FROM sales SHOW SUM(net_sales) GROUP BY month SINCE -6m UNTIL today
  FROM sessions SHOW sessions GROUP BY device_type SINCE -7d UNTIL today
  FROM products SHOW SUM(net_quantity) AS units_sold ORDER BY units_sold DESC SINCE -30d

Requires the 'read_reports' access scope on your access token.

Examples:
  shopify-admin analytics query "FROM sales SHOW SUM(net_sales) SINCE -30d UNTIL today"
  shopify-admin analytics query "FROM orders SHOW COUNT(order_id) GROUP BY day SINCE -7d" --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := client.RunShopifyQL(args[0])
		if err != nil {
			return err
		}
		if len(result.ParseErrors) > 0 {
			msgs := make([]string, len(result.ParseErrors))
			for i, e := range result.ParseErrors {
				msgs[i] = fmt.Sprintf("[%s] %s", e.Code, e.Message)
			}
			return fmt.Errorf("ShopifyQL parse errors: %s", strings.Join(msgs, "; "))
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		if result.TableData == nil || len(result.TableData.Rows) == 0 {
			fmt.Println("No data returned.")
			return nil
		}
		// Build headers from column definitions
		headers := make([]string, len(result.TableData.Columns))
		for i, col := range result.TableData.Columns {
			name := col.DisplayName
			if name == "" {
				name = col.Name
			}
			headers[i] = strings.ToUpper(name)
		}
		output.PrintTable(headers, result.TableData.Rows)
		fmt.Printf("\n(%d rows)\n", len(result.TableData.Rows))
		return nil
	},
}

func init() {
	analyticsCmd.AddCommand(analyticsQueryCmd)
	rootCmd.AddCommand(analyticsCmd)
}
