package main

import (
	"fmt"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(hello)
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func main() {}
