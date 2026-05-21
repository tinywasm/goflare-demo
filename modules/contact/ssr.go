//go:build !wasm

package contact

import (
	. "github.com/tinywasm/css"
)

// SSRInstance is the assetmin discovery entry point — see assetmin/ssr_invoke.go.
// Returns a zero ContactForm so RenderCSS can be invoked at build time.
func SSRInstance() *ContactForm { return &ContactForm{} }

// RenderCSS produces the styles for the contact form using design tokens
// from tinywasm/css (theme-aware, light/dark mode automatic).
//
// This method is //go:build !wasm because tinywasm/css is meant for SSR;
// the WASM frontend never includes this code so the binary stays minimal.
func (c *ContactForm) RenderCSS() *Stylesheet {
	return NewStylesheet(
		// Page shell.
		Rule("body",
			Margin(Zero),
			BackgroundColor(ColorBackground),
			Color(ColorOnSurface),
			FontFamily(Str("system-ui, -apple-system, Segoe UI, Roboto, sans-serif")),
			Decl{Prop: "line-height", Val: "1.6"},
			Decl{Prop: "min-height", Val: "100vh"},
			Decl{Prop: "display", Val: "flex"},
			Decl{Prop: "align-items", Val: "center"},
			Decl{Prop: "justify-content", Val: "center"},
		),

		// Card container.
		Rule("#app",
			MaxWidth(Rem(28)),
			Width(Pct(100)),
			Margin(Rem(2), Auto),
			Padding(Rem(2.5)),
			BackgroundColor(ColorSurface),
			BorderRadius(Rem(1)),
			Decl{Prop: "box-shadow", Val: "0 4px 24px rgba(0,0,0,0.10), 0 1px 4px rgba(0,0,0,0.06)"},
			BoxSizing(Str("border-box")),
		),

		// Card heading.
		Rule("#app::before",
			Decl{Prop: "content", Val: `"Contacto"`},
			Display(Block),
			FontSize(Rem(1.6)),
			FontWeight(Str("700")),
			Decl{Prop: "margin-bottom", Val: "0.25rem"},
			Color(ColorPrimary),
			Decl{Prop: "letter-spacing", Val: "-0.02em"},
		),
		Rule("#app::after",
			Decl{Prop: "content", Val: `"Completá el formulario y te respondemos a la brevedad."`},
			Display(Block),
			FontSize(Rem(0.9)),
			Color(ColorMuted),
			Decl{Prop: "margin-bottom", Val: "1.5rem"},
		),

		// Form layout: vertical stack with spacing.
		Rule("form",
			Display(Flex_),
			FlexDirection(Column),
			Gap(Rem(1.1)),
		),

		// Labels above fields.
		Rule("form label",
			Display(Block),
			Decl{Prop: "margin-bottom", Val: "0.3rem"},
			FontWeight(Str("600")),
			FontSize(Rem(0.875)),
			Color(ColorOnSurface),
			Decl{Prop: "letter-spacing", Val: "0.01em"},
		),

		// Inputs and textarea.
		Rule("form input, form textarea",
			Width(Pct(100)),
			Padding(Rem(0.7), Rem(0.9)),
			BorderRadius(Rem(0.5)),
			Decl{Prop: "border", Val: "1.5px solid " + ColorMuted.Var()},
			BackgroundColor(ColorBackground),
			Color(ColorOnSurface),
			FontFamily(Str("inherit")),
			FontSize(Rem(1)),
			Outline(None),
			BoxSizing(Str("border-box")),
			Decl{Prop: "transition", Val: "border-color 0.18s, box-shadow 0.18s"},
		),
		Rule("form input::placeholder, form textarea::placeholder",
			Color(ColorMuted),
			FontSize(Rem(0.93)),
		),
		Rule("form input:focus, form textarea:focus",
			BorderColor(ColorPrimary),
			Decl{Prop: "box-shadow", Val: "0 0 0 3px " + ColorPrimary.Var() + "28"},
		),
		Rule("form textarea",
			MinHeight(Rem(8)),
			Decl{Prop: "resize", Val: "vertical"},
			Decl{Prop: "line-height", Val: "1.6"},
		),

		// Submit button: primary action.
		Rule(`form button[type="submit"]`,
			Width(Pct(100)),
			BackgroundColor(ColorPrimary),
			Color(ColorOnPrimary),
			Border(None),
			BorderRadius(Rem(0.5)),
			Padding(Rem(0.8), Rem(1.5)),
			FontSize(Rem(1)),
			FontWeight(Str("700")),
			Cursor(Pointer),
			Decl{Prop: "letter-spacing", Val: "0.02em"},
			Decl{Prop: "transition", Val: "background-color 0.18s, transform 0.1s, box-shadow 0.18s"},
			Decl{Prop: "margin-top", Val: "0.4rem"},
			Decl{Prop: "box-shadow", Val: "0 2px 8px " + ColorPrimary.Var() + "44"},
		),
		Rule(`form button[type="submit"]:hover`,
			BackgroundColor(ColorSecondary),
			Decl{Prop: "box-shadow", Val: "0 4px 16px " + ColorSecondary.Var() + "44"},
		),
		Rule(`form button[type="submit"]:active`,
			Decl{Prop: "transform", Val: "scale(0.98)"},
		),

		// Result panel.
		Rule("#result",
			Decl{Prop: "margin-top", Val: "0.75rem"},
		),
		Rule(".success-msg",
			Color(ColorPrimary),
			Padding(Rem(0.75), Rem(1)),
			BorderRadius(Rem(0.5)),
			Decl{Prop: "background", Val: ColorPrimary.Var() + "18"},
			FontWeight(Str("600")),
			FontSize(Rem(0.95)),
		),
		Rule(".error-msg",
			Color(Hex("#d63031")),
			Padding(Rem(0.75), Rem(1)),
			BorderRadius(Rem(0.5)),
			Decl{Prop: "background", Val: "#d6303118"},
			FontWeight(Str("600")),
			FontSize(Rem(0.95)),
		),
	)
}
