package store

import (
	"github.com/Financial-Times/neoism"
	"github.com/pkg/errors"

	"github.com/kbariotis/go-discover/internal/model"
)

// Neo store implementation
type Neo struct {
	db *neoism.Database
}

// NewNeo constrcuts a new Neo store given a neoism db
func NewNeo(db *neoism.Database) (Store, error) {
	neo := &Neo{
		db: db,
	}

	return neo, nil
}

// GetRepository returns a repository graph from neo
func (neo *Neo) GetRepository(name string) (*model.Repository, error) {
	// TODO implement neo.GetRepository
	return nil, nil
}

// PutRepository merges a repository's graph in neo
func (neo *Neo) PutRepository(repository *model.Repository) error {
	// merge repository
	repositoryQuery := &neoism.CypherQuery{
		Statement: `MERGE (r:Repository {name: {repository}})`,
		Parameters: map[string]interface{}{
			"repository": repository.Name,
		},
	}
	if err := neo.db.Cypher(repositoryQuery); err != nil {
		return errors.Wrap(err, "could not merge repository")
	}

	// add repository's labels
	for _, label := range repository.Labels {
		query := &neoism.CypherQuery{
			Statement: `
				MATCH (r:Repository {name: {repository}})
				MERGE (l:Label {name: {label}})
				MERGE (r)-[:Has]->(l)
			`,
			Parameters: map[string]interface{}{
				"repository": repository.Name,
				"label":      label,
			},
		}
		if err := neo.db.Cypher(query); err != nil {
			return errors.Wrap(err, "could not merge Has")
		}
	}

	// add repository's languages
	for _, language := range repository.Languages {
		query := &neoism.CypherQuery{
			Statement: `
				MATCH (r:Repository {name: {repository}})
				MERGE (l:Language {name: {language}})
				MERGE (r)-[:Contains]->(l)
			`,
			Parameters: map[string]interface{}{
				"repository": repository.Name,
				"language":   language,
			},
		}
		if err := neo.db.Cypher(query); err != nil {
			return errors.Wrap(err, "could not merge Contains")
		}
	}

	return nil
}

// GetUser returns a user graph from neo
func (neo *Neo) GetUser(user string) (*model.User, error) {
	// TODO implement neo.GetUser
	return nil, nil
}

// PutUser merges a user's graph in neo
func (neo *Neo) PutUser(user *model.User) error {
	// merge user
	userQuery := &neoism.CypherQuery{
		Statement: `MERGE (u:User {name: {user}})`,
		Parameters: map[string]interface{}{
			"user": user.Name,
		},
	}
	if err := neo.db.Cypher(userQuery); err != nil {
		return errors.Wrap(err, "could not merge user")
	}

	// add user's followers
	for _, followee := range user.Followees {
		query := &neoism.CypherQuery{
			Statement: `
				MATCH (u:User {name: {user}})
				MERGE (f:User {name: {followee}})
				MERGE (u)-[:IsFollowing]->(f)
			`,
			Parameters: map[string]interface{}{
				"user":     user.Name,
				"followee": followee,
			},
		}
		if err := neo.db.Cypher(query); err != nil {
			return errors.Wrap(err, "could not merge IsFollowing")
		}
	}

	// add starred repositories
	for _, repository := range user.Stars {
		query := &neoism.CypherQuery{
			Statement: `
				MATCH (u:User {name: {user}})
				MERGE (r:Repository {name: {repository}})
				MERGE (u)-[:HasStarred]->(r)
			`,
			Parameters: map[string]interface{}{
				"user":       user.Name,
				"repository": repository,
			},
		}
		if err := neo.db.Cypher(query); err != nil {
			return errors.Wrap(err, "could not merge HasStarred")
		}
	}

	return nil
}
