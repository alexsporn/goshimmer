package peerregister

import (
	"bytes"
	"sync"

	"github.com/iotaledger/goshimmer/packages/accountability"
	"github.com/iotaledger/goshimmer/packages/events"
	"github.com/iotaledger/goshimmer/plugins/autopeering/types/peer"
	"github.com/iotaledger/goshimmer/plugins/autopeering/types/peerlist"
	"github.com/iotaledger/goshimmer/plugins/autopeering/types/request"
)

type PeerRegister struct {
	Peers  map[string]*peer.Peer
	Events peerRegisterEvents
	lock   sync.RWMutex
}

func New() *PeerRegister {
	return &PeerRegister{
		Peers: make(map[string]*peer.Peer),
		Events: peerRegisterEvents{
			Add:    events.NewEvent(peerCaller),
			Update: events.NewEvent(peerCaller),
			Remove: events.NewEvent(peerCaller),
		},
	}
}

// returns true if a new entry was added
func (this *PeerRegister) AddOrUpdate(peer *peer.Peer, lock ...bool) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	// if len(lock) == 0 || lock[0] {
	// 	defer this.Lock()()
	// }

	if peer.Identity == nil || bytes.Equal(peer.Identity.Identifier, accountability.OwnId().Identifier) {
		return false
	}

	if existingPeer, exists := this.Peers[peer.Identity.StringIdentifier]; exists {
		existingPeer.PeerMutex.Lock()
		existingPeer.Address = peer.Address
		existingPeer.GossipPort = peer.GossipPort
		existingPeer.PeeringPort = peer.PeeringPort
		existingPeer.Salt = peer.Salt

		// also update the public key if not yet present
		if existingPeer.Identity.PublicKey == nil {
			existingPeer.Identity.PublicKey = peer.Identity.PublicKey
		}

		this.Events.Update.Trigger(existingPeer)
		existingPeer.PeerMutex.Unlock()
		return false
	} else {
		this.Peers[peer.Identity.StringIdentifier] = peer

		this.Events.Add.Trigger(peer)

		return true
	}
}

func (this *PeerRegister) Lock() {
	this.lock.Lock()
}

func (this *PeerRegister) Unlock() {
	this.lock.Unlock()
}

func (this *PeerRegister) RLock() {
	this.lock.RLock()
}

func (this *PeerRegister) RUnlock() {
	this.lock.RUnlock()
}

func (this *PeerRegister) Remove(key string, lock ...bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if peerEntry, exists := this.Peers[key]; exists {
		if len(lock) == 0 || lock[0] {
			if peerEntry, exists := this.Peers[key]; exists {
				delete(this.Peers, key)

				this.Events.Remove.Trigger(peerEntry)
			}
		} else {
			delete(this.Peers, key)

			this.Events.Remove.Trigger(peerEntry)
		}
	}
}

func (this *PeerRegister) Contains(key string) bool {
	if _, exists := this.Peers[key]; exists {
		return true
	} else {
		return false
	}
}

func (this *PeerRegister) Filter(filterFn func(this *PeerRegister, req *request.Request) *PeerRegister, req *request.Request) *PeerRegister {
	return filterFn(this, req)
}

func (this *PeerRegister) List() peerlist.PeerList {
	peerList := make(peerlist.PeerList, len(this.Peers))

	counter := 0
	for _, currentPeer := range this.Peers {
		peerList[counter] = currentPeer
		counter++
	}

	return peerList
}
