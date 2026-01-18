package cloud

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// calculateMD5 计算数据的MD5
func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// calculateFileMD5 计算文件的MD5
func calculateFileMD5(file *os.File) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
