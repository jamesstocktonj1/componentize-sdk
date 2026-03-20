package main

import (
	"io"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/file/blobstore"
	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	cont, err := blobstore.Create("hello")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cont.Close()

	obj, err := cont.Open("buffer")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	_, err = io.Copy(obj, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(w, obj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {}
