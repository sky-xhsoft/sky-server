package entity

// Masked 返回敏感字段脱敏后的副本（密钥/凭证保留首尾各4位），用于返回给前端
func (c *SysCompanyConf) Masked() *SysCompanyConf {
	if c == nil {
		return nil
	}

	masked := *c
	masked.SecretKey = maskSecret(masked.SecretKey)
	masked.AliyunOSSAccessKeySecret = maskSecret(masked.AliyunOSSAccessKeySecret)
	masked.TencentCOSSecretKey = maskSecret(masked.TencentCOSSecretKey)
	masked.TencentCloudSecretKey = maskSecret(masked.TencentCloudSecretKey)
	masked.TencentCloudCallbackKey = maskSecret(masked.TencentCloudCallbackKey)
	masked.VolcengineAccessKeySecret = maskSecret(masked.VolcengineAccessKeySecret)
	masked.TencentCOSSecretID = maskSecret(masked.TencentCOSSecretID)
	masked.TencentCloudSecretID = maskSecret(masked.TencentCloudSecretID)
	masked.AliyunOSSAccessKeyID = maskSecret(masked.AliyunOSSAccessKeyID)
	masked.VolcengineAccessKeyId = maskSecret(masked.VolcengineAccessKeyId)
	masked.ArkAPIKey = maskSecret(masked.ArkAPIKey)
	return &masked
}

// maskSecret 对密钥进行脱敏（保留首尾各4位，中间替换为 ****）
func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) > 8 {
		return value[:4] + "****" + value[len(value)-4:]
	}
	return "****"
}
