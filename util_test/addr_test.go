package util_test

import (
	"github.com/observe-fi/indexer/util"
	"testing"
)

func TestElectorAddr(t *testing.T) {
	r, e := util.AddressToRawS("Ef8zMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzMzM0vF")
	t.Log(r)
	t.Log(e)
}
