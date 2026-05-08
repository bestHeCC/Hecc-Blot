package log

type Config struct {
	RootDir    string
	ShowLine   bool
	MaxBackups int
	MaxSize    int
	MaxAge     int
	Compress   bool
}
