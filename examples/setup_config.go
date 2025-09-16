package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("🔧 Polymarket CLOB Go SDK 配置助手")
	fmt.Println(strings.Repeat("=", 50))

	reader := bufio.NewReader(os.Stdin)

	// 检查现有配置
	fmt.Println("\n📋 检查现有环境变量:")
	checkEnvVar("PRIVATE_KEY", "以太坊私钥", true)
	checkEnvVar("POLYMARKET_HOST", "API 主机地址", false)
	checkEnvVar("CHAIN_ID", "区块链网络 ID", false)
	checkEnvVar("SIGNATURE_TYPE", "签名类型", false)
	checkEnvVar("TOKEN_ID", "测试代币 ID", false)

	fmt.Println("\n❓ 是否需要创建新的 .env 配置文件? (y/N)")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("✅ 配置检查完成")
		return
	}

	// 创建 .env 文件
	fmt.Println("\n📝 创建 .env 配置文件...")
	envContent := []string{
		"# Polymarket CLOB Go SDK 配置文件",
		"# 由配置助手自动生成于 " + fmt.Sprintf("%v", os.Getenv("USER")),
		"",
		"# ===== 必需配置 =====",
	}

	// 私钥输入
	fmt.Print("\n🔑 请输入你的以太坊私钥 (不包含 0x 前缀): ")
	privateKey, _ := reader.ReadString('\n')
	privateKey = strings.TrimSpace(privateKey)

	if privateKey == "" {
		fmt.Println("❌ 私钥不能为空")
		return
	}

	// 验证私钥格式
	if len(privateKey) != 64 {
		fmt.Printf("⚠️  警告: 私钥长度应为 64 个字符，当前长度: %d\n", len(privateKey))
	}

	envContent = append(envContent, fmt.Sprintf("PRIVATE_KEY=%s", privateKey))
	envContent = append(envContent, "", "# ===== 可选配置 =====")

	// API 主机
	fmt.Print("🌐 API 主机地址 (默认: https://clob.polymarket.com): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "https://clob.polymarket.com"
	}
	envContent = append(envContent, fmt.Sprintf("POLYMARKET_HOST=%s", host))

	// 链 ID
	fmt.Print("⛓️  区块链网络 ID (137=Polygon主网, 80002=Amoy测试网, 默认: 137): ")
	chainID, _ := reader.ReadString('\n')
	chainID = strings.TrimSpace(chainID)
	if chainID == "" {
		chainID = "137"
	}
	envContent = append(envContent, fmt.Sprintf("CHAIN_ID=%s", chainID))

	// 签名类型
	fmt.Print("✍️  签名类型 (0=EOA推荐, 1=POLY_PROXY, 默认: 0): ")
	sigType, _ := reader.ReadString('\n')
	sigType = strings.TrimSpace(sigType)
	if sigType == "" {
		sigType = "0"
	}
	envContent = append(envContent, fmt.Sprintf("SIGNATURE_TYPE=%s", sigType))

	// 代币 ID
	fmt.Print("🎯 测试代币 ID (可选，用于演示): ")
	tokenID, _ := reader.ReadString('\n')
	tokenID = strings.TrimSpace(tokenID)
	if tokenID == "" {
		tokenID = "91094360697357622623953793720402150934374522251651348543981406747516093190659"
	}
	envContent = append(envContent, fmt.Sprintf("TOKEN_ID=%s", tokenID))

	// 添加安全提示
	envContent = append(envContent, "")
	envContent = append(envContent, "# ===== 安全提示 =====")
	envContent = append(envContent, "# 1. 不要将此文件提交到版本控制系统")
	envContent = append(envContent, "# 2. 定期更换私钥")
	envContent = append(envContent, "# 3. 在生产环境使用更安全的密钥管理")

	// 写入文件
	file, err := os.Create(".env")
	if err != nil {
		fmt.Printf("❌ 创建 .env 文件失败: %v\n", err)
		return
	}
	defer file.Close()

	for _, line := range envContent {
		file.WriteString(line + "\n")
	}

	// 设置文件权限 (仅当前用户可读写)
	os.Chmod(".env", 0600)

	fmt.Println("\n✅ .env 文件创建成功!")
	fmt.Println("📁 文件位置: .env")
	fmt.Println("🔒 文件权限已设置为仅当前用户可读写")

	fmt.Println("\n🚀 下一步:")
	fmt.Println("   1. 运行完整演示: go run examples/complete_sdk_demo.go")
	fmt.Println("   2. 或运行余额管理: go run examples/balance_management.go")
	fmt.Println("   3. 查看使用指南: cat GO_SDK_USAGE_GUIDE.md")

	fmt.Println("\n⚠️  重要提醒:")
	fmt.Println("   - 确保钱包中有 USDC 余额才能进行交易")
	fmt.Println("   - 访问 https://polymarket.com 进行充值")
	fmt.Println("   - 首次使用建议先运行演示了解功能")
}

func checkEnvVar(key, description string, required bool) {
	value := os.Getenv(key)
	status := "❌ 未设置"

	if value != "" {
		if key == "PRIVATE_KEY" {
			status = "✅ 已设置 (***隐藏***)"
		} else {
			// 对于长字符串，只显示前面部分
			if len(value) > 30 {
				status = fmt.Sprintf("✅ 已设置: %s...", value[:30])
			} else {
				status = fmt.Sprintf("✅ 已设置: %s", value)
			}
		}
	}

	requiredText := ""
	if required {
		requiredText = " (必需)"
	}

	fmt.Printf("   %s%s: %s\n", description, requiredText, status)
}