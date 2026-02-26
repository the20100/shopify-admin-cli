package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var ordersCmd = &cobra.Command{
	Use:   "orders",
	Short: "Manage Shopify orders",
}

var (
	ordersListFirst int
	ordersListAfter string
	ordersListQuery string
)

var ordersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List orders",
	Long: `List orders in your Shopify store.

Use --query for Shopify search syntax:
  financial_status:paid, fulfillment_status:unfulfilled, created_at:>2024-01-01

Examples:
  shopify-admin orders list
  shopify-admin orders list --query "financial_status:paid"
  shopify-admin orders list --first 20 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListOrders(ordersListFirst, ordersListAfter, ordersListQuery)
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
			fmt.Println("No orders found.")
			return nil
		}
		headers := []string{"ID", "NAME", "FINANCIAL", "FULFILLMENT", "TOTAL", "CUSTOMER", "CREATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			o := e.Node
			customer := "-"
			if o.Customer != nil {
				customer = strings.TrimSpace(o.Customer.FirstName + " " + o.Customer.LastName)
			}
			rows[i] = []string{
				shortID(o.ID),
				o.Name,
				strings.ToLower(o.FinancialStatus),
				strings.ToLower(o.DisplayFulfillmentStatus),
				formatMoney(o.TotalPriceSet.ShopMoney.Amount, o.TotalPriceSet.ShopMoney.CurrencyCode),
				output.Truncate(customer, 24),
				output.FormatTime(o.CreatedAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

var ordersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific order",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		o, err := client.GetOrder(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(o, output.IsPretty(cmd))
		}
		customer := "-"
		if o.Customer != nil {
			customer = strings.TrimSpace(o.Customer.FirstName+" "+o.Customer.LastName) + " (" + o.Customer.Email + ")"
		}
		shippingAddr := "-"
		if o.ShippingAddress != nil {
			a := o.ShippingAddress
			shippingAddr = fmt.Sprintf("%s %s, %s, %s %s %s",
				a.FirstName, a.LastName, a.Address1, a.City, a.Province, a.Country)
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(o.ID)},
			{"Name", o.Name},
			{"Email", o.Email},
			{"Phone", o.Phone},
			{"Financial", strings.ToLower(o.FinancialStatus)},
			{"Fulfillment", strings.ToLower(o.DisplayFulfillmentStatus)},
			{"Total", formatMoney(o.TotalPriceSet.ShopMoney.Amount, o.TotalPriceSet.ShopMoney.CurrencyCode)},
			{"Subtotal", formatMoney(o.SubtotalPriceSet.ShopMoney.Amount, o.SubtotalPriceSet.ShopMoney.CurrencyCode)},
			{"Tax", formatMoney(o.TotalTaxSet.ShopMoney.Amount, o.TotalTaxSet.ShopMoney.CurrencyCode)},
			{"Customer", customer},
			{"Shipping", shippingAddr},
			{"Note", o.Note},
			{"Tags", output.FormatLabels(o.Tags)},
			{"Created", output.FormatTime(o.CreatedAt)},
		})
		if len(o.LineItems.Edges) > 0 {
			fmt.Println()
			fmt.Println("Line items:")
			headers := []string{"TITLE", "QTY", "PRICE", "SKU"}
			rows := make([][]string, len(o.LineItems.Edges))
			for i, e := range o.LineItems.Edges {
				li := e.Node
				rows[i] = []string{
					output.Truncate(li.Title, 40),
					fmt.Sprintf("%d", li.Quantity),
					formatMoney(li.OriginalUnitPriceSet.ShopMoney.Amount, li.OriginalUnitPriceSet.ShopMoney.CurrencyCode),
					li.SKU,
				}
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

var ordersCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Mark an order as closed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		o, err := client.CloseOrder(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(o, output.IsPretty(cmd))
		}
		fmt.Printf("Order %s closed.\n", o.Name)
		return nil
	},
}

var (
	orderCancelReason  string
	orderCancelRefund  bool
	orderCancelRestock bool
)

var ordersCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel an order",
	Long: `Cancel an order. Valid reasons: customer, fraud, inventory, declined, other

Examples:
  shopify-admin orders cancel 1234567890
  shopify-admin orders cancel 1234567890 --reason customer --refund --restock`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.CancelOrder(args[0], strings.ToUpper(orderCancelReason), orderCancelRefund, orderCancelRestock); err != nil {
			return err
		}
		fmt.Printf("Order %s cancelled.\n", args[0])
		return nil
	},
}

var ordersMarkPaidCmd = &cobra.Command{
	Use:   "mark-paid <id>",
	Short: "Mark an order as paid",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		o, err := client.MarkOrderAsPaid(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(o, output.IsPretty(cmd))
		}
		fmt.Printf("Order %s marked as paid.\n", o.Name)
		fmt.Printf("Financial status: %s\n", strings.ToLower(o.FinancialStatus))
		return nil
	},
}

func init() {
	ordersListCmd.Flags().IntVar(&ordersListFirst, "first", 50, "Number of orders to return")
	ordersListCmd.Flags().StringVar(&ordersListAfter, "after", "", "Pagination cursor")
	ordersListCmd.Flags().StringVar(&ordersListQuery, "query", "", "Shopify search query")

	ordersCancelCmd.Flags().StringVar(&orderCancelReason, "reason", "other", "Cancel reason: customer, fraud, inventory, declined, other")
	ordersCancelCmd.Flags().BoolVar(&orderCancelRefund, "refund", false, "Refund the order when cancelling")
	ordersCancelCmd.Flags().BoolVar(&orderCancelRestock, "restock", false, "Restock items when cancelling")

	ordersCmd.AddCommand(
		ordersListCmd,
		ordersGetCmd,
		ordersCloseCmd,
		ordersCancelCmd,
		ordersMarkPaidCmd,
	)
	rootCmd.AddCommand(ordersCmd)
}
