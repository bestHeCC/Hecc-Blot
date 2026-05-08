package cache

type IRedisCache interface {
	IBaseCache

	HSet(key string, values ...interface{}) error
	HGet(key, field string) (string, error)
	HDel(key string, fields ...string) error

	Close() error
}
