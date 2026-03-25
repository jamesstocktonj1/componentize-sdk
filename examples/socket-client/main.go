package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/jamesstocktonj1/componentize-sdk/net/socket"
	"github.com/jamesstocktonj1/componentize-sdk/net/wasihttp"
)

func init() {
	wasihttp.HandleFunc(handle)
}

func handle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			data, err := handleConnection([]byte("Hello, World!\n"))
			if err != nil {
				fmt.Fprintf(w, "error - %+v\n", err)
			} else {
				w.Write(data)
			}
		}()
	}
	wg.Wait()

	endTime := time.Now()

	fmt.Fprintf(w, "time taken - %+v", endTime.Sub(startTime).Milliseconds())
}

func handleConnection(payload []byte) ([]byte, error) {
	conn, err := socket.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write(payload)
	if err != nil {
		return nil, err
	}

	data := []byte{}
	for {
		buf := make([]byte, 1024)

		n, err := conn.Read(buf)
		if err == io.EOF {
			data = append(data, buf[:n]...)
			return data, nil
		} else if err != nil {
			return nil, err
		}

		data = append(data, buf[:n]...)
		if buf[n-1] == '\n' {
			return data, nil
		}
	}
}

func main() {}
