# Polymarket CLOB Go SDK 完整使用指南

## 📋 概述

Polymarket CLOB Go SDK 是一个功能完整的 Go 语言客户端库，用于与 Polymarket 中央限价订单簿 (CLOB) 进行交互。支持账户管理、余额查询、市场数据获取、订单创建和提交等全部功能。

## 🚀 快速开始

### 环境要求
- Go 1.19 或更高版本
- 以太坊私钥
- 网络连接到 Polygon 主网

### 安装依赖
```bash
go mod init your-project
go get github.com/ethereum/go-ethereum
go get github.com/shopspring/decimal
```

## 🔧 用户需要提供的信息

### 必需信息
1. **私钥 (PRIVATE_KEY)** - 你的以太坊钱包私钥
   - 格式：64位十六进制字符串，不包含 `0x` 前缀
   - 示例：`abcd1234...` (64个字符)

### 可选配置
2. **API 主机地址** - 默认：`https://clob.polymarket.com`
3. **链 ID** - 默认：`137` (Polygon 主网)
4. **签名类型** - 默认：`0` (EOA 签名)

### 配置方式选择

#### 方式一：环境变量 (推荐)
```bash
# 设置环境变量
export PRIVATE_KEY="your_private_key_here"
export POLYMARKET_HOST="https://clob.polymarket.com"  # 可选
export CHAIN_ID="137"                                 # 可选
export SIGNATURE_TYPE="0"                             # 可选
```

#### 方式二：.env 文件
创建 `.env` 文件：
```bash
PRIVATE_KEY=your_private_key_here
POLYMARKET_HOST=https://clob.polymarket.com
CHAIN_ID=137
SIGNATURE_TYPE=0
```

#### 方式三：代码中直接配置 (不推荐)
```go
privateKey := "your_private_key_here"
host := "https://clob.polymarket.com"
chainID := int64(137)
```

## 📚 核心功能使用说明

### 1. 客户端初始化

```go
import (
    "polymarket-clob-go/pkg/client"
    "polymarket-clob-go/pkg/types"
)

// 创建客户端
host := "https://clob.polymarket.com"
chainID := int64(137)
privateKey := os.Getenv("PRIVATE_KEY")
signatureType := 0

clobClient, err := client.NewClobClient(
    host, 
    chainID, 
    privateKey, 
    nil, 
    &signatureType, 
    nil,
)
if err != nil {
    log.Fatal("创建客户端失败:", err)
}

// 设置 API 凭证
apiCreds, err := clobClient.CreateOrDeriveAPIKey(0)
if err != nil {
    log.Fatal("获取 API 凭证失败:", err)
}
clobClient.SetAPICredentials(apiCreds)
```

### 2. 余额管理

#### 获取 USDC 余额
```go
usdcBalance, err := clobClient.GetBalanceAllowance(&types.BalanceAllowanceParams{
    AssetType:     types.COLLATERAL,
    SignatureType: signatureType,
})
if err != nil {
    log.Printf("获取 USDC 余额失败: %v", err)
} else {
    // 转换为可读格式 (USDC 有 6 位小数)
    if balance, err := strconv.ParseFloat(usdcBalance.Balance, 64); err == nil {
        usdcAmount := balance / 1000000
        fmt.Printf("USDC 余额: %.6f USDC\n", usdcAmount)
    }
}
```

#### 获取代币余额
```go
tokenBalance, err := clobClient.GetBalanceAllowance(&types.BalanceAllowanceParams{
    AssetType:     types.CONDITIONAL,
    TokenID:       "your_token_id",
    SignatureType: signatureType,
})
```

#### 更新余额信息
```go
updatedBalance, err := clobClient.UpdateBalanceAllowance(&types.BalanceAllowanceParams{
    AssetType:     types.COLLATERAL,
    SignatureType: signatureType,
})
```

### 3. 市场数据获取

#### 获取代币价格信息
```go
// 获取 tick size (价格精度)
tickSize, err := clobClient.GetTickSize(tokenID)
if err != nil {
    log.Printf("获取 tick size 失败: %v", err)
}

// 获取 neg risk 标志
negRisk, err := clobClient.GetNegRisk(tokenID)
if err != nil {
    log.Printf("获取 neg risk 失败: %v", err)
}
```

### 4. 价格查询

#### 获取单个代币价格
```go
// 获取买1价格
buyPrice, err := clobClient.GetPrice(tokenID, types.BUY)
if err != nil {
    log.Printf("获取买1价格失败: %v", err)
} else {
    fmt.Printf("买1价格: %s\n", buyPrice.Price)
    // 转换为浮点数进行计算
    if price, err := strconv.ParseFloat(buyPrice.Price, 64); err == nil {
        fmt.Printf("概率: %.2f%%\n", price*100)
    }
}

// 获取卖1价格
sellPrice, err := clobClient.GetPrice(tokenID, types.SELL)
if err != nil {
    log.Printf("获取卖1价格失败: %v", err)
} else {
    fmt.Printf("卖1价格: %s\n", sellPrice.Price)
}
```

#### 批量获取多个价格
```go
// 准备查询参数
priceParams := []types.BookParams{
    {TokenID: tokenID1, Side: types.BUY},
    {TokenID: tokenID1, Side: types.SELL},
    {TokenID: tokenID2, Side: types.BUY},
    {TokenID: tokenID2, Side: types.SELL},
}

// 批量查询
prices, err := clobClient.GetPrices(priceParams)
if err != nil {
    log.Printf("批量获取价格失败: %v", err)
} else {
    for i, price := range prices {
        param := priceParams[i]
        fmt.Printf("%s %s: %s\n", param.Side, param.TokenID, price.Price)
    }
}
```

#### 价格分析工具
```go
// 计算价差
if buyPrice != nil && sellPrice != nil {
    buyPriceFloat, _ := strconv.ParseFloat(buyPrice.Price, 64)
    sellPriceFloat, _ := strconv.ParseFloat(sellPrice.Price, 64)
    
    spread := sellPriceFloat - buyPriceFloat
    midPrice := (buyPriceFloat + sellPriceFloat) / 2
    spreadPercent := (spread / midPrice) * 100
    
    fmt.Printf("买卖价差: %.4f\n", spread)
    fmt.Printf("中间价格: %.4f\n", midPrice)
    fmt.Printf("价差百分比: %.2f%%\n", spreadPercent)
}
```

### 5. 订单创建和提交

#### 创建限价订单
```go
orderArgs := types.OrderArgs{
    TokenID:    "your_token_id",
    Price:      0.55,           // 价格 (0-1 之间)
    Size:       10.0,           // 数量
    Side:       types.BUY,      // 买入或卖出
    FeeRateBps: 0,              // 手续费率 (基点)
    Nonce:      0,              // 随机数
    Expiration: 0,              // 过期时间 (0 = 不过期)
    Taker:      "0x0000000000000000000000000000000000000000", // 接受者地址
}

// 创建订单选项
options := &types.CreateOrderOptions{
    TickSize: tickSize,
    NegRisk:  negRisk,
}

// 创建并签名订单
signedOrder, err := clobClient.CreateOrder(orderArgs, options)
if err != nil {
    log.Fatal("创建订单失败:", err)
}

// 提交订单
result, err := clobClient.PostOrder(signedOrder, types.GTC)
if err != nil {
    log.Fatal("提交订单失败:", err)
}
```

#### 创建市价订单
```go
marketArgs := types.MarketOrderArgs{
    TokenID:   "your_token_id",
    Amount:    50.0,            // 金额 (买入时) 或数量 (卖出时)
    Side:      types.BUY,
    OrderType: types.FOK,       // Fill or Kill
}

signedOrder, err := clobClient.CreateMarketOrder(marketArgs, options)
```

### 5. 性能监控

```go
// 打印性能指标
clobClient.PrintMetrics()

// 获取指标数据
metrics := clobClient.GetMetrics()
for _, metric := range metrics {
    fmt.Printf("%s: %v\n", metric.Operation, metric.Duration)
}

// 清除指标
clobClient.ClearMetrics()
```

## 📖 完整示例脚本

以下是一个包含所有功能的完整示例脚本，保存为 `complete_example.go`：

```go
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

func main() {
    fmt.Println("🚀 Polymarket CLOB Go SDK - 完整功能演示")
    fmt.Println(strings.Repeat("=", 60))
    
    // 1. 配置检查
    privateKey := os.Getenv("PRIVATE_KEY")
    if privateKey == "" {
        log.Fatal("❌ 请设置 PRIVATE_KEY 环境变量")
    }
    
    host := getEnvOrDefault("POLYMARKET_HOST", "https://clob.polymarket.com")
    chainID := getEnvAsIntOrDefault("CHAIN_ID", 137)
    signatureType := int(getEnvAsIntOrDefault("SIGNATURE_TYPE", 0))
    
    fmt.Printf("📋 配置信息:\n")
    fmt.Printf("   Host: %s\n", host)
    fmt.Printf("   Chain ID: %d\n", chainID)
    fmt.Printf("   Signature Type: %d\n", signatureType)
    
    // 2. 创建客户端
    fmt.Println("\n🔧 初始化客户端...")
    clobClient, err := client.NewClobClient(
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
    
    fmt.Printf("✅ 客户端创建成功\n")
    fmt.Printf("   地址: %s\n", clobClient.GetAddress())
    
    // 3. 设置 API 凭证
    fmt.Println("\n🔑 设置 API 凭证...")
    apiCreds, err := clobClient.CreateOrDeriveAPIKey(0)
    if err != nil {
        log.Fatalf("❌ 获取 API 凭证失败: %v", err)
    }
    clobClient.SetAPICredentials(apiCreds)
    fmt.Printf("✅ API 凭证设置完成\n")
    
    // 4. 获取余额信息
    fmt.Println("\n💰 检查账户余额...")
    
    // USDC 余额
    usdcBalance, err := clobClient.GetBalanceAllowance(&types.BalanceAllowanceParams{
        AssetType:     types.COLLATERAL,
        SignatureType: signatureType,
    })
    
    var usdcAmount float64
    var hasBalance bool
    
    if err != nil {
        fmt.Printf("❌ 获取 USDC 余额失败: %v\n", err)
    } else {
        fmt.Printf("📊 USDC 原始余额: %s\n", usdcBalance.Balance)
        if usdcBalance.Balance != "" && usdcBalance.Balance != "0" {
            if balance, err := strconv.ParseFloat(usdcBalance.Balance, 64); err == nil {
                usdcAmount = balance / 1000000 // USDC 有 6 位小数
                hasBalance = usdcAmount > 0
                fmt.Printf("💵 USDC 余额: %.6f USDC\n", usdcAmount)
            }
        } else {
            fmt.Printf("💵 USDC 余额: 0.000000 USDC\n")
        }
    }
    
    // 5. 更新余额
    fmt.Println("\n🔄 更新余额信息...")
    updatedBalance, err := clobClient.UpdateBalanceAllowance(&types.BalanceAllowanceParams{
        AssetType:     types.COLLATERAL,
        SignatureType: signatureType,
    })
    if err != nil {
        fmt.Printf("⚠️  更新余额警告: %v\n", err)
    } else {
        fmt.Printf("✅ 余额更新完成\n")
    }
    
    // 6. 获取市场数据 (使用示例代币 ID)
    tokenID := getEnvOrDefault("TOKEN_ID", "91094360697357622623953793720402150934374522251651348543981406747516093190659")
    fmt.Printf("\n📊 获取市场数据 (Token: %s...)...\n", tokenID[:20])
    
    // 获取 tick size
    tickSize, err := clobClient.GetTickSize(tokenID)
    if err != nil {
        fmt.Printf("⚠️  获取 tick size 失败: %v\n", err)
        tickSize = types.TickSize001 // 使用默认值
    } else {
        fmt.Printf("📏 Tick Size: %s\n", tickSize)
    }
    
    // 获取 neg risk
    negRisk, err := clobClient.GetNegRisk(tokenID)
    if err != nil {
        fmt.Printf("⚠️  获取 neg risk 失败: %v\n", err)
        negRisk = false // 使用默认值
    } else {
        fmt.Printf("⚠️  Neg Risk: %t\n", negRisk)
    }
    
    // 7. 创建示例订单 (不提交)
    fmt.Println("\n📝 创建示例订单...")
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
    
    signedOrder, err := clobClient.CreateOrder(orderArgs, options)
    if err != nil {
        fmt.Printf("❌ 创建订单失败: %v\n", err)
    } else {
        fmt.Printf("✅ 订单创建成功\n")
        fmt.Printf("📋 订单详情:\n")
        fmt.Printf("   价格: %.2f (%.0f%%)\n", orderArgs.Price, orderArgs.Price*100)
        fmt.Printf("   数量: %.1f\n", orderArgs.Size)
        fmt.Printf("   方向: %s\n", orderArgs.Side)
        fmt.Printf("   总价值: $%.2f\n", orderArgs.Price*orderArgs.Size)
        fmt.Printf("   Maker Amount: %s\n", signedOrder.MakerAmount)
        fmt.Printf("   Taker Amount: %s\n", signedOrder.TakerAmount)
        
        // 如果有余额，可以选择提交订单
        if hasBalance {
            fmt.Println("\n❓ 是否要提交此订单? (输入 'yes' 确认)")
            var response string
            fmt.Scanln(&response)
            
            if strings.ToLower(response) == "yes" {
                fmt.Println("📤 提交订单...")
                result, err := clobClient.PostOrder(signedOrder, types.GTC)
                if err != nil {
                    fmt.Printf("❌ 订单提交失败: %v\n", err)
                } else {
                    fmt.Printf("🎉 订单提交成功!\n")
                    fmt.Printf("📋 结果: %+v\n", result)
                }
            } else {
                fmt.Println("⏭️  跳过订单提交")
            }
        } else {
            fmt.Println("\n💡 无法提交订单 - 需要 USDC 余额")
            fmt.Println("   请访问 https://polymarket.com 充值")
        }
    }
    
    // 8. 性能总结
    fmt.Println("\n📈 性能指标:")
    clobClient.PrintMetrics()
    
    fmt.Println("\n✅ 演示完成!")
    fmt.Println("\n💡 提示:")
    fmt.Println("   - 设置 TOKEN_ID 环境变量来交易不同市场")
    fmt.Println("   - 确保有足够的 USDC 余额进行交易")
    fmt.Println("   - 查看文档了解更多功能")
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
```

## 🔧 常用配置

### 订单类型
- `types.GTC` - Good Till Cancelled (一直有效直到取消)
- `types.FOK` - Fill Or Kill (全部成交或取消)
- `types.IOC` - Immediate Or Cancel (立即成交或取消)

### 订单方向
- `types.BUY` - 买入
- `types.SELL` - 卖出

### 签名类型
- `0` - EOA (外部拥有账户) - 推荐
- `1` - POLY_PROXY (Polymarket 代理)

## ⚠️ 注意事项

1. **私钥安全**: 永远不要在代码中硬编码私钥
2. **最小订单**: Polymarket 有最小订单金额要求 (通常 $1)
3. **余额要求**: 交易前确保有足够的 USDC 余额
4. **网络连接**: 确保网络连接稳定
5. **价格范围**: 价格必须在 0-1 之间

## 🐛 错误处理

```go
result, err := clobClient.PostOrder(signedOrder, types.GTC)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "not enough balance"):
        fmt.Println("余额不足，请充值 USDC")
    case strings.Contains(err.Error(), "invalid amount"):
        fmt.Println("订单金额不符合要求")
    case strings.Contains(err.Error(), "insufficient auth level"):
        fmt.Println("认证级别不足")
    default:
        fmt.Printf("订单失败: %v\n", err)
    }
}
```

## 📞 获取帮助

1. 查看示例代码
2. 检查环境变量配置
3. 确认网络连接
4. 验证私钥格式
5. 检查余额是否充足

这份指南涵盖了 SDK 的所有主要功能，用户只需要提供私钥就可以开始使用所有功能。