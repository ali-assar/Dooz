package entity

import "dooz/internal/app/api/dto"

type GameChallengeStatus byte

const (
	GameChallengePending  GameChallengeStatus = 1
	GameChallengeAccepted GameChallengeStatus = 2
	GameChallengeRejected GameChallengeStatus = 3
	GameChallengeCanceled GameChallengeStatus = 4
	GameChallengeExpired  GameChallengeStatus = 5
)

func (s GameChallengeStatus) String() string {
	switch s {
	case GameChallengePending:
		return "pending"
	case GameChallengeAccepted:
		return "accepted"
	case GameChallengeRejected:
		return "rejected"
	case GameChallengeCanceled:
		return "canceled"
	case GameChallengeExpired:
		return "expired"
	default:
		return "unknown"
	}
}

type GameChallenge struct {
	ID          string              `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	RequesterID string              `gorm:"column:requester_id;type:uuid;not null" json:"requester_id"`
	AddresseeID string              `gorm:"column:addressee_id;type:uuid;not null" json:"addressee_id"`
	BoardID     *string             `gorm:"column:board_id;type:uuid" json:"board_id"`
	Status      GameChallengeStatus `gorm:"type:smallint;not null;default:1" json:"status"`
	ExpiresAt   int64               `gorm:"column:expires_at;type:bigint;not null" json:"expires_at"`
}

func (GameChallenge) TableName() string {
	return "game_challenges"
}

func (g *GameChallenge) ToDTO() *dto.GameChallengeDTO {
	boardID := ""
	if g.BoardID != nil {
		boardID = *g.BoardID
	}
	return &dto.GameChallengeDTO{
		ID:          g.ID,
		RequesterID: g.RequesterID,
		AddresseeID: g.AddresseeID,
		BoardID:     boardID,
		Status:      g.Status.String(),
		ExpiresAt:   g.ExpiresAt,
	}
}
