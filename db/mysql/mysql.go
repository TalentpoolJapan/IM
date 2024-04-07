package mysql

import (
	"fmt"

	"imserver/config"

	_ "github.com/go-sql-driver/mysql"

	"xorm.io/xorm"
)

type MysqlDB struct {
	db *xorm.Engine
}

func NewMysqlDB() (*MysqlDB, error) {
	DB, err := xorm.NewEngine("mysql", fmt.Sprintf("root:%s@%s/%s?charset=utf8", config.MYSQL_SECRECT, config.MYSQL_HOST, config.MYSQL_DB))
	return &MysqlDB{
		db: DB,
	}, err
}

type UserProfile struct {
	Uuid     []string
	UserType int
}

type UserBasicInfo struct {
	FullNameEn string `json:"full_name_en" xorm:"default 'NULL' VARCHAR(64)"`
	FullNameJa string `json:"full_name_ja" xorm:"default 'NULL' VARCHAR(64)"`
	Avatar     string `json:"logo" xorm:"default '''' VARCHAR(128)"`
}

func (m *MysqlDB) GetUserProfileByUser(user UserProfile) (userBasicInfo []UserBasicInfo, err error) {
	if user.UserType == 0 {
		err = m.db.SQL("select company_name_en as full_name_en,company_name_ja as full_name_ja,logo as avatar from enterprise_account_basic_info where uuid in (?)", user.Uuid).Find(&userBasicInfo)
	}
	if user.UserType == 1 {
		err = m.db.SQL("select nick_name_oversea as full_name_en,nick_name_ja as full_name_ja,logo as avatar from job_seeker_basic_info where uuid in (?)", user.Uuid).Find(&userBasicInfo)
	}
	return
}
