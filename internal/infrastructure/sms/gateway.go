package sms

import "context"

type Gateway interface {
	SendOTP(ctx context.Context, phone string, code string) error
}
