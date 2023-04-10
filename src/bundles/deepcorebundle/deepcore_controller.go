package deepcorebundle

import (
	"reflect"

	"github.com/sc-js/pour"

	"gorm.io/gorm"
)

type Controller struct {
}

var autoMigrate bool
var db *gorm.DB

// Initialize the parent controller for all other bundle controllers.
// This handles auto migration of models into the relational database, and can be used to define interface funcs.
func Init(database *gorm.DB, migrate bool) {
	autoMigrate = migrate
	db = database
}

// Register a certain Interface, auto migrates it into the database and adds specified allowed filters,
// e.G. for getter Handlers
func RegisterModel(m interface{}, filters []string) {
	if autoMigrate {
		err := db.AutoMigrate(&m)
		if err != nil {
			pour.LogColor(false, pour.ColorRed, "Error auto migrating:", err)
			return
		}
	}
	addAllowedFilters(m, filters)
}

func addAllowedFilters(m interface{}, filters []string) {
	allowedFilters[reflect.TypeOf(m).Name()] = filters
}

var allowedFilters = make(map[string][]string)

func GetAllowedFilters(m interface{}) []string {
	return allowedFilters[reflect.TypeOf(m).Name()]
}
