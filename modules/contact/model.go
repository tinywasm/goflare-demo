//go:build wasm

package contact

// ormc:formonly
type ContactForm struct {
    Nombre  string `input:"required,min=2"`
    Email   string `input:"email,required"`
    Mensaje string `input:"textarea,required,min=10"`
}
