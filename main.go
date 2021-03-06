package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const domain = "galax.be"

func main() {
	initDatabase()
	defer db.Close()
	http.HandleFunc("/", handler)
	println("Listening 8765")
	http.ListenAndServe("127.0.0.1:8765", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Host", r.Host, "Path", r.URL.Path)
	if r.Host == domain && r.URL.Path == "/new" {
		decoder := json.NewDecoder(r.Body)
		var link tLink
		err := decoder.Decode(&link)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		link.save()
		fmt.Fprintln(w, link.Hash)
		return
	} else if r.Host != domain {
		read, write := io.Pipe()
		hash := strings.TrimSuffix(r.Host, "."+domain)
		log.Println("Searching hash", hash)
		w.Header().Add("Access-Control-Allow-Origin", r.Host)
		go func() {
			defer write.Close()
			if r.URL.Path == "/" {
				log.Println("Path is /")
				link := tLink{
					Hash: hash,
				}
				if len(hash) == 0 {
					log.Println("hash len is 0, return", hash)
					return
				}
				link.load()
				resp, err := http.Get(link.URL)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				io.Copy(write, resp.Body)
			} else {
				log.Println("Path is not /")
				link := tLink{
					Hash: hash,
				}
				if len(hash) == 0 {
					log.Println("hash len is 0, return", hash)
					return
				}
				link.load()
				url, err := url.Parse(link.URL)
				if len(hash) == 0 {
					log.Println("error parsing url", link.URL)
					return
				}
				log.Println("Proxying URL", url.Scheme+"://"+url.Host+r.URL.Path)
				resp, err := http.Get(url.Scheme + "://" + url.Host + r.URL.Path)
				if err != nil {
					return
				}
				defer resp.Body.Close()
				io.Copy(write, resp.Body)
			}
		}()
		io.Copy(w, read)
		return
	} else {
		http.ServeFile(w, r, path.Join("public", r.URL.Path))
		return
	}
}
