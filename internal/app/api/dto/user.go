package dto

type UserDTO struct {
	ID         string `json:"id"`
	Phone      string `json:"phone"`
	Email      string `json:"email,omitempty"`
	Fullname   string `json:"fullname"`
	Role       string `json:"role"`
	Avatar     string `json:"avatar"`
	Coins      int    `json:"coins"`
	Gems       int    `json:"gems"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	IsOnline   bool   `json:"is_online"`
	LastSeenAt int64  `json:"last_seen_at"`
}

type UpdateUserRequest struct {
	UserID   string `json:"user_id,omitempty"`
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=user admin super_admin"`
	Phone    string `json:"phone,omitempty" binding:"omitempty,iranian_phone"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Fullname string `json:"fullname,omitempty" binding:"omitempty,min=2,max=100"`
	Avatar   string `json:"avatar,omitempty"`
}

type ChangePasswordRequest struct {
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8,max=72"`
}

type GetUserByIDRequest struct {
	ID string `uri:"id" binding:"required"`
}
