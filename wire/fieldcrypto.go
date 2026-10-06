package wire

import (
	"github.com/sky-xhsoft/sky-server/internal/config"
	"github.com/sky-xhsoft/sky-server/internal/pkg/fieldcrypto"
	"go.uber.org/zap"
)

// setupFieldCrypto 初始化敏感字段加密器（须在任何实体读写之前调用）
func setupFieldCrypto(cfg *config.Config, log *zap.Logger) {
	key := cfg.Security.FieldEncryptKey
	if key == "" {
		log.Warn("security.fieldEncryptKey 未配置，敏感字段加密回退使用 jwt.secret；生产环境请单独配置且不要随意更换")
		key = cfg.JWT.Secret
	}
	fieldcrypto.Setup(key)
}
