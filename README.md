# Polymarket CLOB Go SDK

A high-performance Go SDK for interacting with the Polymarket CLOB (Central Limit Order Book) API. This SDK provides comprehensive functionality for creating, signing, and submitting maker orders with detailed performance metrics tracking.

## Features

- **Multi-level Authentication**: Support for L0 (public), L1 (private key), and L2 (API credentials) authentication
- **Order Management**: Create and sign limit orders and market orders
- **EIP712 Signing**: Proper EIP712 signature implementation for order authentication
- **HMAC Authentication**: Secure API request signing for Level 2 operations
- **Performance Metrics**: Detailed timing metrics for all operations
- **Caching**: Intelligent caching of tick sizes and neg risk flags
- **Error Handling**: Comprehensive error handling and validation
- **Type Safety**: Strong typing throughout the SDK

## 🚀 快速开始

### 第一步：安装依赖
```bash
# 克隆项目
git clone https://github.com/your-repo/polymarket-clob-go
cd polymarket-clob-go

# 安装依赖
make deps
```

### 第二步：配置环境
```bash
# 运行配置助手
make config

# 或者手动设置环境变量
export PRIVATE_KEY="your_ethereum_private_key_here"
```

### 第三步：运行演示
```bash
# 运行完整功能演示 (推荐)
make demo

# 或者运行其他示例
make balance    # 余额管理
make run-example # 完整工作流
```

### 基本使用代码

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "polymarket-clob-go/pkg/client"
    "polymarket-clob-go/pkg/types"
)

func main() {
    // 从环境变量读取私钥
    privateKey := os.Getenv("PRIVATE_KEY")
    
    // 创建客户端
    clobClient, err := client.NewClobClient(
        "https://clob.polymarket.com",
        137, // Polygon 主网
        privateKey,
        nil, nil, nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // 设置 API 凭证
    creds, err := clobClient.CreateOrDeriveAPIKey(0)
    if err != nil {
        log.Fatal(err)
    }
    clobClient.SetAPICredentials(creds)
    
    // 检查余额
    balance, err := clobClient.GetBalanceAllowance(&types.BalanceAllowanceParams{
        AssetType:     types.COLLATERAL,
        SignatureType: 0,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("USDC 余额: %s\n", balance.Balance)
    
    // 创建订单
    orderArgs := types.OrderArgs{
        TokenID:    "your_token_id",
        Price:      0.55,
        Size:       10.0,
        Side:       types.BUY,
        FeeRateBps: 0,
        Nonce:      0,
        Expiration: 0,
        Taker:      "0x0000000000000000000000000000000000000000",
    }
    
    signedOrder, err := clobClient.CreateOrder(orderArgs, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // 提交订单
    result, err := clobClient.PostOrder(signedOrder, types.GTC)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("订单提交成功: %+v\n", result)
}
```

### 📚 详细文档
- [完整使用指南](GO_SDK_USAGE_GUIDE.md) - 详细的功能说明和示例
- [示例代码](examples/) - 各种使用场景的完整示例

## Architecture

### Core Components

1. **Client** (`pkg/client`): Main client interface with all API operations
2. **Signer** (`pkg/signer`): Cryptographic operations and EIP712 signing
3. **Auth** (`pkg/auth`): Authentication header generation (L1 and L2)
4. **OrderBuilder** (`pkg/orderbuilder`): Order creation and signing logic
5. **Types** (`pkg/types`): Type definitions and data structures
6. **Utils** (`pkg/utils`): Utility functions for crypto and calculations

### Authentication Levels

- **Level 0**: Public endpoints only (market data, server time)
- **Level 1**: Private key authentication (API key creation/derivation)
- **Level 2**: Full authentication (trading operations)

### Order Flow

1. **Client Creation**: Initialize with private key and chain ID
2. **API Key Generation**: Create or derive API credentials using EIP712 signing
3. **Market Data**: Fetch tick size and neg risk information
4. **Order Creation**: Build order data with proper amount calculations
5. **Order Signing**: Sign order using EIP712 with exchange contract
6. **Order Submission**: Submit signed order with HMAC authentication

## Performance Metrics

The SDK tracks detailed performance metrics for all operations:

```go
// Get all metrics
metrics := client.GetMetrics()

// Print formatted metrics
client.PrintMetrics()

// Clear metrics
client.ClearMetrics()
```

Metrics include:
- Operation name and duration
- Success/failure status
- Error messages
- Start timestamps

## Examples

### Complete Workflow

```bash
# Set your private key
export PRIVATE_KEY="your_ethereum_private_key"

# Run complete workflow example
make run-example
```

### Simple Order Creation

```bash
# Run simple example
make run-simple
```

### Using Environment File

```bash
# Create environment template
make env-template

# Edit .env file with your credentials
cp .env.template .env
# Edit .env file...

# Run with environment file
make run-with-env
```

## API Reference

### Client Methods

#### Authentication
- `CreateAPIKey(nonce int64) (*types.ApiCreds, error)`
- `DeriveAPIKey(nonce int64) (*types.ApiCreds, error)`
- `CreateOrDeriveAPIKey(nonce int64) (*types.ApiCreds, error)`
- `SetAPICredentials(creds *types.ApiCreds)`

#### Market Data
- `GetTickSize(tokenID string) (types.TickSize, error)`
- `GetNegRisk(tokenID string) (bool, error)`

#### Order Operations
- `CreateOrder(orderArgs types.OrderArgs, options *types.CreateOrderOptions) (*types.SignedOrder, error)`
- `CreateMarketOrder(orderArgs types.MarketOrderArgs, options *types.CreateOrderOptions) (*types.SignedOrder, error)`
- `PostOrder(signedOrder *types.SignedOrder, orderType types.OrderType) (map[string]interface{}, error)`
- `CreateAndPostOrder(orderArgs types.OrderArgs, options *types.CreateOrderOptions) (map[string]interface{}, error)`

#### Metrics
- `GetMetrics() []types.PerformanceMetrics`
- `PrintMetrics()`
- `ClearMetrics()`

### Order Types

```go
// Limit Order
orderArgs := types.OrderArgs{
    TokenID:    "token_id",
    Price:      0.55,           // Price per share
    Size:       10.0,           // Number of shares
    Side:       types.BUY,      // BUY or SELL
    FeeRateBps: 0,              // Fee rate in basis points
    Nonce:      time.Now().Unix(),
    Expiration: time.Now().Add(24 * time.Hour).Unix(),
    Taker:      "0x0000000000000000000000000000000000000000", // Zero address for public orders
}

// Market Order
marketArgs := types.MarketOrderArgs{
    TokenID:   "token_id",
    Amount:    50.0,            // Dollar amount (for BUY) or shares (for SELL)
    Side:      types.BUY,
    OrderType: types.FOK,       // Fill or Kill
}
```

## Development

### Setup

```bash
# Initialize project
make init

# Install dependencies
make deps

# Install development tools
make dev-tools
```

### Building

```bash
# Build for current platform
make build

# Build for Linux
make build-linux

# Format code
make fmt

# Vet code
make vet

# Run tests
make test

# Run all checks
make all
```

### Testing

```bash
# Run tests
make test

# Run benchmarks
make benchmark

# Security check
make security
```

## Configuration

### Environment Variables

- `PRIVATE_KEY`: Your Ethereum private key (required)
- `CLOB_API_URL`: CLOB API URL (default: https://clob.polymarket.com)
- `CLOB_API_KEY`: Existing API key (optional)
- `CLOB_SECRET`: Existing API secret (optional)
- `CLOB_PASS_PHRASE`: Existing API passphrase (optional)

### Supported Networks

- **Polygon Mainnet** (Chain ID: 137)
- **Amoy Testnet** (Chain ID: 80002)

## Error Handling

The SDK provides comprehensive error handling:

```go
signedOrder, err := client.CreateOrder(orderArgs, nil)
if err != nil {
    // Handle specific error types
    switch {
    case strings.Contains(err.Error(), "insufficient auth level"):
        // Handle authentication error
    case strings.Contains(err.Error(), "invalid price"):
        // Handle price validation error
    default:
        // Handle other errors
    }
}
```

## Performance Optimization

- **Caching**: Tick sizes and neg risk flags are cached
- **Connection Pooling**: HTTP client with proper timeouts
- **Batch Operations**: Support for multiple order operations
- **Metrics Tracking**: Identify bottlenecks with detailed timing

## Security

- Private keys are never transmitted over the network
- All API requests use HMAC-SHA256 signatures
- EIP712 signing for order authentication
- Proper nonce handling to prevent replay attacks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make all` to ensure everything passes
6. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
1. Check the examples in the `examples/` directory
2. Review the API documentation
3. Open an issue on GitHub

## Changelog

### v1.0.0
- Initial release
- Complete order creation and signing
- Multi-level authentication
- Performance metrics tracking
- Comprehensive error handling