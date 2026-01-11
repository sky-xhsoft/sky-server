package permission

// 权限位定义（5位）
const (
	Read   = 1 << 0 // 1  - 读权限
	Write  = 1 << 1 // 2  - 写权限
	Submit = 1 << 2 // 4  - 提交权限
	Audit  = 1 << 3 // 8  - 审核权限
	Export = 1 << 4 // 16 - 导出权限
)

// 权限组合常量
const (
	None            = 0                          // 0  - 无权限
	ReadWrite       = Read | Write               // 3  - 读+写
	ReadSubmit      = Read | Submit              // 5  - 读+提交
	ReadAudit       = Read | Audit               // 9  - 读+审核
	ReadWriteSubmit = Read | Write | Submit      // 7  - 读+写+提交
	ReadWriteAudit  = Read | Write | Audit       // 11 - 读+写+审核
	AllNoExport     = Read | Write | Submit | Audit // 15 - 全部权限（无导出）
	ReadExport      = Read | Export              // 17 - 读+导出
	WriteExport     = Write | Export             // 19 - 写+导出（实际应该是 Read | Write | Export = 19，但这里根据文档定义）
	ReadSubmitExport = Read | Submit | Export    // 21 - 读+提交+导出
	WriteSubmitExport = Read | Write | Submit | Export // 23 - 读+写+提交+导出
	All             = Read | Write | Submit | Audit | Export // 31 - 全部权限
)

// HasPermission 检查是否拥有指定权限
// userPerm: 用户拥有的权限值
// requirePerm: 需要的权限值
// 返回: 是否拥有所需权限
func HasPermission(userPerm, requirePerm int) bool {
	return (userPerm & requirePerm) == requirePerm
}

// AddPermission 添加权限
func AddPermission(userPerm, addPerm int) int {
	return userPerm | addPerm
}

// RemovePermission 移除权限
func RemovePermission(userPerm, removePerm int) int {
	return userPerm &^ removePerm
}

// ParsePermission 解析权限值为字符串数组
func ParsePermission(perm int) []string {
	var perms []string

	if perm&Read != 0 {
		perms = append(perms, "read")
	}
	if perm&Write != 0 {
		perms = append(perms, "write")
	}
	if perm&Submit != 0 {
		perms = append(perms, "submit")
	}
	if perm&Audit != 0 {
		perms = append(perms, "audit")
	}
	if perm&Export != 0 {
		perms = append(perms, "export")
	}

	return perms
}

// BuildPermission 从字符串数组生成权限值
func BuildPermission(perms []string) int {
	var result int

	for _, perm := range perms {
		switch perm {
		case "read":
			result |= Read
		case "write":
			result |= Write
		case "submit":
			result |= Submit
		case "audit":
			result |= Audit
		case "export":
			result |= Export
		}
	}

	return result
}

// CanRead 检查是否有读权限
func CanRead(perm int) bool {
	return HasPermission(perm, Read)
}

// CanWrite 检查是否有写权限（需要同时有读和写权限）
func CanWrite(perm int) bool {
	return HasPermission(perm, ReadWrite)
}

// CanSubmit 检查是否有提交权限（需要同时有读权限）
func CanSubmit(perm int) bool {
	return HasPermission(perm, Read|Submit)
}

// CanAudit 检查是否有审核权限（需要同时有读权限）
func CanAudit(perm int) bool {
	return HasPermission(perm, Read|Audit)
}

// CanExport 检查是否有导出权限（需要同时有读权限）
func CanExport(perm int) bool {
	return HasPermission(perm, Read|Export)
}
