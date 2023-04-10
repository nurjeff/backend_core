package tools

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetSingle[T any](db *gorm.DB) (T, error) {

	t := new(T)
	return *t, db.First(&t).Error
}

func GetAll[T any](db *gorm.DB) ([]T, error) {

	t := new([]T)
	return *t, db.Find(&t).Error
}

func GetSingleById[T any](c *gin.Context, db *gorm.DB) (T, error) {

	t := new(T)
	return *t, db.Where("id=?", Decode(c.Param("hid"))).First(&t).Error
}

func GetSingleByIdAndSend[T any](c *gin.Context, db *gorm.DB) (T, error) {

	b, err := GetSingleById[T](c, db)
	if err != nil {
		RespondError(err, http.StatusForbidden, c)
		return *new(T), err
	}
	RespondWithJSON(c, http.StatusOK, &b)
	return b, nil
}

func Update[T any](obj interface{}, db *gorm.DB, c *gin.Context) (T, error) {

	t := obj.(T)
	c.Bind(&t)
	o, err := updateObject(t, db, c)
	if err != nil {
		RespondError(err, http.StatusBadRequest, c)
		return t, err
	}
	oc := o.(T)
	if err := db.Save(&oc).Error; err != nil {
		RespondError(err, http.StatusBadRequest, c)
		return t, err
	}
	RespondWithJSON(c, http.StatusOK, &oc)
	return oc, err
}

func updateObject(object interface{}, db *gorm.DB, c *gin.Context) (interface{}, error) {
	id := Decode(c.Param("hid"))
	objectType := reflect.TypeOf(object)
	copy := reflect.New(objectType).Interface()
	if err := db.First(&copy, "id=?", id).Error; err != nil {
		return object, err
	}
	ret := autoPatch(object, copy)
	return ret, nil
}

func autoPatch(x interface{}, y interface{}) interface{} {
	valOld := reflect.ValueOf(y)
	valNew := reflect.ValueOf(x)

	fieldsOld := make(map[string]*typeValue)
	fieldsNew := make(map[string]*typeValue)

	var wg sync.WaitGroup
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(j int) {
			defer wg.Done()
			var z reflect.Value
			var v reflect.Value
			if j == 0 {
				v = valOld
			} else {
				v = valNew
			}
			if v.Kind() == reflect.Ptr {
				z = v.Elem()
			} else {
				z = v
			}
			t := z.Type()
			for i := 0; i < z.NumField(); i++ {
				field := z.Field(i)
				if !reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) && t.Field(i).Tag.Get("update") != "false" {
					if j == 0 {
						fieldsOld[t.Field(i).Name] = &typeValue{Type: field.Type().Kind(), Value: field.Interface()}
					} else {
						fieldsNew[t.Field(i).Name] = &typeValue{Type: field.Type().Kind(), Value: field.Interface()}
					}
				}
			}
		}(i)
	}

	wg.Wait()

	if valOld.Kind() == reflect.Ptr {
		valOld = valOld.Elem()
	}
	if valNew.Kind() == reflect.Ptr {
		valNew = valNew.Elem()
	}

	wg.Add(len(fieldsNew))
	for key, val := range fieldsNew {
		go func(km string, vm *typeValue) {
			defer wg.Done()
			if valOld.FieldByName(km).CanSet() {
				valOld.FieldByName(km).Set(valNew.FieldByName(km))
			}
		}(key, val)
	}
	wg.Wait()

	return valOld.Interface()
}

/*func write(oldField reflect.Value, newField reflect.Value, fieldType reflect.Kind) {
	//Sieht schlimm aus, wird aber vom compiler zur hashmap optimiert
	oldField.Set(newField)
	/*switch fieldType {
	case reflect.String:
		oldField.SetString(newField.String())
	case reflect.Int:
		oldField.SetInt(newField.Int())
	case reflect.Int8:
		oldField.SetInt(newField.Int())
	case reflect.Int16:
		oldField.SetInt(newField.Int())
	case reflect.Int32:
		oldField.SetInt(newField.Int())
	case reflect.Int64:
		oldField.SetInt(newField.Int())
	case reflect.Float32:
		oldField.SetFloat(newField.Float())
	case reflect.Float64:
		oldField.SetFloat(newField.Float())
	case reflect.Bool:
		oldField.SetBool(newField.Bool())
	case reflect.Uint:
		oldField.SetUint(newField.Uint())
	case reflect.Uint8:
		oldField.SetUint(newField.Uint())
	case reflect.Uint16:
		oldField.SetUint(newField.Uint())
	case reflect.Uint32:
		oldField.SetUint(newField.Uint())
	case reflect.Uint64:
		oldField.SetUint(newField.Uint())
	}
}*/

type typeValue struct {
	Type  reflect.Kind `json:"type"`
	Value interface{}  `json:"value"`
}
