package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var count = 0
var passMap = map[int]string{}
var passMapChan = make(chan ReqPass)
var requestAverage = 0
var server http.Server

type ReqPass struct {
	Count int
	Pass  string
	Time  int
}

type Stats struct {
	Total   int
	Average int
}

func postHash(w http.ResponseWriter, r *http.Request) {
	log.Println("got /hash request")
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	pass := r.Header.Get("password")

	if pass == "" {
		http.Error(w, fmt.Sprintf("Error: you must set a password value in header: '%s'", pass), http.StatusBadRequest)
		return
	}
	count += 1
	current := count

	duration := time.Now().Sub(start)
	log.Println(duration.Microseconds())
	go func() {
		reqPass := ReqPass{Count: current, Pass: pass, Time: int(duration.Microseconds())}
		time.Sleep(5 * time.Second)
		passMapChan <- reqPass
	}()
	io.WriteString(w, strconv.Itoa(current))
}

func getHash(w http.ResponseWriter, r *http.Request) {
	log.Println("got /hash/:id request")
	if r.Method != http.MethodGet {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	p := strings.Split(r.URL.Path, "/")
	if len(p) == 1 {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	} else if len(p) > 1 {
		reqCount, err := strconv.Atoi(p[2])
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: invalid value for: '%s'", p[2]), http.StatusBadRequest)
			return
		} else {
			passEncode, ok := passMap[reqCount]
			if ok {
				w.WriteHeader(http.StatusOK)
				io.WriteString(w, passEncode)
				return
			}
			// The request has been accepted for processing, but the processing has not been completed
			w.WriteHeader(http.StatusAccepted)
			return
		}
	} else {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
}

func getStats(w http.ResponseWriter, r *http.Request) {
	log.Println("got /stats request")
	w.Header().Set("Content-Type", "application/json")
	stats := Stats{
		Total:   count,
		Average: requestAverage,
	}

	json.NewEncoder(w).Encode(stats)
}

func postShutdown(w http.ResponseWriter, r *http.Request) {
	log.Println("got /shutdown request")

	if err := server.Shutdown(r.Context()); err != nil {
		log.Fatal(err)
	}
}

func main() {
	go startEncodePass(passMapChan)

	server = http.Server{Addr: ":8000", Handler: getHandler()}

	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func getHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hash", postHash)
	mux.HandleFunc("/hash/", getHash)
	mux.HandleFunc("/stats", getStats)
	mux.HandleFunc("/shutdown", postShutdown)

	return mux
}

func startEncodePass(c chan ReqPass) {
	for {
		v, ok := <-c
		if ok == false {
			fmt.Println("BREAK")
			break
		}

		h := sha512.New()
		h.Write([]byte(v.Pass))
		sha := h.Sum(nil)

		sEnc := base64.StdEncoding.EncodeToString([]byte(sha))
		passMap[v.Count] = sEnc

		if v.Count == 1 {
			requestAverage = v.Time
		} else {
			requestAverage = (requestAverage + v.Time) / 2
		}
	}
}
