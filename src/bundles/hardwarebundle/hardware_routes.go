package hardwarebundle

import (
	"net/http"

	"github.com/gin-gonic/gin"
	t "github.com/sc-js/core_backend/src/tools"
)

var routes []t.GinRoute

func InitBundle(r *gin.RouterGroup, wrap *t.DataWrap, autoMigrate bool, settings map[string]string) {
	controller := initialize(wrap, autoMigrate, settings)

	routes = []t.GinRoute{
		{Method: http.MethodGet, Endpoint: "/hardware/configuration", Handler: controller.getHardwareConfigurationHandler, Permission: t.PERM_ADMIN},
		{Method: http.MethodGet, Endpoint: "/hardware/usage", Handler: controller.getHardwareUsageHandler, Permission: t.PERM_ADMIN},
	}

	t.InitHandlers(r, routes)
}
