package authbundle

import (
	"net/http"

	"github.com/gin-gonic/gin"
	t "github.com/sc-js/backend_core/src/tools"
)

var routes []t.GinRoute

func InitBundle(r *gin.RouterGroup, wrap *t.DataWrap, withRegister bool, settings map[string]string) {
	controller := initialize(wrap, settings)

	routes = []t.GinRoute{
		{Method: http.MethodPost, Endpoint: "/auth/login", Handler: controller.loginHandler, Permission: t.PERM_ZERO},
		{Method: http.MethodPost, Endpoint: "/auth/refresh", Handler: controller.refreshHandler},
		{Method: http.MethodPost, Endpoint: "/auth/logout", Handler: controller.logoutHandler},
		{Method: http.MethodGet, Endpoint: "/auth/user", Handler: controller.getUserHandler},
		{Method: http.MethodGet, Endpoint: "/auth/user/:hid", Handler: controller.getUserByIdHandler},
		{Method: http.MethodPatch, Endpoint: "/auth/user/:hid", Handler: controller.updateUserHandler},

		//Images
		{Method: http.MethodGet, Endpoint: "/auth/user/image", Handler: controller.getUserImageHandler},
		{Method: http.MethodPost, Endpoint: "/auth/user/image", Handler: controller.uploadUserImageHandler},

		{Method: http.MethodPost, Endpoint: "/auth/user/:hid/image", Handler: controller.uploadUserByIdImageHandler},
		{Method: http.MethodGet, Endpoint: "/auth/user/:hid/image", Handler: controller.getUserImageByIdHandler},
		{Method: http.MethodDelete, Endpoint: "/auth/user/:hid/image", Handler: controller.deleteUserByIdImageHandler},
	}

	if withRegister {
		routes = append(routes, t.GinRoute{Method: http.MethodPost, Endpoint: "/auth/register", Handler: controller.registerHandler, Permission: t.PERM_ZERO})
	}

	t.InitHandlers(r, routes)
}
