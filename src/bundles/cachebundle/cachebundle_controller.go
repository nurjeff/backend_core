package cachebundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/aerospike/aerospike-client-go"
	"github.com/go-redis/redis"
	"github.com/sc-js/core_backend/src/errs"
	"github.com/sc-js/core_backend/src/tools"
	"github.com/sc-js/pour"
	"golang.org/x/net/context"
)

const (
	AerospikeDefaultWorkspace = "aero_default"
	AeroSpike                 = 0
	Redis                     = 1
)

var (
	connectedModule = -1
	redisClient     *redis.Client
	aeroClient      *aerospike.Client
	errNf           = errors.New("key not found")
	workspace       = ""
)

// This function initializes the cache bundle, allowing the user to choose between Aerospike or Redis.
// It also initializes the necessary callbacks for the application.
// Reads Translations for locales form a given path and stores it into the cache.
func InitCache(val int, address string, portOverride uint, username string, password string, space string) {
	workspace = space

	ctx, ctxFunc := context.WithTimeout(context.TODO(), time.Second*5)

	defer ctxFunc()
	for {
		select {
		case <-ctx.Done():
			if redisClient == nil && aeroClient == nil {
				pour.LogErr(errors.New("connection to cache timed out"))
			}
			return
		default:
			if val == AeroSpike {
				initAero(address, portOverride, username, password, ctxFunc)
			}
			if val == Redis {
				initRedis(address, portOverride, username, password, ctxFunc)
			}
			connectedModule = val
			ReadTSJson(translation_path, true)
			tools.TranslationCallback = TranslateStruct
			tools.ValidatorCallback = validateLocale
			tools.SingleTranslationCallback = GetTS
			return
		}
	}
}

// Initializes AeroSpike as Cache Engine
func initAero(address string, portOverride uint, username string, password string, ctxFunc context.CancelFunc) {
	var err error
	port := 3000
	if portOverride > 0 {
		port = int(portOverride)
	}
	aeroClient, err = aerospike.NewClient(address, port)
	if err != nil {
		aeroClient = nil
		pour.LogPanicKill(1, err)
	}
	pour.LogColor(false, pour.ColorPurple, "AeroSpike connected at", address+":"+fmt.Sprint(port))
	ctxFunc()
}

// Initializes Redis as Cache Engine
func initRedis(address string, portOverride uint, username string, password string, ctxFunc context.CancelFunc) {

	port := 6379
	if portOverride > 0 {
		port = int(portOverride)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     address + ":" + fmt.Sprint(port), //redis port
		Password: password,
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		redisClient = nil
		pour.LogPanicKill(1, err)
	} else {
		pour.LogColor(false, pour.ColorPurple, "Redis connected at", address+":"+fmt.Sprint(port))
		ctxFunc()
	}
}

// This method works like th Put Method, but it also takes in an expiration time,
// after which the record will be automatically removed from the cache
func PutExpire[T any](name string, key string, val T, expiration time.Duration) error {
	defer errs.Defer()
	switch connectedModule {
	case AeroSpike:
		internalKey, err := aerospike.NewKey(workspace, name, key)
		if err != nil {
			return err
		}
		bin := aerospike.BinMap{"a": val}
		policy := aerospike.NewWritePolicy(0, uint32(expiration/time.Second))
		return aeroClient.Put(policy, internalKey, bin)
	case Redis:
		return redisClient.Set(name+key, val, expiration).Err()
	}

	return errors.New("no module connected")
}

// This method takes in a string, a key, and a value of any type. It then checks which connected module is in use.
// If the module is AeroSpike, it will create an internal key from the workspace, name, and key values and store the value in a bin.
// If the module is Redis, it will set the name, key, and value and return the result of the Set method.
// If none of the modules are connected, it will return an error.
func Put[T any](name string, key string, val T) error {
	defer errs.Defer()
	switch connectedModule {
	case AeroSpike:
		internalKey, err := aerospike.NewKey(workspace, name, key)
		if err != nil {
			return err
		}
		bin := aerospike.BinMap{"a": val}
		return aeroClient.Put(nil, internalKey, bin)
	case Redis:
		return redisClient.Set(name+key, val, 0).Err()
	}

	return errors.New("no module connected")
}

// This method retrieves a value of a specified type (T) from either an AeroSpike or Redis database, depending on the connected module.
// It uses the workspace name and key provided to search for the value and returns the value or an error if the value is not found.
// It also handles different types of values and can convert them to the desired type.
func Get[T any](name string, key string) (T, error) {
	defer errs.Defer()
	var result T
	switch connectedModule {
	case AeroSpike:
		internalKey, err := aerospike.NewKey(workspace, name, key)
		if err != nil {
			return result, err
		}
		rec, err := aeroClient.Get(nil, internalKey)
		if err != nil {
			return result, err
		}
		for _, v := range rec.Bins {
			result = v.(T)
			return result, nil
		}
	case Redis:
		v := reflect.ValueOf(new(T))
		switch reflect.TypeOf(result).Name() {
		case "float64":
			data, err := redisClient.Get(name + key).Float64()
			if err != nil {
				return result, errNf
			}
			v.Elem().Set(reflect.ValueOf(data))
			return v.Elem().Interface().(T), nil
		case "int":
			data, err := redisClient.Get(name + key).Int()
			if err != nil {
				return result, errNf
			}
			v.Elem().Set(reflect.ValueOf(data))
			return v.Elem().Interface().(T), nil
		case "int64":
			data, err := redisClient.Get(name + key).Int64()
			if err != nil {
				return result, errNf
			}
			v.Elem().Set(reflect.ValueOf(data))
			return v.Elem().Interface().(T), nil
		case "string":
			err := redisClient.Get(name + key).Err()
			if err == nil {
				return result, errNf
			}
			data := redisClient.Get(name + key).String()
			v.Elem().Set(reflect.ValueOf(data))
			return v.Elem().Interface().(T), nil
		case "uint64":
			data, err := redisClient.Get(name + key).Uint64()
			if err != nil {
				return result, errNf
			}
			v.Elem().Set(reflect.ValueOf(data))
			return v.Elem().Interface().(T), nil
		case "float32":
			data, err := redisClient.Get(name + key).Float32()
			if err != nil {
				return result, errNf
			}
			v.Elem().Set(reflect.ValueOf(data))
			return v.Elem().Interface().(T), nil
		case "bool":
			data, err := redisClient.Get(name + key).Int()
			if err != nil {
				return result, errNf
			}
			v.Elem().Set(reflect.ValueOf(data != 0))
			return v.Elem().Interface().(T), nil

		default:
			data, err := redisClient.Get(name + key).Bytes()
			if err != nil {
				return result, errNf
			}
			refValue := reflect.ValueOf(result)
			if refValue.Kind() == reflect.Slice {
				if refValue.Type().Elem().Kind() == reflect.Uint8 {
					v.Elem().Set(reflect.ValueOf(data))
					return v.Elem().Interface().(T), nil
				}
			}
			var v T
			err = json.Unmarshal(data, &v)
			return v, err
		}
	}
	pour.LogPanicKill(1, errors.New("no module connected"))
	return result, errors.New("no module connected")
}

// This method deletes a given workspace/key from the cache, depending on the connected module.
func Del(name string, key string) error {
	defer errs.Defer()
	switch connectedModule {
	case AeroSpike:
		{
			internalKey, err := aerospike.NewKey(workspace, name, key)
			if err != nil {
				return err
			}
			_, err = aeroClient.Delete(nil, internalKey)
			return err
		}
	case Redis:
		{
			_, err := redisClient.Del(name + key).Result()
			return err
		}
	}
	return errors.New("no module connected")
}
