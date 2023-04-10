package websocketbundle

import (
	"github.com/sc-js/core_backend/src/bundles/deepcorebundle"
	"github.com/sc-js/core_backend/src/errs"
	"github.com/sc-js/core_backend/src/tools"
)

type websocketController struct {
	deepcorebundle.Controller
	DataWrap *tools.DataWrap
}

var wshub *hub
var allowConnections = PERM_LOGIN

const (
	PERM_LOGIN = "PERM_LOGIN"
	PERM_ADMIN = "PERM_ADMIN"
	PERM_NONE  = "PERM_NONE"
)

func initialize(wrap *tools.DataWrap, settings map[string]string) *websocketController {
	defer errs.Defer()
	c := &websocketController{Controller: deepcorebundle.Controller{}, DataWrap: wrap}
	handleSettings(settings, wrap)
	wshub = newHub(wrap)
	go wshub.run()

	return c
}

func handleSettings(settings map[string]string, wrap *tools.DataWrap) {
	if settings == nil {
		return
	}
	allowConnections = PERM_ADMIN
	switch settings["permission"] {
	case (PERM_LOGIN):
		allowConnections = PERM_LOGIN
	case (PERM_NONE):
		allowConnections = PERM_NONE
	}
}
