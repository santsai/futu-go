package futu

import (
	"time"
)

// Options are futu client options.
type clientOptions struct {
	openDAddr  string
	clientId   string
	privateKey []byte
	recvNotify bool
	numWorkers int
	numBuffers int
	timeout    time.Duration
}

type ClientOption func(o *clientOptions)

// NewOptions creates options with defaults.
func newClientOptions(opts []ClientOption) clientOptions {
	opt := &clientOptions{
		openDAddr:  ":11111",
		clientId:   "futu-go",
		recvNotify: true,
		numBuffers: 100,
		numWorkers: 2,
		timeout:    5 * time.Second,
	}

	for _, o := range opts {
		o(opt)
	}

	return *opt
}

// WithID sets client id.
func WithClientID(id string) ClientOption {
	return func(o *clientOptions) {
		o.clientId = id
	}
}

// WithAddr sets futu OpenD address.
func WithOpenDAddr(addr string) ClientOption {
	return func(o *clientOptions) {
		o.openDAddr = addr
	}
}

// WithPrivateKey sets private key.
func WithPrivateKey(privateKey []byte) ClientOption {
	return func(o *clientOptions) {
		o.privateKey = privateKey
	}
}

// WithRecvNotify sets whether to receive notifications.
func WithRecvNotify(recvNotify bool) ClientOption {
	return func(o *clientOptions) {
		o.recvNotify = recvNotify
	}
}

// WithNumBuffer sets response channel size.
func WithNumBuffers(size int) ClientOption {
	return func(o *clientOptions) {
		o.numBuffers = size
	}
}

func WithTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

func WithNumWorkers(n int) ClientOption {
	return func(o *clientOptions) {
		o.numWorkers = n
	}
}
