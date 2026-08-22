//go:build wasm

package main

import (
	"github.com/tinywasm/ddl"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/goflare-demo/routes"
	"github.com/tinywasm/goflare/d1"
	"github.com/tinywasm/goflare/edge"
	"github.com/tinywasm/goflare/files"
	"github.com/tinywasm/goflare/r2"
	"github.com/tinywasm/router"
)

// filesPrefix is where the file routes hang. It MUST end in "/": files.New
// matches by prefix and the server-generated key is what hangs off it.
const filesPrefix = "/api/files/"

func main() {
	// Sin el secreto, la subida denegaría a todo el mundo. Mejor no arrancar que arrancar
	// roto en silencio.
	if err := requireToken(); err != nil {
		fmt.Println(err)
		return
	}

	db, err := d1.NewEdge("DB")
	if err != nil {
		fmt.Println("d1:", err)
		return
	}
	conn := db.RawConn()
	ddlCompiler, ok := d1.DDLCompiler(conn)
	if !ok {
		fmt.Println("migrate: d1 connection has no DDL compiler")
		return
	}
	if err := ddl.New(conn, ddlCompiler).Sync(&contact.Contact{}); err != nil {
		fmt.Println("migrate:", err)
		return
	}

	// La política es del demo: authn dice QUIÉN llama, authorize dice si PUEDE. Ver access.go.
	r := edge.NewRouter(edge.Config{Authn: authn, Authorize: authorize})
	routes.Register(r, db)

	bucket, err := r2.NewEdge("FILES")
	if err != nil {
		fmt.Println("r2:", err)
		return
	}
	store, err := files.New(bucket, filesPrefix)
	if err != nil {
		fmt.Println("files:", err)
		return
	}

	// Tracing for the file routes: goflare only logs failures, so a request that
	// goes WELL leaves no trace. These are the lines that tell that story.
	// The size comes from the header, not from len(ctx.Body()): reading the body
	// here would buffer it and defeat the lazy 413 check inside files.upload.
	r.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx router.Context) {
			if fmt.HasPrefix(ctx.Path(), filesPrefix) {
				switch ctx.Method() {
				case "PUT":
					fmt.Println("demo: upload recibido", ctx.GetHeader("Content-Length"), "bytes")
				case "GET":
					if key := ctx.Path()[len(filesPrefix):]; key != "" {
						fmt.Println("demo: sirviendo", key)
					}
				}
			}
			next(ctx)
		}
	})

	store.Mount(r)

	edge.Serve(r)
}
