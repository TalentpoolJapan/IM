package config

import (
	appUser "imserver/internal/application/user"
	"imserver/internal/domain/imfriend"
	"imserver/internal/domain/immessage"
	domainUser "imserver/internal/domain/user"
	"imserver/internal/infrastructure/persistence"
	"xorm.io/xorm"
)

var (
	UserAppServ   appUser.AppService
	ImFriendRepo  imfriend.ImFriendRepository
	ImMessageRepo immessage.ImMessageRepository
	UserRepo      domainUser.IUserRepository
)

func InitIoc(manticore *xorm.Engine, mysql *xorm.Engine) {
	// repo
	UserRepo = persistence.NewMysqlUserRepository(mysql)
	ImMessageRepo = persistence.NewManticoreImMessageRepo(manticore)
	ImFriendRepo = persistence.NewManticoreImFriendRepo(manticore)

	// app
	UserAppServ = appUser.NewUserAppService(ImFriendRepo, UserRepo, ImMessageRepo)
}
