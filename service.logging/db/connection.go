package db

import (
	"time"

	"github.com/gocql/gocql"
)

// Cassandra configuration
const (
	cassandraHosts      = "logging_db"
	cassandraKeyspace   = "logging"
	cassandraTable      = "auth_log"
	cassandraTimeoutSec = 5
)

// createCluster creates and configures a Cassandra cluster session
func createCluster() (*gocql.ClusterConfig, error) {
	cluster := gocql.NewCluster(cassandraHosts)
	cluster.Keyspace = cassandraKeyspace
	cluster.Timeout = time.Second * cassandraTimeoutSec
	return cluster, nil
}

// getSession creates a Cassandra session
func GetSession() (*gocql.Session, error) {
	cluster, err := createCluster()
	if err != nil {
		return nil, err
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}
