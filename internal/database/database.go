package database

import (
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-log/v2"
)

var logger = log.Logger("database")
var DBCon *badger.DB

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

func Close() {
	err := DBCon.Close()
	if err != nil {
		logger.Fatal(err)
	}
}

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

// Put save a key-pair in the DB with a TTL of 10 minutes
func Put(key []byte, val []byte) error {
	err := DBCon.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, val).WithTTL(time.Minute * 10)
		updateErr := txn.SetEntry(entry)
		return updateErr
	})
	logger.Debugf("successfully wrote key -> %v , pair -> %s to the database", key, val)
	return err
}
