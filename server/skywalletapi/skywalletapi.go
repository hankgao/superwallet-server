package skywalletapi

import (
	"encoding/json"

	"github.com/skycoin/skycoin/src/cipher"
	bip39 "github.com/tyler-smith/go-bip39"
)

// GetSupportedCoins returns a list of coins that are currently supported, in JSON format
func GetSupportedCoins() (string, error) {
	return "", nil
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
	return "", nil
}

// GetBalance returns balances of addresses
func GetBalance(coinType, addresses string) (string, error) {
	// check to see if coinType is bitcoin, if it is, then go to Bitcoin code
	return "", nil
}
