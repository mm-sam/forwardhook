package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"time"
	"log"
	"runtime"
)

var conf *Config

// mirrorRequest will POST through body and headers from an
// incoming http.Request.
// Failures are retried up to 10 times.
func mirrorRequest(method string, h http.Header, body []byte, url string) {
	attempt := 1
	for {
		logger.Info.Printf("Attempting %s try=%d\n", url, attempt)

		client := &http.Client{}

		rB := bytes.NewReader(body)
		req, err := http.NewRequest(method, url, rB)
		if err != nil {
			logger.Error.Println("[error] http.NewRequest:", err)
		}

		// Set headers from request
		req.Header = h

		logger.Info.Println(h)

		resp, err := client.Do(req)
		if err != nil {
			logger.Error.Println("[error] client.Do:", err)
			time.Sleep(10 * time.Second)
		} else {
			resp.Body.Close()
			logger.Info.Printf("[success] %s status=%d\n", url, resp.StatusCode)
			break
		}

		attempt++
		if attempt > conf.MaxRetries {
			logger.Error.Println("[error] maxRetries reached")
			break
		}
	}
}

func handleHook(sites []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Fail on ReadAll")
		}
		defer r.Body.Close()

		for _, url := range sites {
			go mirrorRequest(r.Method, r.Header, rB, url)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"status": 1, "data":true}`))
	})
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	confPath := flag.String("c", "config.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	err, conf := NewConfig(*confPath)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	for _, mapping := range conf.Mappings {
		logger.Info.Printf("Will forward hooks to: %v , on Path:%v", mapping.Sites,mapping.Path)
		http.Handle(mapping.Path, handleHook(mapping.Sites))
	}
	http.HandleFunc("/status", handleHealthCheck)
	http.NotFoundHandler()

	logger.Info.Printf("Listening on: http://%v\n", conf.Listen)
	err = http.ListenAndServe(conf.Listen, nil)
	if err != nil {
		logger.Error.Fatal(err)
	}
}
