//go:build wasm

package main

import (
	"syscall/js"

	"github.com/tinywasm/dom"
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/form"
	"github.com/tinywasm/json"

	"github.com/tinywasm/goflare-demo/modules/contact"
)

func main() {
	// Worker URL set in index.html as window.WORKER_URL.
	// Fallback to "/contact" for local development with proxy.
	workerURL := js.Global().Get("WORKER_URL").String()
	if workerURL == "" {
		workerURL = "/contact"
	}

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

		fetch.Post(workerURL).
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

	if err := dom.Render("app", f); err != nil {
		fmt.Println("render error:", err)
		return
	}

	select {}
}
