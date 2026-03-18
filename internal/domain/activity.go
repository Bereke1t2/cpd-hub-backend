package domain

type Activity struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type ActivityRepository interface {
	List() ([]*Activity, error)
}
