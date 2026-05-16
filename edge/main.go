//go:build wasm

package main

import (
	"github.com/tinywasm/goflare/workers"
	"github.com/tinywasm/json"

	"github.com/tinywasm/goflare-demo/modules/contact"
)

func main() {
	workers.Handle(handler)
}

func handler(w *workers.Response, r *workers.Request) {
	w.Header()["Access-Control-Allow-Origin"] = "*"

	if r.Method == "OPTIONS" {
		// CORS preflight — no body, no Content-Type
		w.Header()["Access-Control-Allow-Methods"] = "POST, OPTIONS"
		w.Header()["Access-Control-Allow-Headers"] = "Content-Type"
		w.WriteHeader(204)
		return
	}

	// All responses with body are JSON
	w.Header()["Content-Type"] = "application/json"

	writeJSON := func(status int, body string) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}

	if r.Method != "POST" {
		writeJSON(405, `{"error":"method not allowed"}`)
		return
	}

	var data contact.ContactForm
	if err := json.Decode(r.Body(), &data); err != nil {
		writeJSON(400, `{"error":"invalid json"}`)
		return
	}

	if err := data.Validate(0); err != nil {
		writeJSON(422, `{"error":"`+err.Error()+`"}`)
		return
	}

	// TODO: send email notification using tinywasm/fetch to Resend/Mailgun
	// fetch.Post("https://api.resend.com/emails").
	//     Header("Authorization", "Bearer "+env.Get("RESEND_API_KEY")).
	//     ContentTypeJSON().Body(emailPayload).Send(...)

	writeJSON(200, `{"message":"¡Mensaje recibido! Te contactamos pronto."}`)
}
