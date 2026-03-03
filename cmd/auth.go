package cmd

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/the20100/shopify-admin-cli/internal/config"
)

// defaultScopes covers all operations the CLI supports.
const defaultScopes = "read_products,write_products,read_orders,write_orders,read_customers,write_customers,read_discounts,write_discounts,read_inventory,write_inventory,read_fulfillments,write_fulfillments,read_analytics,read_markets,write_markets,read_metaobjects,write_metaobjects"

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Shopify Admin authentication",
}

// ── auth setup ────────────────────────────────────────────────────────────────

var authSetupCmd = &cobra.Command{
	Use:   "setup <shop> <access-token>",
	Short: "Save a Shopify shop domain and access token directly",
	Long: `Manually save a shop domain and Admin API access token.

<shop> can be the full domain (mystore.myshopify.com) or just the store name.

To get an access token via OAuth instead, use:
  shopify-admin auth configure <client-id> <client-secret>
  shopify-admin auth login --shop <shop>

Credentials are stored at:
  macOS:   ~/Library/Application Support/shopify-admin/config.json
  Linux:   ~/.config/shopify-admin/config.json
  Windows: %AppData%\shopify-admin\config.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		shop := args[0]
		token := args[1]

		if !strings.Contains(shop, ".") {
			shop = shop + ".myshopify.com"
		}
		if len(token) < 16 {
			return fmt.Errorf("access token looks too short")
		}
		c, _ := config.Load()
		if c == nil {
			c = &config.Config{}
		}
		c.Shop = shop
		c.AccessToken = token
		if err := config.Save(c); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Credentials saved to %s\n", config.Path())
		fmt.Printf("Shop:  %s\n", shop)
		fmt.Printf("Token: %s\n", maskOrEmpty(token))
		return nil
	},
}

// ── auth configure ────────────────────────────────────────────────────────────

var authConfigureCmd = &cobra.Command{
	Use:   "configure <client-id> <client-secret>",
	Short: "Save Shopify app Client ID and Client Secret",
	Long: `Save your Shopify app credentials (Client ID and Client Secret).

These are found in:
  - Shopify Partners dashboard → Your app → API credentials
  - Or in the custom app settings in the Shopify admin

After configuring, run:
  shopify-admin auth login --shop <shop>`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID := args[0]
		clientSecret := args[1]

		c, _ := config.Load()
		if c == nil {
			c = &config.Config{}
		}
		c.ClientID = clientID
		c.ClientSecret = clientSecret
		if err := config.Save(c); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("App credentials saved to %s\n", config.Path())
		fmt.Printf("Client ID:     %s\n", clientID)
		fmt.Printf("Client Secret: %s\n", maskOrEmpty(clientSecret))
		return nil
	},
}

// ── auth login ────────────────────────────────────────────────────────────────

var (
	loginShop      string
	loginScopes    string
	loginNoBrowser bool
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate via Shopify OAuth (Authorization Code Grant)",
	Long: `Start the OAuth flow to obtain a permanent offline access token.

Requires app credentials configured via:
  shopify-admin auth configure <client-id> <client-secret>

Default (browser) mode:
  1. A local HTTP server is started to capture the OAuth callback.
  2. Your browser opens to Shopify's authorization page.
  3. You approve the app; Shopify redirects back automatically.
  4. The CLI exchanges the code for an access token and saves it.

  IMPORTANT: register the printed redirect URI in your app settings.

--no-browser mode (remote / headless environments):
  1. The CLI prints the authorization URL — open it in any browser.
  2. After you approve, Shopify redirects to http://localhost/callback?...
     which will fail in your browser (that's expected).
  3. Copy the full URL from your browser's address bar and paste it here.
  4. The CLI extracts the code, exchanges it, and saves the token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if c.ClientID == "" || c.ClientSecret == "" {
			return fmt.Errorf("app credentials not set\n\nRun first: shopify-admin auth configure <client-id> <client-secret>")
		}

		shop := loginShop
		if shop == "" {
			shop = c.Shop
		}
		if shop == "" {
			return fmt.Errorf("shop not specified\n\nUse: shopify-admin auth login --shop <shop>")
		}
		if !strings.Contains(shop, ".") {
			shop = shop + ".myshopify.com"
		}

		scopes := loginScopes

		// Generate a random state nonce to guard against CSRF.
		stateBytes := make([]byte, 16)
		if _, err := rand.Read(stateBytes); err != nil {
			return fmt.Errorf("generating state: %w", err)
		}
		state := hex.EncodeToString(stateBytes)

		if loginNoBrowser {
			return loginManual(c, shop, scopes, state)
		}
		return loginWithServer(c, shop, scopes, state)
	},
}

// loginWithServer is the default browser flow: starts a local HTTP server,
// opens the browser, and waits for Shopify's redirect callback.
func loginWithServer(c *config.Config, shop, scopes, state string) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	authURL := fmt.Sprintf(
		"https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		shop,
		url.QueryEscape(c.ClientID),
		url.QueryEscape(scopes),
		url.QueryEscape(redirectURI),
		state,
	)

	type oauthResult struct {
		tr  *tokenResponse
		err error
	}
	ch := make(chan oauthResult, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if q.Get("state") != state {
			ch <- oauthResult{err: fmt.Errorf("state mismatch: possible CSRF attack")}
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}
		if err := verifyHMAC(q, c.ClientSecret); err != nil {
			ch <- oauthResult{err: fmt.Errorf("HMAC validation failed: %w", err)}
			http.Error(w, "HMAC validation failed", http.StatusBadRequest)
			return
		}
		code := q.Get("code")
		if code == "" {
			ch <- oauthResult{err: fmt.Errorf("no authorization code in callback")}
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}
		tr, err := exchangeAuthCode(q.Get("shop"), c.ClientID, c.ClientSecret, code)
		if err != nil {
			ch <- oauthResult{err: fmt.Errorf("exchanging code: %w", err)}
			fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>%s</p><p>You can close this tab.</p></body></html>", err)
			return
		}
		ch <- oauthResult{tr: tr}
		fmt.Fprintf(w, "<html><body><h2>Authentication successful!</h2><p>You can close this tab and return to the terminal.</p></body></html>")
	})

	go srv.Serve(listener) //nolint:errcheck

	fmt.Printf("Redirect URI for your app settings:\n  %s\n\n", redirectURI)
	fmt.Println("Opening browser for Shopify authorization...")
	fmt.Printf("If the browser does not open, visit:\n  %s\n\n", authURL)
	openBrowser(authURL)
	fmt.Println("Waiting for authorization (5 min timeout)...")

	select {
	case res := <-ch:
		srv.Close()
		if res.err != nil {
			return res.err
		}
		return saveToken(c, shop, res.tr)
	case <-time.After(5 * time.Minute):
		srv.Close()
		return fmt.Errorf("timed out waiting for Shopify to redirect back")
	}
}

// loginManual is the --no-browser flow for remote/headless environments.
// It prints the auth URL, then prompts the user to paste back the full
// callback URL that Shopify redirected to (which fails in the browser but
// whose URL is visible in the address bar).
func loginManual(c *config.Config, shop, scopes, state string) error {
	// Use a fixed placeholder redirect URI — the exact value doesn't matter
	// for parsing, but it must be registered in the app settings.
	redirectURI := "http://localhost/callback"

	authURL := fmt.Sprintf(
		"https://%s/admin/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		shop,
		url.QueryEscape(c.ClientID),
		url.QueryEscape(scopes),
		url.QueryEscape(redirectURI),
		state,
	)

	fmt.Printf("Redirect URI to register in your app settings:\n  %s\n\n", redirectURI)
	fmt.Printf("Open this URL in your browser:\n  %s\n\n", authURL)
	fmt.Println("After approving, your browser will redirect to http://localhost/callback?...")
	fmt.Println("That page will fail to load — that's expected.")
	fmt.Println("Copy the full URL from your browser's address bar and paste it below.\n")
	fmt.Print("Callback URL: ")

	reader := bufio.NewReader(os.Stdin)
	raw, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("no URL provided")
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parsing callback URL: %w", err)
	}
	q := parsed.Query()

	if q.Get("state") != state {
		return fmt.Errorf("state mismatch: the URL does not belong to this login session")
	}
	if err := verifyHMAC(q, c.ClientSecret); err != nil {
		return fmt.Errorf("HMAC validation failed: %w", err)
	}
	code := q.Get("code")
	if code == "" {
		return fmt.Errorf("no authorization code found in the URL")
	}
	callbackShop := q.Get("shop")
	if callbackShop == "" {
		callbackShop = shop
	}

	tr, err := exchangeAuthCode(callbackShop, c.ClientID, c.ClientSecret, code)
	if err != nil {
		return fmt.Errorf("exchanging code: %w", err)
	}
	return saveToken(c, shop, tr)
}

// saveToken writes shop + token (+ optional expiry) to config and prints confirmation.
func saveToken(c *config.Config, shop string, tr *tokenResponse) error {
	c.Shop = shop
	c.AccessToken = tr.AccessToken
	if tr.ExpiresIn > 0 {
		c.TokenExpiresAt = time.Now().Unix() + int64(tr.ExpiresIn)
	} else {
		c.TokenExpiresAt = 0 // no expiry (permanent offline token)
	}
	if err := config.Save(c); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}
	fmt.Printf("\nAuthenticated successfully!\n")
	fmt.Printf("Shop:  %s\n", shop)
	fmt.Printf("Token: %s\n", maskOrEmpty(tr.AccessToken))
	if tr.ExpiresIn > 0 {
		fmt.Printf("Expires: in ~%d hours (auto-refreshed on next use)\n", tr.ExpiresIn/3600)
	}
	return nil
}

// ── auth status ───────────────────────────────────────────────────────────────

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Printf("Config: %s\n\n", config.Path())

		envShop := os.Getenv("SHOPIFY_SHOP")
		envToken := os.Getenv("SHOPIFY_ACCESS_TOKEN")

		if envShop != "" && envToken != "" {
			fmt.Println("Source: env vars (take priority over config)")
			fmt.Printf("Shop:   %s\n", envShop)
			fmt.Printf("Token:  %s\n", maskOrEmpty(envToken))
		} else if c.Shop != "" && c.AccessToken != "" {
			fmt.Println("Source: config file")
			fmt.Printf("Shop:   %s\n", c.Shop)
			fmt.Printf("Token:  %s\n", maskOrEmpty(c.AccessToken))
		} else {
			fmt.Println("Status: not authenticated")
			fmt.Printf("\nOption A – OAuth flow (recommended):\n")
			fmt.Printf("  shopify-admin auth configure <client-id> <client-secret>\n")
			fmt.Printf("  shopify-admin auth login --shop <shop>\n")
			fmt.Printf("\nOption B – manual token:\n")
			fmt.Printf("  shopify-admin auth setup <shop> <access-token>\n")
		}

		if c.ClientID != "" {
			fmt.Printf("\nApp Client ID: %s\n", c.ClientID)
		}

		return nil
	},
}

// ── auth logout ───────────────────────────────────────────────────────────────

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved credentials from the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Println("Credentials removed from config.")
		return nil
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

// verifyHMAC validates the HMAC-SHA256 Shopify includes in OAuth callbacks.
// All query params except "hmac" are sorted and joined as key=value pairs,
// then signed with the app's client secret.
func verifyHMAC(q url.Values, secret string) error {
	provided := q.Get("hmac")
	if provided == "" {
		return fmt.Errorf("missing hmac parameter")
	}

	pairs := make([]string, 0, len(q))
	for k, vs := range q {
		if k == "hmac" {
			continue
		}
		pairs = append(pairs, k+"="+vs[0])
	}
	sort.Strings(pairs)
	message := strings.Join(pairs, "&")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(provided), []byte(expected)) {
		return fmt.Errorf("signature does not match")
	}
	return nil
}

// tokenResponse is the shape of every /admin/oauth/access_token response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"` // seconds; 0 if not present (offline token)
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func parseTokenResponse(resp *http.Response) (*tokenResponse, error) {
	defer resp.Body.Close()
	var r tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	if r.Error != "" {
		if r.ErrorDesc != "" {
			return nil, fmt.Errorf("%s: %s", r.Error, r.ErrorDesc)
		}
		return nil, fmt.Errorf("%s", r.Error)
	}
	if r.AccessToken == "" {
		return nil, fmt.Errorf("empty access token in response (HTTP %d)", resp.StatusCode)
	}
	return &r, nil
}

// exchangeAuthCode exchanges a Shopify OAuth authorization code for an access token.
func exchangeAuthCode(shop, clientID, clientSecret, code string) (*tokenResponse, error) {
	endpoint := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	resp, err := http.PostForm(endpoint, form)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", endpoint, err)
	}
	return parseTokenResponse(resp)
}

// ClientCredentialsGrant obtains an access token directly from client credentials
// (no user interaction required). Tokens last ~24 h (expires_in: 86399).
// Exported so root.go can call it for auto-refresh.
func ClientCredentialsGrant(shop, clientID, clientSecret string) (*tokenResponse, error) {
	if !strings.Contains(shop, ".") {
		shop = shop + ".myshopify.com"
	}
	endpoint := fmt.Sprintf("https://%s/admin/oauth/access_token", shop)
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("grant_type", "client_credentials")
	resp, err := http.PostForm(endpoint, form)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", endpoint, err)
	}
	return parseTokenResponse(resp)
}

// openBrowser opens the given URL in the user's default browser.
func openBrowser(rawURL string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{rawURL}
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler", rawURL}
	default:
		cmd, args = "xdg-open", []string{rawURL}
	}
	exec.Command(cmd, args...).Start() //nolint:errcheck
}

func init() {
	authLoginCmd.Flags().StringVar(&loginShop, "shop", "", "Shopify shop domain (e.g. mystore or mystore.myshopify.com)")
	authLoginCmd.Flags().StringVar(&loginScopes, "scopes", defaultScopes, "OAuth scopes to request")
	authLoginCmd.Flags().BoolVar(&loginNoBrowser, "no-browser", false, "Print the auth URL and prompt for the callback URL instead of opening a browser (use in remote/headless environments)")

	authCmd.AddCommand(authSetupCmd, authConfigureCmd, authLoginCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
