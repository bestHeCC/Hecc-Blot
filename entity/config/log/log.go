package log

type Config struct {
	Local LocalConfig
	Sls   SlsConfig
}

type LocalConfig struct {
	Enable     bool
	RootDir    string
	ShowLine   bool
	MaxBackups int
	MaxSize    int
	MaxAge     int
	Compress   bool
}

type SlsConfig struct {
	Enable      bool
	Endpoint    string
	AccessKey   string
	SecretKey   string
	SecretToken string
	Project     string
	LogStore    string
}
