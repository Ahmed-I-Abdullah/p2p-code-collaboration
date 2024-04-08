// Package database provides functionalities for storing and retreiving data from a BadgerDB database
// BadgerDB is a key-value store written in Go
package database

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-log/v2"
)

var logger = log.Logger("database")

// DBCon represents the connection to the BadgerDB database
var DBCon *badger.DB

// Init initializes the BadgerDB database connection
// It takes an integer id as input for the badgerDB folder path and initializes the database connection
// It returns an error if there's any issue initializing the database connection
func Init(id int) error {
	err := log.SetLogLevel("database", "info")
	if err != nil {
		logger.Error("could not set log level for the db logger")
	}
	var dbErr error = nil
	folderPath := fmt.Sprintf("/tmp/badger/%v", id)
	DBCon, dbErr = badger.Open(badger.DefaultOptions(folderPath))
	if dbErr != nil {
		logger.Fatalf("Error connecting to DB: %v", dbErr)
		return dbErr
	}
	logger.Infof("successfully connected to db. Folder can be found at : %v", folderPath)
	return nil
}

// Close closes the BadgerDB database connection
func Close() {
	err := DBCon.Close()
	if err != nil {
		logger.Fatal(err)
	}
}

// Get retrieves the value associated with the given key from the database
// It returns the value and any error encountered during the retrieval process
func Get(key []byte) ([]byte, error) {
	var valCpy []byte
	err := DBCon.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		valCpy, err = item.ValueCopy(nil)
		if err != nil {
			logger.Errorf("could not copy value for key %v, error: %v", key, err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	logger.Debugf("value returned for key %v : %s", key, valCpy)
	return valCpy, err
}

// Put saves a key-value pair in the database with a Time-To-Live (TTL) of 10 minutes
// It returns an error if there's any issue during the write operation
func Put(key []byte, val []byte) error {
	err := DBCon.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, val).WithTTL(time.Minute * 10)
		updateErr := txn.SetEntry(entry)
		return updateErr
	})
	logger.Debugf("successfully wrote key -> %v , pair -> %s to the database", key, val)
	return err
}
