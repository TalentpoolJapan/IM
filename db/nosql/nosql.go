package nosql

import (
	"time"

	rdb "github.com/rosedblabs/rosedb/v2"
)

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
	batch.Commit()

	// s, _ := batch.Get([]byte(k))
	// fmt.Println(string(s))
	return true, nil
}

// func (n *NoSqlDB) TRUNCATE() {
// 	//queue := make(chan string)
// 	n.db.AscendKeys(nil, true, func(k []byte) (bool, error) {
// 		fmt.Println("key = ", string(k))
// 		// go func() {
// 		// 	queue <- string(k)
// 		// }()
// 		// // n.db.Delete(k)
// 		// _, err := n.db.Get(k)
// 		//fmt.Println("x")
// 		return true, nil
// 	})

// 	//key := <-queue
// 	//n.db.Delete([]byte(key))

// }
