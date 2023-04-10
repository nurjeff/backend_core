package cachebundle

import (
	"reflect"
	"sync"
)

var cacheFileLock = sync.RWMutex{}

func TranslateStruct(transStruct reflect.Value, languageCode string) interface{} {
	return translate(transStruct, languageCode)
}

func translate(obj reflect.Value, languageCode string) interface{} {
	// Wrap the original in a reflect.Value
	copy := reflect.New(obj.Type()).Elem()
	translateRecursive(copy, obj, languageCode, true)
	// Remove the reflection wrapper
	return copy
}

func translateRecursive(copy, original reflect.Value, languageCode string, doTranslate bool) {
	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		if copy.CanSet() {
			copy.Set(reflect.New(originalValue.Type()))
			// Unwrap the newly created pointer
			translateRecursive(copy.Elem(), originalValue, languageCode, doTranslate)
		}

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		if originalValue.IsValid() && !originalValue.IsZero() {
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue, languageCode, doTranslate)
			copy.Set(copyValue)
		}

	// If it is a struct we translate each field
	case reflect.Struct:
		if original.Type().Name() == "Time" {
			copy.Set(original)
			return
		}
		val := reflect.Indirect(original)
		for i := 0; i < original.NumField(); i += 1 {
			_, ok := val.Type().Field(i).Tag.Lookup("trans")
			translateRecursive(copy.Field(i), original.Field(i), languageCode, ok)
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			translateRecursive(copy.Index(i), original.Index(i), languageCode, doTranslate)
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			translateRecursive(copyValue, originalValue, languageCode, doTranslate)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string translate it (yay finally we're doing what we came for)
	case reflect.String:
		if doTranslate {
			value := original.Interface().(string)
			if value != "" {
				trans, err := GetTS(languageCode, value)
				if err != nil {
					copy.SetString(value)
					WriteNewTSEntry(languageCode, value)
				} else {
					if len(trans) > 0 {
						copy.SetString(trans)
						return
					}
					copy.SetString(value)
				}
			}

		} else {
			copy.Set(original)
		}

	// And everything else will simply be taken from the original
	default:
		if copy.CanSet() {
			copy.Set(original)
		}
	}
}
