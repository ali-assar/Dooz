package email

import (
	"html/template"
	"log/slog"
	"os"
)

var OTPTemplate *template.Template

type OTPTemplateData struct {
	Code string
}

func LoadTemplates(logger *slog.Logger) {
	templatePath := "internal/infrastructure/email/templates/OTP.html"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		logger.Warn("OTP template not found, emails will use plaintext fallback", "path", templatePath)
		return
	}
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		logger.Warn("failed to parse OTP template", "error", err)
		return
	}
	OTPTemplate = t
}
