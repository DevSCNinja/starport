package network

import (
	"context"

	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
	"github.com/tendermint/starport/starport/pkg/cosmosclient"
	"github.com/tendermint/starport/starport/pkg/events"
)

// Network is network builder.
type Network struct {
	ev      events.Bus
	cosmos  cosmosclient.Client
	account cosmosaccount.Account
}

type Chain interface {
	ID() (string, error)
	Name() string
	SourceURL() string
	SourceHash() string
	GentxPath() (string, error)
	GenesisPath() (string, error)
	Peer(ctx context.Context, addr string) (string, error)
}

type Option func(*Network)

// CollectEvents collects events from the network builder.
func CollectEvents(ev events.Bus) Option {
	return func(b *Network) {
		b.ev = ev
	}
}

// New creates a Builder.
func New(cosmos cosmosclient.Client, account cosmosaccount.Account, options ...Option) (Network, error) {
	n := Network{
		cosmos:  cosmos,
		account: account,
	}
	for _, opt := range options {
		opt(&n)
	}
	return n, nil
}
