package tangle

import (
	"sync"
	"testing"

	"github.com/iotaledger/goshimmer/packages/events"
	"github.com/iotaledger/goshimmer/packages/model/value_transaction"
	"github.com/iotaledger/goshimmer/packages/ternary"
	"github.com/iotaledger/goshimmer/plugins/gossip"
)

func TestSolidifier(t *testing.T) {
	// initialize plugin
	configureDatabase(nil)
	configureSolidifier(nil)

	// create transactions and chain them together
	transaction1 := value_transaction.New()
	transaction1.SetNonce(ternary.Trinary("99999999999999999999999999A"))
	transaction2 := value_transaction.New()
	transaction2.SetBranchTransactionHash(transaction1.GetHash())
	transaction3 := value_transaction.New()
	transaction3.SetBranchTransactionHash(transaction2.GetHash())
	transaction4 := value_transaction.New()
	transaction4.SetBranchTransactionHash(transaction3.GetHash())

	// setup event handlers
	var wg sync.WaitGroup
	Events.TransactionSolid.Attach(events.NewClosure(func(transaction *value_transaction.ValueTransaction) {
		wg.Done()
	}))

	// issue transactions
	wg.Add(4)
	gossip.Events.ReceiveTransaction.Trigger(transaction1.MetaTransaction)
	gossip.Events.ReceiveTransaction.Trigger(transaction2.MetaTransaction)
	gossip.Events.ReceiveTransaction.Trigger(transaction3.MetaTransaction)
	gossip.Events.ReceiveTransaction.Trigger(transaction4.MetaTransaction)

	// wait until all are solid
	wg.Wait()
}
