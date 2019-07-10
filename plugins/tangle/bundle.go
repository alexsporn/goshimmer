package tangle

import (
	"github.com/dgraph-io/badger"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/packages/datastructure"
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/model/bundle"
	"github.com/iotaledger/goshimmer/packages/node"
	"github.com/iotaledger/goshimmer/packages/unsafeconvert"
	"github.com/iotaledger/iota.go/trinary"
)

// region global public api ////////////////////////////////////////////////////////////////////////////////////////////

// GetBundle retrieves bundle from the database.
func GetBundle(headerTransactionHash trinary.Trytes, computeIfAbsent ...func(trinary.Trytes) (*bundle.Bundle, errors.IdentifiableError)) (result *bundle.Bundle, err errors.IdentifiableError) {
	if cacheResult := bundleCache.ComputeIfAbsent(headerTransactionHash, func() interface{} {
		if dbBundle, dbErr := getBundleFromDatabase(headerTransactionHash); dbErr != nil {
			err = dbErr

			return nil
		} else if dbBundle != nil {
			return dbBundle
		} else {
			if len(computeIfAbsent) >= 1 {
				if computedBundle, computedErr := computeIfAbsent[0](headerTransactionHash); computedErr != nil {
					err = computedErr
				} else {
					return computedBundle
				}
			}

			return nil
		}
	}); cacheResult != nil && cacheResult.(*bundle.Bundle) != nil {
		result = cacheResult.(*bundle.Bundle)
	}

	return
}

func ContainsBundle(headerTransactionHash trinary.Trytes) (result bool, err errors.IdentifiableError) {
	if bundleCache.Contains(headerTransactionHash) {
		result = true
	} else {
		result, err = databaseContainsBundle(headerTransactionHash)
	}

	return
}

func StoreBundle(bundle *bundle.Bundle) {
	bundleCache.Set(bundle.GetHash(), bundle)
}

// region lru cache ////////////////////////////////////////////////////////////////////////////////////////////////////

var bundleCache = datastructure.NewLRUCache(BUNDLE_CACHE_SIZE, &datastructure.LRUCacheOptions{
	EvictionCallback: onEvictBundle,
})

func onEvictBundle(_ interface{}, value interface{}) {
	if evictedBundle := value.(*bundle.Bundle); evictedBundle.GetModified() {
		go func(evictedBundle *bundle.Bundle) {
			if err := storeBundleInDatabase(evictedBundle); err != nil {
				panic(err)
			}
		}(evictedBundle)
	}
}

const (
	BUNDLE_CACHE_SIZE = 50000
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region database /////////////////////////////////////////////////////////////////////////////////////////////////////

var bundleDatabase database.Database

func configureBundleDatabase(plugin *node.Plugin) {
	if db, err := database.Get("bundle"); err != nil {
		panic(err)
	} else {
		bundleDatabase = db
	}
}

func storeBundleInDatabase(bundle *bundle.Bundle) errors.IdentifiableError {
	if bundle.GetModified() {
		if err := bundleDatabase.Set(unsafeconvert.StringToBytes(bundle.GetHash()), bundle.Marshal()); err != nil {
			return ErrDatabaseError.Derive(err, "failed to store bundle")
		}

		bundle.SetModified(false)
	}

	return nil
}

func getBundleFromDatabase(transactionHash trinary.Trytes) (*bundle.Bundle, errors.IdentifiableError) {
	bundleData, err := bundleDatabase.Get(unsafeconvert.StringToBytes(transactionHash))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, nil
		}

		return nil, ErrDatabaseError.Derive(err, "failed to retrieve bundle")
	}

	var result bundle.Bundle
	if err = result.Unmarshal(bundleData); err != nil {
		panic(err)
	}

	return &result, nil
}

func databaseContainsBundle(transactionHash trinary.Trytes) (bool, errors.IdentifiableError) {
	if contains, err := bundleDatabase.Contains(unsafeconvert.StringToBytes(transactionHash)); err != nil {
		return false, ErrDatabaseError.Derive(err, "failed to check if the bundle exists")
	} else {
		return contains, nil
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
