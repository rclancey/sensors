package api

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ReadJSONFile(fn string, obj interface{}) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}
