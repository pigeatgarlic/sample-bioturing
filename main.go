package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
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

func UploadInfoFunction(update_host, msg string) {
	fmt.Printf("%s\n", msg)
	return

	body, _ := json.Marshal(struct {
		Data  string `json:"data"`
		Token string `json:"token"`
		Body  string `json:"body"`
	}{
		Data:  Encrypt([]byte(msg)),
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
}

func captureLog(update_host string, proc *exec.Cmd) {
	stdout, err := proc.StdoutPipe()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}
	stderr, err := proc.StderrPipe()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	proc.Start()
	buffer := make([]byte, 2048)
	end := make(chan error, 2)
	go func() {
		for {
			size, err := stdout.Read(buffer)
			if err != nil {
				end <- err
			}
			for _,lines  := range strings.Split(string(buffer[:size]), "\n") {
				for _,line  := range strings.Split(lines, "\t") {
					timestamp := time.Now().Format(time.DateTime)
					UploadInfoFunction(update_host, fmt.Sprintf("%s : %s",timestamp,line))
				}
			}

		}
	}()
	go func() {
		for {
			size, err := stderr.Read(buffer)
			if err != nil {
				end <- err
			}
			UploadInfoFunction(update_host, string(buffer[:size]))
		}
	}()
	fmt.Printf("file watcher self closed %s\n", (<-end).Error())
}

func main() {
	watch_folder := "./data/"
	manifest := "./.image-info"
	update_host, found := os.LookupEnv("BIOTURING_T2D_HOST")
	if !found {
		fmt.Errorf("bioturing update host not found") // TODO
	}

	go func() {
		file, err := os.OpenFile(manifest, os.O_RDONLY, 0755)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		for {
			data, err := io.ReadAll(file)
			if err != nil {
				panic(err)
			}

			UploadInfoFunction(update_host, string(data))
			time.Sleep(time.Minute)
		}
	}()

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Printf("failed to initialize file watcher %s\n", err.Error())
			return
		}
		defer watcher.Close()

		watch_process := map[string]*os.Process{}
		watching_files := []string{}
		watcher.Add(watch_folder)
		err = filepath.Walk(watch_folder, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("failed to initialize file watcher %s\n", err.Error())
				return err
			} else if info.IsDir() {
				return nil
			}

			watching_files = append(watching_files, path)
			return nil
		})

		fmt.Printf("watching files %v\n",watching_files)
		if err != nil {
			fmt.Printf("failed to initialize file watcher %s\n", err.Error())
		}

		go func() {
			for {
				delete_files := []string{}
				for watching_file, process := range watch_process { // kill process
					found := false
					for _, shouldwatch_file := range watching_files {
						if shouldwatch_file == watching_file {
							found = true
						}
					}
					if found {
						continue
					}

					process.Kill()
					delete_files = append(delete_files, watching_file)
					fmt.Printf("deleting file watcher for %s\n",watching_file)
				}

				for _, file := range delete_files {
					delete(watch_process, file)
				}
				keys := make([]string, 0, len(watch_process)) // start tail -f process
				for k := range watch_process {
					keys = append(keys, k)
				}

				for _, shouldwatch_file := range watching_files {
					found := false
					for _, watching_file := range keys {
						if watching_file == shouldwatch_file {
							found = true
						}
					}

					if found {
						continue
					}

					fmt.Printf("starting tail -f %s\n", shouldwatch_file)
					cmd := exec.Command("tail", "-f", shouldwatch_file)
					go captureLog(update_host, cmd)
					watch_process[shouldwatch_file] = cmd.Process
				}

				time.Sleep(time.Second)
			}
		}()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					fmt.Printf("error watch event\n")
				}

				if event.Op == fsnotify.Create {
					watching_files = append(watching_files, event.Name)
					fmt.Printf("new watching file: %v\n", watching_files)
				} else if event.Op == fsnotify.Remove {
					temp := []string{}
					for _, file := range watching_files {
						if event.Name == file {
							continue
						}
						temp = append(temp, file)
					}
					watching_files = temp
				}
			case event, ok := <-watcher.Errors:
				if !ok {
					fmt.Printf("error watch error\n")
				}

				fmt.Printf("error watch file : %s\n", event.Error())
			}
		}
	}()

	for {
		time.Sleep(time.Minute)
	}
}
