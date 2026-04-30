package constants

import "time"

const (
	DefaultRequestTimeout = 5 * time.Second
	CronRequestTimeout    = 30 * time.Second
	CronDeleteTimeout     = 10 * time.Second
	DBSetupTimeout        = 10 * time.Second
	ShutdownTimeout       = 10 * time.Second
)

const (
	AccessTokenExpiration      = 15 * time.Minute
	RefreshTokenExpiration     = 24 * time.Hour
	RefreshTokenLongExpiration = 90 * 24 * time.Hour
)

const (
	OTPExpiration     = 10 * time.Minute
	OTPLength         = 5
	MaxOTPRetries     = 3
	OTPBatchSize      = 100
	OTPRetryBaseDelay = 5 * time.Second
	OTPRetryMaxDelay  = 60 * time.Second
)

const (
	OTPWorkerCount     = 5
	OTPWorkerBatchSize = 100
	OTPSendTimeout     = 10 * time.Second
)

const (
	DeleteExpiredOTPInterval = 10 * time.Minute
	ProcessOTPInterval       = 15 * time.Second
	CronHTTPTimeout          = 10 * time.Second
	ExpireMatchesInterval    = 30 * time.Second
)

const (
	MaxOpenConns    = 25
	MaxIdleConns    = 5
	ConnMaxLifetime = 5 * time.Minute
	ConnMaxIdleTime = 10 * time.Minute
)

const (
	SMTPConnectionTimeout = 10 * time.Second
	EmailSendTimeout      = 30 * time.Second
)

const (
	SMTPPoolMaxSize         = 10
	SMTPPoolIdleTimeout     = 30 * time.Second
	SMTPPoolMaxLifetime     = 5 * time.Minute
	SMTPPoolCleanupInterval = 30 * time.Second
)

const (
	DefaultPaginationLimit uint32 = 20
	MaxPaginationLimit     uint32 = 100
)

// MatchmakingTimeout is how long we wait for an opponent before assigning a bot.
const MatchmakingTimeout = 5 * time.Second

// MatchmakingPollInterval is how often we check the Redis queue for a new opponent.
const MatchmakingPollInterval = 500 * time.Millisecond

// CoinRewardWin is awarded to the winner of a game.
const CoinRewardWin = 25

// CoinRewardDraw is awarded to both players on a draw.
const CoinRewardDraw = 10

// WinStreakKey is the Redis key prefix for tracking win streaks.
const WinStreakKey = "win_streak:"
