package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var customersCmd = &cobra.Command{
	Use:   "customers",
	Short: "Manage Shopify customers",
}

var (
	customersListFirst int
	customersListAfter string
	customersListQuery string
)

var customersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List customers",
	Long: `List customers in your Shopify store.

Use --query for Shopify search syntax, e.g.: email:john@example.com, state:enabled

Examples:
  shopify-admin customers list
  shopify-admin customers list --query "email:john@example.com"
  shopify-admin customers list --first 20 --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListCustomers(customersListFirst, customersListAfter, customersListQuery)
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
			fmt.Println("No customers found.")
			return nil
		}
		headers := []string{"ID", "NAME", "EMAIL", "ORDERS", "SPENT", "STATE", "CREATED"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			c := e.Node
			rows[i] = []string{
				shortID(c.ID),
				output.Truncate(strings.TrimSpace(c.FirstName+" "+c.LastName), 28),
				output.Truncate(c.Email, 32),
				c.NumberOfOrders,
				formatMoney(c.AmountSpent.Amount, c.AmountSpent.CurrencyCode),
				strings.ToLower(c.State),
				output.FormatTime(c.CreatedAt),
			}
		}
		output.PrintTable(headers, rows)
		if conn.PageInfo.HasNextPage {
			fmt.Printf("\n(more results — use --after %s)\n", conn.PageInfo.EndCursor)
		}
		return nil
	},
}

var customersGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.GetCustomer(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(c, output.IsPretty(cmd))
		}
		addr := "-"
		if c.DefaultAddress != nil {
			a := c.DefaultAddress
			addr = strings.TrimSpace(fmt.Sprintf("%s, %s %s %s", a.Address1, a.City, a.Province, a.Country))
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(c.ID)},
			{"Name", strings.TrimSpace(c.FirstName + " " + c.LastName)},
			{"Email", c.Email},
			{"Phone", c.Phone},
			{"State", strings.ToLower(c.State)},
			{"Orders", c.NumberOfOrders},
			{"Total Spent", formatMoney(c.AmountSpent.Amount, c.AmountSpent.CurrencyCode)},
			{"Tags", output.FormatLabels(c.Tags)},
			{"Address", addr},
			{"Created", output.FormatTime(c.CreatedAt)},
			{"Updated", output.FormatTime(c.UpdatedAt)},
		})
		return nil
	},
}

var (
	customerCreateFirst string
	customerCreateLast  string
	customerCreateEmail string
	customerCreatePhone string
	customerCreateTags  string
)

var customersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new customer",
	Long: `Create a new customer. At least one of --email or --phone is required by Shopify.

Examples:
  shopify-admin customers create --email john@example.com --first John --last Doe
  shopify-admin customers create --email john@example.com --phone "+15551234567"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if customerCreateEmail == "" && customerCreatePhone == "" {
			return fmt.Errorf("--email or --phone is required")
		}
		c, err := client.CreateCustomer(
			customerCreateFirst,
			customerCreateLast,
			customerCreateEmail,
			customerCreatePhone,
			splitTags(customerCreateTags),
		)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(c, output.IsPretty(cmd))
		}
		fmt.Printf("Customer created: %s %s\n", c.FirstName, c.LastName)
		fmt.Printf("ID:    %s\n", shortID(c.ID))
		fmt.Printf("Email: %s\n", c.Email)
		return nil
	},
}

var (
	customerUpdateFirst string
	customerUpdateLast  string
	customerUpdateEmail string
	customerUpdatePhone string
	customerUpdateTags  string
)

var customersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a customer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.UpdateCustomer(
			args[0],
			customerUpdateFirst,
			customerUpdateLast,
			customerUpdateEmail,
			customerUpdatePhone,
			splitTags(customerUpdateTags),
		)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(c, output.IsPretty(cmd))
		}
		fmt.Printf("Customer updated: %s %s\n", c.FirstName, c.LastName)
		fmt.Printf("ID:    %s\n", shortID(c.ID))
		return nil
	},
}

var customersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a customer (irreversible)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.DeleteCustomer(args[0]); err != nil {
			return err
		}
		fmt.Printf("Customer %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	customersListCmd.Flags().IntVar(&customersListFirst, "first", 50, "Number of customers to return")
	customersListCmd.Flags().StringVar(&customersListAfter, "after", "", "Pagination cursor")
	customersListCmd.Flags().StringVar(&customersListQuery, "query", "", "Shopify search query")

	customersCreateCmd.Flags().StringVar(&customerCreateFirst, "first", "", "First name")
	customersCreateCmd.Flags().StringVar(&customerCreateLast, "last", "", "Last name")
	customersCreateCmd.Flags().StringVar(&customerCreateEmail, "email", "", "Email address")
	customersCreateCmd.Flags().StringVar(&customerCreatePhone, "phone", "", "Phone number")
	customersCreateCmd.Flags().StringVar(&customerCreateTags, "tags", "", "Comma-separated tags")

	customersUpdateCmd.Flags().StringVar(&customerUpdateFirst, "first", "", "New first name")
	customersUpdateCmd.Flags().StringVar(&customerUpdateLast, "last", "", "New last name")
	customersUpdateCmd.Flags().StringVar(&customerUpdateEmail, "email", "", "New email address")
	customersUpdateCmd.Flags().StringVar(&customerUpdatePhone, "phone", "", "New phone number")
	customersUpdateCmd.Flags().StringVar(&customerUpdateTags, "tags", "", "New comma-separated tags")

	customersCmd.AddCommand(
		customersListCmd,
		customersGetCmd,
		customersCreateCmd,
		customersUpdateCmd,
		customersDeleteCmd,
	)
	rootCmd.AddCommand(customersCmd)
}
