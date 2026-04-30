package dto

type FriendshipDTO struct {
	ID          string `json:"id"`
	RequesterID string `json:"requester_id"`
	AddresseeID string `json:"addressee_id"`
	Status      string `json:"status"`
	UpdatedAt   int64  `json:"updated_at"`
}

type FriendWithUserDTO struct {
	FriendshipID string  `json:"friendship_id"`
	Friend       UserDTO `json:"friend"`
	Status       string  `json:"status"`
}

type PendingFriendRequestDTO struct {
	ID          string  `json:"id"`
	Requester   UserDTO `json:"requester"`
	AddresseeID string  `json:"addressee_id"`
	Status      string  `json:"status"`
	UpdatedAt   int64   `json:"updated_at"`
}

type SendFriendRequestDTO struct {
	AddresseeID   string `json:"addressee_id,omitempty"`
	AddresseeCode *int   `json:"addressee_code,omitempty" binding:"omitempty,min=100000,max=999999"`
}

type RespondFriendRequestDTO struct {
	FriendshipID string `uri:"id" binding:"required"`
}
