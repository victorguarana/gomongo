package mongo

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

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

func InterfaceToStruct(in interface{}, out interface{}) error {
	data, _ := dataToBSON(in)
	data["id"] = data["_id"]

	marshal, err := bson.Marshal(data)
	if err != nil {
		return err
	}

	err = bson.Unmarshal(marshal, out)
	if err != nil {
		return err
	}

	return nil
}

func dataToBSON(data interface{}) (bson.M, error) {
	dataMarshal, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("convert data: %w", err)
	}

	var dataBSON bson.M
	if err := bson.Unmarshal(dataMarshal, &dataBSON); err != nil {
		return nil, fmt.Errorf("convert data: %w", err)
	}

	return dataBSON, nil
}
