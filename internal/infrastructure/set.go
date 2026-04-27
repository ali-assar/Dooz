package infrastructure

import (
	"dooz/internal/infrastructure/email"
	"dooz/internal/infrastructure/godotenv"
	"dooz/internal/infrastructure/messaging"
	"dooz/internal/infrastructure/sms"

	"github.com/google/wire"
)

var InfrastructureSet = wire.NewSet(
	godotenv.NewEnv,
	sms.SMSGatewaySet,
	email.EmailGatewaySet,
	messaging.MessagingSet,
)
