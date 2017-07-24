package base

/* GitHubUser representation of a GitHub user
 * with id and name fields
 */
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"login"`
}

type Client interface {
	Search(location string, count int) ([]User, error)
}
