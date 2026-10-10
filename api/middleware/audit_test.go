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

// 过程请求与回调不审计（原先记成 unknown 刷满最新动态）
func TestShouldAudit_SkipsNoise(t *testing.T) {
	for _, path := range []string{
		"/api/v1/live/push-url/by-room", "/api/v1/live/pull-url/by-room", "/api/v1/live/play-url/generate",
		"/api/v1/cloud/files/multipart/upload", "/api/v1/cloud/files/multipart/12/chunk", "/api/v1/cloud/files/multipart/12/presigned",
		"/api/v1/cloud/files/multipart/init", "/api/v1/live/callback/push",
	} {
		if shouldAudit("POST", path) {
			t.Errorf("POST %s 不应审计", path)
		}
	}
	if !shouldAudit("POST", "/api/v1/cloud/files/multipart/complete") {
		t.Error("分片上传完成应审计（记一次上传）")
	}
}

func TestParseActionAndResource_Business(t *testing.T) {
	cases := []struct{ method, path, action, resource string }{
		{"POST", "/api/v1/cloud/files/multipart/complete", "upload", "cloud_file"},
		{"POST", "/api/v1/cloud/files/upload", "upload", "cloud_file"},
		{"PUT", "/api/v1/cloud/items/2606/rename", "rename", "cloud_file"},
		{"POST", "/api/v1/cloud/items", "create", "cloud_file"},
		{"DELETE", "/api/v1/cloud/items/2600", "delete", "cloud_file"},
		{"POST", "/api/v1/cloud/shares", "share", "cloud_share"},
		{"POST", "/api/v1/live/rooms/92/start", "start_live", "live_room"},
		{"POST", "/api/v1/live/rooms/92/end", "end_live", "live_room"},
		{"POST", "/api/v1/live/pull-stream/tasks", "create", "pull_stream"},
		{"POST", "/api/v1/live/domains", "create", "live_domain"},
		{"PUT", "/api/v1/company/config", "update_config", "company"},
		// 通用规则不受影响
		{"POST", "/api/v1/workflow/instances/5/complete", "complete_task", "workflow"},
		{"POST", "/api/v1/data/SYS_USER", "create", "table"},
	}
	for _, c := range cases {
		a, r := parseActionAndResource(c.method, c.path)
		if a != c.action || r != c.resource {
			t.Errorf("%s %s = (%s, %s)，期望 (%s, %s)", c.method, c.path, a, r, c.action, c.resource)
		}
	}
}
