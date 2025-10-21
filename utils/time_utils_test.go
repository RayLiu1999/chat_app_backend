package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatTime(t *testing.T) {
	tm := time.Date(2023, 10, 26, 15, 4, 5, 0, time.UTC)
	tests := []struct {
		name   string
		t      time.Time
		layout []string
		want   string
	}{
		{"預設版面", tm, nil, "2023-10-26 15:04:05"},
		{"自定義版面", tm, []string{time.RFC3339}, "2023-10-26T15:04:05Z"},
		{"空時間", time.Time{}, nil, "0001-01-01 00:00:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatTime(tt.t, tt.layout...))
		})
	}
}

func TestParseTime(t *testing.T) {
	expectedTime := time.Date(2023, 10, 26, 15, 4, 5, 0, time.Local)
	tests := []struct {
		name    string
		timeStr string
		layout  []string
		want    time.Time
		wantErr bool
	}{
		{"預設版面", "2023-10-26 15:04:05", nil, expectedTime, false},
		{"自定義版面", "2023-10-26T15:04:05Z", []string{time.RFC3339}, time.Date(2023, 10, 26, 15, 4, 5, 0, time.UTC), false},
		{"無效時間字符串", "not-a-time", nil, time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.timeStr, tt.layout...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// 成功解析後，檢查時間是否正確，忽略預設情況的位置
				if len(tt.layout) == 0 {
					// 比較元件，因為位置可能不同
					assert.Equal(t, tt.want.Year(), got.Year())
					assert.Equal(t, tt.want.Month(), got.Month())
					assert.Equal(t, tt.want.Day(), got.Day())
				} else {
					assert.True(t, got.Equal(tt.want))
				}
			}
		})
	}
}

func TestTimeDiff(t *testing.T) {
	t1 := time.Date(2023, 10, 26, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 10, 27, 11, 1, 30, 0, time.UTC) // 1 天，1 小時，1 分鐘，30 秒後
	tests := []struct {
		name string
		t1   time.Time
		t2   time.Time
		unit []string
		want float64
	}{
		{"預設（秒）", t1, t2, nil, 90090},
		{"秒", t1, t2, []string{"second"}, 90090},
		{"分鐘", t1, t2, []string{"minute"}, 1501.5},
		{"小時", t1, t2, []string{"hour"}, 25.025},
		{"天", t1, t2, []string{"day"}, 1.0427083333333334},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, TimeDiff(tt.t1, tt.t2, tt.unit...), 0.000000001)
		})
	}
}

func TestGetDateRange(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Taipei")
	assert.NoError(t, err)
	tm := time.Date(2023, 10, 26, 10, 0, 0, 0, loc) // A Thursday

	t.Run("day", func(t *testing.T) {
		start, end := GetDateRange(tm, "day")
		wantStart := time.Date(2023, 10, 26, 0, 0, 0, 0, loc)
		wantEnd := time.Date(2023, 10, 26, 23, 59, 59, 999999999, loc)
		assert.True(t, start.Equal(wantStart))
		assert.True(t, end.Equal(wantEnd))
	})

	t.Run("week", func(t *testing.T) {
		start, end := GetDateRange(tm, "week")
		wantStart := time.Date(2023, 10, 23, 0, 0, 0, 0, loc)          // 星期一
		wantEnd := time.Date(2023, 10, 29, 23, 59, 59, 999999999, loc) // 星期日
		assert.True(t, start.Equal(wantStart))
		assert.True(t, end.Equal(wantEnd))
	})

	t.Run("month", func(t *testing.T) {
		start, end := GetDateRange(tm, "month")
		wantStart := time.Date(2023, 10, 1, 0, 0, 0, 0, loc)
		wantEnd := time.Date(2023, 10, 31, 23, 59, 59, 999999999, loc)
		assert.True(t, start.Equal(wantStart))
		assert.True(t, end.Equal(wantEnd))
	})

	t.Run("year", func(t *testing.T) {
		start, end := GetDateRange(tm, "year")
		wantStart := time.Date(2023, 1, 1, 0, 0, 0, 0, loc)
		wantEnd := time.Date(2023, 12, 31, 23, 59, 59, 999999999, loc)
		assert.True(t, start.Equal(wantStart))
		assert.True(t, end.Equal(wantEnd))
	})
}
