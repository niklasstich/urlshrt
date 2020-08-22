package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Entry struct {
	Shorthand string
	Url       string
}

var indexhtml []byte
var mongoclient *mongo.Client

func homePage(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(indexhtml)
	if err != nil {
		log.Fatal(err)
	}
}

func addEntry(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Only JSON data is accepted for POST", http.StatusUnsupportedMediaType)
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	var entry Entry
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&entry)
	if err != nil {
		//differentiate between errors, because not all errors are user errors (400 or 500)
		//borrowed from https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			http.Error(w, "Bad JSON format", http.StatusBadRequest)

		case errors.As(err, unmarshalTypeErr):
			http.Error(w, "JSON type mismatch (should be string/string)", http.StatusBadRequest)

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			http.Error(w, "JSON body has unexpected field", http.StatusBadRequest)

		case errors.Is(err, io.EOF):
			http.Error(w, "JSON body can't be empty", http.StatusBadRequest)

		case err.Error() == "http: request body too large":
			http.Error(w, "JSON body too large, 1MB at max", http.StatusBadRequest)

		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}
	entry.Shorthand = strings.TrimSpace(entry.Shorthand)
	entry.Url = strings.TrimSpace(entry.Url)
	match, err := regexp.MatchString("[\\w\\-]+", entry.Shorthand)
	if entry.Shorthand == "" || entry.Url == "" {
		http.Error(w, "Shorthand and Url can't be empty.", http.StatusBadRequest)
		return
	} else if !match {
		http.Error(w, "Shorthand didn't match RegEx [\\w\\-]+", http.StatusBadRequest)
		return
	}
	//check if Shorthand is already in database, if not insert and give back 200
	collection := mongoclient.Database("sh").Collection("redirects")

	filter := bson.D{{Key: "shorthand", Value: entry.Shorthand}}
	result := collection.FindOne(r.Context(), filter)
	if result.Err() != mongo.ErrNoDocuments {
		http.Error(w, "There already exists an entry with the specified Shorthand.", http.StatusBadRequest)
		return
	}
	_, err = collection.InsertOne(r.Context(), entry)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(http.StatusCreated)
	_, _ = fmt.Fprintf(w, "<h1>The association for %s to %s has been created!</h1>\n<p>Click <a href=\"%s/%s\">here</a> to try out the redirect!", entry.Shorthand, entry.Url, os.Getenv("HOSTNAME"), entry.Shorthand)
}

func redirectByKey(w http.ResponseWriter, r *http.Request) {
	//Shorthand := mux.Vars(r)["Shorthand"]
	//see if Shorthand is in database, 301
}

func initializeDBClient() {
	cs := fmt.Sprintf("mongodb://%s:%s@%s:%s", os.Getenv("MONGO_USER"),
		os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_URI"), os.Getenv("MONGO_PORT"))
	clientOptions := options.Client().ApplyURI(cs)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	mongoclient = client

	collection := mongoclient.Database("sh").Collection("redirects")

	index := mongo.IndexModel{
		Keys:    bson.D{{"shorthand", 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(context.TODO(), index)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var err error
	indexhtml, err = ioutil.ReadFile("index.html")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connecting to mongoDB...")
	initializeDBClient()
	defer mongoclient.Disconnect(context.TODO())
	log.Println("Connected!")

	router := mux.NewRouter()
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/", addEntry).Methods("POST")
	router.HandleFunc("/{Shorthand:[\\s\\-]+}", redirectByKey).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")), nil))
}
