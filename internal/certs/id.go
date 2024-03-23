package certs

import "github.com/google/uuid"

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
		return ID{}, err
	}
	return ID{value: parsedID.String()}, nil
}

func (id ID) String() string {
	return string(id.value)
}
