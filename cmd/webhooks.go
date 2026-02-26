package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var webhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Manage Shopify webhook subscriptions",
}

var (
	webhooksListFirst int
	webhooksListAfter string
)

var webhooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhook subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListWebhooks(webhooksListFirst, webhooksListAfter)
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
			fmt.Println("No webhooks found.")
			return nil
		}
		headers := []string{"ID", "TOPIC", "FORMAT", "CALLBACK URL", "CREATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			w := e.Node
			rows[i] = []string{
				shortID(w.ID),
				w.Topic,
				w.Format,
				output.Truncate(w.Endpoint.CallbackURL, 50),
				output.FormatTime(w.CreatedAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

var (
	webhookCreateTopic  string
	webhookCreateURL    string
	webhookCreateFormat string
)

var webhooksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new webhook subscription",
	Long: `Create an HTTP webhook subscription for a Shopify event topic.

Common topics: PRODUCTS_CREATE, PRODUCTS_UPDATE, ORDERS_CREATE, ORDERS_UPDATED,
               CUSTOMERS_CREATE, CUSTOMERS_UPDATE, INVENTORY_LEVELS_UPDATE,
               CHECKOUTS_CREATE, REFUNDS_CREATE, APP_UNINSTALLED

Examples:
  shopify-admin webhooks create --topic ORDERS_CREATE --url https://example.com/webhooks/orders
  shopify-admin webhooks create --topic PRODUCTS_UPDATE --url https://myapp.com/hook --format JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if webhookCreateTopic == "" {
			return fmt.Errorf("--topic is required")
		}
		if webhookCreateURL == "" {
			return fmt.Errorf("--url is required")
		}
		w, err := client.CreateWebhook(webhookCreateTopic, webhookCreateURL, webhookCreateFormat)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(w, output.IsPretty(cmd))
		}
		fmt.Printf("Webhook created\n")
		fmt.Printf("ID:      %s\n", shortID(w.ID))
		fmt.Printf("Topic:   %s\n", w.Topic)
		fmt.Printf("URL:     %s\n", w.Endpoint.CallbackURL)
		fmt.Printf("Format:  %s\n", w.Format)
		return nil
	},
}

var webhooksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a webhook subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteWebhook(args[0]); err != nil {
			return err
		}
		fmt.Printf("Webhook %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	webhooksListCmd.Flags().IntVar(&webhooksListFirst, "first", 50, "Number of webhooks to return")
	webhooksListCmd.Flags().StringVar(&webhooksListAfter, "after", "", "Pagination cursor")

	webhooksCreateCmd.Flags().StringVar(&webhookCreateTopic, "topic", "", "Webhook topic (required, e.g. ORDERS_CREATE)")
	webhooksCreateCmd.Flags().StringVar(&webhookCreateURL, "url", "", "Callback URL (required)")
	webhooksCreateCmd.Flags().StringVar(&webhookCreateFormat, "format", "JSON", "Payload format: JSON or XML")

	webhooksCmd.AddCommand(webhooksListCmd, webhooksCreateCmd, webhooksDeleteCmd)
	rootCmd.AddCommand(webhooksCmd)
}
