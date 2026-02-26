package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

var marketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "View Shopify Markets",
}

var marketsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Shopify Markets",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := client.ListMarkets(50)
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
			fmt.Println("No markets found.")
			return nil
		}
		headers := []string{"ID", "NAME", "HANDLE", "ENABLED", "PRIMARY", "REGIONS"}
		rows := make([][]string, len(conn.Edges))
		for i, e := range conn.Edges {
			m := e.Node
			regions := make([]string, 0, len(m.Regions.Edges))
			for _, r := range m.Regions.Edges {
				regions = append(regions, r.Node.Name)
			}
			regionStr := strings.Join(regions, ", ")
			if regionStr == "" {
				regionStr = "-"
			}
			rows[i] = []string{
				shortID(m.ID),
				m.Name,
				m.Handle,
				output.FormatBool(m.Enabled),
				output.FormatBool(m.Primary),
				output.Truncate(regionStr, 40),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

var marketsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get details of a specific market",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := client.GetMarket(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(m, output.IsPretty(cmd))
		}
		regions := make([]string, 0, len(m.Regions.Edges))
		for _, r := range m.Regions.Edges {
			regions = append(regions, r.Node.Name)
		}
		output.PrintKeyValue([][]string{
			{"ID", shortID(m.ID)},
			{"Name", m.Name},
			{"Handle", m.Handle},
			{"Enabled", output.FormatBool(m.Enabled)},
			{"Primary", output.FormatBool(m.Primary)},
			{"Regions", output.FormatLabels(regions)},
		})
		return nil
	},
}

func init() {
	marketsCmd.AddCommand(marketsListCmd, marketsGetCmd)
	rootCmd.AddCommand(marketsCmd)
}
