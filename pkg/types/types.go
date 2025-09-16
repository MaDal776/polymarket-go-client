package types

import (
	"math/big"
	"time"
)

// OrderSide represents the side of an order
type OrderSide string

const (
	BUY  OrderSide = "BUY"
	SELL OrderSide = "SELL"
)

// OrderType represents the type of an order
type OrderType string

const (
	GTC OrderType = "GTC" // Good Till Cancelled
	FOK OrderType = "FOK" // Fill Or Kill
	GTD OrderType = "GTD" // Good Till Date
	FAK OrderType = "FAK" // Fill And Kill
)

// AuthLevel represents the authentication level
type AuthLevel int

const (
	L0 AuthLevel = iota // No auth
	L1                  // Private key auth
	L2                  // Full API auth
)

// ApiCreds holds API credentials
type ApiCreds struct {
	ApiKey        string `json:"api_key"`
	ApiSecret     string `json:"api_secret"`
	ApiPassphrase string `json:"api_passphrase"`
}

// OrderArgs represents order arguments
type OrderArgs struct {
	TokenID     string    `json:"token_id"`
	Price       float64   `json:"price"`
	Size        float64   `json:"size"`
	Side        OrderSide `json:"side"`
	FeeRateBps  int       `json:"fee_rate_bps"`
	Nonce       int64     `json:"nonce"`
	Expiration  int64     `json:"expiration"`
	Taker       string    `json:"taker"`
}

// MarketOrderArgs represents market order arguments
type MarketOrderArgs struct {
	TokenID     string    `json:"token_id"`
	Amount      float64   `json:"amount"`
	Side        OrderSide `json:"side"`
	Price       float64   `json:"price,omitempty"`
	FeeRateBps  int       `json:"fee_rate_bps"`
	Nonce       int64     `json:"nonce"`
	Taker       string    `json:"taker"`
	OrderType   OrderType `json:"order_type"`
}

// OrderData represents the order data structure for signing
type OrderData struct {
	Maker         string   `json:"maker"`
	Taker         string   `json:"taker"`
	TokenID       string   `json:"tokenId"`
	MakerAmount   *big.Int `json:"makerAmount"`
	TakerAmount   *big.Int `json:"takerAmount"`
	Side          int      `json:"side"` // 0 for BUY, 1 for SELL (internal use)
	FeeRateBps    string   `json:"feeRateBps"`
	Nonce         string   `json:"nonce"`
	Signer        string   `json:"signer"`
	Expiration    string   `json:"expiration"`
	SignatureType int      `json:"signatureType"`
}

// SignedOrder represents a signed order
type SignedOrder struct {
	Salt      int64 `json:"salt"`  // Should be int like Python
	Maker     string `json:"maker"`
	Signer    string `json:"signer"`
	Taker     string `json:"taker"`
	TokenID   string `json:"tokenId"`
	MakerAmount string `json:"makerAmount"`
	TakerAmount string `json:"takerAmount"`
	Expiration  string `json:"expiration"`
	Nonce       string `json:"nonce"`
	FeeRateBps  string `json:"feeRateBps"`
	Side        OrderSide `json:"side"`  // Use OrderSide type for proper JSON serialization
	SignatureType int  `json:"signatureType"`
	Signature   string `json:"signature"`
}

// OrderRequest represents the request body for posting an order
type OrderRequest struct {
	Order     SignedOrder `json:"order"`
	Owner     string      `json:"owner"`
	OrderType OrderType   `json:"orderType"`
}

// TickSize represents valid tick sizes
type TickSize string

const (
	TickSize01   TickSize = "0.1"
	TickSize001  TickSize = "0.01"
	TickSize0001 TickSize = "0.001"
	TickSize00001 TickSize = "0.0001"
)

// RoundConfig represents rounding configuration
type RoundConfig struct {
	Price  int `json:"price"`
	Size   int `json:"size"`
	Amount int `json:"amount"`
}

// CreateOrderOptions represents options for creating orders
type CreateOrderOptions struct {
	TickSize TickSize `json:"tick_size"`
	NegRisk  bool     `json:"neg_risk"`
}

// ContractConfig represents contract configuration
type ContractConfig struct {
	Exchange           string `json:"exchange"`
	Collateral         string `json:"collateral"`
	ConditionalTokens  string `json:"conditional_tokens"`
}

// RequestArgs represents request arguments for signing
type RequestArgs struct {
	Method      string      `json:"method"`
	RequestPath string      `json:"request_path"`
	Body        interface{} `json:"body,omitempty"`
}

// ClobAuth represents the EIP712 structure for CLOB authentication
type ClobAuth struct {
	Address   string `json:"address"`
	Timestamp string `json:"timestamp"`
	Nonce     int64  `json:"nonce"`
	Message   string `json:"message"`
}

// PerformanceMetrics tracks timing for operations
type PerformanceMetrics struct {
	Operation string        `json:"operation"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
}

// OrderBookSummary represents order book data
type OrderBookSummary struct {
	Market    string         `json:"market"`
	AssetID   string         `json:"asset_id"`
	Timestamp string         `json:"timestamp"`
	Bids      []OrderSummary `json:"bids"`
	Asks      []OrderSummary `json:"asks"`
	Hash      string         `json:"hash"`
}

// OrderSummary represents a single order in the book
type OrderSummary struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// AssetType represents the type of asset
type AssetType string

const (
	COLLATERAL AssetType = "COLLATERAL"
	CONDITIONAL AssetType = "CONDITIONAL"
)

// BalanceAllowanceParams represents parameters for balance/allowance queries
type BalanceAllowanceParams struct {
	AssetType     AssetType `json:"asset_type,omitempty"`
	TokenID       string    `json:"token_id,omitempty"`
	SignatureType int       `json:"signature_type,omitempty"`
}

// BalanceAllowanceResponse represents balance and allowance information
type BalanceAllowanceResponse struct {
	Balance   string `json:"balance"`
	Allowance string `json:"allowance"`
}

// PriceResponse represents the price response for a token
type PriceResponse struct {
	Price string `json:"price"`
}

// BookParams represents parameters for book-related queries
type BookParams struct {
	TokenID string    `json:"token_id"`
	Side    OrderSide `json:"side"`
}

// PricesRequest represents a request for multiple prices
type PricesRequest struct {
	TokenID string    `json:"token_id"`
	Side    OrderSide `json:"side"`
}