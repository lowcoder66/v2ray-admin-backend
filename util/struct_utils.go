package util

import "reflect"

func CopyFields(src, dst interface{}, ignoreFields ...string) {
	sVal := reflect.ValueOf(src).Elem()
	dVal := reflect.ValueOf(dst).Elem()

	for i := 0; i < dVal.NumField(); i++ {
		fieldName := dVal.Type().Field(i).Name
		if contains(ignoreFields, fieldName) {
			continue
		}

		srcVal := sVal.FieldByName(fieldName)
		if srcVal.IsValid() {
			dstVal := dVal.Field(i)
			dstVal.Set(srcVal)
		}
	}
}

func contains(slice []string, ele string) bool {
	c := false
	for i := 0; i < len(slice); i++ {
		if slice[i] == ele {
			c = true
			break
		}
	}
	return c
}
