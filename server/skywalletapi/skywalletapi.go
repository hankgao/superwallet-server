package skywalletapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/skycoin/skycoin/src/cipher"
	bip39 "github.com/tyler-smith/go-bip39"
)

var httpClient http.Client

const (
	dialTimeout         time.Duration = 60 * time.Second
	tlsHandshakeTimeout time.Duration = 60 * time.Second
	httpClientTimeout   time.Duration = 120 * time.Second

	serverURL           string = "http://127.0.0.1:6789"
	GET_SUPPORTED_COINS string = "getSupportedCoins"
	GET_BALANCE         string = "getBalance"
)

func init() {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: dialTimeout,
		}).Dial,
		TLSHandshakeTimeout: tlsHandshakeTimeout,
	}
	httpClient = http.Client{
		Transport: transport,
		Timeout:   httpClientTimeout,
	}
}

// GetSupportedCoins returns a list of coins that are currently supported, in JSON format
func GetSupportedCoins() (string, error) {
	path := fmt.Sprintf("%s/%s", serverURL, GET_SUPPORTED_COINS)
	r, err := httpClient.Get(path)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()

	bytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// NewSeed returns a randomly generated seed which is unique globally
func NewSeed() (string, error) {
	// TODO: support 256 bits in the future
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}

	sd, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	return sd, nil
}

// GenerateNewAddresses creates qty new addresses using a seed provided
func GenerateNewAddresses(lastSeed string, qty int) (string, error) {

	sd, seckeys := cipher.GenerateDeterministicKeyPairsSeed([]byte(lastSeed), qty)
	entries := make([]AddressEntry, qty)
	for i, sec := range seckeys {
		pub := cipher.PubKeyFromSecKey(sec)
		entries[i].Address = cipher.AddressFromPubKey(pub).String()
		entries[i].Public = pub.Hex()
		entries[i].Secret = sec.Hex()
	}

	nar := NewsAddressesResult{
		LastSeed: string(sd),
		Addrs:    entries,
	}

	jsonBytes, err := json.MarshalIndent(nar, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// SendCoin sends coins from a list of addresses to a target address
func SendCoin(coinType, inputAddresses, targetAddress string, amount float64) (string, error) {
	// get outputs
	// create raw transaction
	// inject transaction
	// note the server is just a proxy, api.Client will not be used
	return "", nil
}

// GetBalance returns balances of addresses
func GetBalance(coinType, addresses string) (string, error) {
	// check to see if coinType is bitcoin, if it is, then go to Bitcoin code
	path := fmt.Sprintf("%s/%s/%s", serverURL, coinType, GET_BALANCE)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("addrs", addresses)

	req.URL.RawQuery = q.Encode()

	path = req.URL.String()

	return httpGet(path)
}

func httpGet(path string) (string, error) {
	r, err := httpClient.Get(path)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	return string(bytes), nil

}
