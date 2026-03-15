package main

import (
	"fmt"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Blobby!")
}

func main() {}
