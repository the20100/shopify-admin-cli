package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var fulfillmentsCmd = &cobra.Command{
	Use:   "fulfillments",
	Short: "Manage Shopify fulfillments and fulfillment orders",
}

var fulfillmentsListCmd = &cobra.Command{
	Use:   "list <order-id>",
	Short: "List fulfillment orders for an order",
	Long: `List fulfillment orders for a specific order.

Examples:
  shopify-admin fulfillments list 1234567890
  shopify-admin fulfillments list 1234567890 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListFulfillmentOrders(args[0])
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
			fmt.Println("No fulfillment orders found.")
			return nil
		}
		headers := []string{"ID", "STATUS", "REQUEST STATUS", "LOCATION", "FULFILL AT"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			fo := e.Node
			rows[i] = []string{
				shortID(fo.ID),
				strings.ToLower(fo.Status),
				strings.ToLower(fo.RequestStatus),
				fo.AssignedLocation.Name,
				output.FormatTime(fo.FulfillAt),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

var (
	fulfillmentCreateTracking string
	fulfillmentCreateNumber   string
	fulfillmentCreateURL      string
)

var fulfillmentsCreateCmd = &cobra.Command{
	Use:   "create <fulfillment-order-id>",
	Short: "Create a fulfillment for a fulfillment order",
	Long: `Create a fulfillment for all line items of a fulfillment order.

Examples:
  shopify-admin fulfillments create 1234567890
  shopify-admin fulfillments create 1234567890 --tracking UPS --number 1Z999AA1234567890
  shopify-admin fulfillments create 1234567890 --tracking FedEx --number 123456789 --url https://track.example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := client.CreateFulfillment(args[0], fulfillmentCreateTracking, fulfillmentCreateNumber, fulfillmentCreateURL)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(f, output.IsPretty(cmd))
		}
		fmt.Printf("Fulfillment created\n")
		fmt.Printf("ID:     %s\n", shortID(f.ID))
		fmt.Printf("Status: %s\n", strings.ToLower(f.Status))
		if f.TrackingCompany != "" {
			fmt.Printf("Carrier: %s\n", f.TrackingCompany)
		}
		if len(f.TrackingNumbers) > 0 {
			fmt.Printf("Tracking: %s\n", strings.Join(f.TrackingNumbers, ", "))
		}
		return nil
	},
}

func init() {
	fulfillmentsCreateCmd.Flags().StringVar(&fulfillmentCreateTracking, "tracking", "", "Tracking company name")
	fulfillmentsCreateCmd.Flags().StringVar(&fulfillmentCreateNumber, "number", "", "Tracking number")
	fulfillmentsCreateCmd.Flags().StringVar(&fulfillmentCreateURL, "url", "", "Tracking URL")

	fulfillmentsCmd.AddCommand(fulfillmentsListCmd, fulfillmentsCreateCmd)
	rootCmd.AddCommand(fulfillmentsCmd)
}
