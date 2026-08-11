//go:build !wasm

package contact

import (
	. "github.com/tinywasm/css"
)

// RenderCSS produces the styles for the contact form using design tokens
// from tinywasm/css (theme-aware, light/dark mode automatic).
//
// This method is //go:build !wasm because tinywasm/css is meant for SSR;
// the WASM frontend never includes this code so the binary stays minimal.
func (c *Contact) RenderCSS() *Stylesheet {
	return NewStylesheet(
		// Page shell.
		Raw("body { margin: 0; background-color: " + ColorBackground.Var() +
			"; color: " + ColorOnSurface.Var() +
			"; font-family: system-ui, -apple-system, Segoe UI, Roboto, sans-serif" +
			"; line-height: 1.6; min-height: 100vh; display: flex; align-items: center; justify-content: center; }"),

		// Card container.
		Raw("#app { max-width: 28rem; width: 100%; margin: 2rem auto; padding: 2.5rem; " +
			"background-color: " + ColorSurface.Var() + "; border-radius: 1rem; " +
			"box-shadow: 0 4px 24px rgba(0,0,0,0.10), 0 1px 4px rgba(0,0,0,0.06); box-sizing: border-box; }"),

		// Card heading.
		Raw("#app::before { content: \"Contacto\"; display: block; font-size: 1.6rem; font-weight: 700; " +
			"margin-bottom: 0.25rem; color: " + ColorPrimary.Var() + "; letter-spacing: -0.02em; }"),
		Raw("#app::after { content: \"Completá el formulario y te respondemos a la brevedad.\"; display: block; " +
			"font-size: 0.9rem; color: " + ColorMuted.Var() + "; margin-bottom: 1.5rem; }"),

		// Form layout: vertical stack with spacing.
		Raw("form { display: flex; flex-direction: column; gap: 1.1rem; }"),

		// Labels above fields.
		Raw("form label { display: block; margin-bottom: 0.3rem; font-weight: 600; font-size: 0.875rem; " +
			"color: " + ColorOnSurface.Var() + "; letter-spacing: 0.01em; }"),

		// Inputs and textarea.
		Raw("form input, form textarea { width: 100%; padding: 0.7rem 0.9rem; border-radius: 0.5rem; " +
			"border: 1.5px solid " + ColorMuted.Var() + "; background-color: " + ColorBackground.Var() +
			"; color: " + ColorOnSurface.Var() + "; font-family: inherit; font-size: 1rem; outline: none; " +
			"box-sizing: border-box; transition: border-color 0.18s, box-shadow 0.18s; }"),
		Raw("form input::placeholder, form textarea::placeholder { color: " + ColorMuted.Var() + "; font-size: 0.93rem; }"),
		Raw("form input:focus, form textarea:focus { border-color: " + ColorPrimary.Var() +
			"; box-shadow: 0 0 0 3px " + ColorPrimary.Var() + "28; }"),
		Raw("form textarea { min-height: 8rem; resize: vertical; line-height: 1.6; }"),

		// Submit button: primary action.
		Raw(`form button[type="submit"] { width: 100%; background-color: ` + ColorPrimary.Var() +
			"; color: " + ColorOnPrimary.Var() + "; border: none; border-radius: 0.5rem; " +
			"padding: 0.8rem 1.5rem; font-size: 1rem; font-weight: 700; cursor: pointer; " +
			"letter-spacing: 0.02em; transition: background-color 0.18s, transform 0.1s, box-shadow 0.18s; " +
			"margin-top: 0.4rem; box-shadow: 0 2px 8px " + ColorPrimary.Var() + "44; }"),
		Raw(`form button[type="submit"]:hover { background-color: ` + Hover(ColorPrimary) +
			"; box-shadow: 0 4px 16px " + Hover(ColorPrimary) + "44; }"),
		Raw(`form button[type="submit"]:active { transform: scale(0.98); }`),

		// Result panel.
		Raw("#result { margin-top: 0.75rem; }"),
		Raw(".success-msg { color: " + ColorPrimary.Var() + "; padding: 0.75rem 1rem; border-radius: 0.5rem; " +
			"background: " + ColorPrimary.Var() + "18; font-weight: 600; font-size: 0.95rem; }"),
		Raw(".error-msg { color: #d63031; padding: 0.75rem 1rem; border-radius: 0.5rem; background: #d6303118; " +
			"font-weight: 600; font-size: 0.95rem; }"),
	)
}
