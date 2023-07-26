package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const DEFAULT_PORT = "3458"
const DEFAULT_HOST = "0.0.0.0"
const MAX_NAME_SIZE = 50
const KEY_CHARS = "A-Za-z0-9-_"

type Passway struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func (p Passway) String() string {
	var append string
	if p.Port != "" {
		append = ":" + p.Port
	}
	return (p.IP + append)
}

var registry = make(map[string]string)

func main() {
	port := os.Getenv("PASSWAY_PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	host := os.Getenv("PASSWAY_HOST")
	if host == "" {
		host = DEFAULT_HOST
	}

	http.HandleFunc("/", handler)
	log.Printf("Passwayd listening in http://%s:%s", host, port)
	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pattern := fmt.Sprintf("^\\/?[%s]{1,%d}\\/?$", KEY_CHARS, MAX_NAME_SIZE)
	re := regexp.MustCompile(pattern)
	clientIP := strings.Split(r.RemoteAddr, ":")[0]

	if r.Method == http.MethodPost {
		var req Passway

		if r.URL.Path != "/" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// TODO: validate IP address
		// TODO: validate port
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil || !re.Match([]byte(req.Name)) {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if req.IP == "" {
			req.IP = clientIP
		}

		logMsg := fmt.Sprintf("Update from %v: (%s) %+v", clientIP, req.Name, req.String())
		log.Println(logMsg)
		registry[req.Name] = req.String()

		w.WriteHeader(http.StatusOK)
		return
	}

	if !re.Match([]byte(r.URL.Path)) {
		http.Error(w, "Bad key name", http.StatusBadRequest)
		return
	}

	passwayName := r.URL.Path[1:]
	passway, ok := registry[passwayName]
	if !ok {
		logFetch(passwayName, clientIP, "not found")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	length, err := w.Write([]byte(passway))
	if err != nil {
		log.Printf("[write error] %v", err)
		http.Error(w, "Write error", http.StatusInternalServerError)
		return
	}

	logFetch(passwayName, clientIP, fmt.Sprintf("%d bytes", length))
}

func logFetch(key, remoteIP, result string) {
	log.Printf("Request key `%s` from %v: %s\n", key, remoteIP, result)
}
