package skywalletapi

// GetSupportedCoins returns a list of coins that are currently supported, in JSON format
func GetSupportedCoins() (string, error) {
	return "", nil
}

// NewSeed returns a randomly generated seed which is unique globally
func NewSeed() (string, error) {
	return "", nil
}

// AddNewAddresses creates qty new addresses using a seed provided
func AddNewAddresses(lastSeed string, qty int) (string, error) {
	return "", nil
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
