package authbundle

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sc-js/core_backend/src/bundles/cachebundle"
	"github.com/sc-js/core_backend/src/errs"
	"github.com/sc-js/core_backend/src/tools"
	"github.com/sc-js/pour"
	"github.com/twinj/uuid"
)

var signSecret = ""
var refreshSecret = ""

// Create a JWT token for a specific user
func CreateToken(userid uint64) (*TokenDetails, error) {
	defer errs.Defer()
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 24 * 30).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 60).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	os.Setenv("SIGN_SECRET", signSecret) //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(signSecret))
	if err != nil {
		return nil, err
	}
	os.Setenv("REFRESH_SECRET", refreshSecret) //this should be in an env file
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(refreshSecret))
	if err != nil {
		return nil, err
	}

	return td, nil
}

// This creates an auth object and saves it into the cache for fast access
// the auth can be revoked at any time by deleting it from the cache, even if the jwt itself is still valid
func CreateAuth(userid uint64, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()
	if err := cachebundle.PutExpire("user_session", td.AccessUuid, int(userid), at.Sub(now)); err != nil {
		log.Println("err 1")
		pour.LogColor(true, pour.ColorRed, err)
		return err
	}
	return cachebundle.PutExpire("user_session", td.RefreshUuid, int(userid), rt.Sub(now))
}

// Extract the JWT from an incoming request
func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// Checks whether or not the JWT is still valid
func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Decrypt and return the JWT Token object from an incoming string
func ExtractJWTTokenFromToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Wrapper function to simplify token checking
func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}

	if !token.Valid {
		return err
	}
	return nil
}

// Get Token information from an incoming WebSocket connection
func ExtractTokenMetadataWS(r *http.Request) (*AccessDetails, error) {
	defer errs.Defer()
	token, err := ExtractJWTTokenFromToken(r.URL.Query().Get("token"))
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}

		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, err
}

// Extract the whole Metadata information (UUID and UserID) from an incoming request
func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	defer errs.Defer()
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}

		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, err
}

// Check the cache if the auth is still valid
func FetchAuth(authD *AccessDetails) (int, error) {
	userid, err := cachebundle.Get[int]("user_session", authD.AccessUuid)
	if err != nil {
		return 0, err
	}
	return userid, nil
}

// Revoke auth by deleting it from the cache
func DeleteAuth(givenUuid string, c *authController) (int, error) {
	userid, err := cachebundle.Get[int]("user_session", givenUuid)
	if err != nil {
		return 0, err
	}

	user := AuthUser{}
	if err := user.GetFromId(tools.ModelID(userid), c); err != nil {
		return 0, err
	}

	pour.LogColor(false, pour.ColorCyan, "AUTH -> Revoking login for User '"+user.Username+"'")
	err = cachebundle.Del("user_session", givenUuid)
	if err != nil {
		return 0, err
	}
	return userid, nil
}

// Wrapper func to simplify auth checking from an incoming Request inside a Gin Handler
func CheckAuth(c *gin.Context) error {
	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		_, err := extractClient(c)
		if err == nil {
			return nil
		}
		tools.RespondWithError(c, http.StatusUnauthorized, "not_authorized")
		return errors.New("not_authorized")
	}
	_, err = FetchAuth(tokenAuth)
	if err != nil {
		tools.RespondWithError(c, http.StatusUnauthorized, "not_authorized")
		return errors.New("not_authorized")
	}

	return nil
}

// Get the users ID from an incoming Gin request
func extractClient(c *gin.Context) (tools.ModelID, error) {
	xclient := c.GetHeader("X-CLIENT")
	if len(xclient) > 0 {
		clientID, err := cachebundle.Get[tools.ModelID]("client_session", xclient+c.GetHeader("Authorization"))
		if err != nil {
			return 0, errors.New("no_client_auth")
		}
		return clientID, nil
	}
	return 0, errors.New("no_client")
}
