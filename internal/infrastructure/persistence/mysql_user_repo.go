package persistence

import (
	"errors"
	"imserver/internal/domain/user"
	"xorm.io/xorm"
)

type UserBasicInfoPO struct {
	Uuid       string `json:"uuid"`
	FullNameEn string `json:"full_name_en"`
	FullNameJa string `json:"full_name_ja"`
	Avatar     string `json:"avatar"`
}

func convertUserEntity(po *UserBasicInfoPO) *user.UserBasicInfo {
	return &user.UserBasicInfo{
		Uuid:       po.Uuid,
		FullNameEn: po.FullNameEn,
		FullNameJa: po.FullNameJa,
		Avatar:     po.Avatar,
	}
}

// region implement

func NewMysqlUserRepository(engine *xorm.Engine) user.IUserRepository {
	return &MysqlUserRepository{
		mysql: engine,
	}
}

type MysqlUserRepository struct {
	mysql *xorm.Engine
}

func (r MysqlUserRepository) ListUserBasicInfoByUuid(uuids []string) ([]*user.UserBasicInfo, error) {
	var userBasicInfos []*user.UserBasicInfo
	for _, uuid := range uuids {
		userBasicInfo, err := r.GetUserBasicInfoByUuid(uuid)
		if err == nil {
			userBasicInfos = append(userBasicInfos, userBasicInfo)
		}
	}
	return userBasicInfos, nil
}

func (r MysqlUserRepository) GetUserBasicInfoByUuid(uuid string) (*user.UserBasicInfo, error) {
	var userPO UserBasicInfoPO
	ok, err := r.mysql.SQL("select uuid,company_name_en as full_name_en,company_name_ja as full_name_ja,logo as avatar from enterprise_account_basic_info where uuid=?", uuid).Get(&userPO)
	if ok && err == nil {
		entity := convertUserEntity(&userPO)
		entity.UserType = user.Enterprise
		return entity, nil
	}

	ok, err = r.mysql.SQL("select uuid,nick_name_oversea as full_name_en,nick_name_jp as full_name_ja,logo as avatar from job_seeker_basic_info where uuid=?", uuid).Get(&userPO)
	if ok && err == nil {
		entity := convertUserEntity(&userPO)
		entity.UserType = user.JobSeeker
		return entity, nil
	}

	return nil, errors.New("can't find user")
}

// endregion
