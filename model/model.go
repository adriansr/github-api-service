// Package model includes types and interfaces to model the various
// relationships between packages in the project
package model

// User is the representation of a GitHub
// user, already prepared to be serialised
// to json
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"name"`
}

// TopContributorGetter is an interface for a type that implements
// the GetTopContributors method
type TopContributorGetter interface {
	// GetTopContributors returns a list of the top `count` contributors
	// in a given `location`
	GetTopContributors(location string, count int) ([]User, error)
}
