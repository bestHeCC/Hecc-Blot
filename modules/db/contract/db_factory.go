package db

import (
	"context"

	dbEnum "github.com/bestHeCC/hecc-db/enum/db"
)

type IDbFactory interface {
	Build(ctx context.Context, value ...dbEnum.Value) IDb
	SetDefault(dbEnum.Value)
}
