package localizationbundle

import (
	"io"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/gin-gonic/gin"
	"github.com/sc-js/pour"
)

var gabsLocale *gabs.Container
var localeKeyCache map[string]string

func InitLocales() {
	jsonFile, err := os.Open("./locales.json")
	if err != nil {
		return
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	gabsLocale, err = gabs.ParseJSON(byteValue)

	if err != nil {
		pour.LogColor(false, pour.ColorRed, "Parsing locales failed")
	}

	localeKeyCache = make(map[string]string)
}

func LocRes(key string, c *gin.Context) string {
	locale := "en_EN"
	if c.Request != nil {
		locale = c.Request.Header.Get("X-LOCALE")
		if len(locale) <= 0 {
			locale = "en_EN"
		}
	}
	localePath := locale + "." + key

	cached, cacheOk := localeKeyCache[localePath]
	if cacheOk {
		return cached
	}

	val, ok := gabsLocale.Path(localePath).Data().(string)
	if !ok {
		return "LOCALE " + localePath + " NOT FOUND"
	} else {
		localeKeyCache[localePath] = val
		return val
	}
}

func GetLocaleFCM(key string, locale string) string {
	if locale == "de_AT" || locale == "de_CH" {
		locale = "de_DE"
	}
	if locale != "de_DE" && locale != "en_EN" {
		locale = "en_EN"
	}

	localePath := locale + "." + key

	cached, cacheOk := localeKeyCache[localePath]
	if cacheOk {
		return cached
	}

	val, ok := gabsLocale.Path(localePath).Data().(string)
	if !ok {
		return "LOCALE " + localePath + " NOT FOUND"
	} else {
		localeKeyCache[localePath] = val
		return val
	}
}
