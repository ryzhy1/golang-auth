package middlewares

import (
	"github.com/google/uuid"
)

func UUIDGenerator() (uuid.UUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return id, err
	}
	return id, nil
}
