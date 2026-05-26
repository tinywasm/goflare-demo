package routes

import (
	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/goflare-demo/modules/contact"
)

func Register(r router.Router) {
	r.Post("/api/contacto", contact.Handle)
	r.Get("/api/contacto", contact.HandleList)
	r.Options("/api/contacto", contact.Handle) // CORS preflight
}
