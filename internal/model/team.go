package model

type Member struct {
	ID       string
	Name     string
	IsActive bool
}

type Team struct {
	Name    string
	Members []Member
}
