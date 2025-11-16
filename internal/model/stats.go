package model

type UserStatistics struct {
	UserId   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	TeamName string `json:"team_name" db:"team_name"`
	IsActive bool   `json:"is_active" db:"is_active"`

	TotalReviewAssignments  int `json:"total_review_assignments" db:"total_review_assignments"`
	OpenReviewAssignments   int `json:"open_review_assignments" db:"open_review_assignments"`
	MergedReviewAssignments int `json:"merged_review_assignments" db:"merged_review_assignments"`

	TotalAuthoredPrs  int `json:"total_authored_prs" db:"total_authored_prs"`
	OpenAuthoredPrs   int `json:"open_authored_prs" db:"open_authored_prs"`
	MergedAuthoredPrs int `json:"merged_authored_prs" db:"merged_authored_prs"`
}

type StatsResponse struct {
	Statistics []UserStatistics `json:"statistics"`
}
