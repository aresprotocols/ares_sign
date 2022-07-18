package wallet

import (
	"fmt"
	"testing"
)

func TestStrOToBigInt(t *testing.T) {
	fmt.Println(StrToBigInt("0x000000000000000000000000000000000000000000000d3c21bcecceda100000"))
}
