package negocio

import "github.com/google/uuid"

func setID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}
