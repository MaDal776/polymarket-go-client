package signer

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"polymarket-clob-go/pkg/types"
	"polymarket-clob-go/pkg/utils"
)

// Signer handles cryptographic operations
type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	chainID    int64
	metrics    []types.PerformanceMetrics
}

// NewSigner creates a new signer instance
func NewSigner(privateKeyHex string, chainID int64) (*Signer, error) {
	start := time.Now()
	
	// Remove 0x prefix if present
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}
	
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	
	signer := &Signer{
		privateKey: privateKey,
		address:    address,
		chainID:    chainID,
		metrics:    make([]types.PerformanceMetrics, 0),
	}
	
	// Record performance metric
	signer.recordMetric("signer_creation", start, true, "")
	
	return signer, nil
}

// Address returns the signer's address
func (s *Signer) Address() common.Address {
	return s.address
}

// AddressHex returns the signer's address as hex string
func (s *Signer) AddressHex() string {
	return s.address.Hex()
}

// ChainID returns the chain ID
func (s *Signer) ChainID() int64 {
	return s.chainID
}

// Sign signs a message hash
func (s *Signer) Sign(messageHash []byte) ([]byte, error) {
	start := time.Now()
	
	signature, err := crypto.Sign(messageHash, s.privateKey)
	if err != nil {
		s.recordMetric("message_signing", start, false, err.Error())
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}
	
	// For Ethereum signatures, we need to adjust the recovery ID
	// go-ethereum returns recovery ID in range [0, 1]
	// But Ethereum standard expects [27, 28]
	if signature[64] < 27 {
		signature[64] += 27
	}
	
	s.recordMetric("message_signing", start, true, "")
	return signature, nil
}

// SignEIP712 signs an EIP712 message
func (s *Signer) SignEIP712(domainSeparator, structHash []byte) ([]byte, error) {
	start := time.Now()
	
	// Create EIP712 hash
	eip712Hash := utils.CreateEIP712Hash(domainSeparator, structHash)
	
	// Sign the hash
	signature, err := s.Sign(eip712Hash)
	if err != nil {
		s.recordMetric("eip712_signing", start, false, err.Error())
		return nil, err
	}
	
	s.recordMetric("eip712_signing", start, true, "")
	return signature, nil
}

// SignClobAuth signs a CLOB authentication message
func (s *Signer) SignClobAuth(timestamp int64, nonce int64) (string, error) {
	start := time.Now()
	
	// Create CLOB auth message
	clobAuth := types.ClobAuth{
		Address:   s.AddressHex(),
		Timestamp: fmt.Sprintf("%d", timestamp),
		Nonce:     nonce,
		Message:   "This message attests that I control the given wallet",
	}
	
	// Create EIP712 domain separator and struct hash
	domainSeparator := utils.CreateClobAuthDomain(s.chainID)
	structHash := utils.EncodeClobAuth(clobAuth)
	
	// Sign the message
	signature, err := s.SignEIP712(domainSeparator, structHash)
	if err != nil {
		s.recordMetric("clob_auth_signing", start, false, err.Error())
		return "", err
	}
	
	signatureHex := fmt.Sprintf("0x%x", signature)
	s.recordMetric("clob_auth_signing", start, true, "")
	
	return signatureHex, nil
}

// GetMetrics returns performance metrics
func (s *Signer) GetMetrics() []types.PerformanceMetrics {
	return s.metrics
}

// ClearMetrics clears performance metrics
func (s *Signer) ClearMetrics() {
	s.metrics = make([]types.PerformanceMetrics, 0)
}

// recordMetric records a performance metric
func (s *Signer) recordMetric(operation string, startTime time.Time, success bool, errorMsg string) {
	metric := types.PerformanceMetrics{
		Operation: operation,
		StartTime: startTime,
		Duration:  time.Since(startTime),
		Success:   success,
		Error:     errorMsg,
	}
	s.metrics = append(s.metrics, metric)
}