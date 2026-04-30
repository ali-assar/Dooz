package entity

type DeviceType byte

const (
	DeviceWeb      DeviceType = 1
	DeviceMobile   DeviceType = 2
	DeviceTelegram DeviceType = 3
)

type Session struct {
	ID               string     `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID           string     `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	SessionToken     string     `gorm:"column:session_token;type:text" json:"-"`
	JwtID            string     `gorm:"column:jwt_id;type:text;uniqueIndex" json:"-"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash;type:text;not null" json:"-"`
	ExpiresAt        int64      `gorm:"column:expires_at;type:bigint;not null" json:"expires_at"`
	IPAddress        *string    `gorm:"column:ip_address;type:inet" json:"ip_address,omitempty"`
	DeviceType       DeviceType `gorm:"column:device_type;type:smallint;not null" json:"device_type"`
	LastActivityAt   int64      `gorm:"column:last_activity_at;type:bigint;not null" json:"last_activity_at"`
}

func (Session) TableName() string {
	return "user_sessions"
}
