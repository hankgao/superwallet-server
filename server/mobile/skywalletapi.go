package mobile

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hankgao/superwallet-server/server/mobile/bitcoin"
	log "github.com/sirupsen/logrus"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/coin"
	"github.com/skycoin/skycoin/src/util/droplet"
	"github.com/skycoin/skycoin/src/visor"
	bip39 "github.com/tyler-smith/go-bip39"
)

var httpClient http.Client

const (
	packageVersion                    = "1.0.2"
	dialTimeout         time.Duration = 60 * time.Second
	tlsHandshakeTimeout time.Duration = 60 * time.Second
	httpClientTimeout   time.Duration = 120 * time.Second
	coinHourFee         uint64        = 50 // percent, i.e, 50%

	GET_SUPPORTED_COINS = "getSupportedCoins"
	GET_BALANCE         = "getBalance"
	GET_OUTPUTS         = "getOutputs"
	INJECT_TRANSACTION  = "injectTransaction"
	GET_TRANSACTION     = "transaction"
)

var superwalletServer = "http://127.0.0.1:6789"

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

	log.SetLevel(log.InfoLevel)
}

// SetServer allows client to change back-end server, for example, for testing purpose
func SetServer(url string) {
	superwalletServer = url
}

// GetApiVersion returns the version that is being used
func GetApiVersion() string {
	return packageVersion
}

// GetSupportedCoins returns a list of coins that are currently supported, in JSON format
func GetSupportedCoins() (string, error) {
	path := fmt.Sprintf("%s/%s", superwalletServer, GET_SUPPORTED_COINS)
	r, err := httpClient.Get(path)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()

	rawBytes, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return "", err
	}

	return string(rawBytes), nil
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

func bitcoinGenerateAddrs(lastSeed string, qty int) NewAddressesResult {
	stub, addrs := bitcoin.GenerateAddresses([]byte(lastSeed), qty)

	entries := make([]AddressEntry, qty)
	for i, addr := range addrs {
		entries[i].Address = addr.Address
		entries[i].Public = addr.Public
		entries[i].Secret = addr.Secret
	}

	return NewAddressesResult{
		LastSeed: stub,
		Addrs:    entries,
	}

}

func skycoinGenerateAddrs(lastSeed string, qty int) NewAddressesResult {
	sd, seckeys := cipher.GenerateDeterministicKeyPairsSeed([]byte(lastSeed), qty)
	entries := make([]AddressEntry, qty)
	for i, sec := range seckeys {
		pub := cipher.PubKeyFromSecKey(sec)
		entries[i].Address = cipher.AddressFromPubKey(pub).String()
		entries[i].Public = pub.Hex()
		entries[i].Secret = sec.Hex()
	}

	return NewAddressesResult{
		LastSeed: hex.EncodeToString(sd),
		Addrs:    entries,
	}

}

// GenerateNewAddresses creates qty new addresses using a seed provided
func GenerateNewAddresses(coinType, lastSeed string, qty int) (string, error) {

	nar := NewAddressesResult{}
	if coinType == "bitcoin" {
		nar = bitcoinGenerateAddrs(lastSeed, qty)
	} else {
		nar = skycoinGenerateAddrs(lastSeed, qty)
	}

	jsonBytes, err := json.MarshalIndent(nar, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// SendCoin sends coins from a list of addresses to a target address
func SendCoin(coinType, inputAddresses, privateKeys, targetAddress string, amount float64) (string, error) {

	r, err := createRawTx(coinType, inputAddresses, privateKeys, targetAddress, amount)
	if err != nil {
		return "", err
	}

	rawtx := struct {
		Rawtx string `json:"rawtx"`
	}{
		Rawtx: r,
	}

	rawBytes, err := json.MarshalIndent(rawtx, "", "    ")

	url := fmt.Sprintf("%s/%s/%s", superwalletServer, coinType, INJECT_TRANSACTION)
	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(rawBytes))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(string(bodyBytes))
	}

	return string(bodyBytes), nil
}

func bitcoinGetBalance(addrs string) (string, error) {
	b := bitcoin.Bitcoin{}
	balance, err := b.GetBalance(strings.Split(addrs, ","))
	if err != nil {
		return "", err
	}

	return balance.String(), nil
}

func bitcoinSendcoin() {

}

// GetBalance returns balances of addresses
func GetBalance(coinType, addresses string) (string, error) {
	// check to see if coinType is bitcoin, if it is, then go to Bitcoin code
	path := fmt.Sprintf("%s/%s/%s", superwalletServer, coinType, GET_BALANCE)

	if coinType == "bitcoin" {
		return bitcoinGetBalance(addresses)
	}

	// skycoin based coin
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

// GetOutputs is called by Send method as inputs to create a raw transtion, which is then be injected
func GetOutputs(coinType, addrs string) (string, error) {
	// check to see if coinType is bitcoin, if it is, then go to Bitcoin code
	path := fmt.Sprintf("%s/%s/%s", superwalletServer, coinType, GET_OUTPUTS)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("addrs", addrs)

	req.URL.RawQuery = q.Encode()

	path = req.URL.String()

	return httpGet(path)
}

func GetTransaction(coinType, txID string) (string, error) {
	// superwallet.shellpay2.com/mzcoin/transaction
	path := fmt.Sprintf("%s/%s/%s", superwalletServer, coinType, GET_TRANSACTION)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("txid", txID)

	req.URL.RawQuery = q.Encode()

	path = req.URL.String()
	log.Info("send HTTP request: ", path)
	// superwallet.shellpay2.com/mzcoin/transaction?txid=<txID>

	return httpGet(path)
}

func httpGet(path string) (string, error) {
	r, err := httpClient.Get(path)
	if err != nil {
		log.Error("failed to request ", path, " => ", err.Error())
		return "", err
	}

	defer r.Body.Close()
	rawBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("failed to read response data =>", err.Error())
		return "", err
	}

	return string(rawBytes), nil

}

func AddrSecKeyMapFromString(inputAddresses, privateKeys string) (map[string]string, error) {

	asm := make(map[string]string)
	keys := strings.Split(inputAddresses, ",")
	values := strings.Split(privateKeys, ",")

	if len(keys) != len(values) {
		return asm, fmt.Errorf("private keys not match addresses")
	}
	for i, key := range keys {
		asm[key] = values[i]
	}

	return asm, nil
}

func createRawTx(coinType, inputAddresses, privateKeys, targetAddress string, amount float64) (string, error) {

	asm, err := AddrSecKeyMapFromString(inputAddresses, privateKeys)
	if err != nil {
		return "", err
	}

	// Step 1: get all spendable outputs as input
	outputs, err := GetOutputs(coinType, inputAddresses)
	if err != nil {
		return "", err
	}

	o := visor.ReadableOutputs{}
	err = json.Unmarshal([]byte(outputs), &o)
	if err != nil {
		return "", err
	}

	balance, err := o.Balance()
	if err != nil {
		return "", err
	}

	droplets2Transfer := uint64(math.Round(amount*1000) * 1000) // only three decimals supported !!!

	if balance.Coins < droplets2Transfer {
		return "", fmt.Errorf("not enough coins [%d vs %d]", balance.Coins, droplets2Transfer)
	}

	sortUx(o)

	tx := coin.Transaction{}

	// choose the smallest upspents as inputs
	var inputDroplets, inputHours uint64
	var signKeys []cipher.SecKey
	for _, ux := range o {
		d, err := droplet.FromString(ux.Coins)
		if err != nil {
			return "", err
		}

		inputDroplets += d

		inputHours += ux.CalculatedHours

		tx.PushInput(cipher.MustSHA256FromHex(ux.Hash))
		signKeys = append(signKeys, cipher.MustSecKeyFromHex(asm[ux.Address]))

		if inputDroplets >= droplets2Transfer {
			break
		}
	}

	inputHours = inputHours - (inputHours * coinHourFee / 100)

	change := inputDroplets - droplets2Transfer
	if change > 0 {
		changeAddr := o[0].Address // use address of the first ux as change address
		tx.PushOutput(cipher.MustDecodeBase58Address(changeAddr), change, inputHours*9/10)
	}

	tx.PushOutput(cipher.MustDecodeBase58Address(targetAddress), droplets2Transfer, inputHours/10)

	tx.SignInputs(signKeys)

	tx.UpdateHeader()
	d := tx.Serialize()

	return hex.EncodeToString(d), nil

}

func sortUx(input visor.ReadableOutputs) bool {

	sort.Slice(input, func(i, j int) bool {
		a, _ := strconv.ParseUint(input[i].Coins, 10, 64)
		b, _ := strconv.ParseUint(input[j].Coins, 10, 64)

		return a < b
	})

	return true
}

// GetClientID returns a unique ID for a client that will be used to identify client in later call
// func GetClientID() uint64 {
// 	return uint64(0)
// }

// func getClientIDFromRequest(r *http.Request) uint64 {
// 	return uint64(0)
// }
