package middleware

import "testing"

func TestShouldAudit(t *testing.T) {
	for method, want := range map[string]bool{"GET": false, "HEAD": false, "OPTIONS": false, "POST": true, "PUT": true, "PATCH": true, "DELETE": true} {
		if got := shouldAudit(method, "/api/v1/data/SYS_USER"); got != want {
			t.Errorf("shouldAudit(%s) = %v, want %v", method, got, want)
		}
	}
	// POST 形式的查询/搜索是读操作
	for _, path := range []string{"/api/v1/data/SYS_USER/query", "/api/v1/data/SYS_USER/search"} {
		if shouldAudit("POST", path) {
			t.Errorf("POST %s 不应审计", path)
		}
	}
}
