package meta_transaction

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/curl"
	"github.com/iotaledger/goshimmer/packages/ternary"
)

type MetaTransaction struct {
	hash            *ternary.Trinary
	weightMagnitude int

	shardMarker           *ternary.Trinary
	trunkTransactionHash  *ternary.Trinary
	branchTransactionHash *ternary.Trinary
	head                  *bool
	tail                  *bool
	transactionType       *ternary.Trinary
	data                  ternary.Trits
	modified              bool

	hasherMutex                sync.RWMutex
	hashMutex                  sync.RWMutex
	shardMarkerMutex           sync.RWMutex
	trunkTransactionHashMutex  sync.RWMutex
	branchTransactionHashMutex sync.RWMutex
	headMutex                  sync.RWMutex
	tailMutex                  sync.RWMutex
	transactionTypeMutex       sync.RWMutex
	dataMutex                  sync.RWMutex
	bytesMutex                 sync.RWMutex
	modifiedMutex              sync.RWMutex

	trits ternary.Trits
	bytes []byte
}

func New() *MetaTransaction {
	return FromTrits(make(ternary.Trits, MARSHALLED_TOTAL_SIZE))
}

func FromTrits(trits ternary.Trits) *MetaTransaction {
	return &MetaTransaction{
		trits: trits,
	}
}

func FromBytes(bytes []byte) (result *MetaTransaction) {
	result = FromTrits(ternary.BytesToTrits(bytes)[:MARSHALLED_TOTAL_SIZE])
	result.bytes = bytes

	return
}

func (this *MetaTransaction) BlockHasher() {
	this.hasherMutex.RLock()
}

func (this *MetaTransaction) UnblockHasher() {
	this.hasherMutex.RUnlock()
}

func (this *MetaTransaction) ReHash() {
	this.hashMutex.Lock()
	defer this.hashMutex.Unlock()
	this.hash = nil

	this.bytesMutex.Lock()
	defer this.bytesMutex.Unlock()
	this.bytes = nil
}

// retrieves the hash of the transaction
func (this *MetaTransaction) GetHash() (result ternary.Trinary) {
	this.hashMutex.RLock()
	if this.hash == nil {
		this.hashMutex.RUnlock()
		this.hashMutex.Lock()
		defer this.hashMutex.Unlock()
		if this.hash == nil {
			this.hasherMutex.Lock()
			this.parseHashRelatedDetails()
			this.hasherMutex.Unlock()
		}
	} else {
		defer this.hashMutex.RUnlock()
	}

	result = *this.hash

	return
}

// retrieves weight magnitude of the transaction (amount of pow invested)
func (this *MetaTransaction) GetWeightMagnitude() (result int) {
	this.hashMutex.RLock()
	if this.hash == nil {
		this.hashMutex.RUnlock()
		this.hashMutex.Lock()
		defer this.hashMutex.Unlock()
		if this.hash == nil {
			this.hasherMutex.Lock()
			this.parseHashRelatedDetails()
			this.hasherMutex.Unlock()
		}
	} else {
		defer this.hashMutex.RUnlock()
	}

	result = this.weightMagnitude

	return
}

// hashes the transaction using curl (without locking - internal usage)
func (this *MetaTransaction) parseHashRelatedDetails() {
	hashTrits := <-curl.CURLP81.Hash(this.trits)
	hashTrinary := hashTrits.ToTrinary()

	this.hash = &hashTrinary
	this.weightMagnitude = hashTrits.TrailingZeroes()
}

// getter for the shard marker (supports concurrency)
func (this *MetaTransaction) GetShardMarker() (result ternary.Trinary) {
	this.shardMarkerMutex.RLock()
	if this.shardMarker == nil {
		this.shardMarkerMutex.RUnlock()
		this.shardMarkerMutex.Lock()
		defer this.shardMarkerMutex.Unlock()
		if this.shardMarker == nil {
			shardMarker := this.trits[SHARD_MARKER_OFFSET:SHARD_MARKER_END].ToTrinary()

			this.shardMarker = &shardMarker
		}
	} else {
		defer this.shardMarkerMutex.RUnlock()
	}

	result = *this.shardMarker

	return
}

// setter for the shard marker (supports concurrency)
func (this *MetaTransaction) SetShardMarker(shardMarker ternary.Trinary) bool {
	this.shardMarkerMutex.RLock()
	if this.shardMarker == nil || *this.shardMarker != shardMarker {
		this.shardMarkerMutex.RUnlock()
		this.shardMarkerMutex.Lock()
		defer this.shardMarkerMutex.Unlock()
		if this.shardMarker == nil || *this.shardMarker != shardMarker {
			this.shardMarker = &shardMarker

			this.hasherMutex.RLock()
			copy(this.trits[SHARD_MARKER_OFFSET:SHARD_MARKER_END], shardMarker.ToTrits()[:SHARD_MARKER_SIZE])
			this.hasherMutex.RUnlock()

			this.SetModified(true)
			this.ReHash()

			return true
		}
	} else {
		this.shardMarkerMutex.RUnlock()
	}

	return false
}

// getter for the bundleHash (supports concurrency)
func (this *MetaTransaction) GetTrunkTransactionHash() (result ternary.Trinary) {
	this.trunkTransactionHashMutex.RLock()
	if this.trunkTransactionHash == nil {
		this.trunkTransactionHashMutex.RUnlock()
		this.trunkTransactionHashMutex.Lock()
		defer this.trunkTransactionHashMutex.Unlock()
		if this.trunkTransactionHash == nil {
			trunkTransactionHash := this.trits[TRUNK_TRANSACTION_HASH_OFFSET:TRUNK_TRANSACTION_HASH_END].ToTrinary()

			this.trunkTransactionHash = &trunkTransactionHash
		}
	} else {
		defer this.trunkTransactionHashMutex.RUnlock()
	}

	result = *this.trunkTransactionHash

	return
}

// setter for the trunkTransactionHash (supports concurrency)
func (this *MetaTransaction) SetTrunkTransactionHash(trunkTransactionHash ternary.Trinary) bool {
	this.trunkTransactionHashMutex.RLock()
	if this.trunkTransactionHash == nil || *this.trunkTransactionHash != trunkTransactionHash {
		this.trunkTransactionHashMutex.RUnlock()
		this.trunkTransactionHashMutex.Lock()
		defer this.trunkTransactionHashMutex.Unlock()
		if this.trunkTransactionHash == nil || *this.trunkTransactionHash != trunkTransactionHash {
			this.trunkTransactionHash = &trunkTransactionHash

			this.hasherMutex.RLock()
			copy(this.trits[TRUNK_TRANSACTION_HASH_OFFSET:TRUNK_TRANSACTION_HASH_END], trunkTransactionHash.ToTrits()[:TRUNK_TRANSACTION_HASH_SIZE])
			this.hasherMutex.RUnlock()

			this.SetModified(true)
			this.ReHash()

			return true
		}
	} else {
		this.trunkTransactionHashMutex.RUnlock()
	}

	return false
}

// getter for the bundleHash (supports concurrency)
func (this *MetaTransaction) GetBranchTransactionHash() (result ternary.Trinary) {
	this.branchTransactionHashMutex.RLock()
	if this.branchTransactionHash == nil {
		this.branchTransactionHashMutex.RUnlock()
		this.branchTransactionHashMutex.Lock()
		defer this.branchTransactionHashMutex.Unlock()
		if this.branchTransactionHash == nil {
			branchTransactionHash := this.trits[BRANCH_TRANSACTION_HASH_OFFSET:BRANCH_TRANSACTION_HASH_END].ToTrinary()

			this.branchTransactionHash = &branchTransactionHash
		}
	} else {
		defer this.branchTransactionHashMutex.RUnlock()
	}

	result = *this.branchTransactionHash

	return
}

// setter for the trunkTransactionHash (supports concurrency)
func (this *MetaTransaction) SetBranchTransactionHash(branchTransactionHash ternary.Trinary) bool {
	this.branchTransactionHashMutex.RLock()
	if this.branchTransactionHash == nil || *this.branchTransactionHash != branchTransactionHash {
		this.branchTransactionHashMutex.RUnlock()
		this.branchTransactionHashMutex.Lock()
		defer this.branchTransactionHashMutex.Unlock()
		if this.branchTransactionHash == nil || *this.branchTransactionHash != branchTransactionHash {
			this.branchTransactionHash = &branchTransactionHash

			this.hasherMutex.RLock()
			copy(this.trits[BRANCH_TRANSACTION_HASH_OFFSET:BRANCH_TRANSACTION_HASH_END], branchTransactionHash.ToTrits()[:BRANCH_TRANSACTION_HASH_SIZE])
			this.hasherMutex.RUnlock()

			this.SetModified(true)
			this.ReHash()

			return true
		}
	} else {
		this.branchTransactionHashMutex.RUnlock()
	}

	return false
}

// getter for the head flag (supports concurrency)
func (this *MetaTransaction) GetHead() (result bool) {
	this.headMutex.RLock()
	if this.head == nil {
		this.headMutex.RUnlock()
		this.headMutex.Lock()
		defer this.headMutex.Unlock()
		if this.head == nil {
			head := this.trits[HEAD_OFFSET] == 1

			this.head = &head
		}
	} else {
		defer this.headMutex.RUnlock()
	}

	result = *this.head

	return
}

// setter for the head flag (supports concurrency)
func (this *MetaTransaction) SetHead(head bool) bool {
	this.headMutex.RLock()
	if this.head == nil || *this.head != head {
		this.headMutex.RUnlock()
		this.headMutex.Lock()
		defer this.headMutex.Unlock()
		if this.head == nil || *this.head != head {
			this.head = &head

			this.hasherMutex.RLock()
			if head {
				this.trits[HEAD_OFFSET] = 1
			} else {
				this.trits[HEAD_OFFSET] = 0
			}
			this.hasherMutex.RUnlock()

			this.SetModified(true)
			this.ReHash()

			return true
		}
	} else {
		this.headMutex.RUnlock()
	}

	return false
}

// getter for the tail flag (supports concurrency)
func (this *MetaTransaction) GetTail() (result bool) {
	this.tailMutex.RLock()
	if this.tail == nil {
		this.tailMutex.RUnlock()
		this.tailMutex.Lock()
		defer this.tailMutex.Unlock()
		if this.tail == nil {
			tail := this.trits[TAIL_OFFSET] == 1

			this.tail = &tail
		}
	} else {
		defer this.tailMutex.RUnlock()
	}

	result = *this.tail

	return
}

// setter for the tail flag (supports concurrency)
func (this *MetaTransaction) SetTail(tail bool) bool {
	this.tailMutex.RLock()
	if this.tail == nil || *this.tail != tail {
		this.tailMutex.RUnlock()
		this.tailMutex.Lock()
		defer this.tailMutex.Unlock()
		if this.tail == nil || *this.tail != tail {
			this.tail = &tail

			this.hasherMutex.RLock()
			if tail {
				this.trits[TAIL_OFFSET] = 1
			} else {
				this.trits[TAIL_OFFSET] = 0
			}
			this.hasherMutex.RUnlock()

			this.SetModified(true)
			this.ReHash()

			return true
		}
	} else {
		this.tailMutex.RUnlock()
	}

	return false
}

// getter for the transaction type (supports concurrency)
func (this *MetaTransaction) GetTransactionType() (result ternary.Trinary) {
	this.transactionTypeMutex.RLock()
	if this.transactionType == nil {
		this.transactionTypeMutex.RUnlock()
		this.transactionTypeMutex.Lock()
		defer this.transactionTypeMutex.Unlock()
		if this.transactionType == nil {
			transactionType := this.trits[TRANSACTION_TYPE_OFFSET:TRANSACTION_TYPE_END].ToTrinary()

			this.transactionType = &transactionType
		}
	} else {
		defer this.transactionTypeMutex.RUnlock()
	}

	result = *this.transactionType

	return
}

// setter for the transaction type (supports concurrency)
func (this *MetaTransaction) SetTransactionType(transactionType ternary.Trinary) bool {
	this.transactionTypeMutex.RLock()
	if this.transactionType == nil || *this.transactionType != transactionType {
		this.transactionTypeMutex.RUnlock()
		this.transactionTypeMutex.Lock()
		defer this.transactionTypeMutex.Unlock()
		if this.transactionType == nil || *this.transactionType != transactionType {
			this.transactionType = &transactionType

			this.hasherMutex.RLock()
			copy(this.trits[TRANSACTION_TYPE_OFFSET:TRANSACTION_TYPE_END], transactionType.ToTrits()[:TRANSACTION_TYPE_SIZE])
			this.hasherMutex.RUnlock()

			this.SetModified(true)
			this.ReHash()

			return true
		}
	} else {
		this.transactionTypeMutex.RUnlock()
	}

	return false
}

// getter for the data slice (supports concurrency)
func (this *MetaTransaction) GetData() (result ternary.Trits) {
	this.dataMutex.RLock()
	if this.data == nil {
		this.dataMutex.RUnlock()
		this.dataMutex.Lock()
		defer this.dataMutex.Unlock()
		if this.data == nil {
			this.data = this.trits[DATA_OFFSET:DATA_END]
		}
	} else {
		defer this.dataMutex.RUnlock()
	}

	result = this.data

	return
}

func (this *MetaTransaction) GetTrits() (result ternary.Trits) {
	result = make(ternary.Trits, len(this.trits))

	this.hasherMutex.Lock()
	copy(result, this.trits)
	this.hasherMutex.Unlock()

	return
}

func (this *MetaTransaction) GetBytes() (result []byte) {
	this.bytesMutex.RLock()
	if this.bytes == nil {
		this.bytesMutex.RUnlock()
		this.bytesMutex.Lock()
		defer this.bytesMutex.Unlock()

		this.hasherMutex.Lock()
		this.bytes = this.trits.ToBytes()
		this.hasherMutex.Unlock()
	} else {
		this.bytesMutex.RUnlock()
	}

	result = make([]byte, len(this.bytes))
	copy(result, this.bytes)

	return
}

// returns true if the transaction contains unsaved changes (supports concurrency)
func (this *MetaTransaction) GetModified() bool {
	this.modifiedMutex.RLock()
	defer this.modifiedMutex.RUnlock()

	return this.modified
}

// sets the modified flag which controls if a transaction is going to be saved (supports concurrency)
func (this *MetaTransaction) SetModified(modified bool) {
	this.modifiedMutex.Lock()
	defer this.modifiedMutex.Unlock()

	this.modified = modified
}
