package domain

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type InfoRepository interface {
	List() ([]*Info, error)
}
