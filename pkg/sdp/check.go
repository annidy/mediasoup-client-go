package sdp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func getAllFieldNames(t reflect.Type, prefix string) map[string]bool {
	fieldMap := make(map[string]bool)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tagValue := field.Tag.Get("json")
		fieldName := field.Name
		if tagValue != "" && tagValue != "-" {
			parts := strings.Split(tagValue, ",")
			fieldName = parts[0]
		}
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct {
			// 递归处理嵌套结构体
			nestedFields := getAllFieldNames(fieldType, fieldName)
			for k, v := range nestedFields {
				fieldMap[k] = v
			}
		} else {
			fieldMap[fieldName] = true
		}
	}
	return fieldMap
}

func checkExtraKeys(data map[string]interface{}, fieldMap map[string]bool) {
	for key := range data {
		found := false
		for structKey := range fieldMap {
			if strings.HasPrefix(key, structKey) || strings.HasPrefix(structKey, key) {
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Sprintf("Extra key found: %s", key))
		}
	}
}

func checkMissingKeys[T any](jsonStr string) {
	var data map[string]interface{}

	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		fmt.Println("Error unmarshalling:", err)
		return
	}
	elem := reflect.TypeOf(new(T)).Elem()
	fieldMap := getAllFieldNames(elem, "")

	checkExtraKeys(data, fieldMap)
}
