package users

import "github.com/google/uuid"

// TODO: this is the same ID we use in package certs, make it common.
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
	return id.value
}
