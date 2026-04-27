package godotenv

import (
	"cmp"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Env struct {
	Environment  string
	HTTPPort     int
	DatabaseHost string
	Secret       string
	RedisHost    string
	BaseURL      string

	SMSProvider      int
	SMSApiKey        string
	SMSOTPTemplateID int

	EmailProvider int

	SuperAdminEmail    string
	SuperAdminPhone    string
	SuperAdminPassword string

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPSender   string

	CloudflareTurnstileEnabled   bool
	CloudflareTurnstileSecretKey string
}

func NewEnv() *Env {
	e := &Env{}
	e.Load()
	return e
}

func (e *Env) Load() {
	_ = godotenv.Load(".env")

	e.Environment = cmp.Or(os.Getenv("ENVIRONMENT"), "production")

	httpPort, err := strconv.Atoi(cmp.Or(os.Getenv("HTTP_PORT"), "8080"))
	if err != nil {
		log.Fatalf("Failed to parse HTTP_PORT: %v", err)
	}
	e.HTTPPort = httpPort

	e.DatabaseHost = os.Getenv("DATABASE_HOST")
	e.Secret = cmp.Or(os.Getenv("SECRET"), "choose_better_secret")
	e.RedisHost = cmp.Or(os.Getenv("REDIS_HOST"), "localhost:6379")
	e.BaseURL = cmp.Or(os.Getenv("BASE_URL"), "http://localhost:3000")

	smsProvider, err := strconv.Atoi(cmp.Or(os.Getenv("SMS_PROVIDER"), "2"))
	if err != nil {
		log.Fatalf("Failed to parse SMS_PROVIDER: %v", err)
	}
	e.SMSProvider = smsProvider
	e.SMSApiKey = os.Getenv("SMS_API_KEY")

	smsTemplateID, _ := strconv.Atoi(cmp.Or(os.Getenv("SMS_OTP_TEMPLATE_ID"), "1"))
	e.SMSOTPTemplateID = smsTemplateID

	emailProvider, err := strconv.Atoi(cmp.Or(os.Getenv("EMAIL_PROVIDER"), "2"))
	if err != nil {
		log.Fatalf("Failed to parse EMAIL_PROVIDER: %v", err)
	}
	e.EmailProvider = emailProvider

	e.SuperAdminEmail = cmp.Or(os.Getenv("SUPER_ADMIN_EMAIL"), "")
	e.SuperAdminPhone = os.Getenv("SUPER_ADMIN_PHONE")
	e.SuperAdminPassword = os.Getenv("SUPER_ADMIN_PASSWORD")

	e.SMTPHost = os.Getenv("SMTP_HOST")

	smtpPort, _ := strconv.Atoi(cmp.Or(os.Getenv("SMTP_PORT"), "587"))
	e.SMTPPort = smtpPort

	e.SMTPUser = os.Getenv("SMTP_USER")
	e.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	e.SMTPSender = os.Getenv("SMTP_SENDER")

	turnstileEnabled := os.Getenv("CLOUDFLARE_TURNSTILE_ENABLED")
	e.CloudflareTurnstileEnabled = turnstileEnabled == "true" || turnstileEnabled == "1"
	e.CloudflareTurnstileSecretKey = os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY")
}
