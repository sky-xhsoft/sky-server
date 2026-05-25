package fieldcrypto

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

// EncryptedString GORM 序列化器，自动加密/解密字符串字段
//
// 使用方式：
//   1. 注册序列化器：fieldcrypto.RegisterSerializer(encryptor)
//   2. 在模型字段上使用 tag：
//        type User struct {
//            SSN string `gorm:"column:SSN;serializer:encrypted"`
//        }
type EncryptedString struct{}

// 全局加密器（通过 RegisterSerializer 设置）
var globalEncryptor *Encryptor

// RegisterSerializer 注册 GORM 加密序列化器
// 必须在应用启动时调用一次
func RegisterSerializer(encryptor *Encryptor) {
	globalEncryptor = encryptor
	schema.RegisterSerializer("encrypted", &EncryptedString{})
}

// Scan 数据库 → Go 值（解密）
func (s *EncryptedString) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}

	if globalEncryptor == nil {
		return fmt.Errorf("fieldcrypto: encryptor not registered")
	}

	var encrypted string
	switch v := dbValue.(type) {
	case []byte:
		encrypted = string(v)
	case string:
		encrypted = v
	default:
		return fmt.Errorf("fieldcrypto: unsupported db value type: %T", dbValue)
	}

	if encrypted == "" {
		field.ReflectValueOf(ctx, dst).SetString("")
		return nil
	}

	decrypted, err := globalEncryptor.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf("fieldcrypto: decrypt: %w", err)
	}

	field.ReflectValueOf(ctx, dst).SetString(decrypted)
	return nil
}

// Value Go 值 → 数据库（加密）
func (s *EncryptedString) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}

	if globalEncryptor == nil {
		return nil, fmt.Errorf("fieldcrypto: encryptor not registered")
	}

	str, ok := fieldValue.(string)
	if !ok {
		return nil, fmt.Errorf("fieldcrypto: field value must be string, got %T", fieldValue)
	}

	if str == "" {
		return "", nil
	}

	return globalEncryptor.Encrypt(str)
}
