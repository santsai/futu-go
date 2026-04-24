// Package tcp implements the framework of TCP connection.
package futu

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// conn is for frame reading
type conn struct {
	net.Conn

	reader   protocolReader
	wgReader sync.WaitGroup
	closing  bool
}

// newConn dials tcp Conn and starts readLoop
func newConn(address string, reader protocolReader) (*conn, error) {
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}

	conn := &conn{
		Conn:   c,
		reader: reader,
	}

	conn.wgReader.Add(1)
	go conn.readLoop()

	return conn, nil
}

func (c *conn) readLoop() {
	defer c.wgReader.Done()

	for !c.closing {
		// assume err are fatal
		if err := c.reader.readFrom(c); err != nil {
			break
		}
	}

	// close the conn, make future write err out
	c.Conn.Close()
}

// Close closes conn and stops readLoop
func (c *conn) Close() error {

	c.closing = true

	// make read return asap
	c.Conn.SetReadDeadline(time.Now())

	err := c.Conn.Close()
	c.wgReader.Wait()
	return err
}
