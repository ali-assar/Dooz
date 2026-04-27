package messaging

import (
	"context"

	"dooz/entity"
)

type Sender interface {
	SendOTP(ctx context.Context, channel entity.OTPChannel, recipient string, code string) error
}
