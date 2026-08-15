package api

import (
	apiEntity "github.com/bestHeCC/hecc-core/entity/api"
)

type IValidator interface {
	GetMessages() apiEntity.Messages
}
