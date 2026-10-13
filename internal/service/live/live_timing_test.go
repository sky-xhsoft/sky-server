package live

import (
	"testing"
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

func TestEventTime(t *testing.T) {
	now := time.Date(2026, 10, 13, 15, 0, 0, 0, time.Local)
	ok := now.Add(-30 * time.Second)
	cases := []struct {
		name string
		ts   int64
		want time.Time
	}{
		{"正常事件时间", ok.Unix(), time.Unix(ok.Unix(), 0)},
		{"没有事件时间", 0, now},
		{"早于 2020 年", time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC).Unix(), now},
		{"比服务器时间晚 10 分钟（时钟异常）", now.Add(10 * time.Minute).Unix(), now},
		{"领先 30 秒以内允许", now.Add(30 * time.Second).Unix(), time.Unix(now.Add(30*time.Second).Unix(), 0)},
	}
	for _, c := range cases {
		if got := EventTime(c.ts, now); !got.Equal(c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestEndedDuration(t *testing.T) {
	start := time.Date(2026, 10, 13, 14, 0, 0, 0, time.Local)
	if got := EndedDuration(&start, start.Add(90*time.Minute), 0); got != 5400 {
		t.Errorf("结束-开始: got %d", got)
	}
	if got := EndedDuration(&start, start.Add(90*time.Minute), 4000); got != 4000 {
		t.Errorf("优先用上报时长: got %d", got)
	}
	if got := EndedDuration(&start, start.Add(-time.Minute), 0); got != 0 {
		t.Errorf("结束早于开始不应为负: got %d", got)
	}
	if got := EndedDuration(nil, start, 0); got != 0 {
		t.Errorf("没有开始时间: got %d", got)
	}
}

func TestLiveElapsedSeconds(t *testing.T) {
	now := time.Date(2026, 10, 13, 15, 0, 0, 0, time.Local)
	start := now.Add(-75 * time.Second)
	future := now.Add(time.Hour)
	end := start.Add(10 * time.Minute)
	cases := []struct {
		name string
		room *entity.LiveRoom
		want int64
	}{
		{"直播中按实际开播时间计算", &entity.LiveRoom{Status: "live", StartTime: &start}, 75},
		{"直播中但开播时间在未来（计划时间未被覆盖）不显示负数", &entity.LiveRoom{Status: "live", StartTime: &future}, 0},
		{"已结束用记录的时长", &entity.LiveRoom{Status: "ended", StartTime: &start, Duration: 300}, 300},
		{"已结束没记录时长（旧数据）不推算", &entity.LiveRoom{Status: "ended", StartTime: &start, EndTime: &end}, 0},
		{"草稿", &entity.LiveRoom{Status: "draft", StartTime: &start}, 0},
		{"空", nil, 0},
	}
	for _, c := range cases {
		if got := LiveElapsedSeconds(c.room, now); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}
