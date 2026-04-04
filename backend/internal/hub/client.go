package hub

import (
	"log/slog"
	"sync"
	"time"

	fiberws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type Client struct {
	conn      *fiberws.Conn
	manager   *Manager
	inbound   chan []byte
	outbound  chan []byte
	closeOnce sync.Once
	id        uuid.UUID
}

var (
	pongWait     = 60 * time.Second
	pingInterval = (pongWait * 9) / 10
)

func NewClient(id uuid.UUID, conn *fiberws.Conn, manager *Manager) *Client {
	return &Client{
		id:       id,
		conn:     conn,
		manager:  manager,
		inbound:  make(chan []byte, 256),
		outbound: make(chan []byte, 256),
	}
}

func (c *Client) readMessages() {
	defer c.close()

	c.conn.SetReadLimit(512 * 1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			slog.Info("client disconnected", "client_id", c.id)
			return
		}
		c.inbound <- message
	}
}

func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case event, ok := <-c.outbound:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(fiberws.CloseMessage, []byte("conn shutting down"))
				return
			}
			if err := c.conn.WriteMessage(fiberws.TextMessage, event); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if err := c.conn.WriteMessage(fiberws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		close(c.outbound)
		close(c.inbound)
		_ = c.conn.Close()
		c.manager.removeClient(c.id)
	})
}
