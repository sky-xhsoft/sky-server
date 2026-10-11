package middleware

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/pkg/vendortext"
)

// vendorNeutralKeys 提示类字段：这些字段中的第三方云平台名称与服务地址会被替换。
// 其余字段（存储桶地址、播放地址、上传地址等业务数据）原样返回。
var vendorNeutralKeys = map[string]bool{
	"message": true, "msg": true, "error": true, "errors": true,
	"syncError": true, "cloudError": true, "reason": true, "detail": true, "details": true,
}

// VendorNeutralMessages 返回给前端的 JSON 中，提示类字段不出现腾讯云、火山引擎等第三方平台名称。
// 第三方接口的原始报错（含平台名与服务地址）经常被直接拼进提示，这里统一兜底；日志不受影响。
func VendorNeutralMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &vendorNeutralWriter{ResponseWriter: c.Writer}
		c.Writer = w
		defer func() {
			if r := recover(); r != nil {
				// 交还原始写入器，由外层 Recovery 正常写出错误响应
				c.Writer = w.ResponseWriter
				panic(r)
			}
		}()
		c.Next()
		w.flushBuffered()
	}
}

type vendorNeutralWriter struct {
	gin.ResponseWriter
	decided   bool
	buffering bool
	buf       bytes.Buffer
}

func (w *vendorNeutralWriter) decide() {
	if !w.decided {
		w.decided = true
		w.buffering = strings.Contains(w.Header().Get("Content-Type"), "application/json")
	}
}

func (w *vendorNeutralWriter) Write(b []byte) (int, error) {
	w.decide()
	if w.buffering {
		return w.buf.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func (w *vendorNeutralWriter) WriteString(s string) (int, error) {
	w.decide()
	if w.buffering {
		return w.buf.WriteString(s)
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *vendorNeutralWriter) Flush() {
	if w.buffering {
		return // JSON 整体处理后再写出
	}
	w.ResponseWriter.Flush()
}

func (w *vendorNeutralWriter) flushBuffered() {
	if !w.buffering {
		return
	}
	w.buffering = false
	body := w.buf.Bytes()
	if out, ok := neutralizeJSON(body); ok {
		body = out
	}
	_, _ = w.ResponseWriter.Write(body)
}

// neutralizeJSON 替换提示类字段；没有需要替换的内容时返回 false，原样写出
func neutralizeJSON(body []byte) ([]byte, bool) {
	if !vendortext.Contains(string(body)) {
		return nil, false
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // 保持大整数 ID 精度
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return nil, false
	}
	v, changed := neutralize(v, false)
	if !changed {
		return nil, false
	}
	out, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	return out, true
}

// neutralize 遍历 JSON；inMessage 为 true 表示位于提示类字段之下，其中所有字符串都要替换
func neutralize(v interface{}, inMessage bool) (interface{}, bool) {
	changed := false
	switch t := v.(type) {
	case string:
		if inMessage {
			if s := vendortext.Sanitize(t); s != t {
				return s, true
			}
		}
	case map[string]interface{}:
		for k, child := range t {
			if nv, ok := neutralize(child, inMessage || vendorNeutralKeys[k]); ok {
				t[k] = nv
				changed = true
			}
		}
	case []interface{}:
		for i, child := range t {
			if nv, ok := neutralize(child, inMessage); ok {
				t[i] = nv
				changed = true
			}
		}
	}
	return v, changed
}
