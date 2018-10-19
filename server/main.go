package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	skywallet "github.com/hankgao/superwallet-server/server/skywalletapi"
	log "github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/api"
)

var supportedCoinTypes map[string]skywallet.CoinMeta

func init() {
	err := loadCoinsConfig()
	if err != nil {
		panic(err)
	}
}

func main() {

	//https://github.com/gorilla/mux
	// prepare routing table
	r := mux.NewRouter()
	r.HandleFunc("/{coinType}/getBalance", getBalanceHandler)
	r.HandleFunc("/getSupportedCoins", getSupportedCoinsHanlder)
	r.HandleFunc("{coinType}/sendCoin", sendCoinHandler)
	r.HandleFunc("{coinType}/getOutputs", getOutputsHandler)
	r.PathPrefix("/static/").HandlerFunc(logoRequestHandler)
	http.Handle("/", r)

	// start server
	srv := &http.Server{
		Addr: "0.0.0.0:6789",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Errorf("Failed to start server: %s", err)
	}

}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)

	vars := mux.Vars(r)
	coinType := vars["coinType"]

	if isCoinTypeSupported(coinType) {
		ctm := supportedCoinTypes[coinType]

		// localhost:webInterfacePort
		c := api.NewClient(fmt.Sprintf("http://localhost:%s", ctm.WebInterfacePort))

		values := r.URL.Query()
		// addrs should be comma seperated string
		addrs := values.Get("addrs")

		balance, err := c.Balance(strings.Split(addrs, ","))
		if err != nil {
			//TODO：
			response := fmt.Sprintf("getBalance handler failed: %s", err)
			http.Error(w, response, http.StatusForbidden) // client needs to check status
			return
		}

		bytes, err := json.MarshalIndent(balance, "", "    ")
		if err != nil {
			log.Errorf("failed to marshal return balance result: %s", err)
		}

		w.Write(bytes)

	}

}

func getSupportedCoinsHanlder(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)
	bytes, err := json.MarshalIndent(supportedCoinTypes, "", "    ")
	if err != nil {
		http.Error(w, fmt.Sprintf("getSupported coins failed due to: %s", err), http.StatusForbidden)
		return
	}
	w.Write(bytes)
}

func sendCoinHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("PUT %s", r.URL.Path)
}

func loadCoinsConfig() error {
	// We expect a configuration file in the current working directory
	bytes, err := ioutil.ReadFile("coins.config.json")
	if err != nil {
		return err
	}

	cms := skywallet.CoinMetas{}
	err = json.Unmarshal(bytes, &cms)
	if err != nil {
		return err
	}

	supportedCoinTypes = make(map[string]skywallet.CoinMeta, len(cms))
	for _, cm := range cms {
		supportedCoinTypes[cm.NameInEnglish] = cm
	}

	return nil
}

// example request:
// http:superwallet.shellpay.com:6789/static/mzc.logo.png
func logoRequestHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, strings.TrimLeft(r.URL.Path, "/"))
}

func getOutputs(coinType, addrs string) (string, error) {
	if !isCoinTypeSupported(coinType) {
		return "", fmt.Errorf("%s type is not supported", coinType)
	}

	//

	return "", nil
}

func getOutputsHandler(w http.ResponseWriter, r *http.Request) {

}

func isCoinTypeSupported(coinType string) bool {
	_, ok := supportedCoinTypes[coinType]
	if !ok {
		return false
	}

	return true
}
