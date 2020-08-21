package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Entry struct {
	Key string
	Url string
}

var indexhtml []byte

func homePage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(indexhtml)
	if err != nil {
		log.Fatal(err)
	}
}

func addEntry(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		var entry Entry
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		entry.Key = strings.TrimSpace(entry.Key)
		entry.Url = strings.TrimSpace(entry.Url)
		match, err := regexp.MatchString("[\\w\\-]+", entry.Key)
		if entry.Key == "" || entry.Url == "" {
			http.Error(w, "Key and Url can't be empty", http.StatusBadRequest)
		} else if !match {
			http.Error(w, "Key didn't match RegEx [\\w\\-]+", http.StatusBadRequest)
		} else {
			//check if key is already in database, if not insert and give back 200
			cs := fmt.Sprintf("mongodb://%s:%s@%s:%s", os.Getenv("MONGO_USER"),
				os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_URI"), os.Getenv("MONGO_PORT"))
			client, err := mongo.NewClient(options.Client().ApplyURI(cs))
			if err != nil {
				http.Error(w, "Internal Database Error", http.StatusInternalServerError)
			}
			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()
			err = client.Connect(ctx)
			if err != nil {
				http.Error(w, "Internal Database Error", http.StatusInternalServerError)
			}
			collection := client.Database("urlshrt")
		}
	} else {
		http.Error(w, "Only JSON data is accepted for POST", http.StatusUnsupportedMediaType)
	}

}

func redirectByKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	//see if key is in database, 301
}

func main() {
	var err error
	indexhtml, err = ioutil.ReadFile("index.html")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/", addEntry).Methods("POST")
	router.HandleFunc("/{key:[\\s\\-]+}", redirectByKey).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")), nil))
}
