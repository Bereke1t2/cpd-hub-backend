package domain

type BookmarkRepository interface {
	Add(username, problemID string) error
	Remove(username, problemID string) error
	ListProblemIDs(username string) ([]string, error)
}
