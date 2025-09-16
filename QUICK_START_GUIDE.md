# Polymarket CLOB Go SDK 快速开始指南

## 📋 概述

这是一个完整的 Go SDK，用于与 Polymarket 中央限价订单簿 (CLOB) 进行交互。支持账户管理、余额查询、市场数据获取、订单创建和提交等全部功能。

## 🚀 三步快速开始

### 第一步：准备环境
```bash
# 1. 确保安装了 Go 1.19+
go version

# 2. 克隆项目 (如果还没有)
git clone https://github.com/your-repo/polymarket-clob-go
cd polymarket-clob-go

# 3. 安装依赖
make deps
```

### 第二步：配置私钥
```bash
# 运行配置助手 (推荐)
make config

# 或者直接设置环境变量
export PRIVATE_KEY="your_ethereum_private_key_here"
```

**重要**: 你需要提供一个以太坊私钥 (64位十六进制字符串，不含 0x 前缀)

### 第三步：运行演示
```bash
# 运行完整功能演示 (推荐首次使用)
make demo

# 或者运行其他示例
make balance      # 余额管理示例
make run-example  # 完整工作流示例
```

## 📚 用户需要提供的信息

### 必需信息
1. **以太坊私钥** - 你的钱包私钥
   - 格式：64位十六进制字符串
   - 示例：`abcd1234efgh5678...` (64个字符)
   - 不要包含 `0x` 前缀

### 可选配置 (有默认值)
2. **API 主机** - 默认：`https://clob.polymarket.com`
3. **链 ID** - 默认：`137` (Polygon 主网)
4. **签名类型** - 默认：`0` (EOA 签名，推荐)
5. **代币 ID** - 用于测试的市场代币 ID

## 🔧 配置方式

### 方式一：使用配置助手 (推荐)
```bash
make config
```
- 交互式配置
- 自动生成 .env 文件
- 包含安全提示

### 方式二：手动设置环境变量
```bash
export PRIVATE_KEY="your_private_key_here"
export POLYMARKET_HOST="https://clob.polymarket.com"
export CHAIN_ID="137"
export SIGNATURE_TYPE="0"
```

### 方式三：创建 .env 文件
```bash
# 创建 .env 文件
cat > .env << EOF
PRIVATE_KEY=your_private_key_here
POLYMARKET_HOST=https://clob.polymarket.com
CHAIN_ID=137
SIGNATURE_TYPE=0
TOKEN_ID=91094360697357622623953793720402150934374522251651348543981406747516093190659
EOF
```

## 📊 主要功能演示

### 1. 完整功能演示 (推荐)
```bash
make demo
```
包含所有功能的完整演示：
- ✅ 客户端初始化
- ✅ API 认证设置
- ✅ 余额查询和更新
- ✅ 市场数据获取
- ✅ 订单创建和提交
- ✅ 性能监控

### 2. 余额管理示例
```bash
make balance
```
专注于余额相关功能：
- USDC 余额查询
- 代币余额查询
- 余额更新
- 交易准备检查

### 3. 其他示例
```bash
make run-example     # 完整工作流
make run-simple      # 简单订单
make run-performance # 性能分析
```

## 💰 交易准备

### 充值 USDC
1. 访问 https://polymarket.com
2. 连接你的钱包 (使用相同的私钥)
3. 存入 USDC 到账户
4. 重新运行演示脚本

### 最小交易要求
- 最小订单金额：通常 $1
- 价格范围：0-1 之间 (代表概率)
- 需要足够的 USDC 余额

## 🔍 代码示例

### 基本使用
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
    // 1. 创建客户端
    privateKey := os.Getenv("PRIVATE_KEY")
    clobClient, err := client.NewClobClient(
        "https://clob.polymarket.com",
        137,
        privateKey,
        nil, nil, nil,
    )
    if err != nil {
        log.Fatal(err)
    }

    // 2. 设置 API 凭证
    creds, err := clobClient.CreateOrDeriveAPIKey(0)
    if err != nil {
        log.Fatal(err)
    }
    clobClient.SetAPICredentials(creds)

    // 3. 检查余额
    balance, err := clobClient.GetBalanceAllowance(&types.BalanceAllowanceParams{
        AssetType:     types.COLLATERAL,
        SignatureType: 0,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("USDC 余额: %s\n", balance.Balance)

    // 4. 创建订单
    orderArgs := types.OrderArgs{
        TokenID:    "your_token_id",
        Price:      0.55,  // 55% 概率
        Size:       2.0,   // 2 个单位
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

    // 5. 提交订单 (可选)
    result, err := clobClient.PostOrder(signedOrder, types.GTC)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("订单提交成功: %+v\n", result)
}
```

## 🛠️ 可用命令

### 快速开始
```bash
make deps     # 安装依赖
make config   # 配置环境
make demo     # 运行演示
```

### 示例命令
```bash
make demo         # 完整功能演示
make balance      # 余额管理
make run-example  # 完整工作流
make run-simple   # 简单订单
```

### 开发命令
```bash
make test     # 运行测试
make fmt      # 格式化代码
make build    # 构建项目
make clean    # 清理文件
```

## ⚠️ 注意事项

### 安全提醒
- 永远不要在代码中硬编码私钥
- 不要将 .env 文件提交到版本控制
- 定期更换私钥
- 在生产环境使用更安全的密钥管理

### 交易提醒
- 确保有足够的 USDC 余额
- 注意最小订单金额要求
- 价格必须在 0-1 之间
- 网络连接要稳定

### 错误处理
常见错误和解决方案：
- `not enough balance` - 余额不足，需要充值
- `invalid amount` - 订单金额不符合要求
- `insufficient auth level` - 认证级别不足

## 📞 获取帮助

1. **查看详细文档**: [GO_SDK_USAGE_GUIDE.md](GO_SDK_USAGE_GUIDE.md)
2. **运行示例代码**: `examples/` 目录中的各种示例
3. **检查配置**: 确认环境变量设置正确
4. **验证余额**: 确保钱包中有足够的 USDC

## 🎯 下一步

1. 运行 `make demo` 了解所有功能
2. 查看 `GO_SDK_USAGE_GUIDE.md` 了解详细用法
3. 修改示例代码满足你的需求
4. 开始构建你的交易应用

---

**开始使用**: `make config && make demo`