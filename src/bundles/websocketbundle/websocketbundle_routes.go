package websocketbundle

import (
	"net/http"

	"github.com/gin-gonic/gin"
	t "github.com/sc-js/core_backend/src/tools"
)

var routes []t.GinRoute

func InitBundle(r *gin.RouterGroup, wrap *t.DataWrap, autoMigrate bool, settings map[string]string) {
	controller := initialize(wrap, settings)
	_ = controller

	routes = []t.GinRoute{
		{Method: http.MethodGet, Endpoint: "/", Handler: controller.upgradeWSHandler, Permission: t.PERM_ZERO},
	}

	t.InitHandlers(r, routes)
}
