package manticore

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

type ManticoreMysqlConfig struct {
	Host  string
	Port  int
	MYSQL *xorm.Engine
}

type Manticore struct {
	DB *xorm.Engine
}

func NewManticore(m ManticoreMysqlConfig) *Manticore {
	db, _ := xorm.NewEngine("mysql", fmt.Sprintf("``:``@tcp(%s:%d)/Manticore", m.Host, m.Port))
	return &Manticore{
		DB: db,
	}
}

type Describe struct {
	Field      string
	Type       string
	Properties string
}

func (m Manticore) Describe(table string) ([]Describe, error) {
	var _Describe []Describe
	err := m.DB.SQL(fmt.Sprintf("DESCRIBE %s", table)).Find(&_Describe)
	return _Describe, err
}
