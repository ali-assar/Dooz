package dto

type AchievementDTO struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Icon             string `json:"icon"`
	RequirementType  int    `json:"requirement_type"`
	RequirementValue int    `json:"requirement_value"`
	CoinReward       int    `json:"coin_reward"`
	GemReward        int    `json:"gem_reward"`
}

type UserAchievementDTO struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	AchievementID string `json:"achievement_id"`
	EarnedAt      int64  `json:"earned_at"`
}

type UserAchievementWithDetailDTO struct {
	Achievement AchievementDTO `json:"achievement"`
	EarnedAt    int64          `json:"earned_at"`
}

type CreateAchievementRequest struct {
	Name             string `json:"name" binding:"required,min=1,max=100"`
	Description      string `json:"description" binding:"required,min=1,max=500"`
	Icon             string `json:"icon"`
	RequirementType  int    `json:"requirement_type" binding:"required,min=1,max=5"`
	RequirementValue int    `json:"requirement_value" binding:"required,min=1"`
	CoinReward       int    `json:"coin_reward" binding:"min=0"`
	GemReward        int    `json:"gem_reward" binding:"min=0"`
}
