package client

import (
	"time"
)

const (
	defaultAddr        = ":11111"
	defaultID          = "futu-go"
	defaultResChanSize = 100
	defaultTimeout     = 5 * time.Second
)

// Options are futu client options.
type Options struct {
	Addr        string
	ID          string
	PrivateKey  []byte
	PublicKey   []byte
	RecvNotify  bool
	ResChanSize int
	Timeout     time.Duration
}

// NewOptions creates options with defaults.
func NewOptions(opts ...Option) Options {
	var options = Options{
		Addr:        defaultAddr,
		ID:          defaultID,
		RecvNotify:  true,
		ResChanSize: defaultResChanSize,
		Timeout:     defaultTimeout,
	}

	for _, opt := range opts {
		opt(&options)
	}

	return options
}

// Option is for setting options.
type Option func(*Options)

// WithID sets client id.
func WithID(id string) Option {
	return func(o *Options) {
		o.ID = id
	}
}

// WithAddr sets futu OpenD address.
func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithPrivateKey sets private key.
func WithPrivateKey(privateKey []byte) Option {
	return func(o *Options) {
		o.PrivateKey = privateKey
	}
}

// WithPublicKey sets public key.
func WithPublicKey(publicKey []byte) Option {
	return func(o *Options) {
		o.PublicKey = publicKey
	}
}

// WithRecvNotify sets whether to receive notifications.
func WithRecvNotify(recvNotify bool) Option {
	return func(o *Options) {
		o.RecvNotify = recvNotify
	}
}

// WithResChanSize sets response channel size.
func WithResChanSize(size int) Option {
	return func(o *Options) {
		o.ResChanSize = size
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.Timeout = d
	}
}
