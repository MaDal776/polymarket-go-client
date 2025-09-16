package client

import (
	"testing"
	"time"

	"polymarket-clob-go/pkg/types"
)

const (
	testPrivateKey = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	testChainID    = 137 // Polygon mainnet
	testHost       = "https://clob.polymarket.com"
	testTokenID    = "91094360697357622623953793720402150934374522251651348543981406747516093190659"
)

func TestNewClobClient(t *testing.T) {
	client, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.GetAuthLevel() != types.L1 {
		t.Errorf("Expected auth level L1, got %d", client.GetAuthLevel())
	}

	if client.GetAddress() == "" {
		t.Error("Expected non-empty address")
	}
}

func TestCreateOrder(t *testing.T) {
	client, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	orderArgs := types.OrderArgs{
		TokenID:    testTokenID,
		Price:      0.55,
		Size:       10.0,
		Side:       types.BUY,
		FeeRateBps: 0,
		Nonce:      time.Now().Unix(),
		Expiration: time.Now().Add(24 * time.Hour).Unix(),
		Taker:      "0x0000000000000000000000000000000000000000",
	}

	options := &types.CreateOrderOptions{
		TickSize: types.TickSize001,
		NegRisk:  false,
	}

	signedOrder, err := client.CreateOrder(orderArgs, options)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	if signedOrder.Salt == "" {
		t.Error("Expected non-empty salt")
	}

	if signedOrder.Signature == "" {
		t.Error("Expected non-empty signature")
	}

	if signedOrder.MakerAmount == "" {
		t.Error("Expected non-empty maker amount")
	}

	if signedOrder.TakerAmount == "" {
		t.Error("Expected non-empty taker amount")
	}
}

func TestCreateMarketOrder(t *testing.T) {
	client, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	marketOrderArgs := types.MarketOrderArgs{
		TokenID:   testTokenID,
		Amount:    50.0,
		Side:      types.BUY,
		Price:     0.5,
		FeeRateBps: 0,
		Nonce:     time.Now().Unix(),
		Taker:     "0x0000000000000000000000000000000000000000",
		OrderType: types.FOK,
	}

	options := &types.CreateOrderOptions{
		TickSize: types.TickSize001,
		NegRisk:  false,
	}

	signedOrder, err := client.CreateMarketOrder(marketOrderArgs, options)
	if err != nil {
		t.Fatalf("Failed to create market order: %v", err)
	}

	if signedOrder.Expiration != "0" {
		t.Error("Market orders should have expiration = 0")
	}
}

func TestMetrics(t *testing.T) {
	client, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Perform some operations to generate metrics
	orderArgs := types.OrderArgs{
		TokenID:    testTokenID,
		Price:      0.55,
		Size:       10.0,
		Side:       types.BUY,
		FeeRateBps: 0,
		Nonce:      time.Now().Unix(),
		Expiration: time.Now().Add(24 * time.Hour).Unix(),
		Taker:      "0x0000000000000000000000000000000000000000",
	}

	options := &types.CreateOrderOptions{
		TickSize: types.TickSize001,
		NegRisk:  false,
	}

	_, err = client.CreateOrder(orderArgs, options)
	if err != nil {
		t.Fatalf("Failed to create order: %v", err)
	}

	metrics := client.GetMetrics()
	if len(metrics) == 0 {
		t.Error("Expected metrics to be recorded")
	}

	// Check that metrics have required fields
	for _, metric := range metrics {
		if metric.Operation == "" {
			t.Error("Expected non-empty operation name")
		}
		if metric.Duration == 0 {
			t.Error("Expected non-zero duration")
		}
	}

	// Test clearing metrics
	client.ClearMetrics()
	clearedMetrics := client.GetMetrics()
	if len(clearedMetrics) != 0 {
		t.Error("Expected metrics to be cleared")
	}
}

// Benchmark tests
func BenchmarkCreateOrder(b *testing.B) {
	client, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	orderArgs := types.OrderArgs{
		TokenID:    testTokenID,
		Price:      0.55,
		Size:       10.0,
		Side:       types.BUY,
		FeeRateBps: 0,
		Nonce:      time.Now().Unix(),
		Expiration: time.Now().Add(24 * time.Hour).Unix(),
		Taker:      "0x0000000000000000000000000000000000000000",
	}

	options := &types.CreateOrderOptions{
		TickSize: types.TickSize001,
		NegRisk:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderArgs.Nonce = int64(i) // Unique nonce for each iteration
		_, err := client.CreateOrder(orderArgs, options)
		if err != nil {
			b.Fatalf("Failed to create order: %v", err)
		}
	}
}

func BenchmarkSignerCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
	}
}

func BenchmarkOrderAmountCalculation(b *testing.B) {
	client, err := NewClobClient(testHost, testChainID, testPrivateKey, nil, nil, nil)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}

	orderArgs := types.OrderArgs{
		TokenID:    testTokenID,
		Price:      0.55,
		Size:       10.0,
		Side:       types.BUY,
		FeeRateBps: 0,
		Nonce:      time.Now().Unix(),
		Expiration: time.Now().Add(24 * time.Hour).Unix(),
		Taker:      "0x0000000000000000000000000000000000000000",
	}

	options := &types.CreateOrderOptions{
		TickSize: types.TickSize001,
		NegRisk:  false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This benchmarks the order amount calculation part
		_, err := client.orderBuilder.CreateOrder(orderArgs, *options, "0x1234567890123456789012345678901234567890")
		if err != nil {
			b.Fatalf("Failed to create order: %v", err)
		}
	}
}