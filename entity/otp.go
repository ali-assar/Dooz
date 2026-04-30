package entity

type OTPPurpose byte

const (
	RegistrationPurpose   OTPPurpose = 1
	ForgotPasswordPurpose OTPPurpose = 2
)

type OTPStatus byte

const (
	OTPStatusPending OTPStatus = 0
	OTPStatusSent    OTPStatus = 1
	OTPStatusFailed  OTPStatus = 2
)

type OTPChannel byte

const (
	OTPChannelSMS   OTPChannel = 1
	OTPChannelEmail OTPChannel = 2
)

type OTP struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Recipient   string     `gorm:"column:recipient;type:text;not null" json:"recipient"`
	Channel     OTPChannel `gorm:"column:channel;type:smallint;not null" json:"channel"`
	Code        string     `gorm:"column:otp_code;type:text;not null" json:"-"`
	ExpiresAt   int64      `gorm:"column:expires_at;type:bigint;not null" json:"expires_at"`
	Purpose     OTPPurpose `gorm:"column:purpose;type:smallint;not null" json:"purpose"`
	Status      OTPStatus  `gorm:"column:status;type:smallint;not null;default:0" json:"status"`
	RetryCount  int        `gorm:"column:retry_count;type:smallint;not null;default:0" json:"retry_count"`
	NextRetryAt int64      `gorm:"column:next_retry_at;type:bigint" json:"next_retry_at"`
	ProcessedAt int64      `gorm:"column:processed_at;type:bigint" json:"processed_at"`
}

func (OTP) TableName() string {
	return "otp_outbox"
}
