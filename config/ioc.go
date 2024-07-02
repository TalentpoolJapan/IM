package config

import (
	appUser "imserver/internal/application/user"
	domainUser "imserver/internal/domain/user"
	"imserver/internal/infrastructure/persistence"
)

var (
	UserAppServ  appUser.AppService
	ImFriendRepo domainUser.ImFriendRepository
)

func init() {
	ImFriendRepo = persistence.NewManticoreImFriendRepo(ManticoreDB)
	UserAppServ = appUser.NewUserAppService(ImFriendRepo)
}
