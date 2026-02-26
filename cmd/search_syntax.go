package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

// QueryField describes one filterable field for a Shopify resource.
type QueryField struct {
	Field    string   `json:"field"`
	Type     string   `json:"type"` // string, enum, date, boolean, number
	Examples []string `json:"examples"`
	Notes    string   `json:"notes,omitempty"`
}

// ResourceSyntax holds the query syntax for one resource.
type ResourceSyntax struct {
	Resource  string       `json:"resource"`
	Command   string       `json:"command"`
	Fields    []QueryField `json:"fields"`
	Operators []string     `json:"operators"`
}

var resourceSyntaxMap = map[string]ResourceSyntax{
	"products": {
		Resource: "products",
		Command:  "shopify-admin products list --query \"...\"",
		Fields: []QueryField{
			{Field: "status", Type: "enum", Examples: []string{"status:active", "status:draft", "status:archived"}},
			{Field: "vendor", Type: "string", Examples: []string{"vendor:Nike", "vendor:\"Acme Corp\""}},
			{Field: "title", Type: "string", Examples: []string{"title:shirt", "title:\"Summer T-Shirt\""}},
			{Field: "product_type", Type: "string", Examples: []string{"product_type:Shirts", "product_type:Electronics"}},
			{Field: "tag", Type: "string", Examples: []string{"tag:sale", "tag:summer"}},
			{Field: "handle", Type: "string", Examples: []string{"handle:my-product-handle"}},
			{Field: "sku", Type: "string", Examples: []string{"sku:MY-SKU-001"}, Notes: "Searches variant SKUs"},
			{Field: "barcode", Type: "string", Examples: []string{"barcode:123456789"}, Notes: "Searches variant barcodes"},
			{Field: "created_at", Type: "date", Examples: []string{"created_at:>2024-01-01", "created_at:<2024-12-31", "created_at:2024-06-01"}},
			{Field: "updated_at", Type: "date", Examples: []string{"updated_at:>2024-06-01"}},
			{Field: "published_at", Type: "date", Examples: []string{"published_at:>2024-01-01"}},
		},
		Operators: sharedOperators(),
	},
	"collections": {
		Resource: "collections",
		Command:  "shopify-admin collections list --query \"...\"",
		Fields: []QueryField{
			{Field: "title", Type: "string", Examples: []string{"title:Summer", "title:\"Best Sellers\""}},
			{Field: "handle", Type: "string", Examples: []string{"handle:summer-sale"}},
			{Field: "collection_type", Type: "enum", Examples: []string{"collection_type:smart", "collection_type:custom"}},
			{Field: "updated_at", Type: "date", Examples: []string{"updated_at:>2024-01-01"}},
			{Field: "published_at", Type: "date", Examples: []string{"published_at:>2024-01-01"}},
		},
		Operators: sharedOperators(),
	},
	"orders": {
		Resource: "orders",
		Command:  "shopify-admin orders list --query \"...\"",
		Fields: []QueryField{
			{Field: "financial_status", Type: "enum", Examples: []string{"financial_status:paid", "financial_status:pending", "financial_status:refunded", "financial_status:authorized", "financial_status:voided", "financial_status:partially_paid", "financial_status:partially_refunded"}},
			{Field: "fulfillment_status", Type: "enum", Examples: []string{"fulfillment_status:unfulfilled", "fulfillment_status:fulfilled", "fulfillment_status:partial", "fulfillment_status:restocked"}},
			{Field: "status", Type: "enum", Examples: []string{"status:open", "status:closed", "status:cancelled", "status:any"}},
			{Field: "email", Type: "string", Examples: []string{"email:john@example.com"}},
			{Field: "name", Type: "string", Examples: []string{"name:#1001", "name:1001"}},
			{Field: "tag", Type: "string", Examples: []string{"tag:wholesale", "tag:vip"}},
			{Field: "created_at", Type: "date", Examples: []string{"created_at:>2024-01-01", "created_at:<2024-06-30"}},
			{Field: "processed_at", Type: "date", Examples: []string{"processed_at:>2024-01-01"}},
			{Field: "updated_at", Type: "date", Examples: []string{"updated_at:>2024-01-01"}},
			{Field: "customer_id", Type: "number", Examples: []string{"customer_id:1234567890"}},
			{Field: "gateway", Type: "string", Examples: []string{"gateway:shopify_payments", "gateway:paypal"}},
			{Field: "source_name", Type: "string", Examples: []string{"source_name:web", "source_name:pos", "source_name:iphone"}, Notes: "Order source"},
		},
		Operators: sharedOperators(),
	},
	"customers": {
		Resource: "customers",
		Command:  "shopify-admin customers list --query \"...\"",
		Fields: []QueryField{
			{Field: "email", Type: "string", Examples: []string{"email:john@example.com"}},
			{Field: "phone", Type: "string", Examples: []string{"phone:+15551234567"}},
			{Field: "first_name", Type: "string", Examples: []string{"first_name:John"}},
			{Field: "last_name", Type: "string", Examples: []string{"last_name:Doe"}},
			{Field: "state", Type: "enum", Examples: []string{"state:enabled", "state:disabled", "state:invited", "state:declined"}},
			{Field: "tag", Type: "string", Examples: []string{"tag:vip", "tag:wholesale"}},
			{Field: "created_at", Type: "date", Examples: []string{"created_at:>2024-01-01"}},
			{Field: "updated_at", Type: "date", Examples: []string{"updated_at:>2024-06-01"}},
			{Field: "orders_count", Type: "number", Examples: []string{"orders_count:>5", "orders_count:0"}, Notes: "Number of orders placed"},
			{Field: "total_spent", Type: "number", Examples: []string{"total_spent:>100", "total_spent:<50"}, Notes: "Total amount spent"},
		},
		Operators: sharedOperators(),
	},
	"inventory": {
		Resource: "inventory items",
		Command:  "shopify-admin inventory items --query \"...\"",
		Fields: []QueryField{
			{Field: "sku", Type: "string", Examples: []string{"sku:MY-SKU-001", "sku:\"SKU WITH SPACE\""}},
			{Field: "tracked", Type: "boolean", Examples: []string{"tracked:true", "tracked:false"}},
			{Field: "requires_shipping", Type: "boolean", Examples: []string{"requires_shipping:true"}},
			{Field: "created_at", Type: "date", Examples: []string{"created_at:>2024-01-01"}},
			{Field: "updated_at", Type: "date", Examples: []string{"updated_at:>2024-01-01"}},
		},
		Operators: sharedOperators(),
	},
	"discounts": {
		Resource: "discounts",
		Command:  "shopify-admin discounts list --query \"...\"",
		Fields: []QueryField{
			{Field: "status", Type: "enum", Examples: []string{"status:active", "status:expired", "status:scheduled"}},
			{Field: "title", Type: "string", Examples: []string{"title:SUMMER", "title:\"10% off\""}},
			{Field: "starts_at", Type: "date", Examples: []string{"starts_at:>2024-01-01"}},
			{Field: "ends_at", Type: "date", Examples: []string{"ends_at:<2024-12-31"}},
			{Field: "discount_type", Type: "enum", Examples: []string{"discount_type:code", "discount_type:automatic"}},
		},
		Operators: sharedOperators(),
	},
}

func sharedOperators() []string {
	return []string{
		"field:value              exact match",
		"field:>value             greater than (dates, numbers)",
		"field:<value             less than",
		"field:>=value            greater than or equal",
		"field:<=value            less than or equal",
		"field:value1 field:value2  AND (space-separated)",
		"(field:a OR field:b)     OR",
		"NOT field:value          negation",
		"field:\"multi word\"      phrase with spaces — use quotes",
	}
}

var searchSyntaxCmd = &cobra.Command{
	Use:   "search-syntax [resource]",
	Short: "Show supported --query filter fields for each resource",
	Long: `Show the available search/filter fields for Shopify resource list commands.

When no resource is given, all resources are listed.
Pass a resource name to see its fields in detail.

Available resources: products, collections, orders, customers, inventory, discounts

Examples:
  shopify-admin search-syntax             # show all resources
  shopify-admin search-syntax orders      # show orders fields
  shopify-admin search-syntax products    # show products fields
  shopify-admin search-syntax --json      # machine-readable output`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine which resources to show
		var toShow []ResourceSyntax
		if len(args) == 1 {
			key := strings.ToLower(args[0])
			rs, ok := resourceSyntaxMap[key]
			if !ok {
				keys := make([]string, 0, len(resourceSyntaxMap))
				for k := range resourceSyntaxMap {
					keys = append(keys, k)
				}
				return fmt.Errorf("unknown resource %q — available: %s", args[0], strings.Join(keys, ", "))
			}
			toShow = []ResourceSyntax{rs}
		} else {
			order := []string{"products", "collections", "orders", "customers", "inventory", "discounts"}
			for _, k := range order {
				toShow = append(toShow, resourceSyntaxMap[k])
			}
		}

		if output.IsJSON(cmd) {
			if len(toShow) == 1 {
				return output.PrintJSON(toShow[0], output.IsPretty(cmd))
			}
			return output.PrintJSON(toShow, output.IsPretty(cmd))
		}

		// Human-readable output
		for i, rs := range toShow {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("── %s  (%s)\n", strings.ToUpper(rs.Resource), rs.Command)
			fmt.Println()

			// Fields table
			headers := []string{"FIELD", "TYPE", "EXAMPLES"}
			rows := make([][]string, len(rs.Fields))
			for j, f := range rs.Fields {
				examples := strings.Join(f.Examples[:min(2, len(f.Examples))], "  or  ")
				rows[j] = []string{f.Field, f.Type, examples}
			}
			output.PrintTable(headers, rows)

			// Operators section (only when showing single resource)
			if len(toShow) == 1 {
				fmt.Println()
				fmt.Println("Operators:")
				for _, op := range rs.Operators {
					fmt.Printf("  %s\n", op)
				}
				fmt.Println()
				fmt.Println("Combining filters (AND):")
				fmt.Println("  shopify-admin orders list --query \"financial_status:paid fulfillment_status:unfulfilled\"")
				fmt.Println()
				fmt.Println("Date range:")
				fmt.Println("  shopify-admin orders list --query \"created_at:>2024-01-01 created_at:<2024-12-31\"")
			}
		}

		if len(toShow) > 1 {
			fmt.Println()
			fmt.Println("Operators: >, <, >=, <=, NOT, AND (space), OR (parentheses), \"phrase\"")
			fmt.Println("Run 'shopify-admin search-syntax <resource>' for full operator docs.")
		}

		return nil
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	rootCmd.AddCommand(searchSyntaxCmd)
}
