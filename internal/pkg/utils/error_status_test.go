package utils

import (
	"encoding/json"
	stderrors "errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/errors"
)

func TestHTTPStatusFor(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{errors.New(errors.ErrInvalidParam, "x"), 400},
		{errors.New(errors.ErrValidation, "x"), 400},
		{errors.New(errors.ErrInvalidCredentials, "x"), 400},
		{errors.New(errors.ErrTokenExpired, "x"), 401},
		{errors.New(errors.ErrPermissionDenied, "x"), 403},
		{errors.New(errors.ErrUnauthorized, "x"), 403},
		{errors.New(errors.ErrResourceNotFound, "x"), 404},
		{errors.New(errors.ErrResourceExists, "x"), 409},
		{errors.New(errors.ErrDatabase, "x"), 500},
		{errors.Wrap(errors.ErrResourceNotFound, "记录不存在", stderrors.New("gorm")), 404},
		{stderrors.New("plain"), 500},
	}
	for _, tc := range cases {
		if got := HTTPStatusFor(tc.err); got != tc.want {
			t.Errorf("HTTPStatusFor(%v) = %d, want %d", tc.err, got, tc.want)
		}
	}
}

func TestFailFromError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	run := func(err error) (int, string) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		FailFromError(c, "查询失败: ", err)
		var body Response
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		return w.Code, body.Message
	}

	if code, msg := run(errors.New(errors.ErrResourceNotFound, "记录不存在")); code != 404 || msg != "记录不存在" {
		t.Errorf("4xx 应只返回业务提示: %d %q", code, msg)
	}
	if code, msg := run(errors.New(errors.ErrDatabase, "查询失败")); code != 500 || msg == "查询失败" {
		t.Errorf("5xx 应附带前缀与原始错误: %d %q", code, msg)
	}
}
