package mongowrap

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/sc-js/core_backend/src/errs"
	"github.com/sc-js/pour"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Wrapper struct for all Mongo related fields, to make inserting and retreiving data easier
type Mongo struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type Query struct {
	Where          *Where
	Sort           *Sort
	MongoPaging    *MongoPaging
	CollectionName string
}

type MongoPaging struct {
	Page    int64
	PerPage int64
}

type Where map[string]interface{}

// Sort defined the sorting behaviour of the mongo driver, valid values are
// &mwrap.Sort{"voltage": 1} or &mwrap.Sort{"voltage", -1}, for ascending/descending behaviour
// Values other than 1 and -1 might not work as expected
type Sort map[string]int

// Gets a generic slice of objects from the database, usage:
//
// data, err := mwrap.GetMany[Vehicle](mongoClient, mwrap.Where(Key: "customer_name", Value: "Teltonika"))
//
// Need to pass a generic Datatype to the function, else this will not compile
// Automatically parses the collection name in snake case from reflection name
// You can pass nil for the collectionName, which will automatically infer it from the structs name
func GetMany[T any](m *Mongo, query *Query, collectionName string) ([]T, error) {
	defer errs.Defer()
	var data []T = []T{}
	filter := &bson.D{}
	sort := &bson.D{}

	name, err := primeQuery(m, data, query, filter, sort)
	if err != nil {
		return data, err
	}

	name = toSnakeCase(name)

	var cursor *mongo.Cursor
	var cursorError error

	options := getFindOptions(query)

	if len(collectionName) > 0 {
		name = collectionName
	}

	cursor, cursorError = m.Database.Collection(name).Find(context.TODO(), filter, options)
	if err != nil {
		return data, cursorError
	}

	if err := cursor.All(context.TODO(), &data); err != nil {
		return []T{}, err
	}

	return data, nil
}

func GetTotalCount(m *Mongo, query *Query, collectionName string) (int64, error) {
	return m.Database.Collection(collectionName).CountDocuments(context.TODO(), query.Where, nil)
}

func getFindOptions(query *Query) *options.FindOptions {
	opts := options.FindOptions{}
	if query == nil {
		return nil
	}

	if query.Sort != nil {
		opts.Sort = query.Sort
	}

	if query.MongoPaging != nil {
		if query.MongoPaging.Page < 0 {
			query.MongoPaging.Page = 0
		}
		if query.MongoPaging.PerPage < 0 {
			query.MongoPaging.PerPage = 1
		}
		var skip int64 = 0
		skip = query.MongoPaging.Page * query.MongoPaging.PerPage
		opts.Skip = &skip

		opts.Limit = &query.MongoPaging.PerPage
	}

	return &opts
}

// Gets a generic single object from the database, usage:
//
// data, err := mwrap.Get[Vehicle](mongoClient, mwrap.Where(Key: "customer_name", Value: "Teltonika"))
//
// Need to pass a generic Datatype to the function, else this will not compile
// Automatically parses the collection name in snake case from reflection name
func Get[T any](m *Mongo, query *Query) (T, error) {
	defer errs.Defer()
	var data T
	filter := &bson.D{}
	sort := &bson.D{}

	name, err := primeQuery(m, data, query, filter, sort)
	name = toSnakeCase(name)
	if err != nil {
		return data, err
	}

	if len(query.CollectionName) > 0 {
		name = query.CollectionName
	}

	var cursor *mongo.SingleResult

	if len(*sort) > 0 {
		cursor = m.Database.Collection(name).FindOne(context.TODO(), filter, &options.FindOneOptions{Sort: sort})
	} else {
		cursor = m.Database.Collection(name).FindOne(context.TODO(), filter)
	}

	if err != nil {
		return data, cursor.Err()
	}

	if err := cursor.Decode(&data); err != nil {
		pour.LogErr(err)
		return *new(T), err
	}

	return data, nil
}

// Primes a query, checks nil values and auto fills everything, returns reflected object name for the collection
func primeQuery[T any](m *Mongo, data T, query *Query, filter *bson.D, sort *bson.D) (string, error) {
	if err := m.checkNil(); err != nil {
		return "", err
	}

	name := query.CollectionName
	if len(query.CollectionName) == 0 {
		name = reflect.TypeOf(data).Elem().Name()
		if len(name) <= 0 {
			return "", errors.New("MONGO: error parsing reflection name of generic")
		}
	}
	// Automatically get the name of the getter struct to determine collection name

	// Fill mongo filters and sorts
	if err := fillFilter(filter, query); err != nil {
		return "", err
	}
	if err := fillSort(sort, query); err != nil {
		return "", err
	}

	return name, nil
}

// Auto fill a filter pointer depending on a given query
func fillFilter(filter *bson.D, query *Query) error {
	if query != nil {
		if query.Where != nil {
			if len(*query.Where) > 0 {
				for key, element := range *query.Where {
					if len(fmt.Sprint(element)) <= 0 {
						return errors.New("MONGO: Where conditions has empty column or value")
					}
					*filter = append((*filter), bson.E{Key: key, Value: element})
				}
			}
		}
	}
	return nil
}

// Auto fill a sort pointer depending on a given query
func fillSort(sort *bson.D, query *Query) error {
	if query != nil {
		if query.Sort != nil {
			if len(*query.Sort) > 0 {
				for key, element := range *query.Sort {
					val := fmt.Sprint(element)
					if val != "-1" && val != "1" {
						return errors.New("MONGO: Sort value invalid, must be 1 for ascending or -1 for descending")
					}
					*sort = append((*sort), bson.E{Key: key, Value: element})
				}
			}
		}
	}
	return nil
}

// Inserts a generic slice of objects into the database, usage:
//
// result, err := mwrap.PutMany(mongoClient, *sliceToSave)
//
// Need to pass a pointer to a slice, else this will not compile or throw an error.
// Automatically parses the collection name in snake case from reflection name
func PutMany[T any](m *Mongo, value []T, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	defer errs.Defer()
	if err := m.checkNil(); err != nil {
		return nil, err
	}

	if len(value) <= 0 {
		return nil, errors.New("MONGO: tried inserting empty array")
	}

	name := reflect.TypeOf(value).Elem().Name()
	if len(name) == 0 {
		return nil, errors.New("MONGO: error parsing reflection name of generic")
	}

	newValue := make([]interface{}, len(value))
	for index, element := range value {
		newValue[index] = reflect.ValueOf(element).Interface()
	}

	return m.Database.Collection(toSnakeCase(name)).InsertMany(context.TODO(), newValue, opts...)
}

func PutManyCollection[T any](m *Mongo, collectionName string, value []T, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	defer errs.Defer()
	if err := m.checkNil(); err != nil {
		return nil, err
	}

	if len(value) <= 0 {
		return nil, errors.New("MONGO: tried inserting empty array")
	}

	if len(collectionName) == 0 {
		return nil, errors.New("MONGO: collection name can't be empty")
	}

	newValue := make([]interface{}, len(value))
	for index, element := range value {
		newValue[index] = reflect.ValueOf(element).Interface()
	}

	return m.Database.Collection(collectionName).InsertMany(context.TODO(), newValue, opts...)
}

// Inserts a single object into the database, usage:
//
// result, err := mwrap.PutOne(mongoClient, objectToSave)
//
// Contrary to PutMany[](), this does not take in a pointer of an object
// Automatically parses the collection name in snake case from the reflection name
func PutOne[T any](m *Mongo, value T, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	defer errs.Defer()
	if err := m.checkNil(); err != nil {
		return nil, err
	}
	name := reflect.TypeOf(value).Name()
	if len(name) == 0 {
		return nil, errors.New("MONGO: error parsing reflection name of generic")
	}

	return m.Database.Collection(toSnakeCase(name)).InsertOne(context.TODO(), reflect.ValueOf(value).Interface(), opts...)
}

func (m *Mongo) checkNil() error {
	if m.Client == nil {
		return errors.New("MONGO: Client was nil")
	} else if m.Database == nil {
		return errors.New("MONGO: Database was nil")
	}
	return nil
}

// String compiling for auto-collection
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
