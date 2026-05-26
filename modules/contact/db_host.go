//go:build !wasm

package contact

import "errors"

var errHostOnly = errors.New("d1 only available in wasm")

func saveSubmission(_ *ContactSubmission) error                 { return errHostOnly }
func listSubmissions() (*ContactSubmissionList, error) { return nil, errHostOnly }
