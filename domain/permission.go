package domain

type Permission struct {
	ID       string `json:"id"       db:"id"`
	Resource string `json:"resource" db:"resource"`
	Action   string `json:"action"   db:"action"`
}
