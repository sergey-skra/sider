package main

// This file contains the implementation of Sider's database.

import (
	"encoding/gob"
	"errors"
	"os"
	"regexp"
	"sync"
	"time"
)

// Entry is a pair of a value and an expiration time.
type Entry struct {
	Value   []byte
	Expires int64
}

// DB represents a database.
type DB struct {
	dict map[string]Entry
	mu   sync.RWMutex
}

// Open creates and opens a new database.
func Open() *DB {
	d := make(map[string]Entry)
	return &DB{dict: d}
}

// OpenDBFrom restores a database from the file.
func OpenDBFrom(file string) (*DB, error) {
	f, err := os.Open(file)
	if err != nil {
		return &DB{}, err
	}
	defer f.Close()

	d := gob.NewDecoder(f)
	var entries map[string]Entry
	err = d.Decode(&entries)
	if err != nil {
		return &DB{}, err
	}
	return &DB{dict: entries}, nil
}

// SaveTo file on disk.
// TODO. Write an interface instead of file. io.WriteCloser?
func (db *DB) SaveTo(file string) {
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	ne := gob.NewEncoder(f)
	err = ne.Encode(db.dict)
	if err != nil {
		panic(err)
	}
}

// Set the key to hold the entry.
func (db *DB) Set(k string, e Entry) {
	db.mu.Lock()
	db.dict[k] = e
	db.mu.Unlock()
}

// Get the value of the key. If given key doesn't exist it returns ({nil, 0}, false).
func (db *DB) Get(k string) (Entry, bool) {
	db.mu.Lock()
	e, f := db.dict[k]
	db.mu.Unlock()
	return e, f
}

// Update changes the value related to the key. Ttl remains the same.
func (db *DB) Update(k string, val []byte) error {
	if !db.hasKey(k) {
		return errors.New("key doesn't exist")
	}
	db.mu.Lock()
	ttl := db.dict[k].Expires
	db.dict[k] = Entry{val, ttl}
	db.mu.Unlock()
	return nil
}

// Remove the key from the database.
func (db *DB) Remove(k string) {
	db.mu.Lock()
	delete(db.dict, k)
	db.mu.Unlock()
}

// Keys returns all keys mathing the pattern.
func (db *DB) Keys(pattern string) []string {
	// need to check if expensive isExpensive()
	var a []string
	db.mu.Lock()
	for key := range db.dict {
		if res, _ := regexp.MatchString(pattern, key); res {
			a = append(a, key)
		}
	}
	db.mu.Unlock()
	return a
}

// TTL returns the remaining time to live of a key that has a timeout. Or:
// -1 if the key exists but has no associated expire(default) and
// -2 if the key does not exist.
func (db *DB) TTL(k string) int64 {
	if db.hasKey(k) {
		return db.dict[k].Expires
	}
	return -2
}

// DeleteExpired removes all expired keys from the database.
func (db *DB) DeleteExpired() {
	now := time.Now().UnixNano()
	db.mu.Lock()
	for k, e := range db.dict {
		if e.Expires > 0 && now > e.Expires {
			delete(db.dict, k)
		}
	}
	db.mu.Unlock()
}

// Flush flushes database.
func (db *DB) Flush() {
	db.mu.Lock()
	db.dict = map[string]Entry{}
	db.mu.Unlock()
}

func (db *DB) hasKey(k string) bool {
	_, ok := db.dict[k]
	return ok
}
