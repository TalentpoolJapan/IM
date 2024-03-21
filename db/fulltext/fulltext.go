package fulltext

import (
	"database/sql"
	"fmt"

	"xorm.io/xorm"
)

type FulltextDB struct {
	db *xorm.Engine
}

func NewFulltextDB(dbHost string) (*FulltextDB, error) {
	mysql, err := xorm.NewEngine("mysql", fmt.Sprintf("``:``@tcp(%s)/Manticore", dbHost))
	mysql.ShowSQL(true)
	return &FulltextDB{
		db: mysql,
	}, err
}

func (f *FulltextDB) Exec(sql string) (sql.Result, error) {
	return f.db.Exec(sql)
}

func (f *FulltextDB) Query(sql string, rowsSlicePtr interface{}) (interface{}, error) {
	err := f.db.SQL(sql).Find(rowsSlicePtr)
	return rowsSlicePtr, err
}

func (f *FulltextDB) GetSession() *xorm.Session {
	return f.db.NewSession()
}

func (f *FulltextDB) TRUNCATE(table string) {
	f.db.Exec(fmt.Sprintf(`TRUNCATE TABLE %s`, table))
}
