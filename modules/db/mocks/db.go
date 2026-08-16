package mocks

import (
	"context"

	dbEnum "github.com/bestHeCC/hecc-db/enum/db"
	dbContract "github.com/bestHeCC/hecc-db/contract"
)

// MockDbFactory 是 IDbFactory 接口的 mock 实现，可通过 BuildFn 定制返回的 IDb。
type MockDbFactory struct {
	BuildFn func(ctx context.Context, value ...dbEnum.Value) dbContract.IDb
	Default dbEnum.Value
}

func (m *MockDbFactory) Build(ctx context.Context, value ...dbEnum.Value) dbContract.IDb {
	if m.BuildFn != nil {
		return m.BuildFn(ctx, value...)
	}
	return &MockDb{}
}

func (m *MockDbFactory) SetDefault(v dbEnum.Value) {
	m.Default = v
}

// MockDb 是 IDb 接口的 mock 实现。
// 嵌入 IDb 接口，未显式覆盖的方法会 panic（调用未 mock 的方法即为测试错误）。
type MockDb struct {
	dbContract.IDb
	AddFn    func(entry dbContract.IDbModel) error
	CountFn  func() (int64, error)
	FindFn   func(dst any) error
	TakeFn   func(dst any) error
	WhereFn  func(args ...any) dbContract.IDb
	QueryFn  func(entry dbContract.IDbModel) dbContract.IDb
	SaveFn   func(entry dbContract.IDbModel) error
	RemoveFn func(entry dbContract.IDbModel) error
	BeginFn  func() dbContract.IDb
	CommitFn func() error
}

func (m *MockDb) Add(entry dbContract.IDbModel) error { return m.AddFn(entry) }

func (m *MockDb) Count() (int64, error) { return m.CountFn() }

func (m *MockDb) Find(dst any) error { return m.FindFn(dst) }

func (m *MockDb) Take(dst any) error { return m.TakeFn(dst) }

func (m *MockDb) Where(args ...any) dbContract.IDb { return m.WhereFn(args...) }

func (m *MockDb) Query(entry dbContract.IDbModel) dbContract.IDb { return m.QueryFn(entry) }

func (m *MockDb) Save(entry dbContract.IDbModel) error { return m.SaveFn(entry) }

func (m *MockDb) Remove(entry dbContract.IDbModel) error { return m.RemoveFn(entry) }

func (m *MockDb) Begin() dbContract.IDb { return m.BeginFn() }

func (m *MockDb) Commit() error { return m.CommitFn() }
