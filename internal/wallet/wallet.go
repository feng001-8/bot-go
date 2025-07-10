package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// Wallet 表示一个钱包实例
type Wallet struct {
	Index      int               `json:"index"`
	Address    common.Address    `json:"address"`
	PrivateKey *ecdsa.PrivateKey `json:"-"` // 不序列化私钥
	Account    accounts.Account  `json:"account"`
}

// WalletClient 表示带有以太坊客户端的钱包
type WalletClient struct {
	*Wallet
	Client *ethclient.Client `json:"-"`
}

// GenerateWallets 从助记词生成指定数量的钱包
// mnemonic: BIP39助记词
// count: 要生成的钱包数量，默认50个
func GenerateWallets(mnemonic string, count int) ([]*Wallet, error) {
	if count <= 0 {
		count = 50
	}

	// 验证助记词
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic phrase")
	}

	// 从助记词生成种子
	seed := bip39.NewSeed(mnemonic, "")

	// 创建主密钥
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %v", err)
	}

	wallets := make([]*Wallet, 0, count)

	for i := 0; i < count; i++ {
		// 使用BIP44路径: m/44'/60'/0'/0/i (以太坊标准路径)
		// 44' = 0x8000002C (BIP44)
		// 60' = 0x8000003C (以太坊)
		// 0'  = 0x80000000 (账户)
		// 0   = 0x00000000 (外部链)
		// i   = 地址索引
		path := []uint32{
			0x8000002C, // 44'
			0x8000003C, // 60'
			0x80000000, // 0'
			0x00000000, // 0
			uint32(i),  // address_index
		}

		// 派生子密钥
		childKey := masterKey
		for _, pathElement := range path {
			childKey, err = childKey.NewChildKey(pathElement)
			if err != nil {
				return nil, fmt.Errorf("failed to derive child key at index %d: %v", i, err)
			}
		}

		// 获取私钥
		privateKeyECDSA, err := crypto.ToECDSA(childKey.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to ECDSA private key at index %d: %v", i, err)
		}

		// 获取公钥和地址
		publicKey := privateKeyECDSA.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("failed to cast public key to ECDSA at index %d", i)
		}

		address := crypto.PubkeyToAddress(*publicKeyECDSA)

		// 创建钱包实例
		wallet := &Wallet{
			Index:      i,
			Address:    address,
			PrivateKey: privateKeyECDSA,
			Account: accounts.Account{
				Address: address,
			},
		}

		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

// CreateWalletClients 为钱包创建以太坊客户端
// wallets: 钱包列表
// rpcURL: RPC节点URL
func CreateWalletClients(wallets []*Wallet, rpcURL string) ([]*WalletClient, error) {
	if len(wallets) == 0 {
		return nil, fmt.Errorf("no wallets provided")
	}

	if rpcURL == "" {
		return nil, fmt.Errorf("RPC URL cannot be empty")
	}

	// 创建以太坊客户端
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum client: %v", err)
	}

	walletClients := make([]*WalletClient, 0, len(wallets))

	for _, wallet := range wallets {
		walletClient := &WalletClient{
			Wallet: wallet,
			Client: client,
		}
		walletClients = append(walletClients, walletClient)
	}

	return walletClients, nil
}

// GetPrivateKeyHex 获取钱包的私钥十六进制字符串
func (w *Wallet) GetPrivateKeyHex() string {
	return fmt.Sprintf("%x", crypto.FromECDSA(w.PrivateKey))
}

// GetAddressHex 获取钱包地址的十六进制字符串
func (w *Wallet) GetAddressHex() string {
	return w.Address.Hex()
}

// GetChainID 获取链ID
func (wc *WalletClient) GetChainID() (*big.Int, error) {
	// 创建带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return wc.Client.ChainID(ctx)
}

// Close 关闭客户端连接
func (wc *WalletClient) Close() {
	if wc.Client != nil {
		wc.Client.Close()
	}
}

// 实现助记词的生成
func GenerateMnemonic() (string, error) {
	// 生成12个随机助记词
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}

	// 从熵生成助记词
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}
