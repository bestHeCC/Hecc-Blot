package db

import (
	"context"
	"net/http"
	"testing"

	dbContract "github.com/bestHeCC/hecc-db/contract"
	dbConf "github.com/bestHeCC/hecc-db/config"
	dbEnum "github.com/bestHeCC/hecc-db/enum/db"
	logConf "github.com/bestHeCC/hecc-log/config"
	"github.com/bestHeCC/hecc-log"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

var dbConfig = &dbConf.Config{
	Mysql:    mysqlConf,
	Postgres: postgresConfig,
}

var mysqlConf = dbConf.MysqlConfig{
	Ip:              "127.0.0.1",
	Port:            3306,
	Username:        "root",
	Password:        "123456",
	DbName:          "core",
	ConnectTimeout:  3,
	MaxIdleConn:     10,
	MaxOpenConn:     200,
	ConnMaxLifetime: 3600,
	SlowThreshold:   200,
}

var postgresConfig = dbConf.PostgresConfig{
	Ip:              "127.0.0.1",
	Port:            5432,
	Username:        "admin",
	Password:        "123456",
	DbName:          "core",
	ConnectTimeout:  3,
	MaxIdleConn:     10,
	MaxOpenConn:     200,
	ConnMaxLifetime: 3600,
	SlowThreshold:   200,
}

var localConf = &logConf.Config{
	Local: logConf.LocalConfig{
		Enable:     true,
		RootDir:    "./runtime/logs",
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
	},
	Sls: logConf.SlsConfig{
		Enable: false,
	},
}

// Account 定义model
type Account struct {
	ID          int                   `json:"id" gorm:"primaryKey"`
	AccountName string                `json:"account_name"`
	Password    string                `json:"password"`
	CreatedAt   int64                 `json:"created_at"`
	UpdatedAt   int64                 `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"deleted_at"`
}

func (b Account) GetID() int {
	return b.ID
}

// TestBuildExtractsGinContext 验证 Build 会将 *gin.Context 提取为 Request.Context()。
// 若不提取，GORM otel 插件从 *gin.Context 读不到父 span（gin 的 Value 默认不委托到
// Request.Context()），SQL 会变成独立 trace。
func TestBuildExtractsGinContext(t *testing.T) {
	gormDB, err := gorm.Open(mysql.Open("root:123456@tcp(127.0.0.1:3306)/core"), &gorm.Config{})
	assert.NoError(t, err)

	f := Factory{
		db:        map[dbEnum.Value]dbContract.IDb{dbEnum.Mysql: &BaseDbSvc{db: gormDB}},
		defaultDb: dbEnum.Mysql,
	}

	// 模拟 trace 中间件：父 span 等信息存放在 Request.Context() 中
	req := (&http.Request{}).WithContext(context.WithValue(context.Background(), "trace.parent", "span"))
	ginCtx := &gin.Context{Request: req}

	clone := f.Build(ginCtx)

	stmtCtx := clone.GetInstance().(*gorm.DB).Statement.Context
	_, isGin := stmtCtx.(*gin.Context)
	assert.False(t, isGin)
	assert.Equal(t, "span", stmtCtx.Value("trace.parent"))
}

func TestFactory(t *testing.T) {
	logSvc, err := log.NewLogger(localConf)
	assert.NoError(t, err)
	assert.NotNil(t, logSvc)

	dbFactory, clearUp, err := NewDbFactory(dbConfig, logSvc)
	defer clearUp()
	assert.NoError(t, err)
	assert.NotNil(t, dbFactory)

	ctx := context.Background()
	t.Run("default mysql", func(t *testing.T) {
		newData := Account{
			AccountName: "mysql",
		}
		mysqlDB := dbFactory.Build(ctx)
		err = mysqlDB.Add(&newData)
		assert.NoError(t, err)

		data := Account{}
		err = mysqlDB.
			Where("id = ?", newData.ID).
			Take(&data)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "mysql", data.AccountName)

		t.Logf("data: %+v", data)
	})

	t.Run("default with postgres", func(t *testing.T) {
		data := Account{}
		err = dbFactory.Build(ctx, dbEnum.Postgres).
			Where("id = ?", 1).
			Take(&data)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "pg", data.AccountName)

		t.Logf("data: %+v", data)
	})

	t.Run("set postgres", func(t *testing.T) {
		dbFactory.SetDefault(dbEnum.Postgres)

		data := Account{}
		err = dbFactory.Build(ctx).
			Where("id = ?", 1).
			Take(&data)

		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "pg", data.AccountName)

		t.Logf("data: %+v", data)
	})

	t.Run("set postgres with mysql", func(t *testing.T) {
		dbFactory.SetDefault(dbEnum.Postgres)

		newData := Account{
			AccountName: "mysql",
		}
		mysqlDB := dbFactory.Build(ctx, dbEnum.Mysql)
		err = mysqlDB.Add(&newData)
		assert.NoError(t, err)

		data := Account{}
		err = mysqlDB.
			Where("id = ?", newData.ID).
			Take(&data)

		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "mysql", data.AccountName)

		t.Logf("data: %+v", data)
	})
}
