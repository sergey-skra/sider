package main

import (
	"os"
	"testing"
	"time"
)

var entry = Entry{
	Value:   []byte{1, 2, 3},
	Expires: -1,
}

func TestOpen(t *testing.T) {
	db := Open()
	if db == nil {
		t.Fatal("db was not initialized")
	}
}

func TestPersistence(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	db.SaveTo("file")
	defer os.Remove("file")
	_, err := OpenDBFrom("file")
	if err != nil {
		t.Fatal("can't open db from file")
	}
}

func TestSetGet(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	_, ok := db.Get("key")
	if !ok {
		t.Fatal("can't get an entry")
	}
}

func TestUpdate(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	err := db.Update("key", []byte{1, 3})
	if err != nil {
		t.Fatal("can't update an entry")
	}
}

func TestRemove(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	db.Remove("key")
	_, ok := db.Get("key")
	if ok {
		t.Fatal("can't remove an entry")
	}
}

func TestKeys(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	res := db.Keys("k*")
	if res[0] != "key" {
		t.Fatal("can't find keys")
	}
}

func TestTTL(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	ttl := db.TTL("key")
	if ttl != -1 {
		t.Fatal("ttl is not correct")
	}
}

func TestDeleteExpired(t *testing.T) {
	db := Open()
	entry.Expires = 1
	db.Set("key", entry)
	time.Sleep(2 * time.Second)
	db.DeleteExpired()
	_, ok := db.Get("key")
	if ok {
		t.Fatal("the key wasn't removed")
	}
}

func TestFlush(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	db.Flush()
	_, ok := db.Get("key")
	if ok {
		t.Fatal("can't flush db")
	}
}

func TestHasKey(t *testing.T) {
	db := Open()
	db.Set("key", entry)
	ok := db.hasKey("key")
	if !ok {
		t.Fatal("can't find a key")
	}
}
