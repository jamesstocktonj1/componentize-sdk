package main

import (
	"io"
	"net/http"

	"github.com/jamesstocktonj1/componentize-sdk/net/socket"
	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
	conn, err := socket.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("Hello, World!\n"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		buf := make([]byte, 1024)

		n, err := conn.Read(buf)
		if err == io.EOF {
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(buf[:n])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if buf[n-1] == '\n' {
			return
		}
	}
}

func main() {}
