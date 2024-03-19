package rosedb

import (
	rdb "github.com/rosedblabs/rosedb/v2"
)

type RoseDBConfig struct {
	DBPath string
}

type RoseDB struct {
	DB *rdb.DB
}

func NewRoseDB(c RoseDBConfig) (*RoseDB, error) {
	options := rdb.DefaultOptions
	options.DirPath = c.DBPath
	db, err := rdb.Open(options)
	return &RoseDB{
		DB: db,
	}, err
}

func (m *RoseDB) Set(k string, v string) (err error) {
	return m.DB.Put([]byte(k), []byte(v))
}

func (m *RoseDB) SetIfNotExist(k string, v string) (ok bool, err error) {
	batch := m.DB.NewBatch(rdb.DefaultBatchOptions)
	val, err := batch.Get([]byte(k))
	if err != nil {
		//fmt.Println("x", err.Error())
		if err.Error() == "key not found in database" {

		} else {
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
	batch.Commit()

	// s, _ := batch.Get([]byte(k))
	// fmt.Println(string(s))
	return true, nil
}

func (m *RoseDB) Get(k string) (string, error) {
	val, err := m.DB.Get([]byte(k))

	return string(val), err
}
