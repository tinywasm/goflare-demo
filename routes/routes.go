package routes

import (
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/orm"
	"github.com/tinywasm/router"
)

// Register mounts the contact API.
//
// El .Public() es obligatorio y deliberado: el contrato es PRIVADO POR DEFECTO. Una ruta que
// no declara acceso ni siquiera arranca — el enforcer la rechaza al iniciar, en vez de
// responder 403 en silencio en producción. El demo no tiene login, así que sus rutas de
// contacto son públicas de facto; hay que DECIRLO.
//
// Tres rutas sobre el mismo path, una por método: es lo que ofrece el contrato, y ahora las
// dos implementaciones lo cumplen. Esto solía hacer panic en el servidor nativo —registraba
// en el ServeMux sin el método, así que las tres colapsaban en el mismo patrón— y el demo lo
// rodeaba con un Handle("", ...) y un if ctx.Method() a mano. Ese rodeo ya no existe.
func Register(r router.Router, db *orm.DB) {
	r.Post("/api/contacto", contact.Handle(db)).Public()
	r.Get("/api/contacto", contact.HandleList(db)).Public()
	r.Options("/api/contacto", contact.Handle(db)).Public()
}
