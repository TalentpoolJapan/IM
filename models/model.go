package models

import (
	"fmt"
	"time"

	"github.com/gogf/gf/v2/container/gmap"
	rdb "github.com/rosedblabs/rosedb/v2"
	"xorm.io/xorm"
)

type Model struct {
	MemDB       *gmap.AnyAnyMap
	ManticoreDB *xorm.Engine
	MySQLDB     *xorm.Engine
	NoSqlDB     *NoSqlDB
}

type NoSqlDB struct {
	db *rdb.DB
}

func NewNoSqlDB(dbPath string) (*NoSqlDB, error) {
	options := rdb.DefaultOptions
	options.DirPath = dbPath
	db, err := rdb.Open(options)
	return &NoSqlDB{
		db: db,
	}, err
}

func (n *NoSqlDB) Set(k, v string) error {
	return n.db.Put([]byte(k), []byte(v))
}

func (n *NoSqlDB) Get(k string) (string, error) {
	val, err := n.db.Get([]byte(k))

	return string(val), err
}

func (n *NoSqlDB) SetIfNotExistWithTTL(k, v string, expired time.Duration) (bool, error) {
	batch := n.db.NewBatch(rdb.DefaultBatchOptions)
	val, err := batch.Get([]byte(k))
	if err != nil {
		if err.Error() != "key not found in database" {
			batch.Rollback()
			return false, err
		}

	}

	if val != nil {
		batch.Rollback()
		return false, nil
	}

	err = batch.PutWithTTL([]byte(k), []byte(v), expired)
	if err != nil {
		batch.Rollback()
		return false, err
	}

	// s, _ := batch.Get([]byte(k))
	// fmt.Println(string(s))
	return true, batch.Commit()
}

func (n *NoSqlDB) SetIfNotExist(k, v string) (bool, error) {
	batch := n.db.NewBatch(rdb.DefaultBatchOptions)
	val, err := batch.Get([]byte(k))
	if err != nil {
		if err.Error() != "key not found in database" {
			batch.Rollback()
			return false, err
		}

	}

	if val != nil {
		batch.Rollback()
		return false, nil
	}

	err = batch.Put([]byte(k), []byte(v))
	if err != nil {
		batch.Rollback()
		return false, err
	}
	//batch.Commit()

	// s, _ := batch.Get([]byte(k))
	// fmt.Println(string(s))
	return true, batch.Commit()
}

func (n *NoSqlDB) GetBatch() *rdb.Batch {
	return n.db.NewBatch(rdb.DefaultBatchOptions)
}

func (n *NoSqlDB) TRUNCATE() {
	var keys []string
	n.db.AscendKeys(nil, true, func(k []byte) (bool, error) {
		fmt.Println("key = ", string(k))
		keys = append(keys, string(k))
		return true, nil
	})

	for _, v := range keys {
		n.db.Delete([]byte(v))
	}

}
