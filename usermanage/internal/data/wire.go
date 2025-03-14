package data

import (
	"usermanage/internal/pkg/db"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(db.ProviderSet, NewUserRepo, NewRedisTokenRepo, NewData)
