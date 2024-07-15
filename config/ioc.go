package config

import (
	appUser "imserver/internal/application/user"
	domainUser "imserver/internal/domain/user"
	"imserver/internal/infrastructure/persistence"
	"xorm.io/xorm"
)

var (
	UserAppServ   appUser.AppService
	ImFriendRepo  domainUser.ImFriendRepository
	ImMessageRepo domainUser.ImMessageRepository
	UserRepo      domainUser.IUserRepository
)

func InitIoc(manticore *xorm.Engine, mysql *xorm.Engine) {
	UserRepo = persistence.NewMysqlUserRepository(mysql)
	ImMessageRepo = persistence.NewManticoreImMessageRepo(manticore)
	ImFriendRepo = persistence.NewManticoreImFriendRepo(manticore)
	UserAppServ = appUser.NewUserAppService(ImFriendRepo, UserRepo, ImMessageRepo)
}
