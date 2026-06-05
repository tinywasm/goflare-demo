//go:build wasm

package main

import (
	. "github.com/tinywasm/dom"
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/json"

	"github.com/tinywasm/goflare-demo/modules/contact"
)

func main() {
	// API endpoint for both POST and GET
	apiURL := "/api/contacto"

	data := &contact.Contact{}

	f, err := form.New("app", data)
	if err != nil {
		fmt.Println("form error:", err)
		return
	}

	renderList := func() {
		fetch.Get(apiURL).Send(func(resp *fetch.Response, err error) {
			if err != nil {
				fmt.Println("fetch list error:", err)
				return
			}
			var list contact.ContactList
			if err := json.Decode(resp.Body(), &list); err != nil {
				fmt.Println("decode list error:", err)
				return
			}

			items := []Component{}
			for _, sub := range list {
				// Partially hide email (e.g. ci***@test.com)
				emailParts := fmt.Split(sub.Email, "@")
				hiddenEmail := sub.Email
				if len(emailParts) == 2 {
					prefix := emailParts[0]
					if len(prefix) > 2 {
						hiddenEmail = prefix[:2] + "***@" + emailParts[1]
					} else {
						hiddenEmail = prefix + "***@" + emailParts[1]
					}
				}

				// First 60 chars of message
				shortMsg := sub.Mensaje
				if len(shortMsg) > 60 {
					shortMsg = shortMsg[:57] + "..."
				}

				items = append(items, Div(
					Strong(sub.Nombre),
					Span(" ("+hiddenEmail+"): "),
					Span(shortMsg),
				).Class("submission-item"))
			}

			listItems := make([]any, len(items))
			for i, v := range items {
				listItems[i] = v
			}

			Render("submissions", Div(
				H3(fmt.Convert(len(list)).String()+" solicitudes recibidas"),
				Div(listItems...),
			))
		})
	}

	f.OnSubmit(func(fielder fmt.Fielder, done func(error)) {
		var body []byte
		if err := json.Encode(data, &body); err != nil {
			done(err)
			return
		}

		fetch.Post(apiURL).
			ContentTypeJSON().
			Body(body).
			Send(func(resp *fetch.Response, err error) {
				if err != nil {
					Render("result", P("Error: "+err.Error()).Class("error-msg"))
					done(err)
					return
				}
				Render("result", P("¡Mensaje enviado!").Class("success-msg"))
				renderList()
				done(nil)
			})
	})

	container := Div(
		f,
		Div().ID("result"),
		Hr(),
		Div().ID("submissions"),
	)

	if err := Render("app", container); err != nil {
		fmt.Println("render error:", err)
		return
	}

	renderList()

	select {}
}
