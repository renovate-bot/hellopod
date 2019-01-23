// Copyright 2019 Simon Pasquier
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var (
	help           bool
	ready, healthy int32
	listen, host   string
	ok, nok, hello *bytes.Buffer
)

func init() {
	flag.BoolVar(&help, "help", false, "Help message")
	flag.StringVar(&listen, "listen-address", ":8080", "Listen address")
	ready = 1
	healthy = 1

	var err error
	if host, err = os.Hostname(); err != nil {
		panic(err)
	}
	ok = bytes.NewBufferString(fmt.Sprintf("OK from %s", host))
	nok = bytes.NewBufferString(fmt.Sprintf("NOK from %s", host))
	hello = bytes.NewBufferString(fmt.Sprintf("Hello from %s!\n", host))
}

func main() {
	flag.Parse()
	if help {
		fmt.Fprintln(os.Stderr, "Hello pod!")
		flag.PrintDefaults()
		os.Exit(0)
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if atomic.LoadInt32(&healthy) == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(nok.Bytes())
				return
			}
			w.Write(ok.Bytes())
		case "DELETE":
			atomic.StoreInt32(&healthy, 0)
		case "POST":
			atomic.StoreInt32(&healthy, 1)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			if atomic.LoadInt32(&ready) == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write(nok.Bytes())
				return
			}
			w.Write(ok.Bytes())
		case "DELETE":
			atomic.StoreInt32(&ready, 0)
		case "POST":
			atomic.StoreInt32(&ready, 1)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write(hello.Bytes())
	})

	log.Println("Listening on", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}
