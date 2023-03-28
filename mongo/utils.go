package mongo

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

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
