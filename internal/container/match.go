package container

// MatchStatus 表示应用匹配结果。
type MatchStatus uint8

const (
	// MatchNotFound 表示没有应用匹配请求路径。
	MatchNotFound MatchStatus = iota
	// MatchFound 表示找到匹配应用。
	MatchFound
	// MatchUnavailable 表示容器尚未启动或已关闭。
	MatchUnavailable
)
