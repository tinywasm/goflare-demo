//go:build wasm

package contact

import "github.com/tinywasm/goflare/d1"

func saveSubmission(sub *Contact) error {
	db, err := d1.NewEdge("DB")
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.CreateTable(sub); err != nil {
		return err
	}
	return db.Create(sub)
}

// listSubmissions usa el helper generado por ormc + el query builder real.
func listSubmissions() (*ContactList, error) {
	db, err := d1.NewEdge("DB")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	qb := db.Query(&Contact{}).OrderBy("id").Desc()
	return ReadAllContact(qb)
}
