package main

import (
	"encoding/base64"
	"errors"
	"log"
	"math/rand"
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
		log.Printf("Get: |%v|", l.Hash)
		item, err := txn.Get([]byte(l.Hash))
		if err != nil {
			log.Println("Error Get favicon", err, l.Hash)
			return err
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
		hash := []byte(randomString(15))
		log.Printf("Set: |%v|", string(hash))
		err := txn.Set(hash, data)
		if err != nil {
			log.Println("Error saving", string(hash), err)
		}
		l.Hash = string(hash)
		log.Println("new Save", string(hash))
		return err
	})
	if err != nil {
		log.Println("Error saving favicon", err)
	}
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
