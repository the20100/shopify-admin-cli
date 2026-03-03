# shopify-admin

A CLI tool for the [Shopify Admin GraphQL API](https://shopify.dev/docs/api/admin-graphql/latest).

Outputs JSON when piped (for agent/script use) and human-readable tables in a terminal.

## Installation

```bash
git clone https://github.com/the20100/shopify-admin-cli
cd shopify-admin-cli
go build -o shopify-admin .
mv shopify-admin /usr/local/bin/
```

Requires Go 1.22+.

## Authentication

Since January 2026, Shopify no longer issues permanent tokens from the admin UI. You need a **custom app** created in the [Shopify Partners dashboard](https://partners.shopify.com) or [Dev Dashboard](https://dev.shopify.com) and its **Client ID + Client Secret**.

### OAuth flow (recommended)

```bash
# 1. Save your app's Client ID and Client Secret
shopify-admin auth configure <client-id> <client-secret>

# 2. Install the app and obtain an API token
#    Use --no-browser in remote/headless environments
shopify-admin auth login --shop mystore --no-browser
shopify-admin auth login --shop mystore          # opens browser automatically
```

**How `--no-browser` works:**
1. The CLI prints a Shopify authorization URL — open it in any browser
2. Approve the app in the Shopify admin
3. The browser tries to redirect to `http://localhost/callback?code=...` (will fail — that's expected)
4. Copy the full URL from the browser address bar and paste it into the CLI prompt
5. Done — the CLI exchanges the code and fetches a real API token

**Tokens auto-refresh silently.** Once the app is installed, every command checks the token's expiry and refreshes it via the Client Credentials Grant when needed (tokens last ~24 h). No manual re-authentication required.

### Manual token (legacy)

```bash
shopify-admin auth setup mystore.myshopify.com <access-token>
```

### Environment variables

```bash
export SHOPIFY_SHOP=mystore.myshopify.com
export SHOPIFY_ACCESS_TOKEN=<token>
```

> **Note:** env vars take priority over the config file and **bypass auto-refresh**. Remove them (`unset SHOPIFY_SHOP SHOPIFY_ACCESS_TOKEN`) if you want the OAuth flow to handle tokens automatically.

Credentials are stored in:
- macOS: `~/Library/Application Support/shopify-admin/config.json`
- Linux: `~/.config/shopify-admin/config.json`
- Windows: `%AppData%\shopify-admin\config.json`

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Force JSON output |
| `--pretty` | Force pretty-printed JSON output (implies --json) |

Output is **auto-detected**: JSON when piped, human-readable tables in terminal.

---

## Commands

### `info`
Show binary location, config path, and active credential source.
```bash
shopify-admin info
```

### `auth`
```bash
# OAuth flow (recommended)
shopify-admin auth configure <client-id> <client-secret>   # Save app credentials
shopify-admin auth login --shop <shop> --no-browser        # Install app + get token (remote)
shopify-admin auth login --shop <shop>                     # Install app + get token (local)

# Manual / legacy
shopify-admin auth setup <shop> <access-token>             # Save a token directly

# Utilities
shopify-admin auth status                                  # Show active credential source
shopify-admin auth logout                                  # Remove all saved credentials
```

---

### `search-syntax`
Show which `--query` fields and operators are valid for each resource. No authentication required.
```bash
shopify-admin search-syntax                     # All resources (table)
shopify-admin search-syntax orders              # Orders fields + operators (table)
shopify-admin search-syntax products --json     # Machine-readable JSON
shopify-admin search-syntax customers --pretty  # Pretty JSON
```
Available resources: `products`, `collections`, `orders`, `customers`, `inventory`, `discounts`

**Query operators:**
| Operator | Meaning | Example |
|----------|---------|---------|
| `field:value` | Exact match | `status:active` |
| `field:>value` | Greater than | `created_at:>2024-01-01` |
| `field:<value` | Less than | `total_spent:<100` |
| `field:v1 field:v2` | AND (space) | `financial_status:paid fulfillment_status:unfulfilled` |
| `(field:a OR field:b)` | OR | `(status:open OR status:closed)` |
| `NOT field:value` | Negation | `NOT tag:archived` |
| `field:"multi word"` | Phrase | `vendor:"Acme Corp"` |

---

### `shop`
```bash
shopify-admin shop info         # Show store details (name, plan, currency, etc.)
```

---

### `products`
```bash
shopify-admin products list                            # List products
shopify-admin products list --query "status:active"    # Filter by status
shopify-admin products list --first 10 --after CURSOR  # Pagination
shopify-admin products get <id>                        # Get product details + variants
shopify-admin products create "T-Shirt" --vendor Nike --status active --tags "apparel"
shopify-admin products update <id> --title "New Title" --status archived
shopify-admin products delete <id>
```

**Variant commands:**
```bash
shopify-admin products variants get <variant-id>
shopify-admin products variants update <variant-id> --price 29.99 --sku MY-SKU
```

---

### `collections`
```bash
shopify-admin collections list
shopify-admin collections get <id>
shopify-admin collections create "Summer Sale" --desc "Summer items"
shopify-admin collections update <id> --title "New Title"
shopify-admin collections delete <id>
```

---

### `orders`
```bash
shopify-admin orders list
shopify-admin orders list --query "financial_status:paid"
shopify-admin orders list --query "fulfillment_status:unfulfilled"
shopify-admin orders get <id>
shopify-admin orders close <id>
shopify-admin orders cancel <id> --reason customer --refund --restock
shopify-admin orders mark-paid <id>
```

---

### `customers`
```bash
shopify-admin customers list
shopify-admin customers list --query "email:john@example.com"
shopify-admin customers get <id>
shopify-admin customers create --email john@example.com --first John --last Doe
shopify-admin customers update <id> --email new@example.com --tags "vip,wholesale"
shopify-admin customers delete <id>
```

---

### `inventory`
```bash
shopify-admin inventory locations               # List all locations
shopify-admin inventory items                   # List inventory items
shopify-admin inventory items --query "sku:MY-SKU"
shopify-admin inventory levels --location <id>  # Inventory at a location
shopify-admin inventory adjust --item <id> --location <id> --delta 10
shopify-admin inventory adjust --item <id> --location <id> --delta -5 --reason damaged
```

Valid adjustment reasons: `correction`, `received`, `return`, `damaged`, `theft`, `other`

---

### `metafields`
```bash
shopify-admin metafields list --owner "gid://shopify/Product/123"
shopify-admin metafields get <metafield-id>
shopify-admin metafields set --owner "gid://shopify/Product/123" \
  --namespace custom --key my_key --value "Hello" --type single_line_text_field
shopify-admin metafields delete <metafield-id>
```

Common types: `single_line_text_field`, `multi_line_text_field`, `number_integer`,
`number_decimal`, `boolean`, `date`, `date_time`, `json`, `color`, `url`

---

### `metaobjects`
```bash
shopify-admin metaobjects definitions              # List metaobject definitions
shopify-admin metaobjects list --type my_type      # List metaobjects of a type
shopify-admin metaobjects get <id>
shopify-admin metaobjects create --type my_type --field title="Hello" --field body="World"
shopify-admin metaobjects update <id> --field title="Updated"
shopify-admin metaobjects delete <id>
```

---

### `webhooks`
```bash
shopify-admin webhooks list
shopify-admin webhooks create --topic ORDERS_CREATE --url https://example.com/webhooks
shopify-admin webhooks delete <id>
```

Common topics: `ORDERS_CREATE`, `ORDERS_UPDATED`, `PRODUCTS_CREATE`, `PRODUCTS_UPDATE`,
`CUSTOMERS_CREATE`, `INVENTORY_LEVELS_UPDATE`, `CHECKOUTS_CREATE`, `REFUNDS_CREATE`

---

### `discounts`
```bash
shopify-admin discounts list
shopify-admin discounts list --query "status:active"
shopify-admin discounts deactivate <id>
```

---

### `analytics`
```bash
shopify-admin analytics query "FROM sales SHOW SUM(net_sales) SINCE -30d UNTIL today"
shopify-admin analytics query "FROM sales SHOW SUM(net_sales) GROUP BY month SINCE -6m UNTIL today"
shopify-admin analytics query "FROM sessions SHOW sessions GROUP BY device_type SINCE -7d" --json
```

Requires `read_reports` scope.

---

### `fulfillments`
```bash
shopify-admin fulfillments list <order-id>              # List fulfillment orders for an order
shopify-admin fulfillments create <fulfillment-order-id>
shopify-admin fulfillments create <id> --tracking UPS --number 1Z999AA1234567890
```

---

### `markets`
```bash
shopify-admin markets list
shopify-admin markets get <id>
```

---

### `update` — Self-update

Pull the latest source from GitHub, rebuild, and replace the current binary.

```bash
shopify-admin update
```

Requires `git` and `go` to be installed.

---

## Tips

- **Finding IDs**: Use `--json` and pipe to `jq`:
  ```bash
  shopify-admin products list --json | jq '.[].id'
  shopify-admin orders list --json | jq '.[] | {id, name, total: .totalPriceSet.shopMoney.amount}'
  ```
- **Pagination**: Use `--after CURSOR` with the cursor shown at the bottom of list output
- **Search**: All list commands support `--query` with Shopify's search syntax
- **401 errors**: Run `shopify-admin auth status` — if it shows `env vars`, an old token is overriding the config. Fix with `unset SHOPIFY_ACCESS_TOKEN SHOPIFY_SHOP`
- **Env vars**: Set `SHOPIFY_SHOP` and `SHOPIFY_ACCESS_TOKEN` to bypass stored config (note: this disables auto-refresh)
- **API version**: Uses Shopify Admin API `2026-01`
