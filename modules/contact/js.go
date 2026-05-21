//go:build !wasm

package contact

import (
	"github.com/tinywasm/js"
)

// RenderJS provides optional client-side scripts to be included in the page.
// In this example, it could register a service worker or include analytics.
func (c *ContactForm) RenderJS() []*js.Script {
	return []*js.Script{
		// Example: register a service worker for PWA capabilities.
		// js.ServiceWorker("sw.js", &MyServiceWorker{}),
	}
}
