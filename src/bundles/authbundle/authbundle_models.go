package authbundle

import (
	"encoding/json"

	"github.com/sc-js/backend_core/src/tools"
)

const (
	USERTYPE_USER   = 0
	USERTYPE_CLIENT = 1
)

type AuthUser struct {
	tools.Model
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password,omitempty" update:"false"`
	SystemAdmin bool   `json:"-" update:"false"`
	UserType    int    `json:"-" update:"false"`
	VClientName string `json:"-" update:"false"`
	VClientHash string `json:"-" update:"false"`
}

type UserLogin struct {
	User   AuthUser          `json:"user"`
	Tokens map[string]string `json:"tokens"`
}

func (u AuthUser) MarshalJSON() ([]byte, error) {
	type Alias AuthUser
	return json.Marshal(&struct {
		ID       tools.ModelID `json:"id"`
		Password string        `json:"password,omitempty"`
		Alias
	}{
		ID:       u.ID,
		Password: "",
		Alias:    (Alias)(u),
	})
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func (u *AuthUser) GetFromId(id tools.ModelID, c *authController) error {
	return c.DataWrap.DB.Where("id=?", id).First(&u).Error
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}
