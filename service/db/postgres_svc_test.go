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

	dataId := 0

	// 添加数据
	t.Run("add", func(t *testing.T) {
		newAccount := Account{
			AccountName: "test-add",
		}
		err = postgresSvc.Add(&newAccount)
		assert.NoError(t, err)

		dataId = newAccount.ID
	})

	// find获取多条数据
	t.Run("find", func(t *testing.T) {
		data := new(make([]Account, 0))
		err = postgresSvc.
			Where("id >= ?", dataId).
			Find(data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// find 翻页获取多条数据
	t.Run("find with page", func(t *testing.T) {
		data := new(make([]Account, 0))
		err = postgresSvc.
			Where("id >= ?", dataId).
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
			Where("id = ?", dataId).
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
			Where("id >= ?", dataId).
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
		err = postgresSvc.Where("id = ?", dataId).Save(&newAccount)
		assert.NoError(t, err)
	})

	// 统计数据
	t.Run("count", func(t *testing.T) {
		count, err := postgresSvc.Query(&Account{}).Where("id >= ?", dataId).Count()
		assert.NoError(t, err)

		t.Logf("count: %d", count)
	})

	// 排序
	t.Run("order", func(t *testing.T) {
		data := new(make([]Account, 0))
		err = postgresSvc.
			Select("id, account_name").
			Where("id >= ?", dataId).
			Order("id desc").
			Find(data)
		assert.NoError(t, err)
		assert.NotNil(t, data)

		t.Logf("data: %+v", data)
	})

	// 删除数据
	t.Run("delete", func(t *testing.T) {
		err = postgresSvc.Where("id = ?", dataId).Remove(&Account{})
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
		err = tx.Where("id = ?", newAccount.ID).Save(&updateAccount)
		assert.NoError(t, err)

		err = tx.Commit()
		assert.NoError(t, err)
	})

	t.Run("transaction with rollback", func(t *testing.T) {
		postgresSvc.WithContext(context.Background())
		tx := postgresSvc.Begin()
		newAccount := Account{
			AccountName: "test-transaction-rollback",
		}
		err = tx.Add(&newAccount)
		assert.NoError(t, err)

		tx.Rollback()
	})

	t.Run("take after save", func(t *testing.T) {
		newAccount := Account{
			AccountName: "test-take-after-save",
		}
		err = postgresSvc.Add(&newAccount)
		assert.NoError(t, err)
		t.Logf("newAccount: %+v", newAccount)

		newAccount.Password = "test-take-after-save"
		err = postgresSvc.Where("id = ?", newAccount.ID).Save(&newAccount)
		assert.NoError(t, err)
		t.Logf("newAccount: %+v", newAccount)

		data := Account{}
		err = postgresSvc.
			Where("id >= ?", newAccount.ID).
			Take(&data)
		assert.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "test-take-after-save", data.AccountName)
		assert.Equal(t, "test-take-after-save", data.Password)

		t.Logf("data: %+v", data)
	})
}
