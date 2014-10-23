package main

import (
	"io/ioutil"
	"net/http"
	"fmt"
)

var indexPage string

func webServer() {
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/stop", http.HandlerFunc( StopHandler ))
	http.Handle("/start", http.HandlerFunc( StartHandler ))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", serveTemplate)

	Message("Listening...", false)
	http.ListenAndServe(":5000", nil)
}

func loadIndex() {
	Message("Load index", false)
	content, err := ioutil.ReadFile("web/index.html")
	if err != nil {
		indexPage = err.Error()
	} else {
		indexPage = string(content)
	}
}

func StopHandler(response http.ResponseWriter, request *http.Request){
	response.Header().Set("Content-type", "text/html")
	
	if runningClicker == true {
		quit <- true
		fmt.Fprint(response, "OK");
	} else {
		fmt.Fprint(response, "OK");
	}
}

func StartHandler(response http.ResponseWriter, request *http.Request){
	response.Header().Set("Content-type", "text/html")
	go voter()
	fmt.Fprint(response, "OK");
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	fmt.Fprint(w, indexPage);
}