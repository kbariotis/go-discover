package store

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"
	"time"

	"github.com/Financial-Times/neoism"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kbariotis/go-discover/internal/model"
)

// Neo store implementation
type Neo struct {
	db *neoism.Database
}

const (
	neoPutRepositoryQueryTemplate = `
		MERGE (r:Repository {name: "{{ .Name }}"})
		WITH r
		FOREACH (label IN {{ toStrings .Labels }} |
			MERGE (l:Label {name: label})
			MERGE (r)-[:ContainsTopic]->(l)
		)
		WITH r
		FOREACH (language IN {{ toStrings .Languages }} |
			MERGE (l:Label {name: language})
			MERGE (r)-[:ContainsLanguage]->(l)
		)
	`
	// TODO stars should also be
	neoPutUserQueryTemplate = `
		MERGE (u:User {name: "{{ .Name }}"})
		WITH u
		FOREACH (followee IN {{ toStrings .Followees}} |
			MERGE (f:User {name: followee})
			MERGE (u)-[:IsFollowing]->(f)
		)
		WITH u
		WITH {{ toObject .Stars }} AS stars
		FOREACH (star IN stars |
			MERGE (r:Repository {name: star.repository})
			MERGE (u)-[:HasStarred {starredAt: star.starredAt}]->(r)
		)
	`
)

var (
	neoIndicesQueries = []string{
		"CREATE CONSTRAINT ON (n:User) ASSERT n.name IS UNIQUE",
		"CREATE CONSTRAINT ON (n:Repository) ASSERT n.name IS UNIQUE",
		"CREATE CONSTRAINT ON (n:Label) ASSERT n.name IS UNIQUE",
		"CREATE CONSTRAINT ON (n:Language) ASSERT n.name IS UNIQUE",
	}
)

func neoToNeoObject(m interface{}) string {
	bytes, _ := json.Marshal(m)
	json := string(bytes)
	json = strings.Replace(json, `"repository"`, "`repository`", -1)
	json = strings.Replace(json, `"starredAt"`, "`starredAt`", -1)

	if json == "null" {
		return "[]"
	}

	return json
}

func neoToNeoStrings(m []string) string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// NewNeo constrcuts a new Neo store given a neoism db
func NewNeo(db *neoism.Database) (*Neo, error) {
	neo := &Neo{
		db: db,
	}

	return neo, nil
}

// SetupIndices creates indices for neo
func (neo *Neo) SetupIndices() error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "store/Neo.SetupIndices",
	})

	logger.Info("setting up indices")

	// run queries
	for _, neoIndicesQuery := range neoIndicesQueries {
		cypherQuery := &neoism.CypherQuery{
			Statement:  neoIndicesQuery,
			Parameters: map[string]interface{}{},
		}
		if err := neo.db.Cypher(cypherQuery); err != nil {
			return errors.Wrap(err, "could not setup indices")
		}
	}

	return nil
}

// PutRepository merges a repository's graph in neo
func (neo *Neo) PutRepository(repository *model.Repository) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger":                     "store/Neo.PutRepository",
		"repository.name":            repository.Name,
		"repository.labels.count":    len(repository.Labels),
		"repository.languages.count": len(repository.Languages),
	})

	logger.Info("following user")

	// keep start time for query metrics
	startTime := time.Now()

	// create template for query
	neoPutRepositoryQuery, err := template.
		New("neoPutRepositoryQuery").
		Funcs(template.FuncMap{
			"toObject":  neoToNeoObject,
			"toStrings": neoToNeoStrings,
		}).
		Parse(neoPutRepositoryQueryTemplate)
	if err != nil {
		return errors.Wrap(err, "could not parse template")
	}

	// render query
	query := &bytes.Buffer{}
	if err := neoPutRepositoryQuery.Execute(query, repository); err != nil {
		return errors.Wrap(err, "could not merge repository")
	}

	logger.WithField("query", query).Debug("running query")

	// run query
	cypherQuery := &neoism.CypherQuery{
		Statement:  query.String(),
		Parameters: map[string]interface{}{},
	}
	if err := neo.db.Cypher(cypherQuery); err != nil {
		return errors.Wrap(err, "could not merge repo")
	}

	// log query time
	logger.
		WithField("execution_time", time.Now().Sub(startTime)).
		Debug("query execution finished")

	return nil
}

// PutUser merges a user's graph in neo
func (neo *Neo) PutUser(user *model.User) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger":               "store/Neo.PutUser",
		"user.name":            user.Name,
		"user.followees.count": len(user.Followees),
		"user.stars.count":     len(user.Stars),
	})

	logger.Info("following user")

	// keep start time for query metrics
	startTime := time.Now()

	// create template for query
	neoPutUserQuery, err := template.
		New("neoPutUserQuery").
		Funcs(template.FuncMap{
			"toObject":  neoToNeoObject,
			"toStrings": neoToNeoStrings,
		}).
		Parse(neoPutUserQueryTemplate)
	if err != nil {
		return errors.Wrap(err, "could not parse template")
	}

	// render query
	query := &bytes.Buffer{}
	if err := neoPutUserQuery.Execute(query, user); err != nil {
		return errors.Wrap(err, "could not merge user")
	}

	logger.WithField("query", query).Debug("running query")

	// run query
	cypherQuery := &neoism.CypherQuery{
		Statement:  query.String(),
		Parameters: map[string]interface{}{},
	}
	if err := neo.db.Cypher(cypherQuery); err != nil {
		return errors.Wrap(err, "could not merge user")
	}

	// log query time
	logger.
		WithField("execution_time", time.Now().Sub(startTime)).
		Debug("query execution finished")

	return nil
}
