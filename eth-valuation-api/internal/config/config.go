package config

import "github.com/zeromicro/go-zero/rest"

// Config holds the complete application configuration.
type Config struct {
	rest.RestConf

	// Database connection settings
	Database DatabaseConfig `json:"Database"`

	// Redis connection settings
	Redis RedisConfig `json:"Redis"`

	// External API data source settings
	DataSources DataSourcesConfig `json:"DataSources"`

	// Cache TTL settings (in seconds)
	CacheTTL CacheTTLConfig `json:"CacheTTL"`
}

// DatabaseConfig holds database connection parameters.
type DatabaseConfig struct {
	Driver     string `json:"Driver"`
	DataSource string `json:"DataSource"`
}

// RedisConfig holds Redis connection parameters.
type RedisConfig struct {
	Host string `json:"Host"`
	Pass string `json:"Pass,optional"`
	DB   int    `json:"DB,optional"`
}

// DataSourcesConfig holds all external API data source configurations.
type DataSourcesConfig struct {
	Etherscan   EtherscanConfig   `json:"Etherscan"`
	CoinGecko   CoinGeckoConfig   `json:"CoinGecko"`
	DefiLlama   DefiLlamaConfig   `json:"DefiLlama"`
	Glassnode   GlassnodeConfig   `json:"Glassnode"`
	BeaconChain BeaconChainConfig `json:"BeaconChain"`
	TradFi      TradFiConfig      `json:"TradFi"`
}

// EtherscanConfig holds Etherscan API settings.
type EtherscanConfig struct {
	APIKey  string `json:"APIKey,optional"`
	BaseURL string `json:"BaseURL"`
}

// CoinGeckoConfig holds CoinGecko API settings.
type CoinGeckoConfig struct {
	APIKey  string `json:"APIKey,optional"`
	BaseURL string `json:"BaseURL"`
}

// DefiLlamaConfig holds DefiLlama API settings.
type DefiLlamaConfig struct {
	BaseURL string `json:"BaseURL"`
}

// GlassnodeConfig holds Glassnode API settings.
type GlassnodeConfig struct {
	APIKey  string `json:"APIKey,optional"`
	BaseURL string `json:"BaseURL"`
}

// BeaconChainConfig holds Beacon Chain API settings.
type BeaconChainConfig struct {
	BaseURL string `json:"BaseURL"`
}

// TradFiConfig holds traditional finance data API settings.
type TradFiConfig struct {
	APIKey  string `json:"APIKey,optional"`
	BaseURL string `json:"BaseURL"`
}

// CacheTTLConfig holds cache TTL values in seconds for each data category.
type CacheTTLConfig struct {
	Price             int `json:"Price"`             // 价格数据: 10s
	GasData           int `json:"GasData"`           // Gas 数据: 300s (5min)
	OnChainMetrics    int `json:"OnChainMetrics"`    // 链上指标: 300s (5min)
	TVLData           int `json:"TVLData"`           // TVL 数据: 300s (5min)
	InstitutionalData int `json:"InstitutionalData"` // 机构数据: 3600s (1h)
	NetworkData       int `json:"NetworkData"`       // 网络数据: 300s (5min)
	MacroData         int `json:"MacroData"`         // 宏观数据: 3600s (1h)
	ValuationScore    int `json:"ValuationScore"`    // 估值评分: 300s (5min)
}
