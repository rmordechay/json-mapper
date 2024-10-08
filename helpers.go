package jsonmapper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"unicode"
)

func (m *JsonMapper) getType() JsonType {
	switch {
	case m.IsBool:
		return Bool
	case m.IsInt:
		return Int
	case m.IsFloat:
		return Float
	case m.IsString:
		return String
	case m.IsObject:
		return Object
	case m.IsNull:
		return Null
	case m.IsArray:
		return Array
	}
	return Invalid
}

func convertToSlicePtr(data []any) []*any {
	array := make([]*any, len(data))
	for i, v := range data {
		v := v
		array[i] = &v
	}
	return array
}

func convertToMapValuesPtr(data map[string]any) map[string]*any {
	jsonObject := make(map[string]*any, len(data))
	for k, v := range data {
		v := v
		jsonObject[k] = &v
	}
	return jsonObject
}

func isObjectOrArray(data []byte, brackOrParen byte) bool {
	if len(data) == 0 {
		return false
	}
	var firstChar byte
	for _, d := range data {
		firstChar = d
		if unicode.IsSpace(rune(firstChar)) {
			continue
		}
		return firstChar == brackOrParen
	}
	return false
}

func parseTime(t *any, j jsonEntity) time.Time {
	if t == nil {
		j.SetLastError(createNullConversionErr(stringTypeStr))
		return time.Time{}
	}
	timeAsString, ok := (*t).(string)
	if !ok {
		j.SetLastError(fmt.Errorf("cannot convert type %T to type time.Time\n", t))
		return time.Time{}
	}
	for _, layout := range timeLayouts {
		parsedTime, err := time.Parse(layout, timeAsString)
		if err == nil {
			return parsedTime
		}
	}
	j.SetLastError(fmt.Errorf("the value '%v' could not be converted to type time.Time", timeAsString))
	return time.Time{}
}

func marshal(v any) ([]byte, error) {
	jsonBytes, err := jsonIter.Marshal(v)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func marshalIndent(v any) ([]byte, error) {
	// jsoniter has a bug with indentation
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func unmarshal(data []byte, v any) error {
	return jsonIter.Unmarshal(data, &v)
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

type jc[T any] func(data *any, j jsonEntity) T

func getGenericMap[T any](f jc[T], o JsonObject) map[string]T {
	genericMap := make(map[string]T)
	for k, v := range o.object {
		genericMap[k] = f(v, &o)
	}
	return genericMap
}

func getGenericArray[T any](f jc[T], o JsonArray) []T {
	arr := make([]T, 0, len(o.elements))
	for _, element := range o.elements {
		arr = append(arr, f(element, &o))
	}
	return arr
}

func getObjectScalar[T any](o *JsonObject, f jc[T], key string, typeString string) T {
	var zero T
	v, ok := o.object[key]
	if !ok {
		o.SetLastError(createKeyNotFoundErr(key))
		return zero
	}
	if v == nil {
		o.SetLastError(createNullConversionErr(typeString))
		return zero
	}
	return f(v, o)
}

func getArrayScalar[T any](a *JsonArray, f jc[T], i int, typeString string) T {
	var zero T
	if i >= a.Length() {
		a.SetLastError(createIndexOutOfRangeErr(i, a.Length()))
		return zero
	}
	data := a.elements[i]
	if data == nil {
		a.SetLastError(createNullConversionErr(typeString))
		return zero
	}
	return f(data, a)
}

func convertAnyToString(data *any, j jsonEntity) string {
	if data == nil {
		j.SetLastError(createNullConversionErr(stringTypeStr))
		return ""
	}
	switch (*data).(type) {
	case string:
		return (*data).(string)
	case float64:
		return strconv.FormatFloat((*data).(float64), 'f', -1, 64)
	case int:
		return strconv.Itoa((*data).(int))
	case bool:
		return strconv.FormatBool((*data).(bool))
	default:
		j.SetLastError(createTypeConversionErr(*data, stringTypeStr))
		return ""
	}
}

func convertAnyToInt(data *any, j jsonEntity) int {
	if data == nil {
		j.SetLastError(createNullConversionErr(intTypeStr))
		return 0
	}
	switch (*data).(type) {
	case float64:
		return int((*data).(float64))
	case int:
		return (*data).(int)
	default:
		j.SetLastError(createTypeConversionErr(*data, intTypeStr))
		return 0
	}
}

func convertAnyToFloat(data *any, j jsonEntity) float64 {
	if data == nil {
		j.SetLastError(createNullConversionErr(floatTypeStr))
		return 0
	}
	v, ok := (*data).(float64)
	if !ok {
		j.SetLastError(createTypeConversionErr(*data, floatTypeStr))
		return 0
	}
	return v
}

func convertAnyToBool(data *any, j jsonEntity) bool {
	if data == nil {
		j.SetLastError(createNullConversionErr(boolTypeStr))
		return false
	}
	v, ok := (*data).(bool)
	if !ok {
		j.SetLastError(createTypeConversionErr(*data, boolTypeStr))
		return false
	}
	return v
}

func convertAnyToObject(data *any, j jsonEntity) JsonObject {
	if data == nil {
		j.SetLastError(createNullConversionErr(objectTypeStr))
		return *EmptyObject()
	}
	v, ok := (*data).(map[string]any)
	if !ok {
		j.SetLastError(createTypeConversionErr(data, objectTypeStr))
		return *EmptyObject()
	}

	obj := EmptyObject()
	var object = make(map[string]*any)
	for key, value := range v {
		object[key] = &value
	}
	obj.object = object
	return *obj
}

func convertAnyToArray(data *any, j jsonEntity) JsonArray {
	if data == nil {
		j.SetLastError(createNullConversionErr(arrayTypeStr))
		return *EmptyArray()
	}
	v, ok := (*data).([]any)
	if !ok {
		j.SetLastError(createTypeConversionErr(data, arrayTypeStr))
		return *EmptyArray()
	}

	var array JsonArray
	var elements = make([]*any, 0, len(v))
	for _, value := range v {
		value := value
		elements = append(elements, &value)
	}
	array.elements = elements
	return array
}

func convertSliceToJsonArray[T any](data []T) JsonArray {
	var jsonArray JsonArray
	sliceAnyPtr := make([]*any, 0, len(data))
	for _, v := range data {
		var valAny any = v
		sliceAnyPtr = append(sliceAnyPtr, &valAny)
	}
	jsonArray.elements = sliceAnyPtr
	return jsonArray
}
