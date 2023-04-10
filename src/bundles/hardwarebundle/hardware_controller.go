package hardwarebundle

import (
	"github.com/sc-js/core_backend/src/bundles/deepcorebundle"
	"github.com/sc-js/core_backend/src/tools"
)

type hardwareController struct {
	deepcorebundle.Controller
	DataWrap *tools.DataWrap
}

func initialize(wrap *tools.DataWrap, autoMigrate bool, settings map[string]string) *hardwareController {
	c := &hardwareController{Controller: deepcorebundle.Controller{}, DataWrap: wrap}
	deepcorebundle.RegisterModel(hardwareUsage{}, []string{"memory_total"})
	handleSettings(settings, c.DataWrap)

	return c
}

func handleSettings(settings map[string]string, wrap *tools.DataWrap) {
	if settings == nil {
		return
	}
	if settings["polling"] == "true" {
		go pollUsage(wrap.DB)
	}
}
