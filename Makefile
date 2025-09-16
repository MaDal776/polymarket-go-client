# Polymarket CLOB Go SDK Makefile

.PHONY: build test clean run-example run-simple deps fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build targets
BINARY_NAME=polymarket-clob
BINARY_UNIX=$(BINARY_NAME)_unix

# Default target
all: deps fmt vet test build

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOFMT) -s -w .

# Vet code
vet:
	$(GOVET) ./...

# Run tests
test:
	$(GOTEST) -v ./...

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./examples/complete_workflow.go

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./examples/complete_workflow.go

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run the complete workflow example
run-example:
	@echo "Running complete workflow example..."
	@echo "Make sure to set PRIVATE_KEY environment variable"
	$(GOCMD) run examples/complete_workflow.go

# Run the simple order example
run-simple:
	@echo "Running simple order example..."
	@echo "Make sure to set PRIVATE_KEY environment variable"
	$(GOCMD) run examples/simple_order.go

# Run performance analysis
run-performance:
	@echo "Running performance analysis..."
	@echo "Make sure to set PRIVATE_KEY environment variable"
	$(GOCMD) run examples/performance_analysis.go

# Run complete order demo (create + post)
run-demo:
	@echo "Running complete order demo (create + post)..."
	@echo "Make sure to set PRIVATE_KEY environment variable"
	$(GOCMD) run examples/complete_order_demo.go

# Run complete SDK demo (recommended)
demo:
	@echo "🚀 Running complete SDK demo..."
	@echo "Make sure to set PRIVATE_KEY environment variable"
	$(GOCMD) run examples/complete_sdk_demo.go

# Run balance management example
balance:
	@echo "💰 Running balance management example..."
	@echo "Make sure to set PRIVATE_KEY environment variable"
	$(GOCMD) run examples/balance_management.go

# Run configuration helper
config:
	@echo "🔧 Running configuration helper..."
	$(GOCMD) run examples/setup_config.go

# Run price query demo
price:
	@echo "💰 Running price query demo..."
	@echo "Note: Price queries are public and don't require PRIVATE_KEY"
	$(GOCMD) run examples/get_price_demo.go

# Compare Go SDK vs Python SDK prices
compare-prices:
	@echo "🔍 Comparing Go SDK vs Python SDK prices..."
	$(GOCMD) run examples/compare_python_go_prices.go

# Run with environment file
run-with-env:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && $(GOCMD) run examples/complete_workflow.go; \
	else \
		echo "No .env file found. Please create one with PRIVATE_KEY=your_key"; \
	fi

# Initialize project (run once)
init:
	$(GOMOD) init polymarket-clob-go
	$(GOGET) github.com/ethereum/go-ethereum@v1.13.5
	$(GOGET) github.com/shopspring/decimal@v1.3.1
	$(GOGET) github.com/stretchr/testify@v1.8.4

# Create .env template
env-template:
	@echo "Creating .env template..."
	@echo "# Polymarket CLOB Go SDK Environment Variables" > .env.template
	@echo "PRIVATE_KEY=your_ethereum_private_key_here" >> .env.template
	@echo "CLOB_API_URL=https://clob.polymarket.com" >> .env.template
	@echo "" >> .env.template
	@echo "# Optional: Existing API credentials (if you have them)" >> .env.template
	@echo "# CLOB_API_KEY=your_api_key" >> .env.template
	@echo "# CLOB_SECRET=your_api_secret" >> .env.template
	@echo "# CLOB_PASS_PHRASE=your_passphrase" >> .env.template
	@echo ".env.template created. Copy to .env and fill in your values."

# Benchmark performance
benchmark:
	@echo "Running performance benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all ./pkg/client
	$(GOCMD) doc -all ./pkg/types
	$(GOCMD) doc -all ./pkg/signer

# Check for security issues (requires gosec)
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Install development tools
dev-tools:
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) golang.org/x/tools/cmd/goimports@latest

# Help
help:
	@echo "🚀 Polymarket CLOB Go SDK"
	@echo "=========================="
	@echo ""
	@echo "📋 快速开始:"
	@echo "  make deps     - 安装依赖"
	@echo "  make config   - 配置环境变量"
	@echo "  make demo     - 运行完整演示 (推荐)"
	@echo ""
	@echo "📊 示例命令:"
	@echo "  demo         - 完整 SDK 功能演示"
	@echo "  balance      - 余额管理示例"
	@echo "  price        - 价格查询演示 (无需私钥)"
	@echo "  compare-prices - 对比 Go SDK vs Python SDK 价格"
	@echo "  run-example  - 完整工作流示例"
	@echo "  run-simple   - 简单订单示例"
	@echo "  run-performance - 性能分析"
	@echo ""
	@echo "🔧 开发命令:"
	@echo "  all          - 运行 deps, fmt, vet, test, build"
	@echo "  deps         - 下载和整理依赖"
	@echo "  fmt          - 格式化代码"
	@echo "  vet          - 代码检查"
	@echo "  test         - 运行测试"
	@echo "  build        - 构建二进制文件"
	@echo "  clean        - 清理构建文件"
	@echo ""
	@echo "⚙️  配置命令:"
	@echo "  config       - 运行配置助手"
	@echo "  env-template - 创建 .env 模板文件"
	@echo "  run-with-env - 使用 .env 文件运行"
	@echo ""
	@echo "📈 分析命令:"
	@echo "  benchmark    - 运行性能基准测试"
	@echo "  docs         - 生成文档"
	@echo "  security     - 安全检查"
	@echo "  dev-tools    - 安装开发工具"