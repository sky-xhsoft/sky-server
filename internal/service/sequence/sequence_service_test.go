package sequence

import (
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// helper: 由于 service 的内部方法签名是私有的，我们在同包测试中可以直接访问
func newServiceForTest() *service {
	return &service{}
}

func TestSequenceService_GetCurrentDateStr(t *testing.T) {
	s := newServiceForTest()
	now := time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		cycleType string
		expected  string
	}{
		{"D", "20260315"},
		{"M", "202603"},
		{"Y", "2026"},
		{"N", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := s.getCurrentDateStr(tt.cycleType, now)
		if got != tt.expected {
			t.Errorf("getCurrentDateStr(%q): expected %q, got %q", tt.cycleType, tt.expected, got)
		}
	}
}

func TestSequenceService_FormatSequence_BasicFormat(t *testing.T) {
	s := newServiceForTest()
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)

	seq := &entity.SysSeq{
		VFormat: "{YYYY}{MM}{DD}-{0000}",
		CurNum:  42,
	}

	result := s.formatSequence(seq, now)
	if result != "20260521-0042" {
		t.Errorf("Expected 20260521-0042, got %s", result)
	}
}

func TestSequenceService_FormatSequence_WithPrefixSuffix(t *testing.T) {
	s := newServiceForTest()
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)

	seq := &entity.SysSeq{
		VFormat: "{0000}",
		Prefix:  "INV-",
		Suffix:  "-Z",
		CurNum:  1,
	}

	result := s.formatSequence(seq, now)
	if result != "INV-0001-Z" {
		t.Errorf("Expected INV-0001-Z, got %s", result)
	}
}

func TestSequenceService_FormatSequence_VariousZeroWidths(t *testing.T) {
	s := newServiceForTest()
	now := time.Now()

	tests := []struct {
		vformat  string
		curNum   int
		expected string
	}{
		{"{00}", 5, "05"},
		{"{0000}", 5, "0005"},
		{"{000000}", 5, "000005"},
	}

	for _, tt := range tests {
		seq := &entity.SysSeq{VFormat: tt.vformat, CurNum: tt.curNum}
		result := s.formatSequence(seq, now)
		if result != tt.expected {
			t.Errorf("VFormat=%s CurNum=%d: expected %s, got %s", tt.vformat, tt.curNum, tt.expected, result)
		}
	}
}

func TestSequenceService_FormatSequence_YearMonthFormats(t *testing.T) {
	s := newServiceForTest()
	now := time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC)

	seq := &entity.SysSeq{
		VFormat: "{YY}{MM}{DD}",
	}
	result := s.formatSequence(seq, now)
	if result != "260521" {
		t.Errorf("Expected 260521, got %s", result)
	}
}
