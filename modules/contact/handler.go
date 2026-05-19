package contact

import (
	//"github.com/tinywasm/goflare/cloudflare"
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/goflare/router"
	"github.com/tinywasm/json"
)

func Handle(ctx router.Context) {
	ctx.SetHeader("Content-Type", "application/json")
	ctx.SetHeader("Access-Control-Allow-Origin", "*")

	if ctx.Method() == "OPTIONS" {
		ctx.SetHeader("Access-Control-Allow-Methods", "POST, OPTIONS")
		ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type")
		ctx.WriteStatus(204)
		return
	}
	if ctx.Method() != "POST" {
		ctx.WriteStatus(405)
		ctx.Write([]byte(`{"error":"method not allowed"}`))
		return
	}

	var data ContactForm
	if err := json.Decode(ctx.Body(), &data); err != nil {
		ctx.WriteStatus(400)
		ctx.Write([]byte(`{"error":"invalid json"}`))
		return
	}
	if err := data.Validate(0); err != nil {
		ctx.WriteStatus(422)
		ctx.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	/* 	if err := sendEmail(data, cloudflare.Env("RESEND_API_KEY")); err != nil {
		ctx.WriteStatus(502)
		ctx.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	} */

	ctx.WriteStatus(200)
	ctx.Write([]byte(`{"message":"¡Gracias! Hemos recibido tu solicitud."}`))
}

func sendEmail(data ContactForm, apiKey string) error {
	if apiKey == "" {
		return fmt.Err("RESEND_API_KEY is missing")
	}

	// Escape user input to prevent HTML injection in the email body
	safeNombre := fmt.Convert(data.Nombre).EscapeHTML()
	safeEmail := fmt.Convert(data.Email).EscapeHTML()
	safeMensaje := fmt.Convert(data.Mensaje).EscapeHTML()

	payload := &EmailPayload{
		From:    "onboarding@resend.dev",
		To:      "delivered@resend.dev",
		Subject: "Nuevo contacto de " + safeNombre,
		Html:    "<p><strong>Nombre:</strong> " + safeNombre + "</p><p><strong>Email:</strong> " + safeEmail + "</p><p><strong>Mensaje:</strong> " + safeMensaje + "</p>",
	}

	var body []byte
	if err := json.Encode(payload, &body); err != nil {
		return err
	}

	errChan := make(chan error, 1)

	fetch.Post("https://api.resend.com/emails").
		Header("Authorization", "Bearer "+apiKey).
		ContentTypeJSON().
		Body(body).
		Send(func(resp *fetch.Response, err error) {
			if err != nil {
				errChan <- err
				return
			}
			if resp.Status >= 400 {
				errChan <- fmt.Err("email delivery failed with status: ", resp.Status)
				return
			}
			errChan <- nil
		})

	return <-errChan
}
