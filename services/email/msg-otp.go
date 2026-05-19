package email

import (
	"context"
	"fmt"
	"qolboard-api/services/logging"
)

func SendOTPEmail(ctx context.Context, s EmailClient, to string, otp string) error {
	html := fmt.Sprintf(`
    <!DOCTYPE html>
    <html>
    <body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2>Your login code</h2>
        <p>Use the code below to sign in. It expires in <strong>5 minutes</strong>.</p>
        <div style="
            display: inline-block;
            padding: 16px 32px;
            background-color: #F3F4F6;
            border-radius: 8px;
            font-size: 36px;
            font-weight: bold;
            letter-spacing: 8px;
            color: #111827;
            margin: 16px 0;
        ">%s</div>
        <p style="color: #999; font-size: 12px; margin-top: 24px;">
            Never share this code. If you didn't request it, you can ignore this email.
        </p>
    </body>
    </html>`, otp)

	text := fmt.Sprintf(
		"Your login code: %s\n\nThis code expires in 5 minutes.\nNever share this code with anyone.",
		otp,
	)

	if s != nil {
		if err := s.sendEmail(ctx, to, "Your login code", html, text); err != nil {
			return fmt.Errorf("failed to send otp email: %w", err)
		}
	} else {
		logging.LogInfo("email", "attempted to send otp email with nil EmailClient", nil)
	}
	return nil
}
