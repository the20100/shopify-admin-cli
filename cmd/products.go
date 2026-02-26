package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "Manage Shopify products and variants",
}

// ---- products list ----

var (
	productsListFirst int
	productsListAfter string
	productsListQuery string
)

var productsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List products",
	Long: `List products in your Shopify store.

Use --query for Shopify search syntax, e.g.: status:active, vendor:Nike, title:shirt

Examples:
  shopify-admin products list
  shopify-admin products list --query "status:active"
  shopify-admin products list --first 10 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListProducts(productsListFirst, productsListAfter, productsListQuery)
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
			fmt.Println("No products found.")
			return nil
		}
		headers := []string{"ID", "TITLE", "STATUS", "VENDOR", "TYPE", "INVENTORY", "UPDATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			p := e.Node
			rows[i] = []string{
				shortID(p.ID),
				output.Truncate(p.Title, 40),
				strings.ToLower(p.Status),
				output.Truncate(p.Vendor, 20),
				output.Truncate(p.ProductType, 16),
				fmt.Sprintf("%d", p.TotalInventory),
				output.FormatTime(p.UpdatedAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

// ---- products get ----

var productsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific product",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := client.GetProduct(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(p, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(p.ID)},
			{"Title", p.Title},
			{"Status", strings.ToLower(p.Status)},
			{"Handle", p.Handle},
			{"Vendor", p.Vendor},
			{"Type", p.ProductType},
			{"Inventory", fmt.Sprintf("%d", p.TotalInventory)},
			{"Tags", output.FormatLabels(p.Tags)},
			{"Created", output.FormatTime(p.CreatedAt)},
			{"Updated", output.FormatTime(p.UpdatedAt)},
		})
		if len(p.Variants.Edges) > 0 {
			fmt.Println()
			fmt.Println("Variants:")
			headers := []string{"ID", "TITLE", "PRICE", "SKU", "INVENTORY"}
			rows := make([][]string, len(p.Variants.Edges))
			for i, e := range p.Variants.Edges {
				v := e.Node
				rows[i] = []string{
					shortID(v.ID),
					output.Truncate(v.Title, 30),
					v.Price,
					v.SKU,
					fmt.Sprintf("%d", v.InventoryQuantity),
				}
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

// ---- products create ----

var (
	productCreateVendor      string
	productCreateType        string
	productCreateStatus      string
	productCreateDescription string
	productCreateTags        string
)

var productsCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new product",
	Long: `Create a new product in your Shopify store.

Examples:
  shopify-admin products create "My Product"
  shopify-admin products create "T-Shirt" --vendor Nike --status active --tags "apparel,clothing"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := client.CreateProduct(
			args[0],
			productCreateVendor,
			productCreateType,
			productCreateStatus,
			productCreateDescription,
			splitTags(productCreateTags),
		)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(p, output.IsPretty(cmd))
		}
		fmt.Printf("Product created: %s\n", p.Title)
		fmt.Printf("ID:     %s\n", shortID(p.ID))
		fmt.Printf("Handle: %s\n", p.Handle)
		fmt.Printf("Status: %s\n", strings.ToLower(p.Status))
		return nil
	},
}

// ---- products update ----

var (
	productUpdateTitle       string
	productUpdateVendor      string
	productUpdateType        string
	productUpdateStatus      string
	productUpdateDescription string
	productUpdateTags        string
)

var productsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a product",
	Long: `Update fields of an existing product. Only specified flags are changed.

Examples:
  shopify-admin products update 1234567890 --status archived
  shopify-admin products update 1234567890 --title "New Title" --vendor "New Vendor"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := client.UpdateProduct(
			args[0],
			productUpdateTitle,
			productUpdateVendor,
			productUpdateType,
			productUpdateStatus,
			productUpdateDescription,
			splitTags(productUpdateTags),
		)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(p, output.IsPretty(cmd))
		}
		fmt.Printf("Product updated: %s\n", p.Title)
		fmt.Printf("ID:     %s\n", shortID(p.ID))
		fmt.Printf("Status: %s\n", strings.ToLower(p.Status))
		return nil
	},
}

// ---- products delete ----

var productsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a product (irreversible)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteProduct(args[0]); err != nil {
			return err
		}
		fmt.Printf("Product %s deleted.\n", args[0])
		return nil
	},
}

// ---- products variants ----

var variantsCmd = &cobra.Command{
	Use:   "variants",
	Short: "Manage product variants",
}

var variantsGetCmd = &cobra.Command{
	Use:   "get <variant-id>",
	Short: "Get details of a specific variant",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := client.GetVariant(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(v, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(v.ID)},
			{"Title", v.Title},
			{"Price", v.Price},
			{"Compare At", v.CompareAtPrice},
			{"SKU", v.SKU},
			{"Barcode", v.Barcode},
			{"Inventory", fmt.Sprintf("%d", v.InventoryQuantity)},
			{"Weight", fmt.Sprintf("%.3f %s", v.Weight, v.WeightUnit)},
			{"Updated", output.FormatTime(v.UpdatedAt)},
		})
		return nil
	},
}

var (
	variantUpdatePrice   string
	variantUpdateSKU     string
	variantUpdateBarcode string
)

var variantsUpdateCmd = &cobra.Command{
	Use:   "update <variant-id>",
	Short: "Update a product variant",
	Long: `Update price, SKU, or barcode of a variant.

Examples:
  shopify-admin products variants update 1234567890 --price 29.99
  shopify-admin products variants update 1234567890 --sku MY-SKU-001`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := client.UpdateVariant(args[0], variantUpdatePrice, variantUpdateSKU, variantUpdateBarcode)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(v, output.IsPretty(cmd))
		}
		fmt.Printf("Variant updated: %s\n", v.Title)
		fmt.Printf("ID:    %s\n", shortID(v.ID))
		fmt.Printf("Price: %s\n", v.Price)
		fmt.Printf("SKU:   %s\n", v.SKU)
		return nil
	},
}

func init() {
	// list
	productsListCmd.Flags().IntVar(&productsListFirst, "first", 50, "Number of products to return")
	productsListCmd.Flags().StringVar(&productsListAfter, "after", "", "Pagination cursor")
	productsListCmd.Flags().StringVar(&productsListQuery, "query", "", "Shopify search query (e.g. status:active)")

	// create
	productsCreateCmd.Flags().StringVar(&productCreateVendor, "vendor", "", "Product vendor")
	productsCreateCmd.Flags().StringVar(&productCreateType, "type", "", "Product type")
	productsCreateCmd.Flags().StringVar(&productCreateStatus, "status", "", "Status: active, draft, archived")
	productsCreateCmd.Flags().StringVar(&productCreateDescription, "desc", "", "HTML description")
	productsCreateCmd.Flags().StringVar(&productCreateTags, "tags", "", "Comma-separated tags")

	// update
	productsUpdateCmd.Flags().StringVar(&productUpdateTitle, "title", "", "New title")
	productsUpdateCmd.Flags().StringVar(&productUpdateVendor, "vendor", "", "New vendor")
	productsUpdateCmd.Flags().StringVar(&productUpdateType, "type", "", "New product type")
	productsUpdateCmd.Flags().StringVar(&productUpdateStatus, "status", "", "New status: active, draft, archived")
	productsUpdateCmd.Flags().StringVar(&productUpdateDescription, "desc", "", "New HTML description")
	productsUpdateCmd.Flags().StringVar(&productUpdateTags, "tags", "", "New comma-separated tags")

	// variants update
	variantsUpdateCmd.Flags().StringVar(&variantUpdatePrice, "price", "", "New price (e.g. 29.99)")
	variantsUpdateCmd.Flags().StringVar(&variantUpdateSKU, "sku", "", "New SKU")
	variantsUpdateCmd.Flags().StringVar(&variantUpdateBarcode, "barcode", "", "New barcode")

	variantsCmd.AddCommand(variantsGetCmd, variantsUpdateCmd)
	productsCmd.AddCommand(
		productsListCmd,
		productsGetCmd,
		productsCreateCmd,
		productsUpdateCmd,
		productsDeleteCmd,
		variantsCmd,
	)
	rootCmd.AddCommand(productsCmd)
}
