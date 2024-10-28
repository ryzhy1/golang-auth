package middlewares

import (
	"github.com/google/uuid"
)

func UUIDGenerator() (uuid.UUID, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return uid, err
	}
	return uid, nil
}
