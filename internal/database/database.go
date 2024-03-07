package database

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-log/v2"
)

var logger = log.Logger("database")
var DBCon *badger.DB
