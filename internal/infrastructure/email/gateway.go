package email

import "context"

type Param struct {
	To          string
	Subject     string
	Body        string
	ContentType string
	Attachments []Attachment
}

type Attachment struct {
	Filename string
	Data     []byte
}

type Gateway interface {
	SendOTP(ctx context.Context, email string, code string) error
	Send(ctx context.Context, param Param) error
}
