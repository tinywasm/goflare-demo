package contact

import (
	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/orm"
)

func Handle(db *orm.DB) router.HandlerFunc {
	return func(ctx router.Context) {
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

		// NewContact decodifica + valida + fuerza ID=0 (seguro por construcción).
		sub, err := NewContact(ctx.Body())
		if err != nil {
			ctx.WriteStatus(422)
			ctx.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}

		if err := db.Create(sub); err != nil {
			ctx.WriteStatus(502)
			ctx.Write([]byte(`{"error":"db error"}`))
			return
		}

		ctx.WriteStatus(200)
		ctx.Write([]byte(`{"message":"¡Gracias! Hemos recibido tu solicitud."}`))
	}
}
