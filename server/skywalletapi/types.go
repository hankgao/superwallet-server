package skywalletapi

// AddressEntry represents the wallet address
type AddressEntry struct {
	Address string `json:"address"`
	Public  string `json:"pubkey"`
	Secret  string `json:"seckey"`
}

// NewsAddressesResult represents a result returned by function NewAddresses
type NewsAddressesResult struct {
	LastSeed string         `json:"lastseed"`
	Addrs    []AddressEntry `json:"addrs"`
}
