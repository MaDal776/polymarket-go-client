package utils

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"polymarket-clob-go/pkg/types"
)

const (
	ClobDomainName = "ClobAuthDomain"
	ClobVersion    = "1"
)

// CreateEIP712Hash creates an EIP712 hash according to the standard
func CreateEIP712Hash(domainSeparator, structHash []byte) []byte {
	// EIP712 standard: keccak256("\x19\x01" ‖ domainSeparator ‖ hashStruct(message))
	prefix := []byte("\x19\x01")
	data := make([]byte, 0, 2+32+32)
	data = append(data, prefix...)
	data = append(data, domainSeparator...)
	data = append(data, structHash...)
	
	return crypto.Keccak256(data)
}

// CreateClobAuthDomain creates the EIP712 domain separator for CLOB auth
func CreateClobAuthDomain(chainID int64) []byte {
	// EIP712Domain(string name,string version,uint256 chainId)
	domainTypeHash := crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId)"))
	
	nameHash := crypto.Keccak256([]byte(ClobDomainName))
	versionHash := crypto.Keccak256([]byte(ClobVersion))
	
	chainIDBytes := make([]byte, 32)
	big.NewInt(chainID).FillBytes(chainIDBytes)
	
	// Encode domain: keccak256(domainTypeHash ‖ nameHash ‖ versionHash ‖ chainId)
	domain := make([]byte, 0, 128)
	domain = append(domain, domainTypeHash...)
	domain = append(domain, nameHash...)
	domain = append(domain, versionHash...)
	domain = append(domain, chainIDBytes...)
	
	return crypto.Keccak256(domain)
}

// EncodeClobAuth encodes CLOB auth message for EIP712
func EncodeClobAuth(auth types.ClobAuth) []byte {
	// ClobAuth(address address,string timestamp,uint256 nonce,string message)
	typeHash := crypto.Keccak256([]byte("ClobAuth(address address,string timestamp,uint256 nonce,string message)"))
	
	// Encode address (20 bytes padded to 32 bytes)
	addressBytes := common.HexToAddress(auth.Address).Bytes()
	addressPadded := make([]byte, 32)
	copy(addressPadded[12:], addressBytes)
	
	// Encode timestamp as string hash
	timestampHash := crypto.Keccak256([]byte(auth.Timestamp))
	
	// Encode nonce as uint256
	nonceBytes := make([]byte, 32)
	big.NewInt(auth.Nonce).FillBytes(nonceBytes)
	
	// Encode message as string hash
	messageHash := crypto.Keccak256([]byte(auth.Message))
	
	// Combine all: keccak256(typeHash ‖ address ‖ keccak256(timestamp) ‖ nonce ‖ keccak256(message))
	encoded := make([]byte, 0, 160)
	encoded = append(encoded, typeHash...)
	encoded = append(encoded, addressPadded...)
	encoded = append(encoded, timestampHash...)
	encoded = append(encoded, nonceBytes...)
	encoded = append(encoded, messageHash...)
	
	return crypto.Keccak256(encoded)
}

// ToTokenDecimals converts a float to token decimals (6 decimals)
func ToTokenDecimals(amount float64) *big.Int {
	// Polymarket uses 6 decimals
	multiplier := big.NewInt(1000000)
	amountBig := big.NewFloat(amount)
	amountBig.Mul(amountBig, big.NewFloat(0).SetInt(multiplier))
	
	result, _ := amountBig.Int(nil)
	return result
}

// RoundDown rounds down to specified decimal places
func RoundDown(value float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return float64(int(value*multiplier)) / multiplier
}

// RoundUp rounds up to specified decimal places
func RoundUp(value float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return float64(int(value*multiplier)+1) / multiplier
}

// RoundNormal rounds to specified decimal places
func RoundNormal(value float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}
	return float64(int(value*multiplier+0.5)) / multiplier
}

// DecimalPlaces returns the number of decimal places
func DecimalPlaces(value float64) int {
	str := fmt.Sprintf("%.10f", value)
	// Remove trailing zeros
	for len(str) > 0 && str[len(str)-1] == '0' {
		str = str[:len(str)-1]
	}
	// Find decimal point
	for i, c := range str {
		if c == '.' {
			return len(str) - i - 1
		}
	}
	return 0
}

// ValidatePrice validates if price is within tick size bounds
func ValidatePrice(price float64, tickSize types.TickSize) bool {
	tickSizeFloat := ParseTickSize(tickSize)
	return price >= tickSizeFloat && price <= (1.0-tickSizeFloat)
}

// ParseTickSize converts TickSize to float64
func ParseTickSize(tickSize types.TickSize) float64 {
	switch tickSize {
	case types.TickSize01:
		return 0.1
	case types.TickSize001:
		return 0.01
	case types.TickSize0001:
		return 0.001
	case types.TickSize00001:
		return 0.0001
	default:
		return 0.01
	}
}

// GetRoundingConfig returns rounding configuration for tick size
func GetRoundingConfig(tickSize types.TickSize) types.RoundConfig {
	switch tickSize {
	case types.TickSize01:
		return types.RoundConfig{Price: 1, Size: 2, Amount: 3}
	case types.TickSize001:
		return types.RoundConfig{Price: 2, Size: 2, Amount: 4}
	case types.TickSize0001:
		return types.RoundConfig{Price: 3, Size: 2, Amount: 5}
	case types.TickSize00001:
		return types.RoundConfig{Price: 4, Size: 2, Amount: 6}
	default:
		return types.RoundConfig{Price: 2, Size: 2, Amount: 4}
	}
}

// CreateOrderEIP712Hash creates an EIP712 hash for order signing
// This implements the exact same structure as py_order_utils
func CreateOrderEIP712Hash(orderData types.OrderData, salt int64, exchangeAddress string, chainID int64) []byte {
	// Create domain separator for "Polymarket CTF Exchange"
	domainSeparator := CreatePolymarketDomain(chainID, exchangeAddress)
	
	// Create order struct hash
	orderStructHash := CreateOrderStructHash(orderData, salt)
	
	// Create final EIP712 hash
	return CreateEIP712Hash(domainSeparator, orderStructHash)
}

// CreatePolymarketDomain creates the EIP712 domain separator for Polymarket CTF Exchange
func CreatePolymarketDomain(chainID int64, exchangeAddress string) []byte {
	// EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)
	domainTypeHash := crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	
	nameHash := crypto.Keccak256([]byte("Polymarket CTF Exchange"))
	versionHash := crypto.Keccak256([]byte("1"))
	
	chainIDBytes := make([]byte, 32)
	big.NewInt(chainID).FillBytes(chainIDBytes)
	
	// Parse exchange address and pad to 32 bytes
	exchangeAddr := common.HexToAddress(exchangeAddress)
	exchangeBytes := make([]byte, 32)
	copy(exchangeBytes[12:], exchangeAddr.Bytes())
	
	// Encode domain: keccak256(domainTypeHash ‖ nameHash ‖ versionHash ‖ chainId ‖ verifyingContract)
	domain := make([]byte, 0, 160)
	domain = append(domain, domainTypeHash...)
	domain = append(domain, nameHash...)
	domain = append(domain, versionHash...)
	domain = append(domain, chainIDBytes...)
	domain = append(domain, exchangeBytes...)
	
	return crypto.Keccak256(domain)
}

// CreateOrderStructHash creates the struct hash for Order
// Order(uint256 salt,address maker,address signer,address taker,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint256 expiration,uint256 nonce,uint256 feeRateBps,uint8 side,uint8 signatureType)
func CreateOrderStructHash(orderData types.OrderData, salt int64) []byte {
	// Order type hash - MUST match the exact field order from py_order_utils
	orderTypeHash := crypto.Keccak256([]byte("Order(uint256 salt,address maker,address signer,address taker,uint256 tokenId,uint256 makerAmount,uint256 takerAmount,uint256 expiration,uint256 nonce,uint256 feeRateBps,uint8 side,uint8 signatureType)"))
	
	// Convert salt to big.Int
	saltBig := big.NewInt(salt)
	saltBytes := make([]byte, 32)
	saltBig.FillBytes(saltBytes)
	
	// Parse addresses and pad to 32 bytes
	makerAddr := common.HexToAddress(orderData.Maker)
	makerBytes := make([]byte, 32)
	copy(makerBytes[12:], makerAddr.Bytes())
	
	signerAddr := common.HexToAddress(orderData.Signer)
	signerBytes := make([]byte, 32)
	copy(signerBytes[12:], signerAddr.Bytes())
	
	takerAddr := common.HexToAddress(orderData.Taker)
	takerBytes := make([]byte, 32)
	copy(takerBytes[12:], takerAddr.Bytes())
	
	// Parse tokenId
	tokenIdBig := new(big.Int)
	tokenIdBig.SetString(orderData.TokenID, 10)
	tokenIdBytes := make([]byte, 32)
	tokenIdBig.FillBytes(tokenIdBytes)
	
	// MakerAmount and TakerAmount are already *big.Int
	makerAmountBytes := make([]byte, 32)
	orderData.MakerAmount.FillBytes(makerAmountBytes)
	
	takerAmountBytes := make([]byte, 32)
	orderData.TakerAmount.FillBytes(takerAmountBytes)
	
	// Parse expiration
	expirationBig := new(big.Int)
	expirationBig.SetString(orderData.Expiration, 10)
	expirationBytes := make([]byte, 32)
	expirationBig.FillBytes(expirationBytes)
	
	// Parse nonce
	nonceBig := new(big.Int)
	nonceBig.SetString(orderData.Nonce, 10)
	nonceBytes := make([]byte, 32)
	nonceBig.FillBytes(nonceBytes)
	
	// Parse feeRateBps
	feeRateBig := new(big.Int)
	feeRateBig.SetString(orderData.FeeRateBps, 10)
	feeRateBytes := make([]byte, 32)
	feeRateBig.FillBytes(feeRateBytes)
	
	// Side as uint8 (padded to 32 bytes)
	sideBytes := make([]byte, 32)
	sideBytes[31] = byte(orderData.Side)
	
	// SignatureType as uint8 (padded to 32 bytes)
	sigTypeBytes := make([]byte, 32)
	sigTypeBytes[31] = byte(orderData.SignatureType)
	
	// Combine all fields in the exact order from py_order_utils
	encoded := make([]byte, 0, 32*13) // 13 fields * 32 bytes each
	encoded = append(encoded, orderTypeHash...)
	encoded = append(encoded, saltBytes...)
	encoded = append(encoded, makerBytes...)
	encoded = append(encoded, signerBytes...)
	encoded = append(encoded, takerBytes...)
	encoded = append(encoded, tokenIdBytes...)
	encoded = append(encoded, makerAmountBytes...)
	encoded = append(encoded, takerAmountBytes...)
	encoded = append(encoded, expirationBytes...)
	encoded = append(encoded, nonceBytes...)
	encoded = append(encoded, feeRateBytes...)
	encoded = append(encoded, sideBytes...)
	encoded = append(encoded, sigTypeBytes...)
	
	return crypto.Keccak256(encoded)
}