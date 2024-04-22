package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
)

func GetFreePort() (port int, err error) {
	if addr, err := net.ResolveTCPAddr("tcp", "localhost:0"); err != nil {
		return addr.Port, nil
	}

	return 0, err
}

func main() {
	port, err := GetFreePort()
	if err != nil {
		panic(err)
	}

	ready_url, found := os.LookupEnv("READY_URL")
	if !found {
		panic(fmt.Errorf("READY_URL not found"))
	}

	secret, found := os.LookupEnv("SECRET_PASSWORD")
	if !found {
		panic(fmt.Errorf("READY_URL not found"))
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%d", ready_url, port), strings.NewReader(""))
	if err != nil {
		panic(err)
	}

	req.Header.Add("btr-secret-password", secret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	} else if resp.StatusCode != 200 {
		panic(fmt.Errorf("failed to call ready url %s", resp.Status))
	}

	server := http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port)}
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.OpenFile("/.image-info", os.O_RDONLY, 0755)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte(fmt.Sprintf("failed to open image info %s,", []byte(err.Error()))))
		}

		bytes := make([]byte, 2048)
		n, err := file.Read(bytes)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte(fmt.Sprintf("failed to open image info %s,", []byte(err.Error()))))
		}

		w.WriteHeader(200)
		w.Write(bytes[:n])
	})

	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {

	})

	fmt.Printf("hello world\n")
	panic(server.ListenAndServe())
}
