package main

import (
	"io"
	"net"

	export_wasi_cli_run "github.com/jamesstocktonj1/componentize-sdk/examples/socket-server/gen/export_wasi_cli_0_2_6_run"
	_ "github.com/jamesstocktonj1/componentize-sdk/examples/socket-server/gen/wit_exports"
	"github.com/jamesstocktonj1/componentize-sdk/net/socket"
)

func init() {
	export_wasi_cli_run.SetRunner(run)
}

func run() {
	listener, err := socket.Listen("tcp", "0.0.0.0:7777")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		err = handleConnection(conn)
		if err != nil {
			panic(err)
		}
	}
}

func handleConnection(conn net.Conn) error {
	defer conn.Close()

	data, err := readLine(conn)
	if err != nil {
		return err
	}

	revData := toLeetSpeak(data)
	_, err = conn.Write(revData)
	if err != nil {
		return err
	}

	return nil
}

func readLine(r io.Reader) ([]byte, error) {
	data := []byte{}
	for {
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
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

func toLeetSpeak(data []byte) []byte {
	for i := range data {
		switch data[i] {
		case 'a':
			data[i] = '4'
		case 'e':
			data[i] = '3'
		case 'i':
			data[i] = '1'
		case 'o':
			data[i] = '0'
		case 's':
			data[i] = '5'
		case 't':
			data[i] = '7'
		case 'l':
			data[i] = '1'
		}
	}
	return data
}

func main() {}
