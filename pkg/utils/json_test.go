package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"testing"
)

type Item struct {
	Name   string    `json:"name"`
	ItemId int `json:"item_id"`
}

type Items struct {
	Items  Item `json:"items"`
}


func TestJsonUnmarshal(t *testing.T) {

	//We can now send our item as either an int or string without getting any error

	jsonData := []byte(`{"name":"item 1","item_id":"30"}`)
	var item Item

	err := json.Unmarshal(jsonData, &item)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", item)
}

func TestJsonUnmarsha2l(t *testing.T) {

	//We can now send our item as either an int or string without getting any error

	jsonData := []byte(`{"name":"item 1","item_id":"30"}`)
	var item Item
	//var result map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(jsonData)))
	//seNumber causes the Decoder to unmarshal a number into an interface{} as a Number instead of as a float64.
	decoder.UseNumber()
	decoder.Decode(&item)
	fmt.Printf("%+v\n", item)
}