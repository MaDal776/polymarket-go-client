package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"polymarket-clob-go/pkg/client"
	"polymarket-clob-go/pkg/types"
)

// 配置结构
type Config struct {
	PrivateKey    string
	Host          string
	ChainID       int64
	SignatureType int
	TokenID       string
}

func main() {
	fmt.Println("🚀 Polymarket CLOB Go SDK - 完整功能演示")
	fmt.Println(strings.Repeat("=", 60))

	// 1. 加载配置
	config := loadConfig()
	printConfig(config)

	// 2. 创建客户端
	fmt.Println("\n🔧 步骤 1: 初始化客户端")
	fmt.Println(strings.Repeat("-", 40))
	
	startTime := time.Now()
	clobClient, err := client.NewClobClient(
		config.Host,
		config.ChainID,
		config.PrivateKey,
		nil,
		&config.SignatureType,
		nil,
	)
	if err != nil {
		log.Fatalf("❌ 创建客户端失败: %v", err)
	}
	
	initDuration := time.Since(startTime)
	fmt.Printf("✅ 客户端创建成功 (耗时: %v)\n", initDuration)
	fmt.Printf("   钱包地址: %s\n", clobClient.GetAddress())
	fmt.Printf("   认证级别: %d\n", clobClient.GetAuthLevel())

	// 3. 设置 API 凭证
	fmt.Println("\n🔑 步骤 2: 设置 API 凭证")
	fmt.Println(strings.Repeat("-", 40))
	
	authStart := time.Now()
	apiCreds, err := clobClient.CreateOrDeriveAPIKey(0)
	if err != nil {
		log.Fatalf("❌ 获取 API 凭证失败: %v", err)
	}
	clobClient.SetAPICredentials(apiCreds)
	authDuration := time.Since(authStart)
	
	fmt.Printf("✅ API 凭证设置完成 (耗时: %v)\n", authDuration)
	fmt.Printf("   认证级别: %d\n", clobClient.GetAuthLevel())
	fmt.Printf("   API Key: %s...\n", apiCreds.Key[:10])

	// 4. 获取余额信息
	fmt.Println("\n💰 步骤 3: 检查账户余额")
	fmt.Println(strings.Repeat("-", 40))
	
	balanceStart := time.Now()
	usdcAmount, hasBalance := checkUSDCBalance(clobClient, config.SignatureType)
	checkTokenBalance(clobClient, config.TokenID, config.SignatureType)
	balanceDuration := time.Since(balanceStart)
	
	fmt.Printf("✅ 余额检查完成 (耗时: %v)\n", balanceDuration)

	// 5. 更新余额
	fmt.Println("\n🔄 步骤 4: 更新余额信息")
	fmt.Println(strings.Repeat("-", 40))
	
	updateStart := time.Now()
	updateBalance(clobClient, config.SignatureType)
	updateDuration := time.Since(updateStart)
	
	fmt.Printf("✅ 余额更新完成 (耗时: %v)\n", updateDuration)

	// 6. 获取市场数据
	fmt.Println("\n📊 步骤 5: 获取市场数据")
	fmt.Println(strings.Repeat("-", 40))
	
	marketStart := time.Now()
	tickSize, negRisk := getMarketData(clobClient, config.TokenID)
	buyPrice, sellPrice := getPriceData(clobClient, config.TokenID)
	marketDuration := time.Since(marketStart)
	
	fmt.Printf("✅ 市场数据获取完成 (耗时: %v)\n", marketDuration)

	// 7. 创建和管理订单
	fmt.Println("\n📝 步骤 6: 订单创建和管理")
	fmt.Println(strings.Repeat("-", 40))
	
	orderStart := time.Now()
	demonstrateOrderCreation(clobClient, config.TokenID, tickSize, negRisk, hasBalance, usdcAmount)
	orderDuration := time.Since(orderStart)
	
	fmt.Printf("✅ 订单演示完成 (耗时: %v)\n", orderDuration)

	// 8. 性能总结
	fmt.Println("\n📈 步骤 7: 性能总结")
	fmt.Println(strings.Repeat("-", 40))
	
	totalDuration := time.Since(startTime)
	printPerformanceSummary(clobClient, map[string]time.Duration{
		"客户端初始化": initDuration,
		"API 认证":   authDuration,
		"余额检查":    balanceDuration,
		"余额更新":    updateDuration,
		"市场数据":    marketDuration,
		"订单操作":    orderDuration,
		"总耗时":     totalDuration,
	})

	fmt.Println("\n🎉 完整演示结束!")
	fmt.Println("\n💡 下一步建议:")
	fmt.Println("   1. 修改 TOKEN_ID 环境变量来交易不同市场")
	fmt.Println("   2. 调整订单参数 (价格、数量) 满足需求")
	fmt.Println("   3. 查看 GO_SDK_USAGE_GUIDE.md 了解更多功能")
}

// 加载配置
func loadConfig() *Config {
	config := &Config{
		PrivateKey:    os.Getenv("PRIVATE_KEY"),
		Host:          getEnvOrDefault("POLYMARKET_HOST", "https://clob.polymarket.com"),
		ChainID:       getEnvAsIntOrDefault("CHAIN_ID", 137),
		SignatureType: int(getEnvAsIntOrDefault("SIGNATURE_TYPE", 1)),
		TokenID:       getEnvOrDefault("TOKEN_ID", "91094360697357622623953793720402150934374522251651348543981406747516093190659"),
	}

	if config.PrivateKey == "" {
		log.Fatal("❌ 请设置 PRIVATE_KEY 环境变量")
	}

	return config
}

// 打印配置信息
func printConfig(config *Config) {
	fmt.Printf("📋 当前配置:\n")
	fmt.Printf("   API 主机: %s\n", config.Host)
	fmt.Printf("   链 ID: %d\n", config.ChainID)
	fmt.Printf("   签名类型: %d\n", config.SignatureType)
	fmt.Printf("   代币 ID: %s...\n", config.TokenID[:20])
	fmt.Printf("   私钥: %s...***\n", config.PrivateKey[:10])
}

// 检查 USDC 余额
func checkUSDCBalance(client *client.ClobClient, signatureType int) (float64, bool) {
	fmt.Println("💵 检查 USDC 余额...")
	
	usdcBalance, err := client.GetBalanceAllowance(&types.BalanceAllowanceParams{
		AssetType:     types.COLLATERAL,
		SignatureType: signatureType,
	})

	var usdcAmount float64
	var hasBalance bool

	if err != nil {
		fmt.Printf("❌ 获取 USDC 余额失败: %v\n", err)
		return 0, false
	}

	fmt.Printf("📊 原始余额数据: %s\n", usdcBalance.Balance)
	fmt.Printf("📊 授权额度: %s\n", usdcBalance.Allowance)

	if usdcBalance.Balance != "" && usdcBalance.Balance != "0" {
		if balance, err := strconv.ParseFloat(usdcBalance.Balance, 64); err == nil {
			usdcAmount = balance / 1000000 // USDC 有 6 位小数
			hasBalance = usdcAmount > 0
			fmt.Printf("💰 USDC 余额: %.6f USDC\n", usdcAmount)
			
			// 计算可交易信息
			if hasBalance {
				minOrderSize := 1.0
				maxOrders := int(usdcAmount / minOrderSize)
				fmt.Printf("📈 可下单数量: %d 个 $1 订单\n", maxOrders)
			}
		}
	} else {
		fmt.Printf("💰 USDC 余额: 0.000000 USDC\n")
	}

	return usdcAmount, hasBalance
}

// 检查代币余额
func checkTokenBalance(client *client.ClobClient, tokenID string, signatureType int) {
	fmt.Printf("🎯 检查代币余额 (ID: %s...)...\n", tokenID[:20])
	
	tokenBalance, err := client.GetBalanceAllowance(&types.BalanceAllowanceParams{
		AssetType:     types.CONDITIONAL,
		TokenID:       tokenID,
		SignatureType: signatureType,
	})

	if err != nil {
		fmt.Printf("❌ 获取代币余额失败: %v\n", err)
		return
	}

	fmt.Printf("📊 代币原始余额: %s\n", tokenBalance.Balance)

	if tokenBalance.Balance != "" && tokenBalance.Balance != "0" {
		if balance, err := strconv.ParseFloat(tokenBalance.Balance, 64); err == nil {
			tokenAmount := balance / 1000000 // 假设 6 位小数
			fmt.Printf("🎯 代币数量: %.6f tokens\n", tokenAmount)
		}
	} else {
		fmt.Printf("🎯 代币数量: 0.000000 tokens\n")
	}
}

// 更新余额
func updateBalance(client *client.ClobClient, signatureType int) {
	fmt.Println("🔄 更新 USDC 余额...")
	
	updatedBalance, err := client.UpdateBalanceAllowance(&types.BalanceAllowanceParams{
		AssetType:     types.COLLATERAL,
		SignatureType: signatureType,
	})

	if err != nil {
		fmt.Printf("⚠️  更新余额警告: %v\n", err)
	} else {
		fmt.Printf("✅ 余额更新成功\n")
		if updatedBalance.Balance != "updated" {
			fmt.Printf("📊 更新后余额: %s\n", updatedBalance.Balance)
		}
	}
}

// 获取市场数据
func getMarketData(client *client.ClobClient, tokenID string) (types.TickSize, bool) {
	fmt.Printf("📊 获取市场数据 (Token: %s...)...\n", tokenID[:20])
	
	// 获取 tick size
	tickSize, err := client.GetTickSize(tokenID)
	if err != nil {
		fmt.Printf("⚠️  获取 tick size 失败: %v\n", err)
		tickSize = types.TickSize001 // 使用默认值
	} else {
		fmt.Printf("📏 Tick Size: %s\n", tickSize)
	}

	// 获取 neg risk
	negRisk, err := client.GetNegRisk(tokenID)
	if err != nil {
		fmt.Printf("⚠️  获取 neg risk 失败: %v\n", err)
		negRisk = false // 使用默认值
	} else {
		fmt.Printf("⚠️  Neg Risk: %t\n", negRisk)
	}

	return tickSize, negRisk
}

// 获取价格数据
func getPriceData(client *client.ClobClient, tokenID string) (*types.PriceResponse, *types.PriceResponse) {
	fmt.Printf("💰 获取价格数据 (Token: %s...)...\n", tokenID[:20])
	
	// 获取买1价格
	buyPrice, err := client.GetPrice(tokenID, types.BUY)
	if err != nil {
		fmt.Printf("⚠️  获取买1价格失败: %v\n", err)
		buyPrice = nil
	} else {
		fmt.Printf("📈 买1价格: %s\n", buyPrice.Price)
	}

	// 获取卖1价格
	sellPrice, err := client.GetPrice(tokenID, types.SELL)
	if err != nil {
		fmt.Printf("⚠️  获取卖1价格失败: %v\n", err)
		sellPrice = nil
	} else {
		fmt.Printf("📉 卖1价格: %s\n", sellPrice.Price)
	}

	// 计算价差
	if buyPrice != nil && sellPrice != nil {
		if buyPriceFloat, err1 := strconv.ParseFloat(buyPrice.Price, 64); err1 == nil {
			if sellPriceFloat, err2 := strconv.ParseFloat(sellPrice.Price, 64); err2 == nil {
				spread := sellPriceFloat - buyPriceFloat
				fmt.Printf("📏 买卖价差: %.4f\n", spread)
			}
		}
	}

	return buyPrice, sellPrice
}

// 演示订单创建
func demonstrateOrderCreation(client *client.ClobClient, tokenID string, tickSize types.TickSize, negRisk bool, hasBalance bool, usdcAmount float64) {
	fmt.Println("📝 创建示例订单...")

	// 创建限价订单
	fmt.Println("\n📋 限价订单示例:")
	createLimitOrderExample(client, tokenID, tickSize, negRisk, hasBalance)

	// 创建市价订单示例 (仅演示，不提交)
	fmt.Println("\n📋 市价订单示例:")
	createMarketOrderExample(client, tokenID, tickSize, negRisk)

	// 如果有余额，提供交互选项
	if hasBalance {
		fmt.Printf("\n💰 当前 USDC 余额: %.6f USDC\n", usdcAmount)
		fmt.Println("❓ 是否要提交一个真实订单? (输入 'yes' 确认，其他任意键跳过)")
		
		var response string
		fmt.Scanln(&response)
		
		if strings.ToLower(response) == "yes" {
			submitRealOrder(client, tokenID, tickSize, negRisk)
		} else {
			fmt.Println("⏭️  跳过真实订单提交")
		}
	} else {
		fmt.Println("\n💡 无法提交真实订单 - 需要 USDC 余额")
		fmt.Println("   请访问 https://polymarket.com 充值后再试")
	}
}

// 创建限价订单示例
func createLimitOrderExample(client *client.ClobClient, tokenID string, tickSize types.TickSize, negRisk bool) {
	orderArgs := types.OrderArgs{
		TokenID:    tokenID,
		Price:      0.55,           // 55% 概率
		Size:       2.0,            // 2 个单位
		Side:       types.BUY,      // 买入
		FeeRateBps: 0,              // 0 手续费
		Nonce:      0,              // 随机数
		Expiration: 0,              // 不过期
		Taker:      "0x0000000000000000000000000000000000000000",
	}

	options := &types.CreateOrderOptions{
		TickSize: tickSize,
		NegRisk:  negRisk,
	}

	signedOrder, err := client.CreateOrder(orderArgs, options)
	if err != nil {
		fmt.Printf("❌ 创建限价订单失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 限价订单创建成功\n")
	printOrderDetails(orderArgs, signedOrder)
}

// 创建市价订单示例
func createMarketOrderExample(client *client.ClobClient, tokenID string, tickSize types.TickSize, negRisk bool) {
	marketArgs := types.MarketOrderArgs{
		TokenID:   tokenID,
		Amount:    10.0,            // $10
		Side:      types.BUY,
		OrderType: types.FOK,       // Fill or Kill
	}

	options := &types.CreateOrderOptions{
		TickSize: tickSize,
		NegRisk:  negRisk,
	}

	signedOrder, err := client.CreateMarketOrder(marketArgs, options)
	if err != nil {
		fmt.Printf("❌ 创建市价订单失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 市价订单创建成功\n")
	fmt.Printf("   金额: $%.2f\n", marketArgs.Amount)
	fmt.Printf("   方向: %s\n", marketArgs.Side)
	fmt.Printf("   类型: %s\n", marketArgs.OrderType)
	fmt.Printf("   Maker Amount: %s\n", signedOrder.MakerAmount)
	fmt.Printf("   Taker Amount: %s\n", signedOrder.TakerAmount)
}

// 提交真实订单
func submitRealOrder(client *client.ClobClient, tokenID string, tickSize types.TickSize, negRisk bool) {
	fmt.Println("📤 创建并提交真实订单...")

	// 创建一个小额测试订单
	orderArgs := types.OrderArgs{
		TokenID:    tokenID,
		Price:      0.50,           // 50% 概率
		Size:       1.0,            // 1 个单位 (约 $0.50)
		Side:       types.BUY,      // 买入
		FeeRateBps: 0,              // 0 手续费
		Nonce:      time.Now().Unix(), // 使用时间戳作为随机数
		Expiration: 0,              // 不过期
		Taker:      "0x0000000000000000000000000000000000000000",
	}

	options := &types.CreateOrderOptions{
		TickSize: tickSize,
		NegRisk:  negRisk,
	}

	signedOrder, err := client.CreateOrder(orderArgs, options)
	if err != nil {
		fmt.Printf("❌ 创建订单失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 订单创建成功，准备提交...\n")
	printOrderDetails(orderArgs, signedOrder)

	// 提交订单
	result, err := client.PostOrder(signedOrder, types.GTC)
	if err != nil {
		fmt.Printf("❌ 订单提交失败: %v\n", err)
		
		// 提供错误解决建议
		if strings.Contains(err.Error(), "not enough balance") {
			fmt.Printf("💡 解决方案: 余额不足，请充值 USDC\n")
		} else if strings.Contains(err.Error(), "invalid amount") {
			fmt.Printf("💡 解决方案: 订单金额不符合要求，请调整价格或数量\n")
		} else if strings.Contains(err.Error(), "insufficient auth level") {
			fmt.Printf("💡 解决方案: 认证级别不足，请检查 API 凭证\n")
		}
	} else {
		fmt.Printf("🎉 订单提交成功!\n")
		fmt.Printf("📋 提交结果: %+v\n", result)
	}
}

// 打印订单详情
func printOrderDetails(orderArgs types.OrderArgs, signedOrder *types.SignedOrder) {
	fmt.Printf("📋 订单详情:\n")
	fmt.Printf("   价格: %.2f (%.0f%%)\n", orderArgs.Price, orderArgs.Price*100)
	fmt.Printf("   数量: %.1f\n", orderArgs.Size)
	fmt.Printf("   方向: %s\n", orderArgs.Side)
	fmt.Printf("   总价值: $%.2f\n", orderArgs.Price*orderArgs.Size)
	fmt.Printf("   Maker: %s\n", signedOrder.Maker)
	fmt.Printf("   Maker Amount: %s\n", signedOrder.MakerAmount)
	fmt.Printf("   Taker Amount: %s\n", signedOrder.TakerAmount)
	fmt.Printf("   Salt: %d\n", signedOrder.Salt)
	fmt.Printf("   Signature Type: %d\n", signedOrder.SignatureType)
}

// 打印性能总结
func printPerformanceSummary(client *client.ClobClient, durations map[string]time.Duration) {
	fmt.Println("⏱️  操作耗时统计:")
	for operation, duration := range durations {
		fmt.Printf("   %s: %v\n", operation, duration)
	}

	fmt.Println("\n📊 详细性能指标:")
	client.PrintMetrics()
}

// 辅助函数
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}