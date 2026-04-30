package entity

import (
	"dooz/internal/app/api/dto"
)

type Role byte

const (
	RoleUser       Role = 1
	RoleAdmin      Role = 2
	RoleSuperAdmin Role = 3
)

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "user"
	case RoleAdmin:
		return "admin"
	case RoleSuperAdmin:
		return "super_admin"
	default:
		return "user"
	}
}

type User struct {
	ID              string `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserCode        int    `gorm:"column:user_code;type:integer;uniqueIndex;not null" json:"user_code"`
	Phone           string `gorm:"type:text;uniqueIndex" json:"phone"`
	Email           string `gorm:"type:text;uniqueIndex;not null" json:"email,omitempty"`
	Fullname        string `gorm:"type:text;not null" json:"fullname"`
	PasswordHash    string `gorm:"column:password;type:text" json:"-"`
	Role            Role   `gorm:"type:smallint;not null;default:1" json:"role"`
	IsEmailVerified bool   `gorm:"column:is_email_verified;type:boolean;not null;default:false" json:"-"`
	IsPhoneVerified bool   `gorm:"column:is_phone_verified;type:boolean;not null;default:false" json:"-"`
	CurrentTheme    int    `gorm:"column:current_theme;type:smallint;not null;default:1" json:"current_theme"`
	CurrentXOShape  int    `gorm:"column:current_xo_shape;type:smallint;not null;default:1" json:"current_xo_shape"`
	CurrentAvatar   int    `gorm:"column:current_avatar;type:smallint;not null;default:1" json:"current_avatar"`
	Coins           int    `gorm:"type:integer;not null;default:0" json:"coins"`
	Gems            int    `gorm:"type:integer;not null;default:0" json:"gems"`
	XCount          int    `gorm:"column:x_count;type:integer;not null;default:0" json:"x_count"`
	OCount          int    `gorm:"column:o_count;type:integer;not null;default:0" json:"o_count"`
	Wins            int    `gorm:"type:integer;not null;default:0" json:"wins"`
	Losses          int    `gorm:"type:integer;not null;default:0" json:"losses"`
	Draws           int    `gorm:"type:integer;not null;default:0" json:"draws"`
	IsOnline        bool   `gorm:"column:is_online;type:boolean;not null;default:false" json:"is_online"`
	LastSeenAt      int64  `gorm:"column:last_seen_at;type:bigint" json:"last_seen_at"`
	UpdatedAt       int64  `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
	DeletedAt       int64  `gorm:"column:deleted_at;type:bigint" json:"deleted_at"`
	BlockedAt       int64  `gorm:"column:blocked_at;type:bigint" json:"blocked_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) ToDTO() *dto.UserDTO {
	return &dto.UserDTO{
		ID:             u.ID,
		UserCode:       u.UserCode,
		Phone:          u.Phone,
		Email:          u.Email,
		Fullname:       u.Fullname,
		Role:           u.Role.String(),
		CurrentTheme:   u.CurrentTheme,
		CurrentXOShape: u.CurrentXOShape,
		CurrentAvatar:  u.CurrentAvatar,
		Coins:          u.Coins,
		Gems:           u.Gems,
		Wins:           u.Wins,
		Losses:         u.Losses,
		Draws:          u.Draws,
		IsOnline:       u.IsOnline,
		LastSeenAt:     u.LastSeenAt,
	}
}
