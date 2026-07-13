package routes

import (
	"github.com/tinywasm/router"
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/orm"
)

func Register(r router.Router, db *orm.DB) {
	r.Handle("", "/api/contacto", func(ctx router.Context) {
		if ctx.Method() == "GET" {
			contact.HandleList(db)(ctx)
		} else {
			contact.Handle(db)(ctx)
		}
	}).Public()
}
