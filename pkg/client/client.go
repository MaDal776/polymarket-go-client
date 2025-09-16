package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"polymarket-clob-go/pkg/auth"
	"polymarket-clob-go/pkg/orderbuilder"
	"polymarket-clob-go/pkg/signer"
	"polymarket-clob-go/pkg/types"
	"polymarket-clob-go/pkg/utils"
)

// API endpoints
const (
	CreateAPIKey    = "/auth/api-key"
	DeriveAPIKey    = "/auth/derive-api-key"
	GetAPIKeys      = "/auth/api-keys"
	DeleteAPIKey    = "/auth/api-key"
	PostOrder       = "/order"
	PostOrders      = "/orders"
	GetOrder        = "/order/"
	GetOrders       = "/orders"
	CancelOrder     = "/order"
	CancelOrders    = "/orders"
	CancelAll       = "/orders/cancel-all"
	GetOrderBook    = "/book"
	GetTrades       = "/trades"
	GetTickSize     = "/tick-size"
	GetNegRisk      = "/neg-risk"
	GetMidpoint     = "/midpoint"
	GetPrice        = "/price"
	GetPrices       = "/prices"
	GetSpread       = "/spread"
	Time            = "/time"
	GetBalanceAllowance     = "/balance-allowance"
	UpdateBalanceAllowance  = "/balance-allowance/update"
)

// Contract addresses for different chains
var contractConfigs = map[int64]types.ContractConfig{
	80002: { // Amoy testnet
		Exchange:          "0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40",
		Collateral:        "0x9c4e1703476e875070ee25b56a58b008cfb8fa78",
		ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
	},
	137: { // Polygon mainnet
		Exchange:          "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
		Collateral:        "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
		ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
	},
}

// Neg risk contract addresses
var negRiskContractConfigs = map[int64]types.ContractConfig{
	80002: { // Amoy testnet
		Exchange:          "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
		Collateral:        "0x9c4e1703476e875070ee25b56a58b008cfb8fa78",
		ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
	},
	137: { // Polygon mainnet
		Exchange:          "0xC5d563A36AE78145C45a50134d48A1215220f80a",
		Collateral:        "0x2791bca1f2de4661ed88a30c99a7a9449aa84174",
		ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
	},
}

// ClobClient represents the main CLOB client
type ClobClient struct {
	host          string
	chainID       int64
	signer        *signer.Signer
	creds         *types.ApiCreds
	authLevel     types.AuthLevel
	headerBuilder *auth.HeaderBuilder
	orderBuilder  *orderbuilder.OrderBuilder
	httpClient    *http.Client
	metrics       []types.PerformanceMetrics
	
	// Cache
	tickSizes map[string]types.TickSize
	negRisks  map[string]bool
}

// NewClobClient creates a new CLOB client
func NewClobClient(host string, chainID int64, privateKey string, creds *types.ApiCreds, signatureType *int, funder *string) (*ClobClient, error) {
	start := time.Now()
	
	// Clean host URL
	if strings.HasSuffix(host, "/") {
		host = host[:len(host)-1]
	}
	
	client := &ClobClient{
		host:       host,
		chainID:    chainID,
		creds:      creds,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		metrics:    make([]types.PerformanceMetrics, 0),
		tickSizes:  make(map[string]types.TickSize),
		negRisks:   make(map[string]bool),
	}
	
	// Initialize signer if private key provided
	if privateKey != "" {
		s, err := signer.NewSigner(privateKey, chainID)
		if err != nil {
			client.recordMetric("client_creation", start, false, err.Error())
			return nil, fmt.Errorf("failed to create signer: %w", err)
		}
		client.signer = s
		client.headerBuilder = auth.NewHeaderBuilder(s)
		client.orderBuilder = orderbuilder.NewOrderBuilder(s, signatureType, funder)
	}
	
	// Determine auth level
	client.authLevel = client.getAuthLevel()
	
	client.recordMetric("client_creation", start, true, "")
	return client, nil
}

// GetAddress returns the signer's address
func (c *ClobClient) GetAddress() string {
	if c.signer == nil {
		return ""
	}
	return c.signer.AddressHex()
}

// GetAuthLevel returns the current authentication level
func (c *ClobClient) GetAuthLevel() types.AuthLevel {
	return c.authLevel
}

// CreateAPIKey creates a new API key
func (c *ClobClient) CreateAPIKey(nonce int64) (*types.ApiCreds, error) {
	start := time.Now()
	
	if c.authLevel < types.L1 {
		c.recordMetric("api_key_creation", start, false, "insufficient auth level")
		return nil, fmt.Errorf("Level 1 authentication required")
	}
	
	// Create headers
	headers, err := c.headerBuilder.CreateLevel1Headers(nonce)
	if err != nil {
		c.recordMetric("api_key_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to create headers: %w", err)
	}
	
	// Make request
	url := c.host + CreateAPIKey
	resp, err := c.makeRequest("POST", url, headers, nil)
	if err != nil {
		c.recordMetric("api_key_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("api_key_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	creds := &types.ApiCreds{
		ApiKey:        result["apiKey"].(string),
		ApiSecret:     result["secret"].(string),
		ApiPassphrase: result["passphrase"].(string),
	}
	
	c.recordMetric("api_key_creation", start, true, "")
	return creds, nil
}

// DeriveAPIKey derives an existing API key
func (c *ClobClient) DeriveAPIKey(nonce int64) (*types.ApiCreds, error) {
	start := time.Now()
	
	if c.authLevel < types.L1 {
		c.recordMetric("api_key_derivation", start, false, "insufficient auth level")
		return nil, fmt.Errorf("Level 1 authentication required")
	}
	
	// Create headers
	headers, err := c.headerBuilder.CreateLevel1Headers(nonce)
	if err != nil {
		c.recordMetric("api_key_derivation", start, false, err.Error())
		return nil, fmt.Errorf("failed to create headers: %w", err)
	}
	
	// Make request
	url := c.host + DeriveAPIKey
	resp, err := c.makeRequest("GET", url, headers, nil)
	if err != nil {
		c.recordMetric("api_key_derivation", start, false, err.Error())
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("api_key_derivation", start, false, err.Error())
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	creds := &types.ApiCreds{
		ApiKey:        result["apiKey"].(string),
		ApiSecret:     result["secret"].(string),
		ApiPassphrase: result["passphrase"].(string),
	}
	
	c.recordMetric("api_key_derivation", start, true, "")
	return creds, nil
}

// CreateOrDeriveAPIKey creates or derives API key
func (c *ClobClient) CreateOrDeriveAPIKey(nonce int64) (*types.ApiCreds, error) {
	// Try to create first
	creds, err := c.CreateAPIKey(nonce)
	if err != nil {
		// If creation fails, try to derive
		return c.DeriveAPIKey(nonce)
	}
	return creds, nil
}

// SetAPICredentials sets API credentials and updates auth level
func (c *ClobClient) SetAPICredentials(creds *types.ApiCreds) {
	c.creds = creds
	c.authLevel = c.getAuthLevel()
}

// GetTickSize gets the tick size for a token
func (c *ClobClient) GetTickSize(tokenID string) (types.TickSize, error) {
	start := time.Now()
	
	// Check cache first
	if tickSize, exists := c.tickSizes[tokenID]; exists {
		c.recordMetric("tick_size_retrieval", start, true, "from_cache")
		return tickSize, nil
	}
	
	// Make request
	url := fmt.Sprintf("%s%s?token_id=%s", c.host, GetTickSize, tokenID)
	resp, err := c.makeRequest("GET", url, nil, nil)
	if err != nil {
		c.recordMetric("tick_size_retrieval", start, false, err.Error())
		return "", fmt.Errorf("failed to get tick size: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("tick_size_retrieval", start, false, err.Error())
		return "", fmt.Errorf("failed to parse tick size response: %w", err)
	}
	
	// Handle both string and float64 responses
	var tickSizeStr string
	switch v := result["minimum_tick_size"].(type) {
	case string:
		tickSizeStr = v
	case float64:
		tickSizeStr = fmt.Sprintf("%.4f", v)
		// Remove trailing zeros
		tickSizeStr = strings.TrimRight(tickSizeStr, "0")
		tickSizeStr = strings.TrimRight(tickSizeStr, ".")
	default:
		c.recordMetric("tick_size_retrieval", start, false, "invalid tick size type")
		return "", fmt.Errorf("invalid tick size type: %T", v)
	}
	
	tickSize := types.TickSize(tickSizeStr)
	
	// Cache the result
	c.tickSizes[tokenID] = tickSize
	
	c.recordMetric("tick_size_retrieval", start, true, "")
	return tickSize, nil
}

// GetNegRisk gets the neg risk flag for a token
func (c *ClobClient) GetNegRisk(tokenID string) (bool, error) {
	start := time.Now()
	
	// Check cache first
	if negRisk, exists := c.negRisks[tokenID]; exists {
		c.recordMetric("neg_risk_retrieval", start, true, "from_cache")
		return negRisk, nil
	}
	
	// Make request
	url := fmt.Sprintf("%s%s?token_id=%s", c.host, GetNegRisk, tokenID)
	resp, err := c.makeRequest("GET", url, nil, nil)
	if err != nil {
		c.recordMetric("neg_risk_retrieval", start, false, err.Error())
		return false, fmt.Errorf("failed to get neg risk: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("neg_risk_retrieval", start, false, err.Error())
		return false, fmt.Errorf("failed to parse neg risk response: %w", err)
	}
	
	negRisk := result["neg_risk"].(bool)
	
	// Cache the result
	c.negRisks[tokenID] = negRisk
	
	c.recordMetric("neg_risk_retrieval", start, true, "")
	return negRisk, nil
}

// GetPrice gets the market price for a specific token and side
func (c *ClobClient) GetPrice(tokenID string, side types.OrderSide) (*types.PriceResponse, error) {
	start := time.Now()
	
	// Make request
	url := fmt.Sprintf("%s%s?token_id=%s&side=%s", c.host, GetPrice, tokenID, side)
	resp, err := c.makeRequest("GET", url, nil, nil)
	if err != nil {
		c.recordMetric("price_retrieval", start, false, err.Error())
		return nil, fmt.Errorf("failed to get price: %w", err)
	}
	
	// Parse response
	var result types.PriceResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("price_retrieval", start, false, err.Error())
		return nil, fmt.Errorf("failed to parse price response: %w", err)
	}
	
	c.recordMetric("price_retrieval", start, true, "")
	return &result, nil
}

// GetPrices gets market prices for multiple tokens and sides
func (c *ClobClient) GetPrices(params []types.BookParams) ([]types.PriceResponse, error) {
	start := time.Now()
	
	// Convert params to request format
	requestBody := make([]types.PricesRequest, len(params))
	for i, param := range params {
		requestBody[i] = types.PricesRequest{
			TokenID: param.TokenID,
			Side:    param.Side,
		}
	}
	
	// Make request
	url := fmt.Sprintf("%s%s", c.host, GetPrices)
	resp, err := c.makeRequest("POST", url, nil, requestBody)
	if err != nil {
		c.recordMetric("prices_retrieval", start, false, err.Error())
		return nil, fmt.Errorf("failed to get prices: %w", err)
	}
	
	// Parse response - try both array and object formats
	var result []types.PriceResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		// If array parsing fails, try parsing as object
		var objResult map[string]interface{}
		if err2 := json.Unmarshal(resp, &objResult); err2 != nil {
			c.recordMetric("prices_retrieval", start, false, err.Error())
			return nil, fmt.Errorf("failed to parse prices response as array or object: %w", err)
		}
		
		// Convert object to array format
		result = make([]types.PriceResponse, 0, len(objResult))
		for _, value := range objResult {
			if priceObj, ok := value.(map[string]interface{}); ok {
				if priceStr, exists := priceObj["price"]; exists {
					result = append(result, types.PriceResponse{
						Price: fmt.Sprintf("%v", priceStr),
					})
				}
			}
		}
	}
	
	c.recordMetric("prices_retrieval", start, true, "")
	return result, nil
}

// GetBalanceAllowance gets balance and allowance information
func (c *ClobClient) GetBalanceAllowance(params *types.BalanceAllowanceParams) (*types.BalanceAllowanceResponse, error) {
	start := time.Now()
	
	if c.authLevel < types.L2 {
		c.recordMetric("balance_retrieval", start, false, "insufficient auth level")
		return nil, fmt.Errorf("Level 2 authentication required")
	}
	
	// Create headers for authenticated request
	requestArgs := types.RequestArgs{
		Method:      "GET",
		RequestPath: GetBalanceAllowance,
		Body:        nil,
	}
	
	headers, err := c.headerBuilder.CreateLevel2Headers(c.creds, requestArgs)
	if err != nil {
		c.recordMetric("balance_retrieval", start, false, err.Error())
		return nil, fmt.Errorf("failed to create headers: %w", err)
	}
	
	// Build URL with query parameters
	url := c.host + GetBalanceAllowance
	if params != nil {
		queryParams := make([]string, 0)
		
		if params.AssetType != "" {
			queryParams = append(queryParams, fmt.Sprintf("asset_type=%s", params.AssetType))
		}
		if params.TokenID != "" {
			queryParams = append(queryParams, fmt.Sprintf("token_id=%s", params.TokenID))
		}
		if params.SignatureType != 0 {
			queryParams = append(queryParams, fmt.Sprintf("signature_type=%d", params.SignatureType))
		}
		
		if len(queryParams) > 0 {
			url += "?" + strings.Join(queryParams, "&")
		}
	}
	
	// Make request
	resp, err := c.makeRequest("GET", url, headers, nil)
	if err != nil {
		c.recordMetric("balance_retrieval", start, false, err.Error())
		return nil, fmt.Errorf("failed to get balance allowance: %w", err)
	}
	
	// Parse response
	var result types.BalanceAllowanceResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("balance_retrieval", start, false, err.Error())
		return nil, fmt.Errorf("failed to parse balance allowance response: %w", err)
	}
	
	c.recordMetric("balance_retrieval", start, true, "")
	return &result, nil
}

// UpdateBalanceAllowance updates balance and allowance information
func (c *ClobClient) UpdateBalanceAllowance(params *types.BalanceAllowanceParams) (*types.BalanceAllowanceResponse, error) {
	start := time.Now()
	
	if c.authLevel < types.L2 {
		c.recordMetric("balance_update", start, false, "insufficient auth level")
		return nil, fmt.Errorf("Level 2 authentication required")
	}
	
	// Create headers for authenticated request
	requestArgs := types.RequestArgs{
		Method:      "GET",
		RequestPath: UpdateBalanceAllowance,
		Body:        nil,
	}
	
	headers, err := c.headerBuilder.CreateLevel2Headers(c.creds, requestArgs)
	if err != nil {
		c.recordMetric("balance_update", start, false, err.Error())
		return nil, fmt.Errorf("failed to create headers: %w", err)
	}
	
	// Build URL with query parameters
	url := c.host + UpdateBalanceAllowance
	if params != nil {
		queryParams := make([]string, 0)
		
		if params.AssetType != "" {
			queryParams = append(queryParams, fmt.Sprintf("asset_type=%s", params.AssetType))
		}
		if params.TokenID != "" {
			queryParams = append(queryParams, fmt.Sprintf("token_id=%s", params.TokenID))
		}
		if params.SignatureType != 0 {
			queryParams = append(queryParams, fmt.Sprintf("signature_type=%d", params.SignatureType))
		}
		
		if len(queryParams) > 0 {
			url += "?" + strings.Join(queryParams, "&")
		}
	}
	
	// Make request
	resp, err := c.makeRequest("GET", url, headers, nil)
	if err != nil {
		c.recordMetric("balance_update", start, false, err.Error())
		return nil, fmt.Errorf("failed to update balance allowance: %w", err)
	}
	
	// Check if response is empty - this might be normal for update operations
	if len(resp) == 0 {
		// For update operations, empty response might indicate success
		// Return a response indicating the update was triggered
		c.recordMetric("balance_update", start, true, "empty_response_success")
		return &types.BalanceAllowanceResponse{
			Balance:   "updated",
			Allowance: "updated",
		}, nil
	}
	
	// Parse response
	var result types.BalanceAllowanceResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("balance_update", start, false, fmt.Sprintf("json parse error: %v", err))
		return nil, fmt.Errorf("failed to parse balance allowance response: %w", err)
	}
	
	c.recordMetric("balance_update", start, true, "")
	return &result, nil
}

// CreateOrder creates and signs a limit order
func (c *ClobClient) CreateOrder(orderArgs types.OrderArgs, options *types.CreateOrderOptions) (*types.SignedOrder, error) {
	start := time.Now()
	
	if c.authLevel < types.L1 {
		c.recordMetric("order_creation", start, false, "insufficient auth level")
		return nil, fmt.Errorf("Level 1 authentication required")
	}
	
	// Resolve options
	resolvedOptions, err := c.resolveOrderOptions(orderArgs.TokenID, options)
	if err != nil {
		c.recordMetric("order_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to resolve order options: %w", err)
	}
	
	// Validate price
	if !utils.ValidatePrice(orderArgs.Price, resolvedOptions.TickSize) {
		c.recordMetric("order_creation", start, false, "invalid price")
		return nil, fmt.Errorf("invalid price %.6f for tick size %s", orderArgs.Price, resolvedOptions.TickSize)
	}
	
	// Get contract config
	var contractConfig types.ContractConfig
	var exists bool
	
	if resolvedOptions.NegRisk {
		contractConfig, exists = negRiskContractConfigs[c.chainID]
	} else {
		contractConfig, exists = contractConfigs[c.chainID]
	}
	
	if !exists {
		c.recordMetric("order_creation", start, false, "unsupported chain")
		return nil, fmt.Errorf("unsupported chain ID: %d", c.chainID)
	}
	
	// Create order
	signedOrder, err := c.orderBuilder.CreateOrder(orderArgs, *resolvedOptions, contractConfig.Exchange)
	if err != nil {
		c.recordMetric("order_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	
	c.recordMetric("order_creation", start, true, "")
	return signedOrder, nil
}

// CreateMarketOrder creates and signs a market order
// func (c *ClobClient) CreateMarketOrder(orderArgs types.MarketOrderArgs, options *types.CreateOrderOptions) (*types.SignedOrder, error) {
// 	start := time.Now()
	
// 	if c.authLevel < types.L1 {
// 		c.recordMetric("market_order_creation", start, false, "insufficient auth level")
// 		return nil, fmt.Errorf("Level 1 authentication required")
// 	}
	
// 	// Resolve options
// 	resolvedOptions, err := c.resolveOrderOptions(orderArgs.TokenID, options)
// 	if err != nil {
// 		c.recordMetric("market_order_creation", start, false, err.Error())
// 		return nil, fmt.Errorf("failed to resolve order options: %w", err)
// 	}
	
// 	// Calculate market price if not provided
// 	if orderArgs.Price <= 0 {
// 		price, err := c.calculateMarketPrice(orderArgs.TokenID, orderArgs.Side, orderArgs.Amount, orderArgs.OrderType)
// 		if err != nil {
// 			c.recordMetric("market_order_creation", start, false, err.Error())
// 			return nil, fmt.Errorf("failed to calculate market price: %w", err)
// 		}
// 		orderArgs.Price = price
// 	}
	
// 	// Validate price
// 	if !utils.ValidatePrice(orderArgs.Price, resolvedOptions.TickSize) {
// 		c.recordMetric("market_order_creation", start, false, "invalid price")
// 		return nil, fmt.Errorf("invalid price %.6f for tick size %s", orderArgs.Price, resolvedOptions.TickSize)
// 	}
	
// 	// Get contract config
// 	var contractConfig types.ContractConfig
// 	var exists bool
	
// 	if resolvedOptions.NegRisk {
// 		contractConfig, exists = negRiskContractConfigs[c.chainID]
// 	} else {
// 		contractConfig, exists = contractConfigs[c.chainID]
// 	}
	
// 	if !exists {
// 		c.recordMetric("market_order_creation", start, false, "unsupported chain")
// 		return nil, fmt.Errorf("unsupported chain ID: %d", c.chainID)
// 	}
	
// 	// Create market order
// 	signedOrder, err := c.orderBuilder.CreateMarketOrder(orderArgs, *resolvedOptions, contractConfig.Exchange)
// 	if err != nil {
// 		c.recordMetric("market_order_creation", start, false, err.Error())
// 		return nil, fmt.Errorf("failed to create market order: %w", err)
// 	}
	
// 	c.recordMetric("market_order_creation", start, true, "")
// 	return signedOrder, nil
// }

// PostOrder posts a signed order
func (c *ClobClient) PostOrder(signedOrder *types.SignedOrder, orderType types.OrderType) (map[string]interface{}, error) {
	start := time.Now()
	
	if c.authLevel < types.L2 {
		c.recordMetric("order_posting", start, false, "insufficient auth level")
		return nil, fmt.Errorf("Level 2 authentication required")
	}
	
	// Create request body
	orderRequest := types.OrderRequest{
		Order:     *signedOrder,
		Owner:     c.creds.ApiKey,
		OrderType: orderType,
	}
	
	// Create headers
	requestArgs := types.RequestArgs{
		Method:      "POST",
		RequestPath: PostOrder,
		Body:        orderRequest,
	}
	
	headers, err := c.headerBuilder.CreateLevel2Headers(c.creds, requestArgs)
	if err != nil {
		c.recordMetric("order_posting", start, false, err.Error())
		return nil, fmt.Errorf("failed to create headers: %w", err)
	}
	
	// Make request
	url := c.host + PostOrder
	resp, err := c.makeRequest("POST", url, headers, orderRequest)
	if err != nil {
		c.recordMetric("order_posting", start, false, err.Error())
		return nil, fmt.Errorf("failed to post order: %w", err)
	}
	
	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		c.recordMetric("order_posting", start, false, err.Error())
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	c.recordMetric("order_posting", start, true, "")
	return result, nil
}

// CreateAndPostOrder creates and posts an order in one call
func (c *ClobClient) CreateAndPostOrder(orderArgs types.OrderArgs, options *types.CreateOrderOptions) (map[string]interface{}, error) {
	start := time.Now()
	
	// Create order
	signedOrder, err := c.CreateOrder(orderArgs, options)
	if err != nil {
		c.recordMetric("create_and_post_order", start, false, err.Error())
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	
	// Post order
	result, err := c.PostOrder(signedOrder, types.GTC)
	if err != nil {
		c.recordMetric("create_and_post_order", start, false, err.Error())
		return nil, fmt.Errorf("failed to post order: %w", err)
	}
	
	c.recordMetric("create_and_post_order", start, true, "")
	return result, nil
}

// Helper methods

func (c *ClobClient) getAuthLevel() types.AuthLevel {
	if c.signer != nil && c.creds != nil {
		return types.L2
	}
	if c.signer != nil {
		return types.L1
	}
	return types.L0
}

func (c *ClobClient) resolveOrderOptions(tokenID string, options *types.CreateOrderOptions) (*types.CreateOrderOptions, error) {
	if options == nil {
		options = &types.CreateOrderOptions{}
	}
	
	// Get tick size if not provided
	if options.TickSize == "" {
		tickSize, err := c.GetTickSize(tokenID)
		if err != nil {
			return nil, err
		}
		options.TickSize = tickSize
	}
	
	// Get neg risk if not set
	negRisk, err := c.GetNegRisk(tokenID)
	if err != nil {
		return nil, err
	}
	options.NegRisk = negRisk
	
	return options, nil
}

// func (c *ClobClient) calculateMarketPrice(tokenID string, side types.OrderSide, amount float64, orderType types.OrderType) (float64, error) {
// 	// This is a simplified implementation
// 	// In production, you'd fetch the order book and calculate the matching price
	
// 	// For now, return a default price
// 	if side == types.BUY {
// 		return 0.5, nil // Default buy price
// 	}
// 	return 0.5, nil // Default sell price
// }

func (c *ClobClient) makeRequest(method, url string, headers map[string]string, body interface{}) ([]byte, error) {
	start := time.Now()
	
	var reqBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			c.recordMetric("http_request", start, false, err.Error())
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewReader(bodyBytes)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		c.recordMetric("http_request", start, false, err.Error())
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.recordMetric("http_request", start, false, err.Error())
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.recordMetric("http_request", start, false, err.Error())
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Check status code
	if resp.StatusCode >= 400 {
		c.recordMetric("http_request", start, false, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	c.recordMetric("http_request", start, true, "")
	return respBody, nil
}

// GetMetrics returns all performance metrics
func (c *ClobClient) GetMetrics() []types.PerformanceMetrics {
	allMetrics := make([]types.PerformanceMetrics, 0)
	
	// Add client metrics
	allMetrics = append(allMetrics, c.metrics...)
	
	// Add signer metrics
	if c.signer != nil {
		allMetrics = append(allMetrics, c.signer.GetMetrics()...)
	}
	
	// Add header builder metrics
	if c.headerBuilder != nil {
		allMetrics = append(allMetrics, c.headerBuilder.GetMetrics()...)
	}
	
	// Add order builder metrics
	if c.orderBuilder != nil {
		allMetrics = append(allMetrics, c.orderBuilder.GetMetrics()...)
	}
	
	return allMetrics
}

// ClearMetrics clears all performance metrics
func (c *ClobClient) ClearMetrics() {
	c.metrics = make([]types.PerformanceMetrics, 0)
	
	if c.signer != nil {
		c.signer.ClearMetrics()
	}
	
	if c.headerBuilder != nil {
		c.headerBuilder.ClearMetrics()
	}
	
	if c.orderBuilder != nil {
		c.orderBuilder.ClearMetrics()
	}
}

// PrintMetrics prints performance metrics in a readable format
func (c *ClobClient) PrintMetrics() {
	metrics := c.GetMetrics()
	
	fmt.Println("\n=== Performance Metrics ===")
	for _, metric := range metrics {
		status := "✓"
		if !metric.Success {
			status = "✗"
		}
		
		fmt.Printf("%s %s: %v", status, metric.Operation, metric.Duration)
		if metric.Error != "" {
			fmt.Printf(" (Error: %s)", metric.Error)
		}
		fmt.Println()
	}
	fmt.Println("===========================\n")
}

// recordMetric records a performance metric
func (c *ClobClient) recordMetric(operation string, startTime time.Time, success bool, errorMsg string) {
	metric := types.PerformanceMetrics{
		Operation: operation,
		StartTime: startTime,
		Duration:  time.Since(startTime),
		Success:   success,
		Error:     errorMsg,
	}
	c.metrics = append(c.metrics, metric)
}