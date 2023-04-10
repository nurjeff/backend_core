package tools

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var routePermissionMap map[string]uint = make(map[string]uint)

func InitHandlers(r *gin.RouterGroup, routes []GinRoute) {

	for _, element := range routes {
		routePermissionMap[element.Endpoint] = element.Permission
		switch element.Method {

		case (http.MethodGet):
			r.GET(element.Endpoint, element.Handler)
		case (http.MethodPost):
			r.POST(element.Endpoint, element.Handler)
		case (http.MethodPatch):
			r.PATCH(element.Endpoint, element.Handler)
		case (http.MethodDelete):
			r.DELETE(element.Endpoint, element.Handler)
		case (http.MethodPut):
			r.PUT(element.Endpoint, element.Handler)
		case (http.MethodOptions):
			r.OPTIONS(element.Endpoint, element.Handler)
		}

	}
}

func CheckRouteNeedsAuth(endpoint string) bool {
	return routePermissionMap[endpoint] == 0 || routePermissionMap[endpoint] > 1
}

func CheckRouteNeedsAdmin(endpoint string) bool {
	return routePermissionMap[endpoint] >= 2
}

func CreateDirectoryTree(path string) string {

	path = DOCKER_PATH + "/" + path
	parts := strings.Split(path, "/")
	soFar := ""
	for _, element := range parts {
		soFar += element + "/"
		if soFar != "./" {
			if !Exists(soFar) {
				os.Mkdir(soFar, 0755)
			}
		}
	}
	return soFar
}
