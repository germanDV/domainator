package common

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrInvalidID = errors.New("invalid ID")

type ID struct {
	value string
}

func NewID() ID {
	id, _ := uuid.NewV7()
	return ID{value: id.String()}
}

func ParseID(id string) (ID, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return ID{}, fmt.Errorf("error parsing id %s: %w", id, ErrInvalidID)
	}
	return ID{value: parsedID.String()}, nil
}

func (id ID) String() string {
	return id.value
}
