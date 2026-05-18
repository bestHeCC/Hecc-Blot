package db

import (
	"context"
	"testing"

	"hecc-blot/service/log"

	"github.com/stretchr/testify/assert"
)

func TestPostgresSvc(t *testing.T) {
	logSvc, err := log.NewLogger(localConf)
	assert.NoError(t, err)
	assert.NotNil(t, logSvc)

	postgresSvc, clearUp, err := newPostgresSvc(&postgresConfig, logSvc)
	assert.NoError(t, err)
	assert.NotNil(t, postgresSvc)

	defer func() {
		clearUp()
	}()

	// 添加数据
	t.Run("add", func(t *testing.T) {
		newAccount := Account{
			AccountName: "test-add",
		}
		err = postgresSvc.Add(&newAccount)
		assert.NoError(t, err)
	})

	// find获取多条数据
	t.Run("find", func(t *testing.T) {
		data := new(make([]Account, 0))
		err = postgresSvc.
			Where("id >= ? and id <= ?", 1, 8).
			Find(data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// find 翻页获取多条数据
	t.Run("find with page", func(t *testing.T) {
		data := new(make([]Account, 0))
		err = postgresSvc.
			Where("id >= ? and id <= ?", 1, 8).
			Offset(0).
			Limit(2).
			Find(data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// take获取一条数据
	t.Run("take", func(t *testing.T) {
		data := Account{}
		err = postgresSvc.
			Where("id = ?", 1).
			Take(&data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// 查询指定字段
	t.Run("select", func(t *testing.T) {
		data := Account{}
		err = postgresSvc.
			Select("id, account_name").
			Where("id >= ? and id <= ?", 1, 8).
			Take(&data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// 更新数据
	t.Run("update", func(t *testing.T) {
		newAccount := Account{
			AccountName: "test-update",
		}
		err = postgresSvc.Where("id = ?", 1).Save(&newAccount)
		assert.NoError(t, err)
	})

	// 统计数据
	t.Run("count", func(t *testing.T) {
		count, err := postgresSvc.Query(&Account{}).Where("id >= ?", 1).Count()
		assert.NoError(t, err)

		t.Logf("count: %d", count)
	})

	// 排序
	t.Run("order", func(t *testing.T) {
		data := new(make([]Account, 0))
		err = postgresSvc.
			Select("id, account_name").
			Where("id >= ? and id <= ?", 1, 8).
			Order("id desc").
			Find(data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// 删除数据
	t.Run("delete", func(t *testing.T) {
		err = postgresSvc.Where("id = ?", 1).Remove(&Account{})
		assert.NoError(t, err)
	})

	t.Run("transaction", func(t *testing.T) {
		postgresSvc.WithContext(context.Background())
		tx := postgresSvc.Begin()
		newAccount := Account{
			AccountName: "test-transaction",
		}
		err = tx.Add(&newAccount)
		assert.NoError(t, err)

		updateAccount := Account{
			Password: "update-transaction",
		}
		err = postgresSvc.Where("id = ?", newAccount.ID).Save(&updateAccount)
		assert.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)
	})

	t.Run("transaction with rollback", func(t *testing.T) {
		postgresSvc.WithContext(context.Background())
		postgresSvc.Begin()
		newAccount := Account{
			AccountName: "test-transaction-rollback",
		}
		err = postgresSvc.Add(&newAccount)
		assert.NoError(t, err)

		postgresSvc.Rollback()
	})

}
