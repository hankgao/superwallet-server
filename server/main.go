package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	skywallet "github.com/hankgao/superwallet-server/server/mobile"
	log "github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/api"
	"github.com/skycoin/skycoin/src/daemon"
	"github.com/skycoin/skycoin/src/visor"
)

var supportedCoinTypes map[string]skywallet.CoinMeta

var (
	nodeServer = "http://localhost"
	serverPort = "6789"
)

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
	r.HandleFunc("/{coinType}/getOutputs", getOutputsHandler)
	r.HandleFunc("/{coinType}/getBalance", getBalanceHandler)
	r.HandleFunc("/getSupportedCoins", getSupportedCoinsHandler)
	r.HandleFunc("/{coinType}/injectTransaction", injectRawTxHandler).Methods("POST")
	r.HandleFunc("/{coinType}/transaction", getTransactionHandler)
	r.PathPrefix("/static/").HandlerFunc(logoRequestHandler)
	http.Handle("/", r)

	// start server
	srv := &http.Server{
		Addr: "0.0.0.0:" + serverPort,
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

func injectRawTxHandler(w http.ResponseWriter, r *http.Request) {

	log.Infof("POST %s", r.URL.Path)

	vars := mux.Vars(r)
	coinType := vars["coinType"]
	if !isCoinTypeSupported(coinType) {
		http.Error(w, fmt.Sprintf("%s is not supported", coinType), http.StatusForbidden)
		return
	}

	rawtx := struct {
		Rawtx string `json:"rawtx"`
	}{}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("[%s] %s", coinType, err), http.StatusForbidden)
		return
	}

	err = json.Unmarshal(bytes, &rawtx)
	if err != nil {
		http.Error(w, fmt.Sprintf("[%s] %s", coinType, err), http.StatusForbidden)
		return
	}

	log.Infof("rawtx: \n%s", rawtx.Rawtx)

	cm := supportedCoinTypes[coinType]

	c := api.NewClient(fmt.Sprintf("%s:%s", nodeServer, cm.WebInterfacePort))

	txid, err := c.InjectTransaction(rawtx.Rawtx)
	if err != nil {
		log.Errorf("failed to inject raw transaction %s", err)
		http.Error(w, fmt.Sprintf("[%s] %s", coinType, err), http.StatusForbidden)
		return
	}

	w.Write([]byte(txid))

}

/**
This is what is returned
{
    "confirmed": {
        "coins": 21000000,
        "hours": 142744
    },
    "predicted": {
        "coins": 21000000,
        "hours": 142744
    }
*/
func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)

	vars := mux.Vars(r)
	coinType := vars["coinType"]

	if isCoinTypeSupported(coinType) {
		ctm := supportedCoinTypes[coinType]

		// localhost:webInterfacePort
		c := api.NewClient(fmt.Sprintf("%s:%s", nodeServer, ctm.WebInterfacePort))

		values := r.URL.Query()
		// addrs should be comma seperated string
		addrs := values.Get("addrs")

		balance, err := c.Balance(strings.Split(addrs, ","))
		if err != nil {
			//TODO：
			log.Errorf("failed to get balance %s", err)
			response := fmt.Sprintf("getBalance handler failed: %s", err)
			http.Error(w, response, http.StatusInternalServerError) // client needs to check status
			return
		}

		bytes, err := json.MarshalIndent(balance, "", "    ")
		if err != nil {
			log.Errorf("failed to marshal return balance result: %s", err)
			response := fmt.Sprintf("getBalance handler failed: %s", err)
			http.Error(w, response, http.StatusInternalServerError) // client needs to check status
		}

		w.Write(bytes)

	} else {
		http.Error(w, fmt.Sprintf("%s is not supported", coinType), http.StatusBadRequest)
	}

}

func getSupportedCoinsHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)
	bytes, err := json.MarshalIndent(supportedCoinTypes, "", "    ")
	if err != nil {
		log.Errorf("failed to get supported coins %s", err)
		http.Error(w, fmt.Sprintf("getSupported coins failed due to: %s", err), http.StatusForbidden)
		return
	}
	w.Write(bytes)
}

func loadCoinsConfig() error {
	// We expect a configuration file in the current working directory
	bytes, err := ioutil.ReadFile("coins.config.json")
	if err != nil {
		log.Errorf("failed to load coin configruation file %s", err)
		return err
	}

	cms := skywallet.CoinMetas{}
	err = json.Unmarshal(bytes, &cms)
	if err != nil {
		log.Errorf("failed to unmarshal configuration data %s", err)
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

func getOutputs(coinType, addrs string) (*visor.ReadableOutputSet, error) {
	if !isCoinTypeSupported(coinType) {
		return nil, fmt.Errorf("%s type is not supported", coinType)
	}

	addr := fmt.Sprintf("%s:%s", nodeServer, supportedCoinTypes[coinType].WebInterfacePort)
	c := api.NewClient(addr)

	aSlice := strings.Split(addrs, ",")

	return c.OutputsForAddresses(aSlice)
}

func getTransaction(coinType, txid string) (*daemon.TransactionResult, error) {
	if !isCoinTypeSupported(coinType) {
		return nil, fmt.Errorf("%s type is not supported", coinType)
	}

	addr := fmt.Sprintf("%s:%s", nodeServer, supportedCoinTypes[coinType].WebInterfacePort)
	c := api.NewClient(addr)

	return c.Transaction(txid)
}

func getOutputsHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)

	vars := mux.Vars(r)
	coinType := vars["coinType"]

	if !isCoinTypeSupported(coinType) {
		http.Error(w, fmt.Sprintf("%s is not supported", coinType), http.StatusForbidden)
		return
	}

	values := r.URL.Query()
	addrs := values.Get("addrs")

	o, err := getOutputs(coinType, addrs)
	if err != nil {
		log.Errorf("failed to get outputs %s", err)
		http.Error(w, fmt.Sprintf("[%s] failed to get outputs: %s ", coinType, err), http.StatusForbidden)
		return
	}

	// spendable outputs
	so := o.SpendableOutputs()
	bytes, err := json.MarshalIndent(so, "", "    ")
	if err != nil {
		log.Errorf("failed to marshal spendable outputs %s", err)
		http.Error(w, fmt.Sprintf("[%s] failed to get outputs: %s ", coinType, err), http.StatusForbidden)
		return
	}

	w.Write(bytes)

}

func isCoinTypeSupported(coinType string) bool {
	_, ok := supportedCoinTypes[coinType]
	if !ok {
		return false
	}

	return true
}

func getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("GET %s", r.URL.Path)

	vars := mux.Vars(r)
	coinType := vars["coinType"]

	if !isCoinTypeSupported(coinType) {
		http.Error(w, fmt.Sprintf("%s is not supported", coinType), http.StatusForbidden)
		return
	}

	values := r.URL.Query()
	txid := values.Get("txid")

	tr, err := getTransaction(coinType, txid)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get transaction information for txid :%s", txid), http.StatusInternalServerError)
		return
	}
	txJSON, err := json.MarshalIndent(tr.Transaction, "", "    ")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal transaction information for txid :%s", txid), http.StatusInternalServerError)
		return
	}

	w.Write(txJSON)

}
