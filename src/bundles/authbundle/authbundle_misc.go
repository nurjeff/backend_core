package authbundle

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sc-js/backend_core/src/bundles/cachebundle"
	t "github.com/sc-js/backend_core/src/tools"
	"gorm.io/gorm"
)

const (
	CLIENT_TYPE_USER    = 0
	CLIENT_TYPE_VCLIENT = 1
	CLIENT_TYPE_DEFAULT = 2
)

func GetUserIdFromRequest(c *gin.Context) (t.ModelID, int) {
	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		return 0, CLIENT_TYPE_VCLIENT
	}
	userid, err := cachebundle.Get[int]("user_session", tokenAuth.AccessUuid)
	if err != nil {
		return 0, CLIENT_TYPE_DEFAULT
	}
	return t.ModelID(userid), CLIENT_TYPE_USER
}

func GetUserFromToken(c *gin.Context, db *gorm.DB, token string) (AuthUser, error) {
	tokenAuth, err := ExtractTokenMetadataWS(c.Request)
	if err != nil {
		return AuthUser{}, err
	}
	userid, err := cachebundle.Get[int]("user_session", tokenAuth.AccessUuid)
	if err != nil {
		return AuthUser{}, err
	}
	user := AuthUser{}
	err = db.Where("id=?", userid).First(&user).Error
	return user, err
}

func GetUserFromRequest(c *gin.Context, db *gorm.DB) (AuthUser, error) {
	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		return AuthUser{}, err
	}
	userid, err := cachebundle.Get[int]("user_session", tokenAuth.AccessUuid)
	if err != nil {
		return AuthUser{}, err
	}
	user := AuthUser{}
	err = db.Where("id=?", userid).First(&user).Error
	return user, err
}

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Connection") == "Upgrade" {
			c.Next()
			return
		}
		url := c.Request.URL.String()
		if t.CheckRouteNeedsAuth(url) {
			err := CheckAuth(c)
			if err != nil {
				c.Abort()
				return
			}
			if t.CheckRouteNeedsAdmin(url) {
				isAdmin, _ := GetIsAdminFromRequest(c, db)
				if isAdmin {
					c.Next()
				} else {
					t.RespondError(errors.New("not_authorized"), http.StatusUnauthorized, c)
					c.Abort()
					return
				}
			}
			c.Next()
		}
		c.Next()
	}
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-LOCALE")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func GetIsAdminFromRequest(c *gin.Context, db *gorm.DB) (bool, t.ModelID) {
	uid, userType := GetUserIdFromRequest(c)
	if userType == CLIENT_TYPE_VCLIENT {
		isAdmin, err := cachebundle.Get[bool]("client_admin", c.GetHeader("X-CLIENT")+c.GetHeader("Authorization"))
		if err != nil {
			return false, 0
		}
		return isAdmin, uid
	}
	var user AuthUser
	err := db.Select("system_admin").First(&user, uid).Error
	if err == nil {
		return user.SystemAdmin, user.ID
	}
	return false, 0
}

func GetIsAdminFromUser(user AuthUser, db *gorm.DB) (bool, t.ModelID) {
	u := AuthUser{}
	err := db.Select("system_admin").First(&u, user.ID).Error
	if err == nil {
		return user.SystemAdmin, user.ID
	}
	return false, 0
}
