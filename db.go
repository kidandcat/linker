package main

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"log"
	"strings"

	"github.com/dgraph-io/badger/v3"
)

var db *badger.DB

func initDatabase() {
	var err error
	db, err = badger.Open(badger.DefaultOptions("linker.db"))
	if err != nil {
		log.Fatal(err)
	}
}

type tLink struct {
	Hash    string `json:"hash,omitempty"`
	URL     string `json:"url,omitempty"`
	Favicon string `json:"favicon,omitempty"`
}

func (l *tLink) load() {
	err := db.View(func(txn *badger.Txn) error {
		hashbyte, err := base64.StdEncoding.DecodeString(l.Hash)
		if err != nil {
			log.Println("Error decoding b64 hash", l.Hash, err)
		}
		item, err := txn.Get(hashbyte)
		if err != nil {
			log.Println("Error Get favicon", err)
		}
		err = item.Value(func(val []byte) error {
			ar := strings.Split(string(val), "-")
			url, err := base64.StdEncoding.DecodeString(ar[0])
			if err != nil {
				log.Println("Error decoding URL of", l.Hash)
				return errors.New("Error decoding URL of " + l.Hash)
			}
			favicon, err := base64.StdEncoding.DecodeString(ar[1])
			if err != nil {
				log.Println("Error decoding favicon of", l.Hash)
				return errors.New("Error decoding favicon of " + l.Hash)
			}
			l.URL = string(url)
			l.Favicon = string(favicon)
			return nil
		})
		return nil
	})
	if err != nil {
		log.Println("Error reading favicon", err)
	}
}

func (l *tLink) save() {
	err := db.Update(func(txn *badger.Txn) error {
		log.Println("load: db Update")
		urlEnc := base64.StdEncoding.EncodeToString([]byte(l.URL))
		faviconEnc := base64.StdEncoding.EncodeToString([]byte(l.Favicon))
		data := []byte(urlEnc + "-" + faviconEnc)
		hashbytes := md5.Sum(data)
		log.Println("hash 16", hashbytes)
		hash := make([]byte, len(hashbytes))
		copy(hash, hashbytes[:])
		err := txn.Set(hash, data)
		l.Hash = base64.StdEncoding.EncodeToString(hash)
		log.Println("hash", l.Hash)
		return err
	})
	if err != nil {
		log.Println("Error saving favicon", err)
	}
}
