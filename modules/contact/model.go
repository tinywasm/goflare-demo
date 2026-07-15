package contact

import (
	"github.com/tinywasm/form/input"
	"github.com/tinywasm/model"
)

// ContactModel es la ÚNICA fuente de verdad del formulario de contacto: dibuja el
// form, valida la entrada y se persiste en D1. ormc genera desde aquí el struct
// Contact y toda su fontanería (Schema/Pointers/codec/Validate/ContactList).
//
// El Kind del campo (el slot `Type:`) es lo único que decide validación y widget:
//   - input.X() valida Y se renderiza como input en el formulario.
//   - model.X() solo valida; no aparece en el formulario.
//
// Por eso el PK lleva model.Int() y no un input: es seguro por construcción —
// tinywasm/form no le dibuja campo, tinywasm/orm deja que D1 lo asigne
// (AUTOINCREMENT) cuando vale 0, y NewContact fuerza ID=0 para que un cliente
// jamás pueda inyectarlo vía JSON.
var ContactModel = model.Definition{
	Name: "contact",
	Fields: model.Fields{
		{Name: "id", Type: model.Int(), DB: &model.FieldDB{PK: true, AutoInc: true}},
		{Name: "nombre", Type: input.Text(), NotNull: true, Permitted: model.Permitted{Minimum: 2}},
		{Name: "email", Type: input.Email(), NotNull: true},
		{Name: "mensaje", Type: input.Textarea(), NotNull: true, Permitted: model.Permitted{Minimum: 10}},
	},
}

// EmailPayloadModel es un DTO de transporte hacia la API de Resend: no lleva
// metadatos DB (no es una tabla) ni inputs (no se dibuja). Existe como Definition
// porque tinywasm/json solo serializa a través del codec generado: un DTO escrito
// a mano no puede viajar.
//
// Html es model.Raw() a propósito: es marcado que ensamblamos nosotros —con la
// entrada del usuario ya escapada en sendEmail—, no entrada de red. El whitelist
// de model.Text() excluye <> y rechazaría nuestro propio HTML.
var EmailPayloadModel = model.Definition{
	Name: "email_payload",
	Fields: model.Fields{
		{Name: "from", Type: model.Text()},
		{Name: "to", Type: model.Text()},
		{Name: "subject", Type: model.Text()},
		{Name: "html", Type: model.Raw()},
	},
}
