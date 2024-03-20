package model

import "fmt"

type Describe struct {
	Field      string
	Type       string
	Properties string
}

func (m Model) GetDescribe(table string) (*[]Describe, error) {
	var data []Describe
	res, err := m.FulltextDB.Query(fmt.Sprintf("DESCRIBE %s", table), &data)
	if err != nil {
		return res.(*[]Describe), err
	}
	return res.(*[]Describe), nil
}
