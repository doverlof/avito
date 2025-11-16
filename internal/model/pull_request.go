package model

import "time"

var (
	StatusOpen  PullRequestStatus = "OPEN"
	StatusMerge PullRequestStatus = "MERGED"
)

type PullRequestStatus string

type CreatePullRequest struct {
	AuthorID        string
	PullRequestID   string
	PullRequestName string
}

type PullRequest struct {
	AuthorID        string
	PullRequestID   string
	PullRequestName string
	Status          PullRequestStatus
	CreatedAt       time.Time
	MergedAt        time.Time
	ReviewerIDs     []string
}
