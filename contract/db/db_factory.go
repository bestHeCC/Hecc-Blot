package db

import (
	"context"

	dbType "core/enum/db"
)

type IDbFactory interface {
	Build(ctx context.Context, value ...dbType.Value) IDb
}
