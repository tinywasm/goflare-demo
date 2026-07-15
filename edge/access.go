//go:build wasm

package main

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare/cloudflare"
	"github.com/tinywasm/goflare/files"
	"github.com/tinywasm/model"
	"github.com/tinywasm/router"
)

// La política de acceso es del DEMO, no de goflare.
//
// La librería aporta el mecanismo y hace la pregunta —¿quién llama? ¿puede?—; responderla es
// trabajo de la app. goflare/files monta la subida como ruta con permiso `files`/Create, y
// hace bien: un bucket abierto a escritura es un formulario de spam. Lo que faltaba no era
// abrir la ruta, era que el borde supiera identificar al llamante.

const (
	// DemoTokenHeader es la cabecera con la que el demo reconoce a su propio subidor.
	DemoTokenHeader = "X-Demo-Token"

	// DemoUploaderID es la única identidad que este demo conoce.
	DemoUploaderID = "demo-uploader"

	// DemoTokenBinding es el secreto en Cloudflare (wrangler secret / dashboard).
	DemoTokenBinding = "DEMO_UPLOAD_TOKEN"
)

// demoToken lee el secreto del entorno del Worker. Nunca va en el código: un token en el
// binario es un token público.
func demoToken() string { return cloudflare.Env(DemoTokenBinding) }

// authn establece la identidad a partir de la petición. Anónimo ("") es un resultado legal,
// no un error: las rutas de contacto son públicas y no necesitan a nadie detrás.
//
// Corre ANTES de la verja. Al revés, la verja leería un UserID que todavía no ha escrito
// nadie y ningún llamante podría autorizarse jamás.
func authn(next router.HandlerFunc) router.HandlerFunc {
	return func(ctx router.Context) {
		if tok := demoToken(); tok != "" && ctx.GetHeader(DemoTokenHeader) == tok {
			ctx.SetUserID(DemoUploaderID)
		}
		next(ctx)
	}
}

// authorize concede al subidor del demo exactamente un permiso: crear archivos. Todo lo demás
// se deniega — incluida la lectura, que no lo necesita: servir es público.
func authorize(userID string, r model.Resource, a model.Action) bool {
	return userID == DemoUploaderID && r == files.Resource && a.Has(files.Action)
}

// requireToken falla RUIDOSAMENTE si el secreto no está configurado. Arrancar sin él dejaría
// la subida denegando a todo el mundo en silencio — justo el fallo que este trabajo existe
// para eliminar.
func requireToken() error {
	if demoToken() == "" {
		return fmt.Err("demo:", DemoTokenBinding, "no está configurado: la subida denegaría a todo el mundo")
	}
	return nil
}
