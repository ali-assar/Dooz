package entity

import "dooz/internal/app/api/dto"

type FriendshipStatus byte

const (
	FriendshipPending  FriendshipStatus = 1
	FriendshipAccepted FriendshipStatus = 2
	FriendshipRejected FriendshipStatus = 3
	FriendshipBlocked  FriendshipStatus = 4
)

func (s FriendshipStatus) String() string {
	switch s {
	case FriendshipPending:
		return "pending"
	case FriendshipAccepted:
		return "accepted"
	case FriendshipRejected:
		return "rejected"
	case FriendshipBlocked:
		return "blocked"
	default:
		return "unknown"
	}
}

type Friendship struct {
	ID          string           `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	RequesterID string           `gorm:"column:requester_id;type:uuid;not null" json:"requester_id"`
	AddresseeID string           `gorm:"column:addressee_id;type:uuid;not null" json:"addressee_id"`
	Status      FriendshipStatus `gorm:"type:smallint;not null;default:1" json:"status"`
	CreatedAt   int64            `gorm:"column:created_at;type:bigint;not null" json:"created_at"`
	UpdatedAt   int64            `gorm:"column:updated_at;type:bigint;not null" json:"updated_at"`
}

func (Friendship) TableName() string {
	return "friendships"
}

func (f *Friendship) ToDTO() *dto.FriendshipDTO {
	return &dto.FriendshipDTO{
		ID:          f.ID,
		RequesterID: f.RequesterID,
		AddresseeID: f.AddresseeID,
		Status:      f.Status.String(),
		CreatedAt:   f.CreatedAt,
		UpdatedAt:   f.UpdatedAt,
	}
}
