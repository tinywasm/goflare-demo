package routes

import (
	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/orm"
)

func Register(r router.Router, db *orm.DB) {
	r.Post("/api/contacto", contact.Handle(db))
	r.Get("/api/contacto", contact.HandleList(db))
	r.Options("/api/contacto", contact.Handle(db)) // CORS preflight
}
