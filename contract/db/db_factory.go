package db

import (
	"context"

	dbEnum "hecc-blot/enum/db"
)

type IDbFactory interface {
	Build(ctx context.Context, value ...dbEnum.Value) IDb
	SetDefault(dbEnum.Value)
}
