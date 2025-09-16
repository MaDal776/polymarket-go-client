package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"polymarket-clob-go/pkg/signer"
	"polymarket-clob-go/pkg/types"
)

const (
	PolyAddress    = "POLY_ADDRESS"
	PolySignature  = "POLY_SIGNATURE"
	PolyTimestamp  = "POLY_TIMESTAMP"
	PolyNonce      = "POLY_NONCE"
	PolyApiKey     = "POLY_API_KEY"
	PolyPassphrase = "POLY_PASSPHRASE"
)

// HeaderBuilder handles authentication header creation
type HeaderBuilder struct {
	signer  *signer.Signer
	metrics []types.PerformanceMetrics
}

// NewHeaderBuilder creates a new header builder
func NewHeaderBuilder(s *signer.Signer) *HeaderBuilder {
	return &HeaderBuilder{
		signer:  s,
		metrics: make([]types.PerformanceMetrics, 0),
	}
}

// CreateLevel1Headers creates Level 1 authentication headers
func (h *HeaderBuilder) CreateLevel1Headers(nonce int64) (map[string]string, error) {
	start := time.Now()
	
	timestamp := time.Now().Unix()
	
	// Sign CLOB auth message
	signature, err := h.signer.SignClobAuth(timestamp, nonce)
	if err != nil {
		h.recordMetric("level1_headers_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to sign CLOB auth: %w", err)
	}
	
	headers := map[string]string{
		PolyAddress:   h.signer.AddressHex(),
		PolySignature: signature,
		PolyTimestamp: fmt.Sprintf("%d", timestamp),
		PolyNonce:     fmt.Sprintf("%d", nonce),
	}
	
	h.recordMetric("level1_headers_creation", start, true, "")
	return headers, nil
}

// CreateLevel2Headers creates Level 2 authentication headers
func (h *HeaderBuilder) CreateLevel2Headers(creds *types.ApiCreds, requestArgs types.RequestArgs) (map[string]string, error) {
	start := time.Now()
	
	timestamp := time.Now().Unix()
	
	// Build HMAC signature
	hmacSig, err := h.buildHMACSignature(creds.ApiSecret, timestamp, requestArgs)
	if err != nil {
		h.recordMetric("level2_headers_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to build HMAC signature: %w", err)
	}
	
	headers := map[string]string{
		PolyAddress:    h.signer.AddressHex(),
		PolySignature:  hmacSig,
		PolyTimestamp:  fmt.Sprintf("%d", timestamp),
		PolyApiKey:     creds.ApiKey,
		PolyPassphrase: creds.ApiPassphrase,
	}
	
	h.recordMetric("level2_headers_creation", start, true, "")
	return headers, nil
}

// buildHMACSignature builds HMAC signature for Level 2 auth
func (h *HeaderBuilder) buildHMACSignature(secret string, timestamp int64, requestArgs types.RequestArgs) (string, error) {
	start := time.Now()
	
	// Decode base64 secret
	decodedSecret, err := base64.URLEncoding.DecodeString(secret)
	if err != nil {
		h.recordMetric("hmac_signature_build", start, false, err.Error())
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}
	
	// Build message to sign
	message := fmt.Sprintf("%d%s%s", timestamp, requestArgs.Method, requestArgs.RequestPath)
	
	// Add body if present
	if requestArgs.Body != nil {
		bodyBytes, err := json.Marshal(requestArgs.Body)
		if err != nil {
			h.recordMetric("hmac_signature_build", start, false, err.Error())
			return "", fmt.Errorf("failed to marshal body: %w", err)
		}
		// Replace single quotes with double quotes to match Python behavior
		bodyStr := strings.ReplaceAll(string(bodyBytes), "'", "\"")
		message += bodyStr
	}
	
	// Create HMAC
	mac := hmac.New(sha256.New, decodedSecret)
	mac.Write([]byte(message))
	signature := mac.Sum(nil)
	
	// Base64 encode
	encodedSignature := base64.URLEncoding.EncodeToString(signature)
	
	h.recordMetric("hmac_signature_build", start, true, "")
	return encodedSignature, nil
}

// GetMetrics returns performance metrics
func (h *HeaderBuilder) GetMetrics() []types.PerformanceMetrics {
	return h.metrics
}

// ClearMetrics clears performance metrics
func (h *HeaderBuilder) ClearMetrics() {
	h.metrics = make([]types.PerformanceMetrics, 0)
}

// recordMetric records a performance metric
func (h *HeaderBuilder) recordMetric(operation string, startTime time.Time, success bool, errorMsg string) {
	metric := types.PerformanceMetrics{
		Operation: operation,
		StartTime: startTime,
		Duration:  time.Since(startTime),
		Success:   success,
		Error:     errorMsg,
	}
	h.metrics = append(h.metrics, metric)
}