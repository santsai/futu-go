package futu

import (
	"time"
)

const (
	defaultOpenDAddr    = ":11111"
	defaultID           = "futu-go"
	defaultRespChanSize = 100
	defaultTimeout      = 5 * time.Second
)

// Options are futu client options.
type ClientOptions struct {
	openDAddr    string
	clientId     string
	privateKey   []byte
	publicKey    []byte
	recvNotify   bool
	respChanSize int
	timeout      time.Duration
}

// NewOptions creates options with defaults.
func NewClientOptions() *ClientOptions {
	return &ClientOptions{
		openDAddr:    defaultOpenDAddr,
		clientId:     defaultID,
		recvNotify:   true,
		respChanSize: defaultRespChanSize,
		timeout:      defaultTimeout,
	}
}

// WithID sets client id.
func (o *ClientOptions) WithClientID(id string) *ClientOptions {
	o.clientId = id
	return o
}

// WithAddr sets futu OpenD address.
func (o *ClientOptions) WithOpenDAddr(addr string) *ClientOptions {
	o.openDAddr = addr
	return o
}

// WithPrivateKey sets private key.
func (o *ClientOptions) WithPrivateKey(privateKey []byte) *ClientOptions {
	o.privateKey = privateKey
	return o
}

// WithPublicKey sets public key.
func (o *ClientOptions) WithPublicKey(publicKey []byte) *ClientOptions {
	o.publicKey = publicKey
	return o
}

// WithRecvNotify sets whether to receive notifications.
func (o *ClientOptions) WithRecvNotify(recvNotify bool) *ClientOptions {
	o.recvNotify = recvNotify
	return o
}

// WithRespChanSize sets response channel size.
func (o *ClientOptions) WithRespChanSize(size int) *ClientOptions {
	o.respChanSize = size
	return o
}

func (o *ClientOptions) WithTimeout(d time.Duration) *ClientOptions {
	o.timeout = d
	return o
}
