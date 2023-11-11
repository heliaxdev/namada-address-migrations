package addrconv

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
	"crypto/sha256"
	"unsafe"

	"github.com/heliaxdev/namada-address-migrations/bech32m"
)

const (
	hrpAddressOld = "atest"
	hrpAddressNew = "tnam"

	hrpMaspAddressOldUnpinned = "patest"
	hrpMaspAddressOldPinned   = "ppatest"
	hrpMaspAddressNew         = "znam"

	hrpSpendingOld = "xsktest"
	hrpSpendingNew = "zsknam"

	hrpViewingOld = "xfvktest"
	hrpViewingNew = "zvknam"

	hrpPubkeyOld = "pktest"
	hrpPubkeyNew = "tpknam"

	hrpDkgOld = "dpktest"
	hrpDkgNew = "dpknam"

	hrpSigOld = "sigtest"
	hrpSigNew = "signam"
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

func PublicKeyToImplicit(publicKey string) (string, error) {
	hrp, pubKeyData, err := bech32m.DecodeAndConvert(publicKey)
	if err != nil {
		return "", err
	}
	if hrp != hrpPubkeyNew {
		return "", fmt.Errorf("invalid pk hrp: %s", hrp)
	}

	pkHash := sha256.Sum256(pubKeyData)
	rawAddress := buildAddress(0, pkHash[:20])

	return encodeAddress(hrpAddressNew, rawAddress)
}

func ConvertAddress(oldAddress string) (string, error) {
	hrp, data, err := bech32m.DecodeAndConvert(oldAddress)
	if err != nil {
		return "", err
	}
	switch hrp {
	case hrpAddressOld:
		return convertTransparentAddress(data)
	case hrpMaspAddressOldUnpinned:
		return convertMaspPayment(false, data)
	case hrpMaspAddressOldPinned:
		return convertMaspPayment(true, data)
	case hrpSpendingOld:
		return encodeAddress(hrpSpendingNew, data)
	case hrpViewingOld:
		return encodeAddress(hrpViewingNew, data)
	case hrpPubkeyOld:
		return encodeAddress(hrpPubkeyNew, data)
	case hrpDkgOld:
		return encodeAddress(hrpDkgNew, data)
	case hrpSigOld:
		return encodeAddress(hrpSigNew, data)
	default:
		return "", fmt.Errorf("invalid hrp: %s", hrp)
	}
}

func convertMaspPayment(pinned bool, data []byte) (string, error) {
	newData := make([]byte, 0, 64)
	if pinned {
		newData = append(newData, 1)
	} else {
		newData = append(newData, 0)
	}
	newData = append(newData, data...)
	return encodeAddress(hrpMaspAddressNew, newData)
}

func convertTransparentAddress(data []byte) (string, error) {
	newData, err := convertTranspAddressData(data)
	if err != nil {
		return "", fmt.Errorf("failed to convert to new addr data format: %w", err)
	}
	return encodeAddress(hrpAddressNew, newData)
}

func convertTranspAddressData(data []byte) ([]byte, error) {
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

func encodeAddress(hrp string, data []byte) (string, error) {
	newAddress, err := bech32m.ConvertAndEncode(hrp, data)
	if err != nil {
		return "", fmt.Errorf("failed to encode addr data with bech32m: %w", err)
	}
	return newAddress, nil
}
