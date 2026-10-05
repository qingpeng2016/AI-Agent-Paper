package apikey

import conf2 "github.com/qingpeng2016/ai-token-mall/conf"

// NewVaultFromConfig 派生服务端密钥材料。
func NewVaultFromConfig(cfg *conf2.Config) *Vault {
	material := "ai-token-mall-api-keys"
	if cfg != nil && cfg.ServerConf != nil {
		material += "|" + cfg.ServerConf.ServiceName + "|" + cfg.ServerConf.Env
	}
	if cfg != nil && cfg.MysqlMasterConf != nil {
		material += "|" + cfg.MysqlMasterConf.DbName
	}
	return NewVault(material)
}
