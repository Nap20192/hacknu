package hub

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var AuthMessage []byte = []byte("asdjfklasdjfkadl1233r21kjjqkrwerfsdfjlkjdslkaf")

type ClientList map[*Client]bool

type Client struct {
	conn          *websocket.Conn
	manager       *Manager
	inbound       chan []byte
	outbound      chan []byte
	closeOnce     sync.Once
	id            uuid.UUID
	authenticated bool
	authDeadline  time.Time
	authMutex     sync.RWMutex
}

var (
	pongWait     = 60 * time.Second
	pingInterval = (pongWait * 9) / 10
	authTimeout  = 5 * time.Second
)

func NewClient(id uuid.UUID, conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		id:            id,
		conn:          conn,
		manager:       manager,
		inbound:       make(chan []byte),
		outbound:      make(chan []byte),
		authenticated: false,
		authDeadline:  time.Now().Add(authTimeout),
	}
}

// IsAuthenticated checks if the client is authenticated
func (c *Client) IsAuthenticated() bool {
	c.authMutex.RLock()
	defer c.authMutex.RUnlock()
	return c.authenticated
}

func (c *Client) SetAuthenticated(auth bool) {
	c.authMutex.Lock()
	defer c.authMutex.Unlock()
	c.authenticated = auth
}

func (c *Client) readMessages() {
	defer func() {
		c.close()
	}()

	c.conn.SetReadLimit(512)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}

	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			slog.Info("Client disconnected", "client_id", c.id)
			return
		}

		c.inbound <- message
	}
}

func (c *Client) writeMessages() {
	fmt.Println("Starting writeMessages for client:", c.id)
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_ = c.conn.WriteMessage(websocket.TextMessage, []byte("connected"))

	ticker := time.NewTicker(pingInterval)
	authTimer := time.NewTimer(authTimeout)

	defer func() {
		ticker.Stop()
		if !authTimer.Stop() {
			select {
			case <-authTimer.C:
			default:
			}
		}
		c.close()
	}()

	for {
		select {
		case <-authTimer.C:
			slog.Warn("Authentication timer expired", "client_id", c.id)
			if !c.IsAuthenticated() {
				slog.Info("Authentication timeout, closing connection", "client_id", c.id)
				c.conn.WriteMessage(websocket.CloseMessage, []byte("authentication timeout"))
				c.conn.Close()
				return
			}

		case event, ok := <-c.outbound:

			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte("conn shutting down"))
				return
			}

			if string(event) == string(AuthMessage) {
				c.SetAuthenticated(true)
				if !authTimer.Stop() {
					select {
					case <-authTimer.C:
					default:
					}
				}
				if err := c.conn.WriteMessage(websocket.TextMessage, []byte("authenticated")); err != nil {
					return
				}
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, event); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(pingInterval))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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
