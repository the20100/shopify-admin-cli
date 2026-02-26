package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var collectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "Manage Shopify collections",
}

var (
	collectionsListFirst int
	collectionsListAfter string
	collectionsListQuery string
)

var collectionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections",
	Long: `List collections in your Shopify store.

Examples:
  shopify-admin collections list
  shopify-admin collections list --query "title:Summer" --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListCollections(collectionsListFirst, collectionsListAfter, collectionsListQuery)
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
			fmt.Println("No collections found.")
			return nil
		}
		headers := []string{"ID", "TITLE", "HANDLE", "PRODUCTS", "UPDATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			c := e.Node
			rows[i] = []string{
				shortID(c.ID),
				output.Truncate(c.Title, 40),
				c.Handle,
				fmt.Sprintf("%d", c.ProductsCount.Count),
				output.FormatTime(c.UpdatedAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

var collectionsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.GetCollection(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(c, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(c.ID)},
			{"Title", c.Title},
			{"Handle", c.Handle},
			{"Products", fmt.Sprintf("%d", c.ProductsCount.Count)},
			{"Updated", output.FormatTime(c.UpdatedAt)},
			{"Description", output.Truncate(c.Description, 80)},
		})
		return nil
	},
}

var (
	collectionCreateDesc string
)

var collectionsCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.CreateCollection(args[0], collectionCreateDesc)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(c, output.IsPretty(cmd))
		}
		fmt.Printf("Collection created: %s\n", c.Title)
		fmt.Printf("ID:     %s\n", shortID(c.ID))
		fmt.Printf("Handle: %s\n", c.Handle)
		return nil
	},
}

var (
	collectionUpdateTitle string
	collectionUpdateDesc  string
)

var collectionsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.UpdateCollection(args[0], collectionUpdateTitle, collectionUpdateDesc)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(c, output.IsPretty(cmd))
		}
		fmt.Printf("Collection updated: %s\n", c.Title)
		fmt.Printf("ID: %s\n", shortID(c.ID))
		return nil
	},
}

var collectionsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a collection (irreversible)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteCollection(args[0]); err != nil {
			return err
		}
		fmt.Printf("Collection %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	collectionsListCmd.Flags().IntVar(&collectionsListFirst, "first", 50, "Number of collections to return")
	collectionsListCmd.Flags().StringVar(&collectionsListAfter, "after", "", "Pagination cursor")
	collectionsListCmd.Flags().StringVar(&collectionsListQuery, "query", "", "Shopify search query")

	collectionsCreateCmd.Flags().StringVar(&collectionCreateDesc, "desc", "", "HTML description")

	collectionsUpdateCmd.Flags().StringVar(&collectionUpdateTitle, "title", "", "New title")
	collectionsUpdateCmd.Flags().StringVar(&collectionUpdateDesc, "desc", "", "New HTML description")

	collectionsCmd.AddCommand(
		collectionsListCmd,
		collectionsGetCmd,
		collectionsCreateCmd,
		collectionsUpdateCmd,
		collectionsDeleteCmd,
	)
	rootCmd.AddCommand(collectionsCmd)
}
