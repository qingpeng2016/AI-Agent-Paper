package conf

import (
	"time"
)

// Work config
type Work struct {
	IntervalTime time.Duration `mapstructure:"interval_time"`
}

// ThirdPlatform config
type ThirdPlatform struct {
	EthRpcUrl   string `mapstructure:"eth_rpc_url"`
	EthRpcWsUrl string `mapstructure:"eth_rpc_ws_url"`
}

type NacosConfig struct {
	IsRemote  bool
	Endpoint  string
	Host      string
	Port      uint64
	Namespace string
	Group     string
	Username  string
	Password  string
}

// AwsSes config
type AwsSes struct {
	AccessKey    string `mapstructure:"access_key"`
	AccessSecret string `mapstructure:"access_secret"`
	Region       string `mapstructure:"region"`
	SenderEmail  string `mapstructure:"sender_email"`
}

// BitcoinServerConfig config
type BitcoinServerConfig struct {
	Host       string `mapstructure:"host"`
	BusinessId string `mapstructure:"business_id"`
}

type ClickHouse struct {
	Dsn string `mapstructure:"dsn"`
}

// TokenList config
type TokenList struct {
	WBTC string `mapstructure:"wbtc"`
	USDC string `mapstructure:"usdc"`
	WETH string `mapstructure:"weth"`
}

type ContractConfig struct {
	PortfolioMarginManager string `mapstructure:"portfolioMarginManager"`
	Vault                  string `mapstructure:"vault"`
	StandardPMRFQ          string `mapstructure:"standardPMRFQ"`
	StrategyQuery          string `mapstructure:"strategyQuery"`
	Oracle                 string `mapstructure:"oracle"`
	MultiQuery             string `mapstructure:"multiQuery"`
	Multicall              string `mapstructure:"multicall"`

	Token TokenList `mapstructure:"tokens"`
}

type AlertConfig struct {
	Webhook string `json:"webhook"`
}
