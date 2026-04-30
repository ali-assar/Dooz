package entity

import "dooz/internal/app/api/dto"

type RequirementType byte

const (
	RequirementWins        RequirementType = 1
	RequirementDraws       RequirementType = 2
	RequirementFriends     RequirementType = 3
	RequirementGamesPlayed RequirementType = 4
	RequirementWinStreak   RequirementType = 5
)

type Achievement struct {
	ID               string          `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Name             string          `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Description      string          `gorm:"type:text;not null" json:"description"`
	Icon             string          `gorm:"type:text;not null;default:''" json:"icon"`
	RequirementType  RequirementType `gorm:"column:requirement_type;type:smallint;not null" json:"requirement_type"`
	RequirementValue int             `gorm:"column:requirement_value;type:integer;not null" json:"requirement_value"`
	CoinReward       int             `gorm:"column:coin_reward;type:integer;not null;default:0" json:"coin_reward"`
	GemReward        int             `gorm:"column:gem_reward;type:integer;not null;default:0" json:"gem_reward"`
}

func (Achievement) TableName() string {
	return "achievements"
}

func (a *Achievement) ToDTO() *dto.AchievementDTO {
	return &dto.AchievementDTO{
		ID:               a.ID,
		Name:             a.Name,
		Description:      a.Description,
		Icon:             a.Icon,
		RequirementType:  int(a.RequirementType),
		RequirementValue: a.RequirementValue,
		CoinReward:       a.CoinReward,
		GemReward:        a.GemReward,
	}
}
