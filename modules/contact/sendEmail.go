package contact

import (
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fmt"
	"github.com/tinywasm/json"
)

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
