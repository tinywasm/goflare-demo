//go:build wasm

package main

import (
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare/d1"
	"github.com/tinywasm/goflare/pages"
	"github.com/tinywasm/goflare-demo/modules/contact"
	"github.com/tinywasm/goflare-demo/routes"
)

func main() {
	db, err := d1.NewEdge("DB")
	if err != nil {
		fmt.Println("d1:", err)
		return
	}
	if err := db.CreateTable(&contact.Contact{}); err != nil {
		fmt.Println("migrate:", err)
		return
	}

	r := pages.NewRouter()
	routes.Register(r, db)
	pages.Serve(r)
}
