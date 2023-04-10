package tools

import (
	"crypto/md5"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

func GetMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func WalkDir(root string, exts []string) ([]string, error) {

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		for _, s := range exts {
			if strings.HasSuffix(path, "."+s) {
				files = append(files, path)
				return nil
			}
		}

		return nil
	})
	return files, err
}

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func FilterStringSlice(slice []string, allowed []string) []string {
	sliceCopy := slice
	for _, element := range slice {
		if !Contains(allowed, element) {
			sliceCopy = RemoveString(sliceCopy, element)
		}
	}
	return sliceCopy
}

func RemoveString(slice []string, target string) []string {
	for i, v := range slice {
		if v == target {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return slice
}

func Contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var dataPath string = "."

func SetDocker(inDocker bool) {
	if inDocker {
		dataPath = "./data"
		if !Exists(dataPath) {
			os.Mkdir(dataPath, 0755)
		}
	}
}
func GetPath() string {
	return dataPath
}

func GetPackageName(temp interface{}) string {
	strs := strings.Split((runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name()), ".")
	strs = strings.Split(strs[len(strs)-2], "/")
	return strs[len(strs)-1]
}
