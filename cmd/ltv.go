package cmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/api"
	"github.com/the20100/shopify-admin-cli/internal/output"
)

// ── flags ─────────────────────────────────────────────────────────────────────

var (
	ltvStart   string
	ltvEnd     string
	ltvPeriod  string
	ltvExclude int
)

// ── result types ──────────────────────────────────────────────────────────────

type LTVResult struct {
	Period     LTVPeriod `json:"period"`
	NetSales   float64   `json:"net_sales"`
	Tax        float64   `json:"tax"`
	NetRevenue float64   `json:"net_revenue"` // net_sales - tax
	Customers  int64     `json:"customers"`
	LTV        float64   `json:"ltv"`
}

type LTVPeriod struct {
	Start         string `json:"start"`
	End           string `json:"end"`
	ExcludeMonths int    `json:"exclude_months,omitempty"`
	Period        string `json:"period,omitempty"`
}

// ── command ───────────────────────────────────────────────────────────────────

var ltvCmd = &cobra.Command{
	Use:   "ltv",
	Short: "Calculate store Lifetime Value (LTV)",
	Long: `Calculate the store's LTV for a given period.

Formula:
  LTV = (net_sales − tax) ÷ unique paying customers

Default period: last 3 years, excluding the last 3 months.

Date flags:
  --start YYYY-MM-DD  Override start date explicitly
  --end   YYYY-MM-DD  Override end date explicitly
  --period <N>        Duration back from end: Ny (years), Nm (months), Nd (days) — default: 3y
  --exclude <N>       Skip the last N months (gives data time to settle) — default: 3

Priority:
  --start / --end always win over --period / --exclude.
  If only --start is set, --end defaults to today minus --exclude.
  If only --end is set, --start defaults to --end minus --period.

Examples:
  shopify-admin ltv                                    # 3 years, excl. last 3 months
  shopify-admin ltv --period 2y --exclude 0            # 2 years up to today
  shopify-admin ltv --period 18m --exclude 1           # 18 months, excl. last month
  shopify-admin ltv --start 2022-01-01 --end 2024-01-01
  shopify-admin ltv --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start, end, err := resolveLTVDates(ltvStart, ltvEnd, ltvPeriod, ltvExclude)
		if err != nil {
			return err
		}

		startStr := start.Format("2006-01-02")
		endStr := end.Format("2006-01-02")

		fmt.Printf("Querying sales %s → %s…\n", startStr, endStr)
		netSales, tax, err := ltvQuerySales(startStr, endStr)
		if err != nil {
			return err
		}

		fmt.Println("Querying paying customers…")
		customers, err := ltvQueryCustomers(startStr, endStr)
		if err != nil {
			return err
		}
		if customers == 0 {
			return fmt.Errorf("no paying customers found for %s → %s", startStr, endStr)
		}

		netRevenue := netSales - tax
		result := LTVResult{
			Period: LTVPeriod{
				Start:         startStr,
				End:           endStr,
				ExcludeMonths: ltvExclude,
				Period:        ltvPeriod,
			},
			NetSales:   round2(netSales),
			Tax:        round2(tax),
			NetRevenue: round2(netRevenue),
			Customers:  customers,
			LTV:        round2(netRevenue / float64(customers)),
		}

		if output.IsJSON(cmd) {
			return output.PrintJSON(result, output.IsPretty(cmd))
		}
		printLTVReport(result)
		return nil
	},
}

// ── ShopifyQL queries ─────────────────────────────────────────────────────────

// ltvQuerySales runs a ShopifyQL query to fetch SUM(net_sales) and SUM(tax).
func ltvQuerySales(start, end string) (netSales, tax float64, err error) {
	q := fmt.Sprintf(
		"FROM sales SHOW SUM(net_sales), SUM(tax) SINCE %s UNTIL %s",
		start, end,
	)
	res, err := client.RunShopifyQL(q)
	if err != nil {
		return 0, 0, fmt.Errorf("sales query: %w", err)
	}
	if e := qlErrors(res); e != "" {
		return 0, 0, fmt.Errorf("ShopifyQL (sales): %s", e)
	}
	if res.TableData == nil || len(res.TableData.Rows) == 0 {
		return 0, 0, nil
	}
	row := res.TableData.Rows[0]
	ns, _ := qlFloat(row, res.TableData.Columns, "net_sales")
	t, _ := qlFloat(row, res.TableData.Columns, "tax")
	return ns, t, nil
}

// ltvQueryCustomers runs a ShopifyQL query to count unique paying customers.
func ltvQueryCustomers(start, end string) (int64, error) {
	q := fmt.Sprintf(
		"FROM customers SHOW COUNT(customer_id) SINCE %s UNTIL %s",
		start, end,
	)
	res, err := client.RunShopifyQL(q)
	if err != nil {
		return 0, fmt.Errorf("customers query: %w", err)
	}
	if e := qlErrors(res); e != "" {
		return 0, fmt.Errorf("ShopifyQL (customers): %s", e)
	}
	if res.TableData == nil || len(res.TableData.Rows) == 0 {
		return 0, nil
	}
	row := res.TableData.Rows[0]
	if len(row) == 0 {
		return 0, nil
	}
	// Try each column; fall back to first column if no named match.
	if n, ok := qlInt(row, res.TableData.Columns, "customer_id", "customers", "count"); ok {
		return n, nil
	}
	// Last resort: first cell.
	return qlParseInt(row[0])
}

// ── helpers ───────────────────────────────────────────────────────────────────

func qlErrors(res *api.ShopifyQLResult) string {
	return strings.TrimSpace(res.ParseErrors)
}

// qlFloat finds a column by partial name match and returns its value as float64.
func qlFloat(row []string, cols []api.ShopifyQLColumn, name string) (float64, bool) {
	for i, col := range cols {
		if i >= len(row) {
			break
		}
		if colMatches(col, name) {
			v := strings.TrimSpace(strings.ReplaceAll(row[i], ",", ""))
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f, true
			}
		}
	}
	return 0, false
}

// qlInt finds a column by any of the given name hints and returns int64.
func qlInt(row []string, cols []api.ShopifyQLColumn, names ...string) (int64, bool) {
	for _, name := range names {
		for i, col := range cols {
			if i >= len(row) {
				break
			}
			if colMatches(col, name) {
				if n, err := qlParseInt(row[i]); err == nil {
					return n, true
				}
			}
		}
	}
	return 0, false
}

func colMatches(col api.ShopifyQLColumn, name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(strings.ToLower(col.Name), n) ||
		strings.Contains(strings.ToLower(col.DisplayName), n)
}

func qlParseInt(s string) (int64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n, nil
	}
	// Some APIs return floats for counts (e.g. "1234.0")
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(math.Round(f)), nil
	}
	return 0, fmt.Errorf("cannot parse %q as integer", s)
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}

// ── date resolution ───────────────────────────────────────────────────────────

func resolveLTVDates(startFlag, endFlag, periodFlag string, excludeMonths int) (start, end time.Time, err error) {
	now := time.Now()

	// Resolve end date.
	if endFlag != "" {
		end, err = time.Parse("2006-01-02", endFlag)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --end %q: expected YYYY-MM-DD", endFlag)
		}
	} else {
		end = now.AddDate(0, -excludeMonths, 0)
	}

	// Resolve start date.
	if startFlag != "" {
		start, err = time.Parse("2006-01-02", startFlag)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid --start %q: expected YYYY-MM-DD", startFlag)
		}
	} else {
		start, err = subtractPeriod(end, periodFlag)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if !start.Before(end) {
		return time.Time{}, time.Time{}, fmt.Errorf(
			"start date %s must be before end date %s",
			start.Format("2006-01-02"), end.Format("2006-01-02"),
		)
	}
	return start, end, nil
}

// subtractPeriod subtracts a period string (e.g. "3y", "18m", "90d") from t.
func subtractPeriod(from time.Time, period string) (time.Time, error) {
	if period == "" {
		period = "3y"
	}
	period = strings.ToLower(strings.TrimSpace(period))
	if len(period) < 2 {
		return time.Time{}, fmt.Errorf("invalid --period %q: use Ny, Nm, or Nd (e.g. 3y, 18m, 90d)", period)
	}
	unit := period[len(period)-1]
	n, err := strconv.Atoi(period[:len(period)-1])
	if err != nil || n <= 0 {
		return time.Time{}, fmt.Errorf("invalid --period %q: use Ny, Nm, or Nd (e.g. 3y, 18m, 90d)", period)
	}
	switch unit {
	case 'y':
		return from.AddDate(-n, 0, 0), nil
	case 'm':
		return from.AddDate(0, -n, 0), nil
	case 'd':
		return from.AddDate(0, 0, -n), nil
	default:
		return time.Time{}, fmt.Errorf("unknown unit %q in --period: use y (years), m (months), d (days)", string(unit))
	}
}

// ── display ───────────────────────────────────────────────────────────────────

func printLTVReport(r LTVResult) {
	sep := strings.Repeat("─", 44)
	fmt.Println()
	fmt.Println("LTV Report")
	fmt.Println(sep)

	var periodDesc string
	if r.Period.Period != "" && r.Period.ExcludeMonths > 0 {
		periodDesc = fmt.Sprintf("  (%s, excl. last %dm)", r.Period.Period, r.Period.ExcludeMonths)
	} else if r.Period.Period != "" {
		periodDesc = fmt.Sprintf("  (%s)", r.Period.Period)
	} else if r.Period.ExcludeMonths > 0 {
		periodDesc = fmt.Sprintf("  (excl. last %dm)", r.Period.ExcludeMonths)
	}
	fmt.Printf("Period:       %s → %s%s\n", r.Period.Start, r.Period.End, periodDesc)
	fmt.Println(sep)
	fmt.Printf("Net Sales:    %14.2f\n", r.NetSales)
	fmt.Printf("Tax:        − %14.2f\n", r.Tax)
	fmt.Printf("Net Revenue:  %14.2f\n", r.NetRevenue)
	fmt.Printf("Customers:    %14d\n", r.Customers)
	fmt.Println(sep)
	fmt.Printf("LTV:          %14.2f  (store currency)\n", r.LTV)
	fmt.Println()
}

// ── init ──────────────────────────────────────────────────────────────────────

func init() {
	ltvCmd.Flags().StringVar(&ltvStart, "start", "", "Start date (YYYY-MM-DD) — overrides --period")
	ltvCmd.Flags().StringVar(&ltvEnd, "end", "", "End date (YYYY-MM-DD) — default: today minus --exclude months")
	ltvCmd.Flags().StringVar(&ltvPeriod, "period", "3y", "Look-back duration from end date: Ny, Nm, Nd (e.g. 3y, 18m, 90d)")
	ltvCmd.Flags().IntVar(&ltvExclude, "exclude", 3, "Exclude last N months from end date (0 = up to today)")
	rootCmd.AddCommand(ltvCmd)
}
