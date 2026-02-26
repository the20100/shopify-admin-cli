package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/api"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var metaobjectsCmd = &cobra.Command{
	Use:   "metaobjects",
	Short: "Manage Shopify metaobjects",
}

// ---- metaobjects definitions ----

var metaobjectDefinitionsCmd = &cobra.Command{
	Use:   "definitions",
	Short: "List metaobject definitions (types)",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListMetaobjectDefinitions(100)
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
			fmt.Println("No metaobject definitions found.")
			return nil
		}
		headers := []string{"ID", "TYPE", "NAME", "DESCRIPTION"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			d := e.Node
			rows[i] = []string{
				shortID(d.ID),
				d.Type,
				d.Name,
				output.Truncate(d.Description, 50),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

// ---- metaobjects list ----

var (
	metaobjectsListType  string
	metaobjectsListFirst int
	metaobjectsListAfter string
)

var metaobjectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List metaobjects of a given type",
	Long: `List metaobjects of a specific type. Use 'definitions' to see available types.

Examples:
  shopify-admin metaobjects list --type my_custom_type
  shopify-admin metaobjects list --type product_feature --first 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if metaobjectsListType == "" {
			return fmt.Errorf("--type is required")
		}
		conn, err := client.ListMetaobjects(metaobjectsListType, metaobjectsListFirst, metaobjectsListAfter)
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
			fmt.Printf("No metaobjects of type '%s' found.\n", metaobjectsListType)
			return nil
		}
		headers := []string{"ID", "HANDLE", "TYPE", "UPDATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			m := e.Node
			rows[i] = []string{
				shortID(m.ID),
				m.Handle,
				m.Type,
				output.FormatTime(m.UpdatedAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

// ---- metaobjects get ----

var metaobjectsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific metaobject",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := client.GetMetaobject(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(m, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(m.ID)},
			{"Handle", m.Handle},
			{"Type", m.Type},
			{"Updated", output.FormatTime(m.UpdatedAt)},
		})
		if len(m.Fields) > 0 {
			fmt.Println()
			fmt.Println("Fields:")
			headers := []string{"KEY", "TYPE", "VALUE"}
			rows := make([][]string, len(m.Fields))
			for i, f := range m.Fields {
				rows[i] = []string{f.Key, f.Type, output.Truncate(f.Value, 60)}
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

// ---- metaobjects create ----

var (
	metaobjectCreateType   string
	metaobjectCreateHandle string
	metaobjectCreateFields []string
)

var metaobjectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new metaobject",
	Long: `Create a new metaobject. Use --field key=value for each field.

Examples:
  shopify-admin metaobjects create --type my_type --field title="Hello" --field description="World"
  shopify-admin metaobjects create --type my_type --handle my-object --field key=value`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if metaobjectCreateType == "" {
			return fmt.Errorf("--type is required")
		}
		fields := parseFieldPairs(metaobjectCreateFields)
		m, err := client.CreateMetaobject(metaobjectCreateType, metaobjectCreateHandle, fields)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(m, output.IsPretty(cmd))
		}
		fmt.Printf("Metaobject created: %s\n", m.Handle)
		fmt.Printf("ID:   %s\n", shortID(m.ID))
		fmt.Printf("Type: %s\n", m.Type)
		return nil
	},
}

// ---- metaobjects update ----

var (
	metaobjectUpdateFields []string
)

var metaobjectsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update metaobject fields",
	Long: `Update fields of a metaobject. Use --field key=value for each field to update.

Examples:
  shopify-admin metaobjects update 1234567890 --field title="New Title"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(metaobjectUpdateFields) == 0 {
			return fmt.Errorf("at least one --field key=value is required")
		}
		fields := parseFieldPairs(metaobjectUpdateFields)
		m, err := client.UpdateMetaobject(args[0], fields)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(m, output.IsPretty(cmd))
		}
		fmt.Printf("Metaobject updated: %s\n", m.Handle)
		fmt.Printf("ID: %s\n", shortID(m.ID))
		return nil
	},
}

// ---- metaobjects delete ----

var metaobjectsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a metaobject (irreversible)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteMetaobject(args[0]); err != nil {
			return err
		}
		fmt.Printf("Metaobject %s deleted.\n", args[0])
		return nil
	},
}

// parseFieldPairs converts ["key=value", "key2=value2"] into MetaobjectField slice.
func parseFieldPairs(pairs []string) []api.MetaobjectField {
	fields := make([]api.MetaobjectField, 0, len(pairs))
	for _, p := range pairs {
		idx := strings.Index(p, "=")
		if idx < 0 {
			continue
		}
		fields = append(fields, api.MetaobjectField{
			Key:   p[:idx],
			Value: p[idx+1:],
		})
	}
	return fields
}

func init() {
	metaobjectsListCmd.Flags().StringVar(&metaobjectsListType, "type", "", "Metaobject type (required)")
	metaobjectsListCmd.Flags().IntVar(&metaobjectsListFirst, "first", 50, "Number of metaobjects to return")
	metaobjectsListCmd.Flags().StringVar(&metaobjectsListAfter, "after", "", "Pagination cursor")

	metaobjectsCreateCmd.Flags().StringVar(&metaobjectCreateType, "type", "", "Metaobject type (required)")
	metaobjectsCreateCmd.Flags().StringVar(&metaobjectCreateHandle, "handle", "", "Handle (optional)")
	metaobjectsCreateCmd.Flags().StringArrayVar(&metaobjectCreateFields, "field", nil, "Field in key=value format (repeatable)")

	metaobjectsUpdateCmd.Flags().StringArrayVar(&metaobjectUpdateFields, "field", nil, "Field in key=value format (repeatable)")

	metaobjectsCmd.AddCommand(
		metaobjectDefinitionsCmd,
		metaobjectsListCmd,
		metaobjectsGetCmd,
		metaobjectsCreateCmd,
		metaobjectsUpdateCmd,
		metaobjectsDeleteCmd,
	)
	rootCmd.AddCommand(metaobjectsCmd)
}
