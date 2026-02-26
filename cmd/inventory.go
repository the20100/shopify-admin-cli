package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Manage Shopify inventory",
}

// ---- inventory locations ----

var (
	inventoryLocationsFirst int
)

var inventoryLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List inventory locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListLocations(inventoryLocationsFirst)
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
			fmt.Println("No locations found.")
			return nil
		}
		headers := []string{"ID", "NAME", "CITY", "COUNTRY", "ACTIVE"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			l := e.Node
			rows[i] = []string{
				shortID(l.ID),
				output.Truncate(l.Name, 36),
				l.Address.City,
				l.Address.Country,
				output.FormatBool(l.IsActive),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

// ---- inventory levels ----

var (
	inventoryLevelsLocation string
	inventoryLevelsFirst    int
)

var inventoryLevelsCmd = &cobra.Command{
	Use:   "levels",
	Short: "List inventory levels for a location",
	Long: `List inventory levels for a specific location.

Examples:
  shopify-admin inventory levels --location <location-id>
  shopify-admin inventory levels --location 12345 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if inventoryLevelsLocation == "" {
			return fmt.Errorf("--location is required")
		}
		conn, err := client.ListInventoryLevels(inventoryLevelsLocation, inventoryLevelsFirst)
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
			fmt.Println("No inventory levels found.")
			return nil
		}
		headers := []string{"ITEM ID", "SKU", "AVAILABLE", "ON HAND"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			lvl := e.Node
			available := "-"
			onHand := "-"
			for _, q := range lvl.Quantities {
				switch strings.ToLower(q.Name) {
				case "available":
					available = fmt.Sprintf("%d", q.Quantity)
				case "on_hand":
					onHand = fmt.Sprintf("%d", q.Quantity)
				}
			}
			rows[i] = []string{
				shortID(lvl.Item.ID),
				lvl.Item.SKU,
				available,
				onHand,
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results available)\n")
		}
		return nil
	},
}

// ---- inventory items ----

var (
	inventoryItemsFirst int
	inventoryItemsAfter string
	inventoryItemsQuery string
)

var inventoryItemsCmd = &cobra.Command{
	Use:   "items",
	Short: "List inventory items",
	Long: `List inventory items. Use --query to filter by SKU or other fields.

Examples:
  shopify-admin inventory items
  shopify-admin inventory items --query "sku:MY-SKU"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListInventoryItems(inventoryItemsFirst, inventoryItemsAfter, inventoryItemsQuery)
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
			fmt.Println("No inventory items found.")
			return nil
		}
		headers := []string{"ID", "SKU", "TRACKED", "REQUIRES SHIPPING"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			item := e.Node
			rows[i] = []string{
				shortID(item.ID),
				item.SKU,
				output.FormatBool(item.Tracked),
				output.FormatBool(item.RequiresShipping),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

// ---- inventory adjust ----

var (
	inventoryAdjustItem     string
	inventoryAdjustLocation string
	inventoryAdjustDelta    int
	inventoryAdjustReason   string
)

var inventoryAdjustCmd = &cobra.Command{
	Use:   "adjust",
	Short: "Adjust inventory quantity for an item at a location",
	Long: `Adjust the available inventory quantity for a specific item at a location.

Valid reasons: correction, received, return, damaged, theft, other

Examples:
  shopify-admin inventory adjust --item 12345 --location 67890 --delta 10
  shopify-admin inventory adjust --item 12345 --location 67890 --delta -5 --reason damaged`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if inventoryAdjustItem == "" {
			return fmt.Errorf("--item is required")
		}
		if inventoryAdjustLocation == "" {
			return fmt.Errorf("--location is required")
		}
		if err := client.AdjustInventory(
			inventoryAdjustItem,
			inventoryAdjustLocation,
			inventoryAdjustDelta,
			inventoryAdjustReason,
		); err != nil {
			return err
		}
		sign := "+"
		if inventoryAdjustDelta < 0 {
			sign = ""
		}
		fmt.Printf("Inventory adjusted: %s%d units for item %s at location %s\n",
			sign, inventoryAdjustDelta, inventoryAdjustItem, inventoryAdjustLocation)
		return nil
	},
}

func init() {
	inventoryLocationsCmd.Flags().IntVar(&inventoryLocationsFirst, "first", 50, "Number of locations to return")

	inventoryLevelsCmd.Flags().StringVar(&inventoryLevelsLocation, "location", "", "Location ID (required)")
	inventoryLevelsCmd.Flags().IntVar(&inventoryLevelsFirst, "first", 100, "Number of inventory levels to return")

	inventoryItemsCmd.Flags().IntVar(&inventoryItemsFirst, "first", 50, "Number of inventory items to return")
	inventoryItemsCmd.Flags().StringVar(&inventoryItemsAfter, "after", "", "Pagination cursor")
	inventoryItemsCmd.Flags().StringVar(&inventoryItemsQuery, "query", "", "Search query (e.g. sku:MY-SKU)")

	inventoryAdjustCmd.Flags().StringVar(&inventoryAdjustItem, "item", "", "Inventory item ID (required)")
	inventoryAdjustCmd.Flags().StringVar(&inventoryAdjustLocation, "location", "", "Location ID (required)")
	inventoryAdjustCmd.Flags().IntVar(&inventoryAdjustDelta, "delta", 0, "Quantity change (positive=add, negative=subtract)")
	inventoryAdjustCmd.Flags().StringVar(&inventoryAdjustReason, "reason", "correction", "Adjustment reason")

	inventoryCmd.AddCommand(
		inventoryLocationsCmd,
		inventoryLevelsCmd,
		inventoryItemsCmd,
		inventoryAdjustCmd,
	)
	rootCmd.AddCommand(inventoryCmd)
}
