package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

var client = http.Client{Transport: &wasihttp.Transport{}}

func init() {
	wasihttp.HandleFunc(handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/hi", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Greeting string `json:"greeting"`
		Name     string `json:"name"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	fmt.Fprintf(w, "Success %d - %s: %s", resp.StatusCode, response.Name, response.Greeting)
}

func main() {}
