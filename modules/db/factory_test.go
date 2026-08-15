package db

import (
	"context"
	"testing"

	dbConf "github.com/bestHeCC/hecc-db/config"
	logConf "github.com/bestHeCC/hecc-log/config"
	dbEnum "github.com/bestHeCC/hecc-core/enum/db"
	"github.com/bestHeCC/hecc-log"

	"github.com/stretchr/testify/assert"
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
