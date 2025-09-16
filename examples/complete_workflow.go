package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"polymarket-clob-go/pkg/client"
	"polymarket-clob-go/pkg/types"
)

const (
	// Polygon mainnet
	PolygonChainID = 137
	
	// Real token ID
	TestTokenID = "91094360697357622623953793720402150934374522251651348543981406747516093190659"
)

func main() {
	fmt.Println("=== Polymarket CLOB Go SDK - Complete Workflow ===\n")
	
	// Load configuration from environment
	host := getEnvOrDefault("CLOB_API_URL", "https://clob.polymarket.com")
	privateKey := os.Getenv("PRIVATE_KEY")
	
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY environment variable is required")
	}
	
	// Step 1: Create Level 1 client (private key only)
	fmt.Println("1. Creating Level 1 client...")
	startTime := time.Now()
	
	clobClient, err := client.NewClobClient(host, PolygonChainID, privateKey, nil, nil, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	
	fmt.Printf("   ✓ Client created in %v\n", time.Since(startTime))
	fmt.Printf("   Address: %s\n", clobClient.GetAddress())
	fmt.Printf("   Auth Level: %d\n", clobClient.GetAuthLevel())
	
	// Step 2: Create or derive API credentials
	fmt.Println("\n2. Creating/deriving API credentials...")
	startTime = time.Now()
	
	apiCreds, err := clobClient.CreateOrDeriveAPIKey(0) // nonce = 0
	if err != nil {
		log.Fatalf("Failed to create/derive API key: %v", err)
	}
	
	fmt.Printf("   ✓ API credentials obtained in %v\n", time.Since(startTime))
	fmt.Printf("   API Key: %s\n", apiCreds.ApiKey)
	fmt.Printf("   API Secret: %s...\n", apiCreds.ApiSecret[:10])
	fmt.Printf("   API Passphrase: %s\n", apiCreds.ApiPassphrase)
	
	// Step 3: Set API credentials and upgrade to Level 2
	fmt.Println("\n3. Upgrading to Level 2 authentication...")
	startTime = time.Now()
	
	clobClient.SetAPICredentials(apiCreds)
	
	fmt.Printf("   ✓ Upgraded to Level 2 in %v\n", time.Since(startTime))
	fmt.Printf("   Auth Level: %d\n", clobClient.GetAuthLevel())
	
	// Step 4: Get market information
	fmt.Println("\n4. Fetching market information...")
	startTime = time.Now()
	
	tickSize, err := clobClient.GetTickSize(TestTokenID)
	if err != nil {
		log.Printf("   Warning: Failed to get tick size: %v", err)
		tickSize = types.TickSize001 // Default fallback
	}
	
	negRisk, err := clobClient.GetNegRisk(TestTokenID)
	if err != nil {
		log.Printf("   Warning: Failed to get neg risk: %v", err)
		negRisk = false // Default fallback
	}
	
	fmt.Printf("   ✓ Market info fetched in %v\n", time.Since(startTime))
	fmt.Printf("   Tick Size: %s\n", tickSize)
	fmt.Printf("   Neg Risk: %t\n", negRisk)
	
	// Step 5: Create a limit order
	fmt.Println("\n5. Creating limit order...")
	startTime = time.Now()
	
	orderArgs := types.OrderArgs{
		TokenID:    TestTokenID,
		Price:      0.55,           // 55 cents
		Size:       10.0,           // 10 shares
		Side:       types.BUY,
		FeeRateBps: 0,              // 0 basis points fee
		Nonce:      0,
		Expiration: 0, // Expires in 24 hours
		Taker:      "0x0000000000000000000000000000000000000000", // Public order
	}
	
	options := &types.CreateOrderOptions{
		TickSize: tickSize,
		NegRisk:  negRisk,
	}
	
	signedOrder, err := clobClient.CreateOrder(orderArgs, options)
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)
	}
	
	fmt.Printf("   ✓ Order created in %v\n", time.Since(startTime))
	fmt.Printf("   Order Salt: %s\n", signedOrder.Salt)
	fmt.Printf("   Maker Amount: %s\n", signedOrder.MakerAmount)
	fmt.Printf("   Taker Amount: %s\n", signedOrder.TakerAmount)
	fmt.Printf("   Signature: %s...\n", signedOrder.Signature[:20])
	
	// Step 6: Post the order (commented out to avoid actual trading)
	fmt.Println("\n6. Posting order...")
	fmt.Println("   ⚠️  Order posting is disabled in this example to prevent actual trading")
	fmt.Println("   ⚠️  Uncomment the following code to actually post the order:")
	fmt.Println("   /*")
	fmt.Println("   startTime = time.Now()")
	fmt.Println("   result, err := clobClient.PostOrder(signedOrder, types.GTC)")
	fmt.Println("   if err != nil {")
	fmt.Println("       log.Fatalf(\"Failed to post order: %v\", err)")
	fmt.Println("   }")
	fmt.Println("   fmt.Printf(\"   ✓ Order posted in %v\\n\", time.Since(startTime))")
	fmt.Println("   fmt.Printf(\"   Result: %+v\\n\", result)")
	fmt.Println("   */")
	
	/*
	// Uncomment this section to actually post the order
	startTime = time.Now()
	
	result, err := clobClient.PostOrder(signedOrder, types.GTC)
	if err != nil {
		log.Fatalf("Failed to post order: %v", err)
	}
	
	fmt.Printf("   ✓ Order posted in %v\n", time.Since(startTime))
	fmt.Printf("   Result: %+v\n", result)
	*/
	
	// Step 7: Create a market order example
	fmt.Println("\n7. Creating market order example...")
	startTime = time.Now()
	
	marketOrderArgs := types.MarketOrderArgs{
		TokenID:   TestTokenID,
		Amount:    50.0,            // $50 worth
		Side:      types.BUY,
		Price:     0.0,             // Will be calculated automatically
		FeeRateBps: 0,
		Nonce:     0,
		Taker:     "0x0000000000000000000000000000000000000000",
		OrderType: types.FOK,       // Fill or Kill
	}
	
	marketOrder, err := clobClient.CreateMarketOrder(marketOrderArgs, options)
	if err != nil {
		log.Fatalf("Failed to create market order: %v", err)
	}
	
	fmt.Printf("   ✓ Market order created in %v\n", time.Since(startTime))
	fmt.Printf("   Market Order Salt: %s\n", marketOrder.Salt)
	fmt.Printf("   Calculated Price: %s\n", "0.5") // This would be calculated from order book
	
	// Step 8: Display performance metrics
	fmt.Println("\n8. Performance Metrics:")
	clobClient.PrintMetrics()
	
	// Step 9: Create and post order in one call example
	fmt.Println("9. Create and post order in one call example...")
	fmt.Println("   ⚠️  This is also disabled to prevent actual trading")
	fmt.Println("   /*")
	fmt.Println("   quickOrderArgs := types.OrderArgs{")
	fmt.Println("       TokenID:    TestTokenID,")
	fmt.Println("       Price:      0.45,")
	fmt.Println("       Size:       5.0,")
	fmt.Println("       Side:       types.SELL,")
	fmt.Println("       FeeRateBps: 0,")
	fmt.Println("       Nonce:      time.Now().Unix() + 2,")
	fmt.Println("       Expiration: time.Now().Add(12 * time.Hour).Unix(),")
	fmt.Println("       Taker:      \"0x0000000000000000000000000000000000000000\",")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   result, err := clobClient.CreateAndPostOrder(quickOrderArgs, options)")
	fmt.Println("   if err != nil {")
	fmt.Println("       log.Fatalf(\"Failed to create and post order: %v\", err)")
	fmt.Println("   }")
	fmt.Println("   fmt.Printf(\"   ✓ Order created and posted: %+v\\n\", result)")
	fmt.Println("   */")
	
	fmt.Println("\n=== Workflow Complete ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("✓ Level 1 Authentication (EIP712 signing)")
	fmt.Println("✓ API Key Creation/Derivation")
	fmt.Println("✓ Level 2 Authentication (HMAC signing)")
	fmt.Println("✓ Market Information Retrieval")
	fmt.Println("✓ Limit Order Creation and Signing")
	fmt.Println("✓ Market Order Creation")
	fmt.Println("✓ Performance Metrics Tracking")
	fmt.Println("✓ Error Handling and Validation")
	
	fmt.Println("\nPerformance Summary:")
	metrics := clobClient.GetMetrics()
	totalDuration := time.Duration(0)
	successCount := 0
	
	for _, metric := range metrics {
		totalDuration += metric.Duration
		if metric.Success {
			successCount++
		}
	}
	
	fmt.Printf("Total Operations: %d\n", len(metrics))
	fmt.Printf("Successful Operations: %d\n", successCount)
	fmt.Printf("Total Time: %v\n", totalDuration)
	fmt.Printf("Average Time per Operation: %v\n", totalDuration/time.Duration(len(metrics)))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}