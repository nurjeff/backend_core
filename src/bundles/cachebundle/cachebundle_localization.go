package cachebundle

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/aerospike/aerospike-client-go"
	"github.com/go-redis/redis"
	"github.com/sc-js/backend_core/src/tools"
	"github.com/sc-js/pour"
)

const translation_path = "./translations/"

// Gets a translation for a key, given a
func GetTS(locale string, key string) (string, error) {

	switch connectedModule {
	case AeroSpike:
		pour.LogPanicKill(1, errors.New("not_implemented"))
	case Redis:
		err := redisClient.Get("translation_" + locale + "_" + key).Err()
		if err == redis.Nil {
			WriteNewTSEntry(locale, key)
			return "", errNf
		}
		return redisClient.Get("translation_" + locale + "_" + key).Val(), nil
	}
	pour.LogPanicKill(1, errors.New("not_implemented"))
	return "", errors.New("not_implemented")
}

// Append a new translation entry (specified by key) into a specific locale data structure (Cache and file).
func WriteNewTSEntry(locale string, key string) {

	go func() {
		cacheFileLock.Lock()
		defer cacheFileLock.Unlock()
		outPath := translation_path
		if _, err := os.Stat(outPath); os.IsNotExist(err) {
			os.Mkdir(outPath, 0755)
		}
		localeFileName := locale + ".json"
		var fileData []byte
		dat, err := os.ReadFile(outPath + localeFileName)
		if err == nil {
			structure := map[string]string{}
			json.Unmarshal(dat, &structure)
			structure[key] = ""
			fileData, err = json.Marshal(&structure)
			if err != nil {
				pour.LogColor(false, pour.ColorRed, "error appending translation data to cache", err)
				return
			}
			f, _ := os.OpenFile(outPath+localeFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			defer f.Close()
			f.Write(fileData)
		} else {
			structure := map[string]string{}
			structure[key] = ""
			b, err := json.Marshal(&structure)
			if err != nil {
				pour.LogColor(false, pour.ColorRed, "error marshalling new transmap", err)
				return
			}
			os.WriteFile(outPath+localeFileName, b, 0644)
		}
		PutTS(key, locale, "")
	}()
}

// Translations
func PutTS(key string, locale string, val string) error {

	switch connectedModule {
	case AeroSpike:
		internalKey, err := aerospike.NewKey(workspace, "translation_"+locale+"_", key)
		if err != nil {
			return err
		}
		bin := aerospike.BinMap{"a": val}
		return aeroClient.Put(nil, internalKey, bin)
	case Redis:
		return redisClient.Set("translation_"+locale+"_"+key, val, 0).Err()
	}
	return nil
}

// Automatically calls readTSJson in a given time period, to refresh translation data, should it be changed during runtime
func ReadTSJson(path string, autoRefresh bool) {
	insertDefaultValues()
	if autoRefresh {
		go func() {
			for {
				time.Sleep(time.Second * 60)
				err := readTSJson(path)
				if err != nil {
					pour.LogColor(false, pour.ColorRed, "Error reading TS Json:", err)
					return
				}
			}
		}()
	}
}

// Reads all locale files in a specific path (normally, this should be the translations directory).
// Removes all Pre- and suffixes from the name to determine the corresponding locale.
// Parses the locales and caches them for faster access.
func readTSJson(path string) error {

	files, err := tools.WalkDir(path, []string{"json"})
	if err != nil {
		pour.LogColor(true, pour.ColorRed, err)
		return err
	}
	for _, element := range files {
		split := strings.Split(element, ".")
		ext := strings.Split(split[0], "/")
		if len(ext) == 1 {
			ext = strings.Split(split[0], "\\")
		}
		if validateLocale(ext[1]) {
			var m map[string]string
			element = strings.Replace(element, "\\", "/", -1)
			dat, err := os.ReadFile(element)
			if err != nil {
				pour.LogColor(false, pour.ColorRed, "error reading translation file:", element)
				continue
			}
			err = json.Unmarshal(dat, &m)
			if err != nil {
				pour.LogColor(false, pour.ColorRed, "error unmarshalling translation file:", element)
				continue
			}
			split = strings.Split(split[0], "-")
			split[0] = strings.Replace(split[0], "\\", "/", -1)
			err = PutTSMap(m, split[0], false)
			if err != nil {
				pour.LogColor(true, pour.ColorRed, err)
				continue
			}
		}
	}
	return nil
}

// Check whether a given locale is valid, e.G. en_EN, de_DE etc.
func validateLocale(loc string) bool {
	return valid_locales[loc]
}

// Insert the default locale values into the cache.
func insertDefaultValues() {
	for key, value := range default_locales {
		valid_locales[key] = true
		err := PutTSMap(value, key, true)
		if err != nil {
			pour.LogColor(true, pour.ColorRed, err)
		}
	}
}

// Auto-creates the translations folder and creates the default translation files which are hardcoded for the moment.
// Translations are cached for faster acccess.
func PutTSMap(m map[string]string, locale string, verbose bool) error {

	if _, err := os.Stat(translation_path); os.IsNotExist(err) {
		err := os.Mkdir(translation_path, 0777)
		if err != nil {
			pour.LogColor(false, pour.ColorRed, "Error creating dir:", err)
			return err
		}
	}
	if strings.Contains(locale, "/") {
		spl := strings.Split(locale, "/")
		locale = spl[len(spl)-1]
	}
	var j []byte
	var existing map[string]string
	if tools.Exists(translation_path + locale + ".json") {
		b, err := os.ReadFile(translation_path + locale + ".json")
		if err != nil {
			pour.LogColor(false, pour.ColorRed, "Can't read locale file for:", locale)
			return err
		}
		err = json.Unmarshal(b, &existing)
		if err != nil {
			pour.LogColor(false, pour.ColorRed, "Can't read locale file for:", locale)
			return err
		}
	} else {
		ff, err := os.Create(translation_path + locale + ".json")
		if err != nil {
			return err
		}
		err = ff.Close()
		if err != nil {
			return err
		}
	}

	if len(existing) > 0 {
		for key, val := range existing {
			m[key] = val
		}
	}

	j, err := json.Marshal(&m)
	if err != nil {
		return err
	}

	if verbose {
		pour.LogColor(false, pour.ColorYellow, "Writing/Caching", len(m), "translations for", locale)
	}

	f, err := os.OpenFile(translation_path+locale+".json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	_, err = f.Write(j)
	if err != nil {
		return err
	}

	for key, element := range m {
		switch connectedModule {
		case AeroSpike:
			internalKey, err := aerospike.NewKey(workspace, "translation_"+locale+"_", key)
			if err != nil {
				pour.LogColor(true, pour.ColorRed, err)
			}
			bin := aerospike.BinMap{"a": element}
			err = aeroClient.Put(nil, internalKey, bin)
			if err != nil {
				pour.LogColor(true, pour.ColorRed, err)
			}
		case Redis:
			err := redisClient.Set("translation_"+locale+"_"+key, element, 0).Err()
			if err != nil {
				pour.LogColor(true, pour.ColorRed, err)
			}
		}
	}

	return nil
}
