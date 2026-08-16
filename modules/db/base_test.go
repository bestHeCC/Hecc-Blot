package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestUseOtelPlugin 验证 useOtelPlugin 注册了 OpenTelemetry 追踪插件，
// 使 SQL 执行生成 span（无需真实数据库连接，gorm.Open 惰性建连）。
func TestUseOtelPlugin(t *testing.T) {
	gormDB, err := gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:3306)/core"), &gorm.Config{})
	assert.NoError(t, err)

	useOtelPlugin(gormDB)

	assert.NotNil(t, gormDB.Config.Plugins["otelgorm"])
}
