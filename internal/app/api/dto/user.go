package dto

type UserDTO struct {
	ID             string         `json:"id"`
	UserCode       int            `json:"user_code"`
	Phone          string         `json:"phone"`
	Email          string         `json:"email,omitempty"`
	Fullname       string         `json:"fullname"`
	Role           string         `json:"role"`
	CurrentTheme   int            `json:"current_theme"`
	CurrentXOShape int            `json:"current_xo_shape"`
	CurrentAvatar  int            `json:"current_avatar"`
	Coins          int            `json:"coins"`
	Gems           int            `json:"gems"`
	Wins           int            `json:"wins"`
	Losses         int            `json:"losses"`
	Draws          int            `json:"draws"`
	IsOnline       bool           `json:"is_online"`
	LastSeenAt     int64          `json:"last_seen_at"`
	OwnedThemes    []OwnedItemDTO `json:"owned_themes,omitempty"`
	OwnedXOShapes  []OwnedItemDTO `json:"owned_xo_shapes,omitempty"`
	OwnedAvatars   []OwnedItemDTO `json:"owned_avatars,omitempty"`
}

type UpdateUserRequest struct {
	UserID         string `json:"user_id,omitempty"`
	Role           string `json:"role,omitempty" binding:"omitempty,oneof=user admin super_admin"`
	Phone          string `json:"phone,omitempty" binding:"omitempty,iranian_phone"`
	Email          string `json:"email,omitempty" binding:"omitempty,email"`
	Fullname       string `json:"fullname,omitempty" binding:"omitempty,min=2,max=100"`
	CurrentTheme   *int   `json:"current_theme,omitempty"`
	CurrentXOShape *int   `json:"current_xo_shape,omitempty"`
	CurrentAvatar  *int   `json:"current_avatar,omitempty"`
}

type ChangePasswordRequest struct {
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
	ConfirmPassword string `json:"confirm_password" binding:"required,min=8,max=72"`
}

type GetUserByIDRequest struct {
	ID string `uri:"id" binding:"required"`
}

type GetUserByCodeRequest struct {
	Code int `uri:"code" binding:"required,min=100000,max=999999"`
}
