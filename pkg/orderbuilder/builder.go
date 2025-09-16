package orderbuilder

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"polymarket-clob-go/pkg/signer"
	"polymarket-clob-go/pkg/types"
	"polymarket-clob-go/pkg/utils"
)

const (
	ZeroAddress = "0x0000000000000000000000000000000000000000"
	EOAType     = 0 // Externally Owned Account signature type
)

// OrderBuilder handles order creation and signing
type OrderBuilder struct {
	signer        *signer.Signer
	signatureType int
	funder        string
	metrics       []types.PerformanceMetrics
}

// NewOrderBuilder creates a new order builder
func NewOrderBuilder(s *signer.Signer, signatureType *int, funder *string) *OrderBuilder {
	sigType := EOAType
	if signatureType != nil {
		sigType = *signatureType
	}
	
	funderAddr := s.AddressHex()
	if funder != nil {
		funderAddr = *funder
	}
	
	return &OrderBuilder{
		signer:        s,
		signatureType: sigType,
		funder:        funderAddr,
		metrics:       make([]types.PerformanceMetrics, 0),
	}
}

// CreateOrder creates and signs a limit order
func (ob *OrderBuilder) CreateOrder(orderArgs types.OrderArgs, options types.CreateOrderOptions, exchangeAddress string) (*types.SignedOrder, error) {
	start := time.Now()
	
	// Get order amounts
	side, makerAmount, takerAmount, err := ob.getOrderAmounts(orderArgs.Side, orderArgs.Size, orderArgs.Price, options.TickSize)
	if err != nil {
		ob.recordMetric("order_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to calculate order amounts: %w", err)
	}
	
	// Create order data
	orderData := types.OrderData{
		Maker:         ob.funder,
		Taker:         orderArgs.Taker,
		TokenID:       orderArgs.TokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Side:          side,
		FeeRateBps:    fmt.Sprintf("%d", orderArgs.FeeRateBps),
		Nonce:         fmt.Sprintf("%d", orderArgs.Nonce),
		Signer:        ob.signer.AddressHex(),
		Expiration:    fmt.Sprintf("%d", orderArgs.Expiration),
		SignatureType: ob.signatureType,
	}
	
	// Sign the order
	signedOrder, err := ob.signOrder(orderData, exchangeAddress)
	if err != nil {
		ob.recordMetric("order_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to sign order: %w", err)
	}
	
	ob.recordMetric("order_creation", start, true, "")
	return signedOrder, nil
}

// CreateMarketOrder creates and signs a market order
func (ob *OrderBuilder) CreateMarketOrder(orderArgs types.MarketOrderArgs, options types.CreateOrderOptions, exchangeAddress string) (*types.SignedOrder, error) {
	start := time.Now()
	
	// Get market order amounts
	side, makerAmount, takerAmount, err := ob.getMarketOrderAmounts(orderArgs.Side, orderArgs.Amount, orderArgs.Price, options.TickSize)
	if err != nil {
		ob.recordMetric("market_order_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to calculate market order amounts: %w", err)
	}
	
	// Create order data (market orders have expiration = 0)
	orderData := types.OrderData{
		Maker:         ob.funder,
		Taker:         orderArgs.Taker,
		TokenID:       orderArgs.TokenID,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Side:          side,
		FeeRateBps:    fmt.Sprintf("%d", orderArgs.FeeRateBps),
		Nonce:         fmt.Sprintf("%d", orderArgs.Nonce),
		Signer:        ob.signer.AddressHex(),
		Expiration:    "0", // Market orders don't expire
		SignatureType: ob.signatureType,
	}
	
	// Sign the order
	signedOrder, err := ob.signOrder(orderData, exchangeAddress)
	if err != nil {
		ob.recordMetric("market_order_creation", start, false, err.Error())
		return nil, fmt.Errorf("failed to sign market order: %w", err)
	}
	
	ob.recordMetric("market_order_creation", start, true, "")
	return signedOrder, nil
}

// getOrderAmounts calculates maker and taker amounts for limit orders
func (ob *OrderBuilder) getOrderAmounts(side types.OrderSide, size, price float64, tickSize types.TickSize) (int, *big.Int, *big.Int, error) {
	start := time.Now()
	
	roundConfig := utils.GetRoundingConfig(tickSize)
	rawPrice := utils.RoundNormal(price, roundConfig.Price)
	
	var sideInt int
	var makerAmount, takerAmount *big.Int
	
	if side == types.BUY {
		sideInt = 0 // BUY = 0
		
		rawTakerAmt := utils.RoundDown(size, roundConfig.Size)
		rawMakerAmt := rawTakerAmt * rawPrice
		
		// Handle precision for maker amount
		if utils.DecimalPlaces(rawMakerAmt) > roundConfig.Amount {
			rawMakerAmt = utils.RoundUp(rawMakerAmt, roundConfig.Amount+4)
			if utils.DecimalPlaces(rawMakerAmt) > roundConfig.Amount {
				rawMakerAmt = utils.RoundDown(rawMakerAmt, roundConfig.Amount)
			}
		}
		
		makerAmount = utils.ToTokenDecimals(rawMakerAmt)
		takerAmount = utils.ToTokenDecimals(rawTakerAmt)
		
	} else if side == types.SELL {
		sideInt = 1 // SELL = 1
		
		rawMakerAmt := utils.RoundDown(size, roundConfig.Size)
		rawTakerAmt := rawMakerAmt * rawPrice
		
		// Handle precision for taker amount
		if utils.DecimalPlaces(rawTakerAmt) > roundConfig.Amount {
			rawTakerAmt = utils.RoundUp(rawTakerAmt, roundConfig.Amount+4)
			if utils.DecimalPlaces(rawTakerAmt) > roundConfig.Amount {
				rawTakerAmt = utils.RoundDown(rawTakerAmt, roundConfig.Amount)
			}
		}
		
		makerAmount = utils.ToTokenDecimals(rawMakerAmt)
		takerAmount = utils.ToTokenDecimals(rawTakerAmt)
		
	} else {
		ob.recordMetric("order_amounts_calculation", start, false, "invalid side")
		return 0, nil, nil, fmt.Errorf("invalid order side: %s", side)
	}
	
	ob.recordMetric("order_amounts_calculation", start, true, "")
	return sideInt, makerAmount, takerAmount, nil
}

// getMarketOrderAmounts calculates maker and taker amounts for market orders
func (ob *OrderBuilder) getMarketOrderAmounts(side types.OrderSide, amount, price float64, tickSize types.TickSize) (int, *big.Int, *big.Int, error) {
	start := time.Now()
	
	roundConfig := utils.GetRoundingConfig(tickSize)
	rawPrice := utils.RoundNormal(price, roundConfig.Price)
	
	var sideInt int
	var makerAmount, takerAmount *big.Int
	
	if side == types.BUY {
		sideInt = 0 // BUY = 0
		
		rawMakerAmt := utils.RoundDown(amount, roundConfig.Size)
		rawTakerAmt := rawMakerAmt / rawPrice
		
		// Handle precision for taker amount
		if utils.DecimalPlaces(rawTakerAmt) > roundConfig.Amount {
			rawTakerAmt = utils.RoundUp(rawTakerAmt, roundConfig.Amount+4)
			if utils.DecimalPlaces(rawTakerAmt) > roundConfig.Amount {
				rawTakerAmt = utils.RoundDown(rawTakerAmt, roundConfig.Amount)
			}
		}
		
		makerAmount = utils.ToTokenDecimals(rawMakerAmt)
		takerAmount = utils.ToTokenDecimals(rawTakerAmt)
		
	} else if side == types.SELL {
		sideInt = 1 // SELL = 1
		
		rawMakerAmt := utils.RoundDown(amount, roundConfig.Size)
		rawTakerAmt := rawMakerAmt * rawPrice
		
		// Handle precision for taker amount
		if utils.DecimalPlaces(rawTakerAmt) > roundConfig.Amount {
			rawTakerAmt = utils.RoundUp(rawTakerAmt, roundConfig.Amount+4)
			if utils.DecimalPlaces(rawTakerAmt) > roundConfig.Amount {
				rawTakerAmt = utils.RoundDown(rawTakerAmt, roundConfig.Amount)
			}
		}
		
		makerAmount = utils.ToTokenDecimals(rawMakerAmt)
		takerAmount = utils.ToTokenDecimals(rawTakerAmt)
		
	} else {
		ob.recordMetric("market_order_amounts_calculation", start, false, "invalid side")
		return 0, nil, nil, fmt.Errorf("invalid order side: %s", side)
	}
	
	ob.recordMetric("market_order_amounts_calculation", start, true, "")
	return sideInt, makerAmount, takerAmount, nil
}

// signOrder signs an order using EIP712
func (ob *OrderBuilder) signOrder(orderData types.OrderData, exchangeAddress string) (*types.SignedOrder, error) {
	start := time.Now()
	
	// Generate salt using Python-compatible method
	// Python: round(datetime.now().timestamp() * random())
	now := float64(time.Now().Unix())
	salt := int64(now * rand.Float64())
	
	// Create order hash for signing using EIP712 (matches py_order_utils)
	orderHash := utils.CreateOrderEIP712Hash(orderData, salt, exchangeAddress, ob.signer.ChainID())
	
	// Sign the hash
	signature, err := ob.signer.Sign(orderHash)
	if err != nil {
		ob.recordMetric("order_signing", start, false, err.Error())
		return nil, fmt.Errorf("failed to sign order hash: %w", err)
	}
	
	// Convert side integer to OrderSide string
	var sideStr types.OrderSide
	if orderData.Side == 0 {
		sideStr = types.BUY
	} else {
		sideStr = types.SELL
	}
	
	// Create signed order
	signedOrder := &types.SignedOrder{
		Salt:          salt,
		Maker:         orderData.Maker,
		Signer:        orderData.Signer,
		Taker:         orderData.Taker,
		TokenID:       orderData.TokenID,
		MakerAmount:   orderData.MakerAmount.String(),
		TakerAmount:   orderData.TakerAmount.String(),
		Expiration:    orderData.Expiration,
		Nonce:         orderData.Nonce,
		FeeRateBps:    orderData.FeeRateBps,
		Side:          sideStr,
		SignatureType: orderData.SignatureType,
		Signature:     fmt.Sprintf("0x%x", signature),
	}
	
	ob.recordMetric("order_signing", start, true, "")
	return signedOrder, nil
}

// GetMetrics returns performance metrics
func (ob *OrderBuilder) GetMetrics() []types.PerformanceMetrics {
	return ob.metrics
}

// ClearMetrics clears performance metrics
func (ob *OrderBuilder) ClearMetrics() {
	ob.metrics = make([]types.PerformanceMetrics, 0)
}

// recordMetric records a performance metric
func (ob *OrderBuilder) recordMetric(operation string, startTime time.Time, success bool, errorMsg string) {
	metric := types.PerformanceMetrics{
		Operation: operation,
		StartTime: startTime,
		Duration:  time.Since(startTime),
		Success:   success,
		Error:     errorMsg,
	}
	ob.metrics = append(ob.metrics, metric)
}