package utils

import (
	"fmt"
	"time"
)

// FormatTime 將時間格式化為指定格式的字串
// 參數：
//   - t: 時間對象
//   - layout: 格式字串，默認為 "2006-01-02 15:04:05"
//
// 返回：
//   - 格式化後的時間字串
func FormatTime(t time.Time, layout ...string) string {
	timeLayout := "2006-01-02 15:04:05"
	if len(layout) > 0 {
		timeLayout = layout[0]
	}

	return t.Format(timeLayout)
}

// ParseTime 解析時間字串為 time.Time 對象
// 參數：
//   - timeStr: 時間字串
//   - layout: 格式字串，默認為 "2006-01-02 15:04:05"
//
// 返回：
//   - 解析後的時間對象和錯誤訊息
func ParseTime(timeStr string, layout ...string) (time.Time, error) {
	timeLayout := "2006-01-02 15:04:05"
	if len(layout) > 0 {
		timeLayout = layout[0]
	}

	return time.Parse(timeLayout, timeStr)
}

// TimeDiff 計算兩個時間的差值並以指定單位返回
// 參數：
//   - t1, t2: 要比較的兩個時間
//   - unit: 返回單位，可選 "second", "minute", "hour", "day"，默認為 "second"
//
// 返回：
//   - 時間差值（按指定單位）
func TimeDiff(t1, t2 time.Time, unit ...string) float64 {
	diff := t2.Sub(t1).Seconds()

	if len(unit) > 0 {
		switch unit[0] {
		case "minute":
			return diff / 60
		case "hour":
			return diff / 3600
		case "day":
			return diff / 86400
		}
	}

	return diff
}

// GetDateRange 獲取指定日期的開始時間和結束時間
// 參數：
//   - t: 時間對象
//   - rangeType: 範圍類型，可選 "day", "week", "month", "year"
//
// 返回：
//   - 起始時間和結束時間
func GetDateRange(t time.Time, rangeType string) (time.Time, time.Time) {
	year, month, day := t.Date()
	location := t.Location()

	switch rangeType {
	case "day":
		start := time.Date(year, month, day, 0, 0, 0, 0, location)
		end := time.Date(year, month, day, 23, 59, 59, 999999999, location)
		return start, end

	case "week":
		weekday := int(t.Weekday())
		if weekday == 0 { // 調整為週一為一週的開始
			weekday = 7
		}
		start := time.Date(year, month, day-weekday+1, 0, 0, 0, 0, location)
		end := time.Date(year, month, day+(7-weekday), 23, 59, 59, 999999999, location)
		return start, end

	case "month":
		start := time.Date(year, month, 1, 0, 0, 0, 0, location)
		end := time.Date(year, month+1, 0, 23, 59, 59, 999999999, location)
		return start, end

	case "year":
		start := time.Date(year, 1, 1, 0, 0, 0, 0, location)
		end := time.Date(year, 12, 31, 23, 59, 59, 999999999, location)
		return start, end
	}

	return t, t
}

// TimeAgo 將時間轉換為人性化的相對時間描述
// 參數：
//   - t: 過去的時間
//
// 返回：
//   - 人性化的時間描述，如 "剛剛"、"5分鐘前"
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "剛剛"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分鐘前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d小時前", int(diff.Hours()))
	} else if diff < 48*time.Hour {
		return "昨天"
	} else if diff < 72*time.Hour {
		return "前天"
	} else if diff < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	} else if diff < 365*24*time.Hour {
		return fmt.Sprintf("%d個月前", int(diff.Hours()/(24*30)))
	}

	return fmt.Sprintf("%d年前", int(diff.Hours()/(24*365)))
}
