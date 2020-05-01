package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	// CoinbaseBaseURL is the base url of the coinbase api
	CoinbaseBaseURL = "https://api.coinbase.com/"
)

var (
	// CoinbaseAPIKey is your api key obtained from your coinbase account setting
	CoinbaseAPIKey = os.Getenv("COINBASE_APIKEY")
	// CoinbaseAPISecret is your api secret obtained from your coinbase account settings
	CoinbaseAPISecret = os.Getenv("COINBASE_APISECRET")
	// CBAccount is the exported singleton to use this package in code
	CBAccount *APIKeyAuthentication
)

// APIKeyAuthentication Struct implements the Authentication interface and takes
// care of authenticating RPC requests for clients with the `CoinbaseAPIKey` & `CoinbaseAPISecret` pair
type APIKeyAuthentication struct {
	Key     string
	Secret  string
	BaseURL string
	Client  http.Client
}

func init() {
	if CoinbaseAPIKey == "" {
		log.Fatal("Failed to init Coinbase Account package missing APIKey value")
	} else if CoinbaseAPISecret == "" {
		log.Fatal("Failed to init Coinbase Account package missing APISecret value")
	}
	CBAccount = &APIKeyAuthentication{
		Key:     CoinbaseAPIKey,
		Secret:  CoinbaseAPISecret,
		BaseURL: CoinbaseBaseURL,
		Client:  *http.DefaultClient,
	}
}

// Authenticate with API Key + Secret authentication requires a request header of the HMAC SHA-256
// signature of the "message" as well as an incrementing nonce and the API key
func (a APIKeyAuthentication) Authenticate(req *http.Request, endpoint string, params []byte) error {
	// The CB-ACCESS-SIGN header is generated by creating a sha256 HMAC
	// using the secret key on the prehash string timestamp + method + requestPath + body
	// (where + represents string concatenation).
	// The timestamp value is the same as the CB-ACCESS-TIMESTAMP header.
	// The CB-ACCESS-TIMESTAMP header MUST be number of seconds since Unix Epoch.
	timestamp := fmt.Sprintf("%v", time.Now().UTC().Unix())
	message := fmt.Sprintf("%v%s/%s", timestamp, req.Method, endpoint)
	if params != nil {
		message += string(params)
	}
	h := hmac.New(sha256.New, []byte(a.Secret))
	h.Write([]byte(message))

	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("CB-ACCESS-KEY", a.Key)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("CB-VERSION", "2017-06-02")

	return nil
}

// GetBaseURL return the coinbase api base URL
func (a APIKeyAuthentication) GetBaseURL() string {
	return a.BaseURL
}

// GetClient return the authenticated http client
func (a APIKeyAuthentication) GetClient() *http.Client {
	return &a.Client
}
