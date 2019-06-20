package address

import (
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/model/balance"
	"github.com/iotaledger/goshimmer/packages/ternary"
	"github.com/magiconair/properties/assert"
)

func TestAddress_SettersGetters(t *testing.T) {
	address := ternary.Trytes("A9999999999999999999999999999999999999999999999999999999999999999999999999999999F")
	shardMarker := ternary.Trytes("NPHTQORL9XKA")
	addressShard := New(address + shardMarker)

	balanceEntries := []*balance.Entry{balance.NewValue(100, 1), balance.NewValue(100, 2)}

	addressShard.Add(balanceEntries...)
	assert.Equal(t, addressShard.GetBalance(), int64(200), "Accumulated")
}

func TestBalance_MarshalUnmarshalGetters(t *testing.T) {
	address := ternary.Trytes("A9999999999999999999999999999999999999999999999999999999999999999999999999999999F")
	shardMarker := ternary.Trytes("NPHTQORL9XKA")
	addressShard := New(address + shardMarker)

	balanceEntries := []*balance.Entry{balance.NewValue(100, 1), balance.NewValue(100, 2)}

	addressShard.Add(balanceEntries...)

	addressShardByte := addressShard.Marshal()
	addressShardUnmarshaled := &Entry{}
	err := addressShardUnmarshaled.Unmarshal(addressShardByte)
	if err != nil {
		fmt.Println(err, len(addressShardByte))
	}
	assert.Equal(t, addressShardUnmarshaled.GetBalance(), addressShard.GetBalance(), "Accumulated")
}
