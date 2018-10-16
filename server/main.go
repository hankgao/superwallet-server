package main

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{coinType}/getBalance", getBalanceHandler)
	r.HandleFunc("/getSupportedCoins", getSupportedCoinsHanlder)
	r.HandleFunc("{coinType}/sendCoin", sendCoinHandler)
}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)
}

func getSupportedCoinsHanlder(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)
}

func sendCoinHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("PUT %s", r.URL.Path)
}
