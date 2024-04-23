package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	public_key = GenerateKey()
)

func GetFreePort() (port int, err error) {
	if addr, err := net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		if listener, err := net.ListenTCP("tcp", addr); err == nil {
			defer listener.Close()
			return listener.Addr().(*net.TCPAddr).Port, nil
		}
	}

	return 0, err
}

func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) string {
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, pub, msg, nil)
	if err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext)

}

func Encrypt(msg []byte) string {
	return EncryptWithPublicKey(msg, public_key)

}

func main() {
	update_host, found := os.LookupEnv("BIOTURING_T2D_HOST")
	if !found {
		panic(fmt.Errorf("bioturing update host not found"))
	}

	go func() {
		file, err := os.OpenFile("/.image-info", os.O_RDONLY, 0755)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		data := make([]byte, 2048)
		for {
			n, err := file.Read(data)
			if err != nil {
				panic(err)
			}

			body, _ := json.Marshal(struct {
				Data  string `json:"data"`
				Token string `json:"token"`
				Body  string `json:"body"`
			}{
				Data:  Encrypt(data[:n]),
				Token: " ",
				Body:  " ",
			})

			resp, err := http.Post(
				fmt.Sprintf("%s/log_upload_encrypted", update_host), 
				"application/json", 
				strings.NewReader(string(body)),
			)
			if err != nil {
				panic(err)
			}

			resp_body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != 200 {
				panic(fmt.Errorf(string(resp_body)))
			}

			fmt.Printf("uploaded log, got response %s\n", string(resp_body))
			time.Sleep(time.Minute)
		}
	}()

	for {
		time.Sleep(time.Minute)
	}
}
