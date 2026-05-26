package contact

// Contact es el ÚNICO modelo del formulario de contacto: dibuja el form,
// valida la entrada y se persiste en D1. Un solo struct para minimizar binario y
// evitar estructuras duplicadas.
//
// El campo ID es seguro por construcción:
//   - tinywasm/form NO lo renderiza (omite los PK auto-increment del formulario).
//   - tinywasm/orm lo persiste y deja que D1 lo asigne (AUTOINCREMENT) cuando vale 0.
//   - el constructor NewContact fuerza ID=0, así un cliente nunca puede inyectarlo
//     vía JSON (tinywasm/json mapea por nombre de campo y "id" está en el schema).
//
// ormc:form
type Contact struct {
	ID      int    `db:"pk,autoinc"`
	Nombre  string `input:"required,min=2"`
	Email   string `input:"email,required"`
	Mensaje string `input:"textarea,required,min=10"`
}

// ormc:formonly
type EmailPayload struct {
	From    string
	To      string
	Subject string
	Html    string
}
