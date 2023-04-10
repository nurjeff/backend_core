package websocketbundle

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sc-js/backend_core/src/bundles/authbundle"
	"github.com/sc-js/backend_core/src/tools"
)

func (con *websocketController) upgradeWSHandler(c *gin.Context) {

	if allowConnections == PERM_NONE {
		tools.RespondWithError(c, http.StatusForbidden, "not_authorized")
		return
	}
	user, err := authbundle.GetUserFromToken(c, con.DataWrap.DB, c.Query("token"))
	if err != nil {
		tools.RespondWithError(c, http.StatusForbidden, "not_authorized")
		return
	}
	switch allowConnections {
	case PERM_LOGIN:
		serveWs(wshub, c.Writer, c.Request, user)
		return
	case PERM_ADMIN:
		{
			admin, _ := authbundle.GetIsAdminFromUser(user, con.DataWrap.DB)
			if !admin {
				tools.RespondWithError(c, http.StatusForbidden, "not_authorized")
				return
			}
			serveWs(wshub, c.Writer, c.Request, user)
			return
		}
	}
	tools.RespondWithError(c, http.StatusForbidden, "not_authorized")
}
