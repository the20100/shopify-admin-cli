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

You need a Shopify store and an Admin API access token.

**To get an access token:**
1. Go to your Shopify Admin → Settings → Apps and sales channels
2. Click "Develop apps" → Create an app
3. Configure Admin API scopes (read_products, write_products, read_orders, etc.)
4. Install the app and copy the Admin API access token

**Set credentials:**
```bash
# Save to config file (recommended)
shopify-admin auth setup mystore.myshopify.com shpat_xxxxx

# Or use environment variables
export SHOPIFY_SHOP=mystore.myshopify.com
export SHOPIFY_ACCESS_TOKEN=shpat_xxxxx
```

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
shopify-admin auth setup <shop> <access-token>  # Save credentials
shopify-admin auth status                        # Show auth status
shopify-admin auth logout                        # Remove saved credentials
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

## Tips

- **Finding IDs**: Use `--json` and pipe to `jq`:
  ```bash
  shopify-admin products list --json | jq '.[].id'
  shopify-admin orders list --json | jq '.[] | {id, name, total: .totalPriceSet.shopMoney.amount}'
  ```
- **Pagination**: Use `--after CURSOR` with the cursor shown at the bottom of list output
- **Search**: All list commands support `--query` with Shopify's search syntax
- **Env vars**: Set `SHOPIFY_SHOP` and `SHOPIFY_ACCESS_TOKEN` to bypass stored config
- **API version**: Uses Shopify Admin API `2026-01`
