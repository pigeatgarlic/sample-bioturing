package main

import (
	"fmt"
	"net/http"
	"os"
)

func main(){
	port := 60000
	server := http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", port)}

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		file, err := os.OpenFile("/.image-info", os.O_RDONLY, 0755)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte(fmt.Sprintf("failed to open image info %s,", []byte(err.Error()))))
		}

		bytes := make([]byte,2048)
		n,err := file.Read(bytes)
		if err != nil {
			w.WriteHeader(503)
			w.Write([]byte(fmt.Sprintf("failed to open image info %s,", []byte(err.Error()))))
		}


		w.WriteHeader(200)
		w.Write(bytes[:n])
	})

	fmt.Printf("hello world\n")
	panic(server.ListenAndServe())
}
