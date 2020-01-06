package util

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

// Itob converts interface{} to []byte for rpc transfer
func Itob(i interface{}) []byte {
	var b bytes.Buffer
	gob.Register(&i)
	e := gob.NewEncoder(&b)
	err := e.Encode(&i)
	if err != nil {
		fmt.Println("Failed gob Encode in Itob() function", err)
	}
	return b.Bytes()
}

// Btoi converts []byte back to interface{}
func Btoi(data []byte) interface{} {
	var i interface{}
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	err := dec.Decode(&i)
	if err != nil {
		log.Fatal("Fail decode in Btoi():", err)
	}
	return i
}
