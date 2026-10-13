package live

import (
	"time"

	"github.com/sky-xhsoft/sky-server/internal/model/entity"
)

// minValidEventTime 早于这个时间的事件时间视为无效（时钟或数据异常）
var minValidEventTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// maxEventClockSkew 事件时间允许领先服务器的最大误差
const maxEventClockSkew = time.Minute

// EventTime 回调事件发生时间：腾讯云 event_time（Unix 秒）。
// 为 0、早于 2020 年、或比服务器时间晚 1 分钟以上（时钟异常）时用服务器时间 now。
func EventTime(unixSeconds int64, now time.Time) time.Time {
	if unixSeconds <= 0 {
		return now
	}
	t := time.Unix(unixSeconds, 0)
	if t.Before(minValidEventTime) || t.After(now.Add(maxEventClockSkew)) {
		return now
	}
	return t
}

// EndedDuration 结束时的直播时长（秒）：优先用腾讯云上报的推流时长，其次用结束时间减实际开播时间；不会为负。
func EndedDuration(start *time.Time, end time.Time, reportedSeconds int64) int {
	if reportedSeconds > 0 {
		return int(reportedSeconds)
	}
	if start == nil || !end.After(*start) {
		return 0
	}
	return int(end.Sub(*start).Seconds())
}

// LiveElapsedSeconds 直播时长（秒），由服务器时钟计算，前端只在此基础上逐秒累加：
//   - 直播中：当前时间 - 实际开播时间（开播时 START_TIME 会被写成实际时间）
//   - 已结束：记录的时长（断流回调、手动结束时写入）；旧数据没有时长时为 0，前端显示 --
//     （不用 结束时间 - 开播时间 兜底：旧数据的结束时间有批量写入的，算出来会是几十小时）
//   - 其他状态：0
func LiveElapsedSeconds(room *entity.LiveRoom, now time.Time) int64 {
	if room == nil {
		return 0
	}
	switch room.Status {
	case "live":
		if room.StartTime == nil || now.Before(*room.StartTime) {
			return 0
		}
		return int64(now.Sub(*room.StartTime).Seconds())
	case "ended", "archived":
		if room.Duration > 0 {
			return int64(room.Duration)
		}
	}
	return 0
}
