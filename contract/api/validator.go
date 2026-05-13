package api

import (
	apiEntity "hecc-blot/entity/api"
)

type IValidator interface {
	GetMessages() apiEntity.Messages
}
