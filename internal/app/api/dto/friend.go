package dto

type FriendshipDTO struct {
	ID          string `json:"id"`
	RequesterID string `json:"requester_id"`
	AddresseeID string `json:"addressee_id"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type FriendWithUserDTO struct {
	FriendshipID string  `json:"friendship_id"`
	Friend       UserDTO `json:"friend"`
	Status       string  `json:"status"`
}

type SendFriendRequestDTO struct {
	AddresseeID string `json:"addressee_id" binding:"required"`
}

type RespondFriendRequestDTO struct {
	FriendshipID string `uri:"id" binding:"required"`
}
