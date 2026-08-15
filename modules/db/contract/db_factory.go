package db

import (
	"context"

	dbEnum "github.com/bestHeCC/hecc-core/enum/db"
)

type IDbFactory interface {
	Build(ctx context.Context, value ...dbEnum.Value) IDb
	SetDefault(dbEnum.Value)
}
