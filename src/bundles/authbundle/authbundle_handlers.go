package authbundle

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	t "github.com/sc-js/core_backend/src/tools"
	"github.com/sc-js/pour"
)

func (con *authController) logoutHandler(c *gin.Context) {
	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		fmt.Println(err)
		t.RespondError(err, http.StatusUnauthorized, c, "not_authorized")
		return
	}
	_, delErr := DeleteAuth(tokenAuth.AccessUuid, con)
	if delErr != nil {
		t.RespondError(err, http.StatusUnauthorized, c, "not_authorized")
		return
	}
	t.RespondWithJSON(c, http.StatusOK, "Successfully logged out")
}

func (con *authController) registerHandler(c *gin.Context) {
	user := AuthUser{}
	if err := c.BindJSON(&user); err != nil {
		t.RespondError(err, http.StatusForbidden, c)
		return
	}
	user.Password = t.GetMD5(user.Password)
	user.UserType = USERTYPE_USER
	con.DataWrap.DB.Create(&user)
	t.RespondWithJSON(c, http.StatusOK, &user)
}

func (con *authController) updateUserHandler(c *gin.Context) {
	t.Update[AuthUser](AuthUser{}, con.DataWrap.DB, c)
}

func (con *authController) uploadUserByIdImageHandler(c *gin.Context) {
	user, err := t.GetSingleById[AuthUser](c, con.DataWrap.DB)
	if err != nil {
		t.RespondError(errors.New("not_found"), http.StatusNotFound, c)
		return
	}
	userId, _ := GetUserIdFromRequest(c)
	if user.ID == userId {
		t.SaveUploadedFile(c, "users/"+t.Encode(userId)+"/images", user.ID, "avatar.jpg", true)
		return
	}
	ad, _ := GetIsAdminFromRequest(c, con.DataWrap.DB)
	if ad {
		t.SaveUploadedFile(c, "users/"+t.Encode(userId)+"/images", user.ID, "avatar.jpg", true)
		return
	}
	t.RespondError(errors.New("not_authorized"), http.StatusUnauthorized, c)
}

// TODO
func (con *authController) deleteUserByIdImageHandler(c *gin.Context) {
	user, err := t.GetSingleById[AuthUser](c, con.DataWrap.DB)
	if err != nil {
		t.RespondError(errors.New("not_found"), http.StatusNotFound, c)
		return
	}
	userId, _ := GetUserIdFromRequest(c)
	if user.ID == userId {
		t.SaveUploadedFile(c, "users/"+t.Encode(userId)+"/images", user.ID, "avatar.jpg", true)
		return
	}
	ad, _ := GetIsAdminFromRequest(c, con.DataWrap.DB)
	if ad {
		t.SaveUploadedFile(c, "users/"+t.Encode(userId)+"/images", user.ID, "avatar.jpg", true)
		return
	}
	t.RespondError(errors.New("not_authorized"), http.StatusUnauthorized, c)
}

func (con *authController) getUserImageHandler(c *gin.Context) {
	userId, _ := GetUserIdFromRequest(c)
	t.ServeFile(c, "users/"+t.Encode(userId)+"/images/avatar.jpg", userId)
}

func (con *authController) getUserImageByIdHandler(c *gin.Context) {
	user, err := t.GetSingleById[AuthUser](c, con.DataWrap.DB)
	if err != nil {
		t.RespondError(err, http.StatusNotFound, c)
		return
	}
	id, clientType := GetUserIdFromRequest(c)
	if clientType != CLIENT_TYPE_USER {
		t.RespondError(errors.New("only_user"), http.StatusNotImplemented, c)
		return
	}
	t.ServeFile(c, "users/"+t.Encode(user.ID)+"/images/avatar.jpg", id)
}

func (con *authController) uploadUserImageHandler(c *gin.Context) {
	userId, clientType := GetUserIdFromRequest(c)
	if clientType != CLIENT_TYPE_USER {
		t.RespondError(errors.New("only_user"), http.StatusNotImplemented, c)
		return
	}
	t.SaveUploadedFile(c, "users/"+t.Encode(userId)+"/images", userId, "avatar.jpg", true)
}

func (con *authController) getUserByIdHandler(c *gin.Context) {
	t.GetSingleByIdAndSend[AuthUser](c, con.DataWrap.DB)
}

func (con *authController) getUserHandler(c *gin.Context) {
	user := AuthUser{}
	details, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		t.RespondError(errors.New("not_authorized"), http.StatusForbidden, c)
		return
	}
	uid, err := FetchAuth(details)

	if err != nil {
		t.RespondError(errors.New("not_authorized"), http.StatusForbidden, c)
		return
	}
	con.DataWrap.DB.Where("id=?", uid).First(&user)
	t.RespondWithJSON(c, http.StatusOK, &user)
}

func (con *authController) loginHandler(c *gin.Context) {
	var user AuthUser
	if err := c.BindJSON(&user); err != nil {
		t.RespondError(err, http.StatusBadRequest, c)
		return
	}

	if len(user.Password) <= 0 || len(user.Username) <= 0 {
		t.RespondError(errors.New("bad_login"), http.StatusBadRequest, c)
		return
	}
	cryptPassword := t.GetMD5(user.Password)
	var u AuthUser
	if err := con.DataWrap.DB.Where("username = ? AND password = ?", user.Username, cryptPassword).First(&u).Error; err != nil {
		t.RespondError(errors.New("not_found"), http.StatusForbidden, c)
		return
	}

	if user.Username != u.Username || cryptPassword != u.Password {
		t.RespondError(errors.New("bad_login"), http.StatusUnauthorized, c)
		return
	}

	token, err := CreateToken(uint64(u.ID))
	if err != nil {
		log.Println("1")
		t.RespondError(errors.New("auth_error"), http.StatusUnauthorized, c)
		return
	}

	saveErr := CreateAuth(uint64(u.ID), token)
	if saveErr != nil {
		log.Println(saveErr)
		t.RespondError(errors.New("auth_error"), http.StatusUnauthorized, c)
		return
	}
	tokens := map[string]string{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
	}

	sendToken := UserLogin{
		User:   u,
		Tokens: tokens,
	}

	t.RespondWithJSON(c, http.StatusOK, sendToken)
	pour.LogColor(false, pour.ColorCyan, "AUTH -> User '"+user.Username+"' logged in")
}

func (con *authController) refreshHandler(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.BindJSON(&mapToken); err != nil {
		t.RespondError(errors.New("internal_error"), http.StatusUnprocessableEntity, c)
		return
	}
	refreshToken := mapToken["refresh_token"]

	os.Setenv("REFRESH_SECRET", refreshSecret) //this should be in an env file
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(refreshSecret), nil
	})
	if err != nil {
		t.RespondError(errors.New("auth_error"), http.StatusUnprocessableEntity, c)
		return
	}
	if !token.Valid {
		t.RespondError(errors.New("not_authorized"), http.StatusUnauthorized, c)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string)
		if !ok {
			t.RespondError(errors.New("internal_error"), http.StatusUnprocessableEntity, c)
			return
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			t.RespondError(errors.New("internal_error"), http.StatusUnprocessableEntity, c)
			return
		}
		deleted, delErr := DeleteAuth(refreshUuid, con)
		if delErr != nil || deleted == 0 {
			t.RespondError(errors.New("not_authorized"), http.StatusUnauthorized, c)
			return
		}
		ts, createErr := CreateToken(userId)
		if createErr != nil {
			t.RespondError(errors.New("not_authorized"), http.StatusUnprocessableEntity, c)
			return
		}
		saveErr := CreateAuth(userId, ts)
		if saveErr != nil {
			t.RespondError(errors.New("not_authorized"), http.StatusUnprocessableEntity, c)
			return
		}
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		t.RespondWithJSON(c, http.StatusOK, &tokens)
	} else {
		t.RespondError(errors.New("auth_error"), http.StatusUnprocessableEntity, c)
	}
}
