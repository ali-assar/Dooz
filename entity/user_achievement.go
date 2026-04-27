package entity

import "dooz/internal/app/api/dto"

type UserAchievement struct {
	ID            string `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	UserID        string `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	AchievementID string `gorm:"column:achievement_id;type:uuid;not null" json:"achievement_id"`
	EarnedAt      int64  `gorm:"column:earned_at;type:bigint;not null" json:"earned_at"`
}

func (UserAchievement) TableName() string {
	return "user_achievements"
}

func (ua *UserAchievement) ToDTO() *dto.UserAchievementDTO {
	return &dto.UserAchievementDTO{
		ID:            ua.ID,
		UserID:        ua.UserID,
		AchievementID: ua.AchievementID,
		EarnedAt:      ua.EarnedAt,
	}
}
