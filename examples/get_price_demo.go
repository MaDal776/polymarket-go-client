package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"polymarket-clob-go/pkg/client"
	"polymarket-clob-go/pkg/types"
)

func main() {
	fmt.Println("🔍 Polymarket CLOB Go SDK - 价格查询演示")
	fmt.Println("==========================================")

	// 配置
	host := getEnvOrDefault("POLYMARKET_HOST", "https://clob.polymarket.com")
	chainID := getEnvAsIntOrDefault("CHAIN_ID", 137)
	privateKey := os.Getenv("PRIVATE_KEY")
	signatureType := int(getEnvAsIntOrDefault("SIGNATURE_TYPE", 0))

	// 示例代币 ID (可以通过环境变量覆盖)
	// 使用 Python 示例中验证有效的 token ID
	tokenID := getEnvOrDefault("TOKEN_ID", "91094360697357622623953793720402150934374522251651348543981406747516093190659")

	fmt.Printf("📋 配置信息:\n")
	fmt.Printf("   Host: %s\n", host)
	fmt.Printf("   Chain ID: %d\n", chainID)
	fmt.Printf("   Token ID: %s...\n", tokenID[:20])

	// 创建客户端 (价格查询不需要私钥，但为了完整性还是包含)
	var clobClient *client.ClobClient
	var err error

	if privateKey != "" {
		clobClient, err = client.NewClobClient(
			host,
			chainID,
			privateKey,
			nil,
			&signatureType,
			nil,
		)
		if err != nil {
			log.Fatalf("❌ 创建客户端失败: %v", err)
		}
		fmt.Printf("✅ 客户端创建成功 (带认证)\n")
	} else {
		// 价格查询是公开的，不需要私钥
		clobClient, err = client.NewClobClient(
			host,
			chainID,
			"", // 空私钥
			nil,
			&signatureType,
			nil,
		)
		if err != nil {
			log.Fatalf("❌ 创建客户端失败: %v", err)
		}
		fmt.Printf("✅ 客户端创建成功 (公开模式)\n")
	}

	// 1. 获取单个代币的买1价格
	fmt.Println("\n💰 获取买1价格 (BUY)...")
	buyPrice, err := clobClient.GetPrice(tokenID, types.BUY)
	if err != nil {
		fmt.Printf("❌ 获取买1价格失败: %v\n", err)
	} else {
		fmt.Printf("📈 买1价格: %s\n", buyPrice.Price)
		if price, err := strconv.ParseFloat(buyPrice.Price, 64); err == nil {
			fmt.Printf("📊 买1概率: %.2f%%\n", price*100)
			fmt.Printf("💵 买1价格: $%.4f per share\n", price)
		}
	}

	// 2. 获取单个代币的卖1价格
	fmt.Println("\n💰 获取卖1价格 (SELL)...")
	sellPrice, err := clobClient.GetPrice(tokenID, types.SELL)
	if err != nil {
		fmt.Printf("❌ 获取卖1价格失败: %v\n", err)
	} else {
		fmt.Printf("📉 卖1价格: %s\n", sellPrice.Price)
		if price, err := strconv.ParseFloat(sellPrice.Price, 64); err == nil {
			fmt.Printf("📊 卖1概率: %.2f%%\n", price*100)
			fmt.Printf("💵 卖1价格: $%.4f per share\n", price)
		}
	}

	// 3. 计算买卖价差
	if buyPrice != nil && sellPrice != nil {
		fmt.Println("\n📊 价差分析...")
		buyPriceFloat, buyErr := strconv.ParseFloat(buyPrice.Price, 64)
		sellPriceFloat, sellErr := strconv.ParseFloat(sellPrice.Price, 64)

		if buyErr == nil && sellErr == nil {
			spread := sellPriceFloat - buyPriceFloat
			spreadPercent := (spread / ((buyPriceFloat + sellPriceFloat) / 2)) * 100

			fmt.Printf("📏 买卖价差: %.4f\n", spread)
			fmt.Printf("📈 价差百分比: %.2f%%\n", spreadPercent)

			if spread > 0 {
				fmt.Printf("💡 市场状态: 正常 (卖价 > 买价)\n")
			} else {
				fmt.Printf("⚠️  市场状态: 异常 (买价 >= 卖价)\n")
			}
		}
	}

	// 4. 批量获取多个代币价格 (如果有多个代币ID)
	fmt.Println("\n📊 批量价格查询演示...")

	// 使用相同代币的买卖价格作为演示
	priceParams := []types.BookParams{
		{TokenID: tokenID, Side: types.BUY},
		{TokenID: tokenID, Side: types.SELL},
	}

	// 如果有第二个代币ID，可以添加
	secondTokenID := os.Getenv("SECOND_TOKEN_ID")
	if secondTokenID != "" {
		priceParams = append(priceParams,
			types.BookParams{TokenID: secondTokenID, Side: types.BUY},
			types.BookParams{TokenID: secondTokenID, Side: types.SELL},
		)
		fmt.Printf("📋 包含第二个代币: %s...\n", secondTokenID[:20])
	}

	prices, err := clobClient.GetPrices(priceParams)
	if err != nil {
		fmt.Printf("❌ 批量获取价格失败: %v\n", err)
	} else {
		fmt.Printf("✅ 批量获取 %d 个价格成功\n", len(prices))
		for i, price := range prices {
			param := priceParams[i]
			fmt.Printf("   %s %s: %s\n", param.Side, param.TokenID[:20]+"...", price.Price)
		}
	}

	// 5. 实用函数演示
	fmt.Println("\n🛠️  实用函数演示...")
	if buyPrice != nil && sellPrice != nil {
		demonstratePriceUtilities(buyPrice, sellPrice)
	}

	// 6. 性能指标
	fmt.Println("\n📈 性能指标:")
	clobClient.PrintMetrics()

	fmt.Println("\n✅ 价格查询演示完成!")
	fmt.Println("\n💡 使用提示:")
	fmt.Println("   - 价格查询是公开的，不需要私钥")
	fmt.Println("   - 价格范围在 0-1 之间，代表概率")
	fmt.Println("   - 买1价格通常低于卖1价格")
	fmt.Println("   - 可以设置 SECOND_TOKEN_ID 环境变量测试多个代币")
}

// 演示价格相关的实用函数
func demonstratePriceUtilities(buyPrice, sellPrice *types.PriceResponse) {
	buyPriceFloat, _ := strconv.ParseFloat(buyPrice.Price, 64)
	sellPriceFloat, _ := strconv.ParseFloat(sellPrice.Price, 64)

	fmt.Printf("🔧 价格分析工具:\n")

	// 中间价格
	midPrice := (buyPriceFloat + sellPriceFloat) / 2
	fmt.Printf("   📊 中间价格: %.4f (%.2f%%)\n", midPrice, midPrice*100)

	// 流动性评估
	spread := sellPriceFloat - buyPriceFloat
	if spread < 0.01 {
		fmt.Printf("   💧 流动性: 高 (价差 < 1%%)\n")
	} else if spread < 0.05 {
		fmt.Printf("   💧 流动性: 中等 (价差 1-5%%)\n")
	} else {
		fmt.Printf("   💧 流动性: 低 (价差 > 5%%)\n")
	}

	// 交易建议
	if buyPriceFloat < 0.3 {
		fmt.Printf("   💡 交易建议: 考虑买入 (低概率事件)\n")
	} else if buyPriceFloat > 0.7 {
		fmt.Printf("   💡 交易建议: 考虑卖出 (高概率事件)\n")
	} else {
		fmt.Printf("   💡 交易建议: 中性 (概率适中)\n")
	}

	// 风险评估
	if spread > 0.1 {
		fmt.Printf("   ⚠️  风险提示: 价差较大，注意流动性风险\n")
	}
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