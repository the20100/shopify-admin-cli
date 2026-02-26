package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var metafieldsCmd = &cobra.Command{
	Use:   "metafields",
	Short: "Manage Shopify metafields",
}

var (
	metafieldsListOwner string
	metafieldsListFirst int
)

var metafieldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List metafields for a resource",
	Long: `List metafields attached to a specific resource.

The --owner flag accepts a Shopify GID or numeric ID with resource type:
  gid://shopify/Product/1234567890
  gid://shopify/Order/1234567890
  gid://shopify/Customer/1234567890

Examples:
  shopify-admin metafields list --owner "gid://shopify/Product/1234567890"
  shopify-admin metafields list --owner "gid://shopify/Shop/1" --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if metafieldsListOwner == "" {
			return fmt.Errorf("--owner is required (e.g. gid://shopify/Product/123)")
		}
		conn, err := client.ListMetafields(metafieldsListOwner, metafieldsListFirst)
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
			fmt.Println("No metafields found.")
			return nil
		}
		headers := []string{"ID", "NAMESPACE", "KEY", "TYPE", "VALUE", "UPDATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			m := e.Node
			rows[i] = []string{
				shortID(m.ID),
				m.Namespace,
				m.Key,
				m.Type,
				output.Truncate(m.Value, 40),
				output.FormatTime(m.UpdatedAt),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

var metafieldsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := client.GetMetafield(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(m, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(m.ID)},
			{"Namespace", m.Namespace},
			{"Key", m.Key},
			{"Type", m.Type},
			{"Value", m.Value},
			{"Created", output.FormatTime(m.CreatedAt)},
			{"Updated", output.FormatTime(m.UpdatedAt)},
		})
		return nil
	},
}

var (
	metafieldSetOwner     string
	metafieldSetNamespace string
	metafieldSetKey       string
	metafieldSetValue     string
	metafieldSetType      string
)

var metafieldsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or update a metafield",
	Long: `Create or update a metafield on a resource.

Common types: single_line_text_field, multi_line_text_field, number_integer,
              number_decimal, boolean, date, date_time, json, color, url,
              rating, dimension, volume, weight

Examples:
  shopify-admin metafields set --owner "gid://shopify/Product/123" \
    --namespace custom --key my_field --value "Hello" --type single_line_text_field`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if metafieldSetOwner == "" {
			return fmt.Errorf("--owner is required")
		}
		if metafieldSetNamespace == "" {
			return fmt.Errorf("--namespace is required")
		}
		if metafieldSetKey == "" {
			return fmt.Errorf("--key is required")
		}
		if metafieldSetValue == "" {
			return fmt.Errorf("--value is required")
		}
		if metafieldSetType == "" {
			return fmt.Errorf("--type is required")
		}
		m, err := client.SetMetafield(metafieldSetOwner, metafieldSetNamespace, metafieldSetKey, metafieldSetValue, metafieldSetType)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(m, output.IsPretty(cmd))
		}
		fmt.Printf("Metafield set: %s.%s\n", m.Namespace, m.Key)
		fmt.Printf("ID:    %s\n", shortID(m.ID))
		fmt.Printf("Type:  %s\n", m.Type)
		fmt.Printf("Value: %s\n", m.Value)
		return nil
	},
}

var metafieldsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a metafield",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteMetafield(args[0]); err != nil {
			return err
		}
		fmt.Printf("Metafield %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	metafieldsListCmd.Flags().StringVar(&metafieldsListOwner, "owner", "", "Resource GID (e.g. gid://shopify/Product/123)")
	metafieldsListCmd.Flags().IntVar(&metafieldsListFirst, "first", 100, "Number of metafields to return")

	metafieldsSetCmd.Flags().StringVar(&metafieldSetOwner, "owner", "", "Resource GID (required)")
	metafieldsSetCmd.Flags().StringVar(&metafieldSetNamespace, "namespace", "", "Metafield namespace (required)")
	metafieldsSetCmd.Flags().StringVar(&metafieldSetKey, "key", "", "Metafield key (required)")
	metafieldsSetCmd.Flags().StringVar(&metafieldSetValue, "value", "", "Metafield value (required)")
	metafieldsSetCmd.Flags().StringVar(&metafieldSetType, "type", "", "Metafield type (required, e.g. single_line_text_field)")

	metafieldsCmd.AddCommand(
		metafieldsListCmd,
		metafieldsGetCmd,
		metafieldsSetCmd,
		metafieldsDeleteCmd,
	)
	rootCmd.AddCommand(metafieldsCmd)
}
