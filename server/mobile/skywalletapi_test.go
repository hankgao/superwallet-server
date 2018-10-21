package mobile

import "testing"

func TestGetSupportedCoins(t *testing.T) {
	_, err := GetSupportedCoins()
	if err != nil {
		t.Error(err)
	}
}
