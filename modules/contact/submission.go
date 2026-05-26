package contact

import "github.com/tinywasm/json"

// NewContact construye un Contact desde el JSON no confiable de una
// petición. Fuerza ID=0 para que el cliente NUNCA pueda fijar la primary key — D1 la
// asigna vía AUTOINCREMENT. Es la única vía sancionada para crear una submission desde
// la red: seguro por construcción (guard contra mass-assignment).
func NewContact(body any) (*Contact, error) {
	s := &Contact{}
	if err := json.Decode(body, s); err != nil {
		return nil, err
	}
	s.ID = 0 // ignora cualquier id provisto por el cliente
	if err := s.Validate('c'); err != nil {
		return nil, err
	}
	return s, nil
}
