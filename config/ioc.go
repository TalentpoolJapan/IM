package config

import (
	appUser "imserver/internal/application/user"
	domainUser "imserver/internal/domain/user"
	"imserver/internal/infrastructure/persistence"
	"xorm.io/xorm"
)

var (
	UserAppServ        appUser.AppService
	ImFriendRepo       domainUser.ImFriendRepository
	ImMessageRepo      domainUser.ImMessageRepository
	UserRepo           domainUser.IUserRepository
	ImFriendDomainServ domainUser.ImFriendDomainService
)

func InitIoc(manticore *xorm.Engine, mysql *xorm.Engine) {
	// repo
	UserRepo = persistence.NewMysqlUserRepository(mysql)
	ImMessageRepo = persistence.NewManticoreImMessageRepo(manticore)
	ImFriendRepo = persistence.NewManticoreImFriendRepo(manticore)

	// domain
	ImFriendDomainServ = domainUser.NewImFriendDomainService(ImFriendRepo, ImMessageRepo)

	// app
	UserAppServ = appUser.NewUserAppService(ImFriendRepo, UserRepo, ImMessageRepo, ImFriendDomainServ)
}
