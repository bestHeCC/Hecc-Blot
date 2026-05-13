package api

import (
	apiEntity "core/entity/api"
)

type IValidator interface {
	GetMessages() apiEntity.Messages
}
