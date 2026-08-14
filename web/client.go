//go:build wasm

package main

import "github.com/tinywasm/model"

import (
	. "github.com/tinywasm/dom"
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	. "github.com/tinywasm/html"
	"github.com/tinywasm/json"
	"github.com/tinywasm/unixid"
	"syscall/js"

	"github.com/tinywasm/goflare-demo/modules/contact"
)

func main() {
	// API endpoints
	apiURL := "/api/contacto"
	filesURL := "/api/files/"

	data := &contact.Contact{}

	ids, err := unixid.NewUnixID()
	if err != nil {
		fmt.Println("unixid error:", err)
		return
	}

	f, err := form.New("app", data, ids)
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

				items = append(items, Div().Class("submission-item").Child(
					Strong().Text(sub.Nombre),
					Span().Text(" ("+hiddenEmail+"): "),
					Span().Text(shortMsg),
				))
			}

			Render("submissions", Div().Child(
				H3().Text(fmt.Convert(len(list)).String()+" solicitudes recibidas"),
				Div().Child(items...),
			))
		})
	}

	f.OnSubmit(func(fielder model.Fielder, done func(error)) {
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
					Render("result", P().Class("error-msg").Text("Error: "+err.Error()))
					done(err)
					return
				}
				Render("result", P().Class("success-msg").Text("¡Mensaje enviado!"))
				renderList()
				done(nil)
			})
	})

	// File Upload Logic
	fileInput := Input("file").ID("file-upload").Attr("accept", "image/*")
	uploadBtn := Button().ID("upload-btn").Text("Subir Imagen")
	preview := Div().ID("preview")

	uploadBtn.On("click", func(e Event) {
		input := js.Global().Get("document").Call("getElementById", "file-upload")
		files := input.Get("files")
		if files.Length() == 0 {
			Render("preview", P().Class("error-msg").Text("Selecciona una imagen primero"))
			return
		}
		file := files.Index(0)

		Render("preview", P().Text("Subiendo..."))

		file.Call("arrayBuffer").Call("then", js.FuncOf(func(this js.Value, args []js.Value) any {
			arrayBuffer := args[0]
			uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
			data := make([]byte, uint8Array.Length())
			js.CopyBytesToGo(data, uint8Array)

			fetch.Put(filesURL).
				Body(data).
				Send(func(resp *fetch.Response, err error) {
					if err != nil {
						Render("preview", P().Class("error-msg").Text("Error subiendo: "+err.Error()))
						return
					}
					if resp.Status != 201 {
						Render("preview", P().Class("error-msg").Text("Error subiendo: "+fmt.Convert(resp.Status).String()+" "+string(resp.Body())))
						return
					}

					key := string(resp.Body())
					Render("preview", Div().Child(
						P().Text("Imagen subida con éxito: "+key),
						NewElement("img").NoCloseTag().Attr("src", filesURL+key).Attr("style", "max-width: 300px; display: block; margin-top: 10px;"),
					))
				})
			return nil
		}))
	})

	container := Div().Child(
		H2().Text("Contacto"),
		f,
		Div().ID("result"),
		Hr(),
		H2().Text("Subida de Imágenes (R2)"),
		Div().Class("upload-section").Child(
			fileInput,
			uploadBtn,
			preview,
		),
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
