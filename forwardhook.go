package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"os"
	"path"
)

var conf *Config

// mirrorRequest will POST through body and headers from an
// incoming http.Request.
// Failures are retried up to 10 times.
func mirrorRequest(method string, h http.Header, body []byte, url string, query string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	attempt := 1
	for {
		logger.Info.Printf("Attempting %s try=%d\n", url, attempt)

		client := &http.Client{}

		rB := bytes.NewReader(body)
		req, err := http.NewRequest(method, url, rB)
		if err != nil {
			logger.Error.Println("http.NewRequest:", err)
		}

		// Set headers from request
		req.Header = h

		req.URL.RawQuery = query
		logger.Info.Println(req.URL)

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

func handleHook(sites []string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rB, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Fail on ReadAll")
		}
		defer r.Body.Close()

		for _, url := range sites {
			go mirrorRequest(r.Method, r.Header, rB, url, r.URL.RawQuery)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"status": 1, "data":true}`))
	}
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NotFoundHandle(writer http.ResponseWriter, request *http.Request) {
	http.Error(writer, `{status:0, "msg":"handle not found!"}`, 404)
}

func main() {
	confPath := flag.String("c", "config.json", "config json file")
	initFlag := flag.Bool("init", false, "init config.json")
	flag.Parse()

	if *initFlag {
		dir, err := os.Getwd()
		if err != nil {
			logger.Error.Println(err)
			return
		}
		tmpFile := path.Join(dir, "/config.json")
		conf := InitConfig(tmpFile)
		err = conf.Save()
		if err != nil {
			logger.Error.Println(err)
		}
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	err, conf := NewConfig(*confPath)
	if err != nil {
		logger.Error.Println(err)
		return
	}

	r := mux.NewRouter()
	for _, mapping := range conf.Mappings {
		logger.Info.Printf("Will forward hooks to: %v , on Path:%v", mapping.Sites, mapping.Path)
		r.PathPrefix(mapping.Path).HandlerFunc(handleHook(mapping.Sites))
	}
	r.HandleFunc("/status", handleHealthCheck)
	r.NotFoundHandler = http.HandlerFunc(NotFoundHandle)

	logger.Info.Printf("Listening on: http://%v\n", conf.Listen)
	err = http.ListenAndServe(conf.Listen, r)
	if err != nil {
		logger.Error.Fatal(err)
	}
}
