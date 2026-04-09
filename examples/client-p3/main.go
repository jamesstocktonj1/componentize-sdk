package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jamesstocktonj1/componentize-sdk/p3/net/wasihttp"
)

var client = http.Client{Transport: &wasihttp.Transport{}}

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handler)
	mux.HandleFunc("/echo", echoHandler)
	wasihttp.Handle(mux)
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
	defer resp.Body.Close()

	response := struct {
		Greeting string `json:"greeting"`
		Name     string `json:"name"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Success %d - %s: %s", resp.StatusCode, response.Name, response.Greeting)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	const payload = "Hello, echo stream!"

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "http://localhost:8001/echo", strings.NewReader(payload))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if string(body) != payload {
		http.Error(w, fmt.Sprintf("echo mismatch: sent %q, got %q", payload, string(body)), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Echo OK: %s", body)
}

func main() {}
