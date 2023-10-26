package namada

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
	"unsafe"

	"replace-addrs/bech32m"
)

const (
	hrpAddressOld = "atest"
	hrpAddressNew = "tnam"
)

const (
	prefixEstablished = "est"
	prefixImplicit    = "imp"
	prefixInternal    = "ano"
	prefixIbc         = "ibc"
	prefixEth         = "eth"
	prefixNut         = "nut"
)

const (
	internal_POS               = "ano::Proof of Stake                          "
	internal_POS_SLASH_POOL    = "ano::Proof of Stake Slash Pool               "
	internal_PARAMETERS        = "ano::Protocol Parameters                     "
	internal_GOVERNANCE        = "ano::Governance                              "
	internal_IBC               = "ibc::Inter-Blockchain Communication          "
	internal_ETH_BRIDGE        = "ano::ETH Bridge Address                      "
	internal_ETH_BRIDGE_POOL   = "ano::ETH Bridge Pool Address                 "
	internal_REPLAY_PROTECTION = "ano::Replay Protection                       "
	internal_MULTITOKEN        = "ano::Multitoken                              "
	internal_PGF               = "ano::Pgf                                     "
)

func convertAddressData(data []byte) ([]byte, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("invalid utf8 data")
	}

	utf8Data := *(*string)(unsafe.Pointer(&data))
	addrParts := strings.SplitN(utf8Data, "::", 2)

	validInternalFormat := len(addrParts) == 2 &&
		len(addrParts[0]) == 3 &&
		len(addrParts[1]) == 40

	if !validInternalFormat {
		return nil, fmt.Errorf("unexpected internal address format")
	}

	switch addrParts[0] {
	case prefixImplicit:
		return buildAddressFromHex(0, addrParts[1])
	case prefixEstablished:
		return buildAddressFromHex(1, addrParts[1])
	case prefixInternal:
		switch utf8Data {
		case internal_POS:
			return buildDatalessAddress(2), nil
		case internal_POS_SLASH_POOL:
			return buildDatalessAddress(3), nil
		case internal_PARAMETERS:
			return buildDatalessAddress(4), nil
		case internal_GOVERNANCE:
			return buildDatalessAddress(5), nil
		case internal_ETH_BRIDGE:
			return buildDatalessAddress(7), nil
		case internal_ETH_BRIDGE_POOL:
			return buildDatalessAddress(8), nil
		case internal_MULTITOKEN:
			return buildDatalessAddress(9), nil
		case internal_PGF:
			return buildDatalessAddress(10), nil
		}
	case prefixIbc:
		if utf8Data == internal_IBC {
			return buildDatalessAddress(6), nil
		}
		return buildAddressFromHex(13, addrParts[1])
	case prefixEth:
		return buildAddressFromHex(11, addrParts[1])
	case prefixNut:
		return buildAddressFromHex(12, addrParts[1])
	}

	return nil, fmt.Errorf("unknown old address data")
}

func buildAddressFromHex(discriminant byte, data string) ([]byte, error) {
	decoded, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return buildAddress(discriminant, decoded), nil
}

func buildDatalessAddress(discriminant byte) []byte {
	return buildAddress(discriminant, make([]byte, 20))
}

func buildAddress(discriminant byte, data []byte) []byte {
	output := make([]byte, 21)
	output[0] = discriminant
	// assume input buf is 20 bytes long :x
	copy(output[1:], data)
	return output
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
		return "", fmt.Errorf("failed to convert to new addr data format: %w", err)
	}

	newAddress, err := bech32m.ConvertAndEncode(hrpAddressNew, newData)
	if err != nil {
		return "", fmt.Errorf("failed to encode addr data with bech32m: %w", err)
	}

	return newAddress, nil
}
