package config

import (
	appUser "imserver/internal/application/user"
	domainUser "imserver/internal/domain/user"
	"imserver/internal/infrastructure/persistence"
	"xorm.io/xorm"
)

var (
	UserAppServ  appUser.AppService
	ImFriendRepo domainUser.ImFriendRepository
)

func InitIoc(manticore *xorm.Engine) {
	ImFriendRepo = persistence.NewManticoreImFriendRepo(manticore)
	UserAppServ = appUser.NewUserAppService(ImFriendRepo)
}
