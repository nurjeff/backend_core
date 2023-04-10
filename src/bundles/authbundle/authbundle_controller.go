package authbundle

import (
	"github.com/sc-js/core_backend/src/bundles/cachebundle"
	"github.com/sc-js/core_backend/src/bundles/deepcorebundle"
	"github.com/sc-js/core_backend/src/tools"
	"github.com/sc-js/pour"
)

type authController struct {
	deepcorebundle.Controller
	DataWrap *tools.DataWrap
}

func initialize(wrap *tools.DataWrap, settings map[string]string) *authController {
	c := &authController{Controller: deepcorebundle.Controller{}, DataWrap: wrap}

	handleSettings(settings, wrap)
	ReloadVClients(wrap)

	deepcorebundle.RegisterModel(AuthUser{}, []string{"first_name"})
	return c
}

func handleSettings(settings map[string]string, warp *tools.DataWrap) {
	if settings == nil {
		return
	} else {
		signSecret = settings["jwt_secret"]
		refreshSecret = settings["refresh_secret"]
	}
}

func ReloadVClients(wrap *tools.DataWrap) {
	users := []AuthUser{}
	if err := wrap.DB.Where("user_type=?", CLIENT_TYPE_VCLIENT).Find(&users).Error; err != nil {
		pour.LogColor(false, pour.ColorYellow, "No VClients registered")
		return
	}

	names := []string{}
	for _, element := range users {
		cachebundle.Put("client_session", element.VClientName+element.VClientHash, element.ID)
		names = append(names, element.VClientName)
	}
	pour.LogColor(false, pour.ColorYellow, "Added VClients:", names)
}
