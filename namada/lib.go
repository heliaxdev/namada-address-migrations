package namada

import (
	"fmt"

	"replace-addrs/bech32m"
)

const (
	hrpAddressOld = "atest"
	hrpAddressNew = "tnam"
)

func convertAddressData(data []byte) ([]byte, error) {
	return nil, nil
}

func ConvertAddress(oldAddress string) (string, error) {
	hrp, data, err := bech32m.DecodeAndConvert(oldAddress)
	if err != nil {
		return "", err
	}
	if hrp != hrpAddressOld {
		return "", fmt.Errorf("invalid hrp: %s", hrp)
	}

	newData, err := convertAddressData(data)
	if err != nil {
		return "", err
	}

	newAddress, err := bech32m.ConvertAndEncode(hrpAddressNew, newData)
	if err != nil {
		return "", fmt.Errorf("failed to encode addr data with bech32m: %w", err)
	}

	return newAddress, nil
}
