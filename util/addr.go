package util

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/xssnick/tonutils-go/address"
)

func AddressesEq(addr1 *address.Address, addr2 *address.Address) bool {
	wc1, d1 := addr1.Workchain(), addr1.Data()
	wc2, d2 := addr2.Workchain(), addr2.Data()
	return wc1 == wc2 && bytes.Equal(d1, d2)
}

func AddressesEqS(a1 string, a2 string) bool {
	addr1, e := address.ParseAddr(a1)
	if e != nil {
		return false
	}
	addr2, e := address.ParseAddr(a2)
	if e != nil {
		return false
	}
	return AddressesEq(addr1, addr2)
}

func AddressToString(addr *address.Address) string {
	a := addr.String()
	addrx, e := address.ParseAddr(a)
	if e != nil {
		return ""
	}
	addrx.SetBounce(true)
	addrx.SetTestnetOnly(false)
	return addrx.String()
}

func AddressFromRaw(a string) (addr *address.Address, e error) {
	ap := strings.Split(a, ":")
	if len(ap) != 2 {
		e = errors.New("invalid raw address")
		return
	}
	i, e := strconv.ParseInt(ap[0], 10, 32)
	if e != nil {
		return
	}
	h, e := hex.DecodeString(ap[1])
	if e != nil {
		return
	}
	addr = address.NewAddress(byte(0), byte(i), h)
	return
}

func AddressToRaw(addr *address.Address) string {
	wc := addr.Workchain()
	hash := addr.Data()
	hh := hex.EncodeToString(hash)
	return fmt.Sprintf("%d:%s", wc, hh)
}

func AddressToRawS(addr string) (string, error) {
	a, e := address.ParseAddr(addr)
	if e != nil {
		return "", e
	}
	return AddressToRaw(a), nil
}
