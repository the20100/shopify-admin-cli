package api

import "encoding/json"

// ---- GraphQL envelope ----

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message   string `json:"message"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
	Path []any `json:"path,omitempty"`
}

// ShopifyError is returned when the API responds with an error.
type ShopifyError struct {
	StatusCode int
	Message    string
}

func (e *ShopifyError) Error() string {
	return e.Message
}

// ---- Common ----

type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

type MoneyV2 struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
}

type MoneyBag struct {
	ShopMoney MoneyV2 `json:"shopMoney"`
}

type UserError struct {
	Field   []string `json:"field"`
	Message string   `json:"message"`
}

type MailingAddress struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address1  string `json:"address1"`
	Address2  string `json:"address2"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Zip       string `json:"zip"`
	Country   string `json:"country"`
	Phone     string `json:"phone"`
}

// ---- Shop ----

type Shop struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Email                string   `json:"email"`
	MyshopifyDomain      string   `json:"myshopifyDomain"`
	PrimaryDomain        Domain   `json:"primaryDomain"`
	CurrencyCode         string   `json:"currencyCode"`
	CountryCode          string   `json:"countryCode"`
	TimezoneAbbreviation string   `json:"timezoneAbbreviation"`
	CreatedAt            string   `json:"createdAt"`
	Plan                 ShopPlan `json:"plan"`
}

type Domain struct {
	URL  string `json:"url"`
	Host string `json:"host"`
}

type ShopPlan struct {
	DisplayName string `json:"displayName"`
	ShopifyPlus bool   `json:"shopifyPlus"`
}

// ---- Products ----

type Product struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	Status         string            `json:"status"`
	Handle         string            `json:"handle"`
	Description    string            `json:"description"`
	TotalInventory int               `json:"totalInventory"`
	Vendor         string            `json:"vendor"`
	ProductType    string            `json:"productType"`
	Tags           []string          `json:"tags"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
	Variants       VariantConnection `json:"variants"`
}

type ProductEdge struct {
	Node   Product `json:"node"`
	Cursor string  `json:"cursor"`
}

type ProductConnection struct {
	Edges    []ProductEdge `json:"edges"`
	PageInfo PageInfo      `json:"pageInfo"`
}

type ProductVariant struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Price             string  `json:"price"`
	CompareAtPrice    string  `json:"compareAtPrice"`
	SKU               string  `json:"sku"`
	InventoryQuantity int     `json:"inventoryQuantity"`
	Barcode           string  `json:"barcode"`
	Weight            float64 `json:"weight"`
	WeightUnit        string  `json:"weightUnit"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type VariantEdge struct {
	Node   ProductVariant `json:"node"`
	Cursor string         `json:"cursor"`
}

type VariantConnection struct {
	Edges    []VariantEdge `json:"edges"`
	PageInfo PageInfo      `json:"pageInfo"`
}

// ---- Collections ----

type Collection struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	Handle        string        `json:"handle"`
	Description   string        `json:"description"`
	UpdatedAt     string        `json:"updatedAt"`
	ProductsCount ProductsCount `json:"productsCount"`
}

type ProductsCount struct {
	Count int `json:"count"`
}

type CollectionEdge struct {
	Node   Collection `json:"node"`
	Cursor string     `json:"cursor"`
}

type CollectionConnection struct {
	Edges    []CollectionEdge `json:"edges"`
	PageInfo PageInfo         `json:"pageInfo"`
}

// ---- Orders ----

type Order struct {
	ID                       string             `json:"id"`
	Name                     string             `json:"name"`
	Email                    string             `json:"email"`
	Phone                    string             `json:"phone"`
	FinancialStatus          string             `json:"financialStatus"`
	DisplayFulfillmentStatus string             `json:"displayFulfillmentStatus"`
	TotalPriceSet            MoneyBag           `json:"totalPriceSet"`
	SubtotalPriceSet         MoneyBag           `json:"subtotalPriceSet"`
	TotalTaxSet              MoneyBag           `json:"totalTaxSet"`
	CreatedAt                string             `json:"createdAt"`
	ProcessedAt              string             `json:"processedAt"`
	Note                     string             `json:"note"`
	Tags                     []string           `json:"tags"`
	Customer                 *OrderCustomer     `json:"customer"`
	ShippingAddress          *MailingAddress    `json:"shippingAddress"`
	LineItems                LineItemConnection `json:"lineItems"`
}

type OrderCustomer struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type LineItem struct {
	ID                   string   `json:"id"`
	Title                string   `json:"title"`
	Quantity             int      `json:"quantity"`
	SKU                  string   `json:"sku"`
	OriginalUnitPriceSet MoneyBag `json:"originalUnitPriceSet"`
}

type LineItemEdge struct {
	Node   LineItem `json:"node"`
	Cursor string   `json:"cursor"`
}

type LineItemConnection struct {
	Edges    []LineItemEdge `json:"edges"`
	PageInfo PageInfo       `json:"pageInfo"`
}

type OrderEdge struct {
	Node   Order  `json:"node"`
	Cursor string `json:"cursor"`
}

type OrderConnection struct {
	Edges    []OrderEdge `json:"edges"`
	PageInfo PageInfo    `json:"pageInfo"`
}

// ---- Customers ----

type Customer struct {
	ID             string          `json:"id"`
	FirstName      string          `json:"firstName"`
	LastName       string          `json:"lastName"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	State          string          `json:"state"`
	Tags           []string        `json:"tags"`
	NumberOfOrders string          `json:"numberOfOrders"`
	AmountSpent    MoneyV2         `json:"amountSpent"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	DefaultAddress *MailingAddress `json:"defaultAddress"`
}

type CustomerEdge struct {
	Node   Customer `json:"node"`
	Cursor string   `json:"cursor"`
}

type CustomerConnection struct {
	Edges    []CustomerEdge `json:"edges"`
	PageInfo PageInfo       `json:"pageInfo"`
}

// ---- Inventory ----

type Location struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Address  LocationAddress `json:"address"`
	IsActive bool            `json:"isActive"`
}

type LocationAddress struct {
	Address1 string `json:"address1"`
	City     string `json:"city"`
	Country  string `json:"country"`
}

type LocationEdge struct {
	Node   Location `json:"node"`
	Cursor string   `json:"cursor"`
}

type LocationConnection struct {
	Edges    []LocationEdge `json:"edges"`
	PageInfo PageInfo       `json:"pageInfo"`
}

type InventoryItem struct {
	ID              string `json:"id"`
	SKU             string `json:"sku"`
	Tracked         bool   `json:"tracked"`
	RequiresShipping bool  `json:"requiresShipping"`
}

type InventoryItemEdge struct {
	Node   InventoryItem `json:"node"`
	Cursor string        `json:"cursor"`
}

type InventoryItemConnection struct {
	Edges    []InventoryItemEdge `json:"edges"`
	PageInfo PageInfo            `json:"pageInfo"`
}

type InventoryLevel struct {
	ID         string             `json:"id"`
	Quantities []InventoryQuantity `json:"quantities"`
	Location   Location           `json:"location"`
	Item       InventoryItem      `json:"item"`
}

type InventoryQuantity struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type InventoryLevelEdge struct {
	Node   InventoryLevel `json:"node"`
	Cursor string         `json:"cursor"`
}

type InventoryLevelConnection struct {
	Edges    []InventoryLevelEdge `json:"edges"`
	PageInfo PageInfo             `json:"pageInfo"`
}

// ---- Metafields ----

type Metafield struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type MetafieldEdge struct {
	Node   Metafield `json:"node"`
	Cursor string    `json:"cursor"`
}

type MetafieldConnection struct {
	Edges    []MetafieldEdge `json:"edges"`
	PageInfo PageInfo        `json:"pageInfo"`
}

// ---- Metaobjects ----

type Metaobject struct {
	ID        string           `json:"id"`
	Handle    string           `json:"handle"`
	Type      string           `json:"type"`
	UpdatedAt string           `json:"updatedAt"`
	Fields    []MetaobjectField `json:"fields"`
}

type MetaobjectField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type MetaobjectEdge struct {
	Node   Metaobject `json:"node"`
	Cursor string     `json:"cursor"`
}

type MetaobjectConnection struct {
	Edges    []MetaobjectEdge `json:"edges"`
	PageInfo PageInfo         `json:"pageInfo"`
}

type MetaobjectDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type MetaobjectDefinitionEdge struct {
	Node   MetaobjectDefinition `json:"node"`
	Cursor string               `json:"cursor"`
}

type MetaobjectDefinitionConnection struct {
	Edges    []MetaobjectDefinitionEdge `json:"edges"`
	PageInfo PageInfo                   `json:"pageInfo"`
}

// ---- Webhooks ----

type WebhookSubscription struct {
	ID          string          `json:"id"`
	Topic       string          `json:"topic"`
	Endpoint    WebhookEndpoint `json:"endpoint"`
	Format      string          `json:"format"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

type WebhookEndpoint struct {
	CallbackURL string `json:"callbackUrl"`
}

type WebhookSubscriptionEdge struct {
	Node   WebhookSubscription `json:"node"`
	Cursor string              `json:"cursor"`
}

type WebhookSubscriptionConnection struct {
	Edges    []WebhookSubscriptionEdge `json:"edges"`
	PageInfo PageInfo                  `json:"pageInfo"`
}

// ---- Discounts ----

type DiscountNode struct {
	ID       string   `json:"id"`
	Discount Discount `json:"discount"`
}

type Discount struct {
	TypeName        string `json:"__typename"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	StartsAt        string `json:"startsAt"`
	EndsAt          string `json:"endsAt"`
	AsyncUsageCount int    `json:"asyncUsageCount"`
}

type DiscountNodeEdge struct {
	Node   DiscountNode `json:"node"`
	Cursor string       `json:"cursor"`
}

type DiscountNodeConnection struct {
	Edges    []DiscountNodeEdge `json:"edges"`
	PageInfo PageInfo           `json:"pageInfo"`
}

// ---- Fulfillment Orders ----

type FulfillmentOrder struct {
	ID               string                             `json:"id"`
	Status           string                             `json:"status"`
	RequestStatus    string                             `json:"requestStatus"`
	FulfillAt        string                             `json:"fulfillAt"`
	Destination      *FulfillmentDestination            `json:"destination"`
	AssignedLocation FulfillmentOrderLocation           `json:"assignedLocation"`
	LineItems        FulfillmentOrderLineItemConnection `json:"lineItems"`
	Order            *FulfillmentOrderRef               `json:"order"`
}

type FulfillmentOrderRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FulfillmentDestination struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Address1  string `json:"address1"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Zip       string `json:"zip"`
}

type FulfillmentOrderLocation struct {
	Name string `json:"name"`
}

type FulfillmentOrderLineItem struct {
	ID                string      `json:"id"`
	RemainingQuantity int         `json:"remainingQuantity"`
	TotalQuantity     int         `json:"totalQuantity"`
	LineItem          LineItemRef `json:"lineItem"`
}

type LineItemRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	SKU   string `json:"sku"`
}

type FulfillmentOrderLineItemEdge struct {
	Node   FulfillmentOrderLineItem `json:"node"`
	Cursor string                   `json:"cursor"`
}

type FulfillmentOrderLineItemConnection struct {
	Edges    []FulfillmentOrderLineItemEdge `json:"edges"`
	PageInfo PageInfo                       `json:"pageInfo"`
}

type FulfillmentOrderEdge struct {
	Node   FulfillmentOrder `json:"node"`
	Cursor string           `json:"cursor"`
}

type FulfillmentOrderConnection struct {
	Edges    []FulfillmentOrderEdge `json:"edges"`
	PageInfo PageInfo               `json:"pageInfo"`
}

type Fulfillment struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
	TrackingCompany string   `json:"trackingCompany"`
	TrackingNumbers []string `json:"trackingNumbers"`
	TrackingURLs    []string `json:"trackingUrls"`
}

// ---- Markets ----

type Market struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Handle  string              `json:"handle"`
	Enabled bool                `json:"enabled"`
	Primary bool                `json:"primary"`
	Regions MarketRegionConnection `json:"regions"`
}

type MarketRegion struct {
	Name string `json:"name"`
}

type MarketRegionEdge struct {
	Node   MarketRegion `json:"node"`
	Cursor string       `json:"cursor"`
}

type MarketRegionConnection struct {
	Edges    []MarketRegionEdge `json:"edges"`
	PageInfo PageInfo           `json:"pageInfo"`
}

type MarketEdge struct {
	Node   Market `json:"node"`
	Cursor string `json:"cursor"`
}

type MarketConnection struct {
	Edges    []MarketEdge `json:"edges"`
	PageInfo PageInfo     `json:"pageInfo"`
}

// ---- Analytics ----

type ShopifyQLResult struct {
	ParseErrors []ShopifyQLError `json:"parseErrors"`
	TableData   *ShopifyQLTable  `json:"tableData"`
}

type ShopifyQLError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ShopifyQLTable struct {
	Columns []ShopifyQLColumn `json:"columnDefinitions"`
	Rows    [][]string        `json:"rows"`
}

type ShopifyQLColumn struct {
	Name        string `json:"name"`
	DataType    string `json:"dataType"`
	DisplayName string `json:"displayName"`
}
