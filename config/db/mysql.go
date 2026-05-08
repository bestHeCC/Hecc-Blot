package db

type MysqlConfig struct {
	Ip              string
	Port            int
	Username        string
	Password        string
	DbName          string
	ConnectTimeout  int
	MaxIdleConn     int
	MaxOpenConn     int
	ConnMaxLifetime int
	SlowThreshold   int // 慢查询阈值
}
