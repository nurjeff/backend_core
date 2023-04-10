package tools

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

type translationOperator func(obj reflect.Value, locale string) interface{}
type localeValidator func(loc string) bool
type singleTranslationOperator func(locale string, str string) (string, error)

var TranslationCallback translationOperator
var ValidatorCallback localeValidator
var SingleTranslationCallback singleTranslationOperator

func tryTranslate(x interface{}, c *gin.Context) interface{} {

	var val reflect.Value
	vo := reflect.ValueOf(x)
	if vo.Kind() == reflect.Ptr {
		val = vo.Elem()
	} else {
		val = vo
	}

	return TranslationCallback(val, getLocaleFromRequest(c))
}
