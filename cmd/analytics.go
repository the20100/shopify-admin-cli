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
  FROM sales SHOW net_sales SINCE -30d UNTIL today
  FROM sales SHOW net_sales GROUP BY month SINCE -6m UNTIL today
  FROM sessions SHOW sessions GROUP BY device_type SINCE -7d UNTIL today
  FROM sales SHOW net_sales, taxes, customers SINCE -1y UNTIL yesterday

Note: ShopifyQL pre-aggregates metrics — no SUM()/COUNT() functions.
Requires the 'read_reports' access scope on your access token.

Examples:
  shopify-admin analytics query "FROM sales SHOW net_sales SINCE -30d UNTIL today"
  shopify-admin analytics query "FROM sales SHOW net_sales GROUP BY month SINCE -6m" --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := client.RunShopifyQL(args[0])
		if err != nil {
			return err
		}
		if len(result.ParseErrors) > 0 {
			return fmt.Errorf("ShopifyQL parse errors: %s", strings.Join(result.ParseErrors, "; "))
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		if result.TableData == nil || len(result.TableData.Rows) == 0 {
			fmt.Println("No data returned.")
			return nil
		}
		// Build headers and convert map rows to ordered [][]string for PrintTable.
		cols := result.TableData.Columns
		headers := make([]string, len(cols))
		for i, col := range cols {
			name := col.DisplayName
			if name == "" {
				name = col.Name
			}
			headers[i] = strings.ToUpper(name)
		}
		tableRows := make([][]string, len(result.TableData.Rows))
		for i, row := range result.TableData.Rows {
			cells := make([]string, len(cols))
			for j, col := range cols {
				cells[j] = row[col.Name]
			}
			tableRows[i] = cells
		}
		output.PrintTable(headers, tableRows)
		fmt.Printf("\n(%d rows)\n", len(result.TableData.Rows))
		return nil
	},
}

func init() {
	analyticsCmd.AddCommand(analyticsQueryCmd)
	rootCmd.AddCommand(analyticsCmd)
}
