package handler

import (
	"bytes"
	stderrors "errors"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-xhsoft/sky-server/internal/service/sequence"
)

// fakeSequenceService 实现 sequence.Service 接口
type fakeSequenceService struct {
	nextValue        string
	nextErr          error
	currentValue     string
	currentErr       error
	resetErr         error
	previewValue     string
	previewErr       error
}

func (m *fakeSequenceService) NextValue(seqName string) (string, error) {
	return m.nextValue, m.nextErr
}

func (m *fakeSequenceService) GetCurrentValue(seqName string) (string, error) {
	return m.currentValue, m.currentErr
}

func (m *fakeSequenceService) ResetSequence(seqName string) error {
	return m.resetErr
}

func (m *fakeSequenceService) PreviewNext(seqName string) (string, error) {
	return m.previewValue, m.previewErr
}

var _ sequence.Service = (*fakeSequenceService)(nil)

func TestSequenceHandler_NextValue_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeSequenceService{nextValue: "INV20260521-0001"}
	h := NewSequenceHandler(svc)

	r := gin.New()
	r.POST("/sequences/:seqName/next", h.NextValue)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sequences/invoice/next", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["value"] != "INV20260521-0001" {
		t.Errorf("Expected value 'INV20260521-0001', got %v", data["value"])
	}
}

func TestSequenceHandler_NextValue_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeSequenceService{nextErr: stderrors.New("redis error")}
	h := NewSequenceHandler(svc)

	r := gin.New()
	r.POST("/sequences/:seqName/next", h.NextValue)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sequences/test/next", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

func TestSequenceHandler_BatchNextValue_InvalidCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeSequenceService{nextValue: "x"}
	h := NewSequenceHandler(svc)

	r := gin.New()
	r.POST("/sequences/batch", h.BatchNextValue)

	// Count = 0
	body, _ := json.Marshal(map[string]interface{}{"seqName": "test", "count": 0})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sequences/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for count=0, got %d", w.Code)
	}

	// Count > 100
	body, _ = json.Marshal(map[string]interface{}{"seqName": "test", "count": 200})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/sequences/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for count>100, got %d", w.Code)
	}
}

func TestSequenceHandler_BatchNextValue_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &fakeSequenceService{nextValue: "v"}
	h := NewSequenceHandler(svc)

	r := gin.New()
	r.POST("/sequences/batch", h.BatchNextValue)

	body, _ := json.Marshal(map[string]interface{}{"seqName": "test", "count": 5})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/sequences/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	values := data["values"].([]interface{})
	if len(values) != 5 {
		t.Errorf("Expected 5 values, got %d", len(values))
	}
}
