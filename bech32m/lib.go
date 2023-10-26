package bech32m

import (
	"fmt"

	"github.com/pactus-project/pactus/util/bech32m"
)

func ConvertAndEncode(hrp string, data []byte) (string, error) {
	converted, err := bech32m.ConvertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("encoding bech32m failed: %w", err)
	}

	return bech32m.Encode(hrp, converted)
}

func DecodeAndConvert(bech string) (string, []byte, error) {
	hrp, data, err := bech32m.DecodeNoLimit(bech)
	if err != nil {
		return "", nil, fmt.Errorf("decoding bech32m failed: %w", err)
	}

	converted, err := bech32m.ConvertBits(data, 5, 8, false)
	if err != nil {
		return "", nil, fmt.Errorf("decoding base32 failed: %w", err)
	}

	return hrp, converted, nil
}
