package live

import "testing"

func TestNormalizeHistoryTotal(t *testing.T) {
	cases := []struct {
		name                         string
		pageNum, pageSize, rows, rep int64
		wantTotal, wantPage          int64
	}{
		{"腾讯云返回正确", 1, 10, 2, 2, 2, 1},
		{"最后一页不满：按实际条数算出精确总数", 3, 10, 4, 100, 24, 3},
		{"最后一页不满且腾讯云偏小", 2, 10, 5, 3, 15, 2},
		{"满页但腾讯云偏小：至少到本页末尾", 2, 10, 10, 0, 20, 2},
		{"满页且腾讯云更大：保持", 1, 10, 10, 35, 35, 4},
		{"页码超出且总数偏大：压到前面各页之和", 3, 10, 0, 100, 20, 2},
		{"页码超出且腾讯云返回 0：保持 0", 2, 10, 0, 0, 0, 0},
		{"第一页没有数据", 1, 10, 0, 0, 0, 0},
	}
	for _, c := range cases {
		total, page := normalizeHistoryTotal(c.pageNum, c.pageSize, c.rows, c.rep)
		if total != c.wantTotal || page != c.wantPage {
			t.Errorf("%s: got total=%d page=%d, want total=%d page=%d", c.name, total, page, c.wantTotal, c.wantPage)
		}
	}
}
