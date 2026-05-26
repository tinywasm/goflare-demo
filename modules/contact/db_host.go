//go:build !wasm

package contact

import (
	"github.com/tinywasm/goflare/d1"
	"github.com/tinywasm/orm"
)

func hostDB() (*orm.DB, error) {
	return d1.NewLocal("goflare-local.db")
}

func saveSubmission(sub *Contact) error {
	db, err := hostDB()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.CreateTable(sub); err != nil {
		return err
	}
	return db.Create(sub)
}

func listSubmissions() (*ContactList, error) {
	db, err := hostDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	qb := db.Query(&Contact{}).OrderBy("id").Desc()
	return ReadAllContact(qb)
}
