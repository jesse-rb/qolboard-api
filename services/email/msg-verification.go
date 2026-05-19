package email

import (
	"context"
	"fmt"
	"os"
	"qolboard-api/services/logging"
)

func SendVerificationEmail(ctx context.Context, s EmailClient, to string, token string) error {
	apiHost := os.Getenv("API_HOST")
	verifyURL := fmt.Sprintf("%s/auth/verify?token=%s", apiHost, token)

	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Verify your email address</h2>
        <p>Click the button below to verify your email. This link expires in <strong>24 hours</strong>.</p>
        <a href="%s" style="
            display: inline-block;
            padding: 12px 24px;
            background-color: #4F46E5;
            color: white;
            text-decoration: none;
            border-radius: 6px;
            font-weight: bold;
        ">Verify Email</a>
        <p style="color: #666; font-size: 14px; margin-top: 24px;">
            Or copy this link: <a href="%s">%s</a>
        </p>
        <p style="color: #999; font-size: 12px;">If you didn't create an account, you can ignore this email.</p>
    </body>
    </html>`, verifyURL, verifyURL, verifyURL)

	text := fmt.Sprintf(
		"Verify your email address\n\nVisit this link (expires in 24 hours):\n%s\n\nIf you didn't create an account, ignore this email.",
		verifyURL,
	)

	if s != nil {

		if err := s.sendEmail(ctx, to, "Verify your email address", html, text); err != nil {
			return fmt.Errorf("failed sending verification email: :w", err)
		}
	} else {
		logging.LogInfo("email", "attempted to send verification email with nil EmailClient", nil)
	}

	return nil
}
