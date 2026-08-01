package family

import "strings"

// Family es la entidad de dominio.
// El ID corresponde al ID del documento en Firestore, por eso se marca
// con firestore:"-" (no se persiste como campo, vive en la ruta del doc).
type Family struct {
	ID   string `json:"id" firestore:"-"`
	Name string `json:"name" firestore:"name"`
}

// Validate aplica las reglas de negocio de la entidad.
func (f *Family) Validate() error {
	if strings.TrimSpace(f.Name) == "" {
		return ErrInvalidName
	}
	return nil
}
