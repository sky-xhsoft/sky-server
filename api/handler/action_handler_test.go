package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/model/entity"
	"github.com/sky-xhsoft/sky-server/internal/service/action"
)

// fakeActionService 实现 action.Service 接口
type fakeActionService struct {
	executeResult       *action.ActionResult
	executeErr          error
	executeByNameResult *action.ActionResult
	executeByNameErr    error
	batchResult         []*action.ActionResult
	batchErr            error
	getResult           *entity.SysAction
	getErr              error
}

func (m *fakeActionService) ExecuteAction(ctx context.Context, actionID uint, params map[string]interface{}, userID uint) (*action.ActionResult, error) {
	return m.executeResult, m.executeErr
}

func (m *fakeActionService) ExecuteActionByName(ctx context.Context, tableName, actionName string, params map[string]interface{}, userID uint) (*action.ActionResult, error) {
	return m.executeByNameResult, m.executeByNameErr
}

func (m *fakeActionService) BatchExecuteAction(ctx context.Context, actionID uint, batchParams []map[string]interface{}, userID uint) ([]*action.ActionResult, error) {
	return m.batchResult, m.batchErr
}

func (m *fakeActionService) GetAction(ctx context.Context, actionID uint) (*entity.SysAction, error) {
	return m.getResult, m.getErr
}

func TestActionHandler_ExecuteAction_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeActionService{}
	h := NewActionHandler(svc)

	r := gin.New()
	r.POST("/actions/:actionId/execute", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.ExecuteAction(c)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/actions/abc/execute", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestActionHandler_ExecuteAction_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeActionService{}
	h := NewActionHandler(svc)

	r := gin.New()
	r.POST("/actions/:actionId/execute", h.ExecuteAction)

	body, _ := json.Marshal(map[string]interface{}{"params": map[string]interface{}{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/actions/1/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestActionHandler_ExecuteAction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeActionService{
		executeResult: &action.ActionResult{Success: true, Message: "done"},
	}
	h := NewActionHandler(svc)

	r := gin.New()
	r.POST("/actions/:actionId/execute", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.ExecuteAction(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"params": map[string]interface{}{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/actions/1/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestActionHandler_ExecuteAction_BusinessFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeActionService{
		executeResult: &action.ActionResult{Success: false, Error: "action failed"},
	}
	h := NewActionHandler(svc)

	r := gin.New()
	r.POST("/actions/:actionId/execute", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.ExecuteAction(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"params": map[string]interface{}{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/actions/1/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 业务失败应返回 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestActionHandler_ExecuteActionByName_EmptyParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeActionService{
		executeByNameResult: &action.ActionResult{Success: true, Message: "ok"},
	}
	h := NewActionHandler(svc)

	r := gin.New()
	r.POST("/actions/:tableName/:actionName/execute", func(c *gin.Context) {
		c.Set("userID", uint(1))
		h.ExecuteActionByName(c)
	})

	body, _ := json.Marshal(map[string]interface{}{"params": map[string]interface{}{}})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/actions/users/SEND_EMAIL/execute", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}
