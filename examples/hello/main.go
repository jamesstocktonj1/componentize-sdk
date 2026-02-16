package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/echo", echo)
	mux.HandleFunc("/greet", greeting)

	wasihttp.Handle(mux)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func echo(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func greeting(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Name string `json:"name"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Hello, %s!", request.Name)
}

func main() {}
