package conf

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var runBot = flag.Bool("run_bot", false, "run bot")
var runConf = flag.String("run_conf", "", "run config")

const defaultHTTPTimeoutSec = 30

func Get_RUN_CONF() string {
	if envRunConf := os.Getenv("RUN_CONF"); envRunConf != "" {
		return envRunConf
	}
	return *runConf
}

type Config struct {
	ServerConf        *Server        `mapstructure:"server"`
	MysqlMasterConf   *Mysql         `mapstructure:"db"`
	RedisConf         *Redis         `mapstructure:"redis"`
	NotificationConf  *Notification  `mapstructure:"notification"`
	AlipayConf        *Alipay        `mapstructure:"alipay"`
	InvoiceLookupConf *InvoiceLookup `mapstructure:"invoice_lookup"`
	LiteratureConf    *Literature    `mapstructure:"literature"`
}

// InvoiceLookup 企业开票抬头外部查询（诚数 API 支持名称/税号；Mgtv 发票抬头库按名称模糊查）。
type InvoiceLookup struct {
	ChengshuAPIKey          string `mapstructure:"chengshu_api_key"`
	MgtvBaseURL             string `mapstructure:"mgtv_base_url"`
	InsecureSkipTLSVerify   bool   `mapstructure:"insecure_skip_tls_verify"`
}

type Server struct {
	ServiceName     string `mapstructure:"service_name"`
	RunMode         string `mapstructure:"run_mode"`
	Port            string `mapstructure:"port"`
	Env             string `mapstructure:"env"`
	ProductName     string `mapstructure:"product_name"`
	ApiHost         string `mapstructure:"api_host"`
	MachineName     string `mapstructure:"machine_name"`
	HTTPTimeoutSec  int    `mapstructure:"http_timeout_sec"`
}

type Mysql struct {
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	Host        string `mapstructure:"host"`
	DbName      string `mapstructure:"db_name"`
	MaxIdleConn int    `mapstructure:"max_idle_conn"`
	MaxOpenConn int    `mapstructure:"max_open_conn"`
	LogMode     bool   `mapstructure:"log_mode"`
}

type Redis struct {
	Mode     string `mapstructure:"mode"` // standalone（默认）| cluster | sentinel
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	UserName string `mapstructure:"username"`
	DB       int    `mapstructure:"db"`
	TLS      bool   `mapstructure:"tls"`
	PoolSize int    `mapstructure:"pool_size"`
	MinIdle  int    `mapstructure:"min_idle"`
}

type Notification struct {
	Lark LarkNotification `mapstructure:"lark"`
}

type LarkNotification struct {
	AlertWebhookURL  string `mapstructure:"alert_webhook_url"`
	NotifyWebhookURL string `mapstructure:"notify_webhook_url"`
	ErrorWebhookURL  string `mapstructure:"error_webhook_url"`
	ServiceName      string `mapstructure:"service_name"`
	Timeout          int    `mapstructure:"timeout"`
}

// Alipay OpenAPI（账务明细等）
type Alipay struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	Gateway    string `mapstructure:"gateway"`
}

const (
	DefaultLiteratureArxivBaseURL           = "https://export.arxiv.org/api/query"
	DefaultLiteratureOpenAlexBaseURL        = "https://api.openalex.org/works"
	DefaultLiteratureSemanticScholarBaseURL = "https://api.semanticscholar.org/graph/v1/paper/search"
)

// Literature 文献数据源 HTTP 配置（Paper Agent RAG）
type Literature struct {
	Arxiv           *LiteratureArxiv           `mapstructure:"arxiv"`
	OpenAlex        *LiteratureOpenAlex        `mapstructure:"openalex"`
	SemanticScholar *LiteratureSemanticScholar `mapstructure:"semantic_scholar"`
}

// LiteratureArxiv 无需 API Key；公开 Atom API
type LiteratureArxiv struct {
	BaseURL string `mapstructure:"base_url"`
}

// LiteratureOpenAlex 无需 Key；建议配置 mailto 写入 User-Agent（礼貌池）
type LiteratureOpenAlex struct {
	BaseURL string `mapstructure:"base_url"`
	Mailto  string `mapstructure:"mailto"`
}

// LiteratureSemanticScholar api_key 可选（x-api-key）；不配可用但限流更严
type LiteratureSemanticScholar struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}

var GlobalConf *Config

func NewCfg() *Config {
	runConfValue := Get_RUN_CONF()
	if runConfValue == "" {
		log.Fatal("run_conf parameter is required (set RUN_CONF environment variable or use --run_conf flag)")
	}

	switch runConfValue {
	case "local":
		return loadLocalConfig("conf/local.conf.yaml")
	case "dev":
		return loadLocalConfig("conf/dev.conf.yaml")
	case "rock":
		return loadLocalConfig("conf/rock.conf.yaml")
	case "test":
		return loadLocalConfig("conf/test.conf.yaml")
	case "prod":
		return loadLocalConfig("conf/prod.conf.yaml")
	default:
		return loadLocalConfig("conf/local.conf.yaml")
	}
}

func loadLocalConfig(filename string) *Config {
	cfg, err := LoadSettingByFilePath(filename)
	if err != nil {
		log.Fatalf("Failed to load config file %s: %v", filename, err)
	}
	SetGlobalConf(cfg)
	return cfg
}

func LoadSettingByFilePath(filename string) (*Config, error) {
	var cfg Config
	viper.SetConfigFile(filename)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	fmt.Printf("Loaded config from: %s (env: %s, service: %s)\n",
		filename,
		cfg.ServerConf.Env,
		cfg.ServerConf.ServiceName)
	return &cfg, nil
}

func SetGlobalConf(cfg *Config) {
	GlobalConf = cfg
}

func GetServerName() string {
	return GlobalConf.ServerConf.ServiceName
}

func GetEnv() string {
	return GlobalConf.ServerConf.Env
}

func IsRunBot() bool {
	return *runBot
}

func GetMysqlMasterConf() *Mysql {
	if GlobalConf == nil || GlobalConf.MysqlMasterConf == nil {
		log.Printf("Warning: MySQL master configuration is not initialized")
		return nil
	}
	return GlobalConf.MysqlMasterConf
}

func GetMysqlReplicaConf() *Mysql {
	return GetMysqlMasterConf()
}

func GetRedisConf() *Redis {
	if GlobalConf == nil || GlobalConf.RedisConf == nil {
		log.Printf("Warning: Redis configuration is not initialized")
		return nil
	}
	return GlobalConf.RedisConf
}

func GetServeConf() *Server {
	if GlobalConf == nil || GlobalConf.ServerConf == nil {
		log.Printf("Warning: Server configuration is not initialized")
		return nil
	}
	return GlobalConf.ServerConf
}

func Get_MACHINE_NAME() string {
	s := GetServeConf()
	if s == nil {
		return ""
	}
	return s.MachineName
}

func GetLarkNotificationConf() *LarkNotification {
	if GlobalConf == nil || GlobalConf.NotificationConf == nil {
		log.Printf("Warning: Notification configuration is not initialized")
		return nil
	}
	return &GlobalConf.NotificationConf.Lark
}

func GetHTTPTimeout() time.Duration {
	if GlobalConf != nil && GlobalConf.ServerConf != nil && GlobalConf.ServerConf.HTTPTimeoutSec > 0 {
		return time.Duration(GlobalConf.ServerConf.HTTPTimeoutSec) * time.Second
	}
	return time.Duration(defaultHTTPTimeoutSec) * time.Second
}

func GetAlipayConf() *Alipay {
	if GlobalConf == nil {
		return nil
	}
	return GlobalConf.AlipayConf
}

func GetLiteratureConf() *Literature {
	out := defaultLiteratureConf()
	if GlobalConf == nil || GlobalConf.LiteratureConf == nil {
		applyLiteratureEnvFallback(out)
		return out
	}
	mergeLiteratureFromFile(out, GlobalConf.LiteratureConf)
	applyLiteratureEnvFallback(out)
	return out
}

func defaultLiteratureConf() *Literature {
	return &Literature{
		Arxiv:           &LiteratureArxiv{BaseURL: DefaultLiteratureArxivBaseURL},
		OpenAlex:        &LiteratureOpenAlex{BaseURL: DefaultLiteratureOpenAlexBaseURL},
		SemanticScholar: &LiteratureSemanticScholar{BaseURL: DefaultLiteratureSemanticScholarBaseURL},
	}
}

func mergeLiteratureFromFile(dst, src *Literature) {
	if src == nil {
		return
	}
	if src.Arxiv != nil {
		if u := stringsTrim(src.Arxiv.BaseURL); u != "" {
			dst.Arxiv.BaseURL = u
		}
	}
	if src.OpenAlex != nil {
		if u := stringsTrim(src.OpenAlex.BaseURL); u != "" {
			dst.OpenAlex.BaseURL = u
		}
		if m := stringsTrim(src.OpenAlex.Mailto); m != "" {
			dst.OpenAlex.Mailto = m
		}
	}
	if src.SemanticScholar != nil {
		if u := stringsTrim(src.SemanticScholar.BaseURL); u != "" {
			dst.SemanticScholar.BaseURL = u
		}
		if k := stringsTrim(src.SemanticScholar.APIKey); k != "" {
			dst.SemanticScholar.APIKey = k
		}
	}
}

func applyLiteratureEnvFallback(l *Literature) {
	if l == nil || l.SemanticScholar == nil {
		return
	}
	if l.SemanticScholar.APIKey == "" {
		l.SemanticScholar.APIKey = os.Getenv("SEMANTIC_SCHOLAR_API_KEY")
	}
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

