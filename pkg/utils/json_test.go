package utils

import (
	"argo-volcano-executor-plugin/pkg/utils/jsonUtil"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"testing"
	batch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

//go:embed test.json
var jsonData []byte

type Item struct {
	Name   string `json:"name"`
	ItemId int    `json:"item_id"`
}

type JobBody struct {
	Job *batch.Job `json:"job"`
}
type VolcanoPluginBody struct {
	JobBody *JobBody `json:"volcano"`
}

type Items struct {
	Items Item `json:"items"`
}

func TestJsonUnmarshal(t *testing.T) {

	//We can now send our item as either an int or string without getting any error

	var item Item

	err := json.Unmarshal(jsonData, &item)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", item)
}

func TestJsonUnmarsha2l(t *testing.T) {

	//We can now send our item as either an int or string without getting any error

	jsonData = []byte(`{"name":"item 1","item_id":"30"}`)
	var item Item
	//var result map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(jsonData)))
	//seNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64.
	decoder.UseNumber()
	decoder.Decode(&item)
	fmt.Printf("%+v\n", item)
}
func TestSjunmar(t *testing.T) {

	//We can now send our item as either an int or string without getting any error

	var dest interface{}
	item := &VolcanoPluginBody{
		JobBody: &JobBody{
			Job: &batch.Job{},
		},
	}
	json.Unmarshal(jsonData, &dest)

	err := jsonUtil.UnmarshalFromMap(dest, &item)
	if err != nil {
		panic(err)
	}

	newJson, err := json.Marshal(item)
	fmt.Printf(string(newJson))
}
