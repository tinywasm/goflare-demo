//go:build wasm

package main

import (
	"github.com/tinywasm/dom"
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/json"

	"github.com/tinywasm/goflare-demo/modules/contact"
)

func main() {
	// Post to relative path /api/contacto (same origin)
	apiURL := "/api/contacto"

	data := &contact.ContactForm{}

	f, err := form.New("app", data)
	if err != nil {
		fmt.Println("form error:", err)
		return
	}

	f.OnSubmit(func(fielder fmt.Fielder) error {
		var body []byte
		if err := json.Encode(data, &body); err != nil {
			return err
		}

		fetch.Post(apiURL).
			ContentTypeJSON().
			Body(body).
			Send(func(resp *fetch.Response, err error) {
				if err != nil {
					dom.Render("result", dom.P("Error: "+err.Error()).Class("error-msg"))
					return
				}
				dom.Render("result", dom.P("¡Mensaje enviado!").Class("success-msg"))
			})

		return nil
	})

	container := dom.Div(f, dom.Div().ID("result"))

	if err := dom.Render("app", container); err != nil {
		fmt.Println("render error:", err)
		return
	}

	select {}
}
