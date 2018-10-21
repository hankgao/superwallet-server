package mobile

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

// CoinMeta represents a structure that holds metadata for a certain coin type
type CoinMeta struct {
	NameInChinese    string `json:"nameInChinese"`
	NameInEnglish    string `json:"nameInEnglish"`
	Symbol           string `json:"symbol"`
	LogoURL          string `json:"logoURL"`
	WebInterfacePort string `json:"webInterfacePort"`
	NodeVersion      string `json:"nodeVersion"`
}

// CoinMetas represents a slice of CoinMeta
type CoinMetas []CoinMeta
