//go:build !wasm

package contact

import "github.com/tinywasm/fmt"

var errHostOnly = fmt.Err("d1 only available in wasm")

func saveSubmission(_ *Contact) error        { return errHostOnly }
func listSubmissions() (*ContactList, error) { return nil, errHostOnly }
