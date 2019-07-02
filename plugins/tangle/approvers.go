package tangle

import (
	"github.com/dgraph-io/badger"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/packages/datastructure"
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/model/approvers"
	"github.com/iotaledger/goshimmer/packages/node"
	"github.com/iotaledger/goshimmer/packages/ternary"
)

// region global public api ////////////////////////////////////////////////////////////////////////////////////////////

// GetApprovers retrieves approvers from the database.
func GetApprovers(transactionHash ternary.Trytes, computeIfAbsent ...func(ternary.Trytes) *approvers.Approvers) (result *approvers.Approvers, err errors.IdentifiableError) {
	if cacheResult := approversCache.ComputeIfAbsent(transactionHash, func() interface{} {
		if dbApprovers, dbErr := getApproversFromDatabase(transactionHash); dbErr != nil {
			err = dbErr

			return nil
		} else if dbApprovers != nil {
			return dbApprovers
		} else {
			if len(computeIfAbsent) >= 1 {
				return computeIfAbsent[0](transactionHash)
			}

			return nil
		}
	}); cacheResult != nil && cacheResult.(*approvers.Approvers) != nil {
		result = cacheResult.(*approvers.Approvers)
	}

	return
}

func ContainsApprovers(transactionHash ternary.Trytes) (result bool, err errors.IdentifiableError) {
	if approversCache.Contains(transactionHash) {
		result = true
	} else {
		result, err = databaseContainsApprovers(transactionHash)
	}

	return
}

func StoreApprovers(approvers *approvers.Approvers) {
	approversCache.Set(approvers.GetHash(), approvers)
}

// region lru cache ////////////////////////////////////////////////////////////////////////////////////////////////////

var approversCache = datastructure.NewLRUCache(APPROVERS_CACHE_SIZE, &datastructure.LRUCacheOptions{
	EvictionCallback: onEvictApprovers,
})

func onEvictApprovers(_ interface{}, value interface{}) {
	if evictedApprovers := value.(*approvers.Approvers); evictedApprovers.GetModified() {
		go func(evictedApprovers *approvers.Approvers) {
			if err := storeApproversInDatabase(evictedApprovers); err != nil {
				panic(err)
			}
		}(evictedApprovers)
	}
}

const (
	APPROVERS_CACHE_SIZE = 50000
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region database /////////////////////////////////////////////////////////////////////////////////////////////////////

var approversDatabase database.Database

func configureApproversDatabase(plugin *node.Plugin) {
	if db, err := database.Get("approvers"); err != nil {
		panic(err)
	} else {
		approversDatabase = db
	}
}

func storeApproversInDatabase(approvers *approvers.Approvers) errors.IdentifiableError {
	if approvers.GetModified() {
		if err := approversDatabase.Set(approvers.GetHash().CastToBytes(), approvers.Marshal()); err != nil {
			return ErrDatabaseError.Derive(err, "failed to store approvers")
		}

		approvers.SetModified(false)
	}

	return nil
}

func getApproversFromDatabase(transactionHash ternary.Trytes) (*approvers.Approvers, errors.IdentifiableError) {
	approversData, err := approversDatabase.Get(transactionHash.CastToBytes())
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}

		return nil, ErrDatabaseError.Derive(err, "failed to retrieve approvers")
	}

	var result approvers.Approvers
	if err = result.Unmarshal(approversData); err != nil {
		panic(err)
	}

	return &result, nil
}

func databaseContainsApprovers(transactionHash ternary.Trytes) (bool, errors.IdentifiableError) {
	if contains, err := approversDatabase.Contains(transactionHash.CastToBytes()); err != nil {
		return false, ErrDatabaseError.Derive(err, "failed to check if the approvers exists")
	} else {
		return contains, nil
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
