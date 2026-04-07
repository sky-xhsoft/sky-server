package contextutil

// ContextKey 是用于 context.WithValue 的自定义类型
// 使用自定义类型可以避免不同包之间的 key 冲突
type ContextKey string

const (
	// CompanyIDKey 用于在 context 中存储公司ID
	CompanyIDKey ContextKey = "companyId"
)
