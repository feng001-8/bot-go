package main

import (
	"fmt"
	"log"

	"bot-go/internal/wallet"
)

func main() {
	// 1. 生成新的助记词
	fmt.Println("=== 生成助记词 ===")
	mnemonic, err := wallet.GenerateMnemonic()
	if err != nil {
		log.Fatalf("生成助记词失败: %v", err)
	}
	fmt.Printf("助记词: %s\n\n", mnemonic)

	// 2. 从助记词生成钱包
	fmt.Println("=== 生成钱包 ===")
	wallets, err := wallet.GenerateWallets(mnemonic, 5) // 生成5个钱包作为示例
	if err != nil {
		log.Fatalf("生成钱包失败: %v", err)
	}

	// 3. 显示钱包信息
	for i, w := range wallets {
		fmt.Printf("钱包 %d:\n", i+1)
		fmt.Printf("  地址: %s\n", w.GetAddressHex())
		fmt.Printf("  私钥: %s\n", w.GetPrivateKeyHex())
		fmt.Println()
	}

	// 4. 创建带有以太坊客户端的钱包（可选）
	fmt.Println("=== 创建钱包客户端 ===")
	rpcURL := "https://eth.llamarpc.com" // LlamaNodes (免费)

	walletClients, err := wallet.CreateWalletClients(wallets, rpcURL)
	if err != nil {
		fmt.Printf("创建钱包客户端失败: %v\n", err)
		fmt.Println("跳过客户端创建，继续执行...")
	} else {
		fmt.Printf("成功创建 %d 个钱包客户端\n", len(walletClients))

		// 获取链ID（如果连接成功）
		if len(walletClients) > 0 {
			chainID, err := walletClients[0].GetChainID()
			if err != nil {
				fmt.Printf("获取链ID失败: %v\n", err)
			} else {
				fmt.Printf("链ID: %s\n", chainID.String())
			}
		}

		// 关闭客户端连接
		for _, wc := range walletClients {
			wc.Close()
		}
	}

	fmt.Println("\n=== 程序执行完成 ===")
}
