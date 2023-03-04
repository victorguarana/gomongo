package mongo

import "reflect"

// TODO: Remove case sensitive for fields
// Because public and private fields have diferent names inside struct
func convertMapToStruct(m map[string]interface{}, s interface{}) {
	stValue := reflect.ValueOf(s).Elem()
	sType := stValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		if value, ok := m[field.Name]; ok {
			if stValue.Field(i).Type().String() == "int" {
				stValue.Field(i).Set(reflect.ValueOf(int(value.(int32))))
			} else {
				stValue.Field(i).Set(reflect.ValueOf(value))
			}
		}
	}
}
