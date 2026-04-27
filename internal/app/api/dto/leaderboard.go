package dto

type LeaderboardEntryDTO struct {
	Rank       int     `json:"rank"`
	User       UserDTO `json:"user"`
	Wins       int     `json:"wins"`
	Losses     int     `json:"losses"`
	Draws      int     `json:"draws"`
	WinRate    float64 `json:"win_rate"`
	TotalGames int     `json:"total_games"`
}
