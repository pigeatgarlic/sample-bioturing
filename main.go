package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

func UploadInfoFunction(update_host, source, timestamp, msg string) {
	raw, _ := json.Marshal(struct {
		Timestamp string `json:"timestamp"`
		Level     string `json:"level"`
		Source    string `json:"source"`
		Data      string `json:"data"`
	}{
		Data:      msg,
		Source:    "unknown",
		Level:     "unknown",
		Timestamp: time.Now().Format(time.RFC3339),
	})

	body, _ := json.Marshal(struct {
		Data  []string `json:"data"`
		Token string   `json:"token"`
	}{
		Data:  Encrypt(raw),
		Token: "token",
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

	fmt.Printf("%s %s %s\n", source, timestamp, msg)
}

func captureLog(update_host string, proc *exec.Cmd) {
	stdout, err := proc.StdoutPipe()
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
			for _, lines := range strings.Split(string(buffer[:size]), "\n") {
				for _, line := range strings.Split(lines, "\t") {
					if len(line) == 0 {
						continue
					}

					timestamp := time.Now().Format(time.DateTime)
					UploadInfoFunction(update_host, proc.Args[3], timestamp, line)
				}
			}

		}
	}()
	fmt.Printf("file watcher self closed %s\n", (<-end).Error())
}

func IterateFolder(path string) []string {
	watching_files := []string{}
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("failed to initialize file watcher %s\n", err.Error())
			return err
		} else if info.IsDir() {
			return nil
		}

		watching_files = append(watching_files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}

	return watching_files
}

func main() {
	watch_folder := "/data/ecosystem/app_data/t2d_logs"
	manifest := "./.image-info"
	update_host, found := os.LookupEnv("BIOTURING_T2D_HOST")
	if !found {
		fmt.Println("bioturing update host not found") // TODO
		update_host = "http://localhost:60000"
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

			// UploadInfoFunction(update_host, string(data))
			_ = data
			time.Sleep(time.Minute)
		}
	}()

	go func() {
		watch_process := map[string]*os.Process{}
		watching_files := IterateFolder(watch_folder)

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

					if process != nil {
						process.Kill()
					}

					delete_files = append(delete_files, watching_file)
					fmt.Printf("deleting file watcher for %s\n", watching_file)
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

					fmt.Printf("watching file %s\n", shouldwatch_file)
					cmd := exec.Command("tail", "-n0", "-f", shouldwatch_file)
					go captureLog(update_host, cmd)
					watch_process[shouldwatch_file] = cmd.Process
				}

				time.Sleep(time.Second)
			}
		}()

		for {
			watching_files = IterateFolder(watch_folder)
			time.Sleep(time.Second)
		}
	}()

	for {
		time.Sleep(time.Minute)
	}
}
