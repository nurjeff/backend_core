package tools

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sc-js/core_backend/src/bundles/deepcorebundle"
	"github.com/sc-js/core_backend/src/errs"
	"github.com/sc-js/pour"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetPagedAndSend[T any](c *gin.Context, db *gorm.DB) (Paging, error) {
	p, err := GetPaged[T](c, db)
	if err != nil {
		RespondError(err, http.StatusBadRequest, c)
		return Paging{}, err
	}
	RespondWithJSON(c, http.StatusOK, &p)
	return p, nil
}

func GetPaged[T any](c *gin.Context, db *gorm.DB) (Paging, error) {
	defer errs.Defer()
	single := new(T)
	multi := new([]T)
	availableOrder := deepcorebundle.GetAllowedFilters(single)
	page, perPage, order, orderDir := getPageInfo(c, availableOrder)
	var count int64

	orderDB := db.Model(single)
	for _, element := range order {
		orderDB = orderDB.Order(element + " " + orderDir)
		break
	}

	orderDB.Count(&count)

	err := orderDB.Offset(int(perPage) * int(page)).Limit(int(perPage)).Find(&multi).Error

	if err != nil {
		return Paging{}, err
	}
	paging := Paging{TotalCount: count, Page: page, PerPage: perPage, Data: multi}

	return paging, nil
}

func SaveUploadedFiles(c *gin.Context, destination string, userId ModelID, filenameOverride string, respond bool) error {
	defer errs.Defer()
	form, _ := c.MultipartForm()
	files := form.File["upload[]"]
	if files == nil {
		if respond {
			RespondError(errors.New("file_nil"), http.StatusBadRequest, c)
		}
		return errors.New("file_nil")
	}
	destination = CreateDirectoryTree(destination)
	for _, file := range files {
		pour.LogColor(true, pour.ColorPurple, "User", userId, "uploaded:", file.Filename, file.Size)
		if len(filenameOverride) > 0 {
			c.SaveUploadedFile(file, destination+filenameOverride)
		} else {
			c.SaveUploadedFile(file, destination+file.Filename)
		}
	}
	if respond {
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	}
	return nil
}

func ServeFile(c *gin.Context, path string, userId ModelID) error {
	defer errs.Defer()
	pour.LogColor(true, pour.ColorPurple, "Serving File:", path, "to", userId)
	c.File(DOCKER_PATH + "/" + path)
	return nil
}

func SaveUploadedFile(c *gin.Context, destination string, userId ModelID, filenameOverride string, respond bool) error {
	defer errs.Defer()
	file, _ := c.FormFile("file")
	if file == nil {
		if respond {
			RespondError(errors.New("file_nil"), http.StatusBadRequest, c)
		}
		return errors.New("file_nil")
	}
	destination = CreateDirectoryTree(destination)
	pour.LogColor(true, pour.ColorPurple, "User", userId, "uploaded:", file.Filename, file.Size)
	if len(filenameOverride) > 0 {
		c.SaveUploadedFile(file, destination+filenameOverride)
	} else {
		c.SaveUploadedFile(file, destination+file.Filename)
	}

	if respond {
		c.String(http.StatusOK, fmt.Sprintf("%d file uploaded!", file.Size))
	}
	return nil
}

func getPageInfo(c *gin.Context, filters []string) (uint64, uint64, []string, string) {
	var page uint64 = 0
	var perPage uint64 = 50
	orderDir := "ASC"
	queryVals := c.Request.URL.Query()
	order := queryVals["order"]
	parsedPage, err := strconv.ParseUint(queryVals.Get("page"), 10, 64)
	if err == nil {
		page = parsedPage
	}
	parsedPerPage, err := strconv.ParseUint(queryVals.Get("per_page"), 10, 64)
	if err == nil {
		perPage = parsedPerPage
	}

	for index, element := range order {
		if strings.Contains(element, ",ASC") {
			orderDir = "ASC"
		}
		if strings.Contains(element, ",DESC") {
			orderDir = "DESC"
		}
		order[index] = strings.ReplaceAll(order[index], ",ASC", "")
		order[index] = strings.ReplaceAll(order[index], ",DESC", "")
	}

	order = FilterStringSlice(order, filters)

	return page, perPage, order, orderDir
}

func RespondWithJSON(c *gin.Context, code int, payload interface{}) {
	defer errs.Defer()
	tr := tryTranslate(payload, c).(reflect.Value)
	c.JSON(code, tr.Interface())
	go logRequestDetails(c, code, tr.Interface())
}

func RespondWithJsonSilent(c *gin.Context, code int, payload interface{}) {
	defer errs.Defer()
	tr := tryTranslate(payload, c).(reflect.Value)
	c.JSON(code, tr.Interface())
	//go logRequestDetails(c, code, tr.Interface())
}

func logRequestDetails(c *gin.Context, code int, payload interface{}) {
	logStr := ""
	logStr += c.Request.Method + ":" + c.Request.RequestURI + ":" + fmt.Sprint(code) + ":" + c.Request.RemoteAddr
	if code != http.StatusOK && code != http.StatusAccepted {
		pour.LogColor(true, pour.ColorRed, logStr)
		return
	}
	pour.LogColor(true, pour.ColorWhite, logStr)
}

func getLocaleFromRequest(c *gin.Context) string {
	loc := c.GetHeader("X-LOCALE")
	if ValidatorCallback(loc) {
		return loc
	}
	return "en_EN"

}

func RespondWithError(c *gin.Context, code int, message string) {
	RespondWithJSON(c, code, map[string]string{"error": message})
}

func RespondError(err error, code int, c *gin.Context, v ...any) {
	defer errs.Defer()
	var msg ErrorMessage
	locale := getLocaleFromRequest(c)
	if len(v) > 0 {
		errMessage := fmt.Sprint(v[0])
		trans, transErr := SingleTranslationCallback(locale, errMessage)
		if transErr != nil || len(trans) == 0 {
			trans = errMessage
		}
		msg = ErrorMessage{Code: code, Error: err.Error(), Localized: trans}
		c.JSON(code, msg)
	} else {
		trans, transErr := SingleTranslationCallback(locale, err.Error())
		if transErr != nil || len(trans) == 0 {
			trans = err.Error()
		}
		msg = ErrorMessage{Code: code, Error: err.Error(), Localized: trans}
		c.JSON(code, msg)
	}
	go logRequestDetails(c, code, msg)
	c.Error(err)
}

func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, ginErr := range c.Errors {
			logger.Error(ginErr.Error())
		}
	}
}
