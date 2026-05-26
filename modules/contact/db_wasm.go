//go:build wasm

package contact

import "github.com/tinywasm/goflare/d1"

func saveSubmission(sub *ContactSubmission) error {
	db, err := d1.New("DB")
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
func listSubmissions() (*ContactSubmissionList, error) {
	db, err := d1.New("DB")
	if err != nil {
		return nil, err
	}
	defer db.Close()
	qb := db.Query(&ContactSubmission{}).OrderBy("id").Desc()
	return ReadAllContactSubmission(qb)
}
