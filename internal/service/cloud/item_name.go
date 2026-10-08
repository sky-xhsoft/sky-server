package cloud

import (
	"strings"
	"unicode/utf8"

	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

// maxItemNameLength 文件/文件夹名称最大长度（与 cloud_item.NAME 列一致）
const maxItemNameLength = 255

// ValidateItemName 校验用户输入的文件/文件夹名称，返回去掉首尾空白后的名称。
// 名称会拼进路径（PATH），含 / 或 \ 会与真实的子目录路径混淆，改名、移动时误改其他条目的路径
func ValidateItemName(name string) (string, error) {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return "", errors.New(errors.ErrInvalidParam, "名称不能为空")
	case name == "." || name == "..":
		return "", errors.New(errors.ErrInvalidParam, "名称不能为 . 或 ..")
	case strings.ContainsAny(name, `/\`):
		return "", errors.New(errors.ErrInvalidParam, "名称不能包含 / 或 \\")
	case utf8.RuneCountInString(name) > maxItemNameLength:
		return "", errors.New(errors.ErrInvalidParam, "名称不能超过 255 个字符")
	}
	return name, nil
}
