package contact

import (
	//"github.com/tinywasm/goflare/cloudflare"

	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/json"
)

func Handle(ctx router.Context) {
	ctx.SetHeader("Content-Type", "application/json")
	ctx.SetHeader("Access-Control-Allow-Origin", "*")

	if ctx.Method() == "OPTIONS" {
		ctx.SetHeader("Access-Control-Allow-Methods", "POST, OPTIONS")
		ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type")
		ctx.WriteStatus(204)
		return
	}
	if ctx.Method() != "POST" {
		ctx.WriteStatus(405)
		ctx.Write([]byte(`{"error":"method not allowed"}`))
		return
	}

	var data ContactForm
	if err := json.Decode(ctx.Body(), &data); err != nil {
		ctx.WriteStatus(400)
		ctx.Write([]byte(`{"error":"invalid json"}`))
		return
	}
	if err := data.Validate(0); err != nil {
		ctx.WriteStatus(422)
		ctx.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	sub := &ContactSubmission{Nombre: data.Nombre, Email: data.Email, Mensaje: data.Mensaje}
	if err := saveSubmission(sub); err != nil {
		ctx.WriteStatus(502)
		ctx.Write([]byte(`{"error":"db error"}`))
		return
	}

	ctx.WriteStatus(200)
	ctx.Write([]byte(`{"message":"¡Gracias! Hemos recibido tu solicitud."}`))
}
