package hub

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Manager struct {
	ctx          context.Context
	clients      map[uuid.UUID]*Client
	read         chan ReadFromWs
	write        chan WriteToWs
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	shutdownOnce sync.Once
	mu           sync.Mutex
}

func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		clients: make(map[uuid.UUID]*Client),
		wg:      sync.WaitGroup{},
		read:    make(chan ReadFromWs),
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (m *Manager) addClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c.id] = c
}

func (m *Manager) removeClient(c uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, c)
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("WebSocket upgrade failed", "client_id", id.String(), "error", err)
		return
	}

	slog.Info("WebSocket connection established", "client_id", id.String())

	client := NewClient(id, conn, m)
	m.addClient(client)
	m.wg.Add(3)
	go func() {
		defer m.wg.Done()
		client.readMessages()
	}()
	go func() {
		defer m.wg.Done()
		client.writeMessages()
	}()
	go func() {
		defer m.wg.Done()
		for message := range client.inbound {
			m.read <- ReadFromWs{
				Payload:    message,
				ProducerID: client.id,
			}
		}
	}()
}

type WriteToWs struct {
	Payload    []byte
	ConsumerID uuid.UUID
}

type ReadFromWs struct {
	Payload    []byte
	ProducerID uuid.UUID
}

func (m *Manager) StartWrite(ctx context.Context) {
	m.mu.Lock()
	if ctx == nil {
		ctx = context.Background()
	}

	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.ctx.Done():
				return
			case message, ok := <-m.write:
				if !ok {
					return
				}
				m.mu.Lock()
				client, exists := m.clients[message.ConsumerID]
				m.mu.Unlock()

				if exists {
					select {
					case client.outbound <- message.Payload:
					default:
						slog.Warn("Client outbound channel full", "client_id", message.ConsumerID)
					}
				} else {
					slog.Warn("Client not found for message", "client_id", message.ConsumerID, "message", string(message.Payload))
				}
			}
		}
	}()
}

func (m *Manager) ReadChannel() <-chan ReadFromWs {
	return m.read
}

func (m *Manager) WithWriteChannel(write chan WriteToWs) {
	m.write = write
}

// Broadcast sends payload to every currently connected client.
// Non-blocking: clients with full buffers are skipped with a warning.
func (m *Manager) Broadcast(payload []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		select {
		case client.outbound <- payload:
		default:
			slog.Warn("Broadcast: client buffer full, skipping", "client_id", client.id)
		}
	}
}

func (m *Manager) Shutdown() {
	m.shutdownOnce.Do(func() {
		if m.cancel != nil {
			m.cancel()
		}

		for _, client := range m.clients {
			client.close()
		}

		m.wg.Wait()

		close(m.read)
	})
}
