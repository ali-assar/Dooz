package seed

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"dooz/entity"
	"dooz/utils/encrypt"

	"gorm.io/gorm"
)

func Run(db *gorm.DB, logger *slog.Logger, count int) error {
	lg := logger.With("component", "seed")
	lg.Info("seeding database", "count", count)

	for i := 0; i < count; i++ {
		now := time.Now().Unix()
		user := &entity.User{
			Email:           fmt.Sprintf("user%d@example.com", i+1),
			Fullname:        fmt.Sprintf("User %d", i+1),
			PasswordHash:    encrypt.HashSHA256("password123"),
			Phone:           fmt.Sprintf("09%09d", rand.Intn(1000000000)),
			UserCode:        100000 + i,
			Role:            entity.RoleUser,
			IsEmailVerified: true,
			IsPhoneVerified: true,
			CurrentTheme:    1,
			CurrentXOShape:  1,
			CurrentAvatar:   1,
			Coins:           rand.Intn(500),
			Gems:            rand.Intn(50),
			Wins:            rand.Intn(100),
			Losses:          rand.Intn(100),
			Draws:           rand.Intn(30),
			UpdatedAt:       now,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := db.WithContext(ctx).Create(user).Error; err != nil {
			lg.Warn("failed to seed user", "index", i, "error", err)
		}
		cancel()
	}

	lg.Info("seeding complete")
	return nil
}
