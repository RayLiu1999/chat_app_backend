// Package version 提供應用程式版本資訊
// 版本資訊可在編譯時透過 ldflags 注入
package version

import (
	"fmt"
	"runtime"
)

// 版本資訊變數，可在編譯時透過 -ldflags 覆蓋
var (
	// Version 應用程式版本號 (語義化版本)
	Version = "1.0.6"

	// BuildTime 編譯時間 (ISO 8601 格式)
	BuildTime = "unknown"

	// GitCommit Git 提交 SHA
	GitCommit = "unknown"

	// GitBranch Git 分支名稱
	GitBranch = "unknown"
)

// Info 版本資訊結構
type Info struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetInfo 取得完整版本資訊
func GetInfo() Info {
	return Info{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String 返回簡短版本字串
func String() string {
	return fmt.Sprintf("v%s", Version)
}

// Full 返回完整版本字串
func Full() string {
	return fmt.Sprintf("v%s (commit: %s, built: %s)", Version, GitCommit, BuildTime)
}
