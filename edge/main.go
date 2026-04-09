//go:build wasm

package main

import (
    "github.com/tinywasm/fmt"
    "github.com/tinywasm/goflare/workers"
    "github.com/tinywasm/json"
)

type contactPayload struct {
    Nombre  string
    Email   string
    Mensaje string
}

func (p *contactPayload) Schema() []fmt.Field {
    return []fmt.Field{
        {Name: "nombre",  Type: fmt.FieldText},
        {Name: "email",   Type: fmt.FieldText},
        {Name: "mensaje", Type: fmt.FieldText},
    }
}

func (p *contactPayload) Pointers() []any {
    return []any{&p.Nombre, &p.Email, &p.Mensaje}
}

type jsonMsg struct {
    Key   string
    Value string
}

func (m *jsonMsg) Schema() []fmt.Field {
    return []fmt.Field{{Name: m.Key, Type: fmt.FieldText}}
}

func (m *jsonMsg) Pointers() []any { return []any{&m.Value} }

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

    writeJSON := func(status int, key, value string) {
        w.WriteHeader(status)
        json.Encode(&jsonMsg{Key: key, Value: value}, w)
    }

    if r.Method != "POST" {
        writeJSON(405, "error", "method not allowed")
        return
    }

    var p contactPayload
    if err := json.Decode(r.Body(), &p); err != nil {
        writeJSON(400, "error", "invalid json")
        return
    }

    if err := validateContact(p); err != nil {
        writeJSON(422, "error", err.Error())
        return
    }

    // TODO: send email notification using tinywasm/fetch to Resend/Mailgun
    // fetch.Post("https://api.resend.com/emails").
    //     Header("Authorization", "Bearer "+env.Get("RESEND_API_KEY")).
    //     ContentTypeJSON().Body(emailPayload).Send(...)

    writeJSON(200, "message", "¡Mensaje recibido! Te contactamos pronto.")
}

func validateContact(p contactPayload) error {
    if len(p.Nombre) < 2 {
        return fmt.Errf("nombre requerido")
    }
    if !containsAt(p.Email) {
        return fmt.Errf("email inválido")
    }
    if len(p.Mensaje) < 10 {
        return fmt.Errf("mensaje muy corto")
    }
    return nil
}

func containsAt(s string) bool {
    for _, c := range s {
        if c == '@' {
            return true
        }
    }
    return false
}
