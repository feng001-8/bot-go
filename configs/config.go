package configs

type Config struct {
	// 基础配置
	MNEMONIC         string // 12 个 注记词
	CONTRACT_ADDRESS string // 合约地址
	TOKEN_ADDRESS    string //代币地址（用于approve）
	Input_DATA       string // 字节码（不含0x）

	// 参数处理
	AUTO_ZERO_MIN_AMOUNT   string // ?
	MIN_AMOUNT_PARAM_INDEX int    // ?

	// gas配置
	INITIAL_GAS_PRICE_GWEI string
	INITIAL_GAS_LIMIT      uint64
}
