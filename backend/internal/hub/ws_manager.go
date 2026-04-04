package hub

import (
	"context"
	"log/slog"
	"sync"

	fiberws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

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
		write:   make(chan WriteToWs, 256),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Write enqueues a message to be sent to a specific connected client.
func (m *Manager) Write(consumerID uuid.UUID, payload []byte) {
	select {
	case m.write <- WriteToWs{Payload: payload, ConsumerID: consumerID}:
	default:
		slog.Warn("Manager.Write: write channel full", "consumer_id", consumerID)
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

// ServeWS is called from the Fiber WebSocket handler with an already-upgraded conn.
// The client is a PRODUCER: it sends telemetry frames that are forwarded to the
// aggregator via the read channel.
// NOTE: this function BLOCKS until the connection is closed — required by gofiber/websocket.
func (m *Manager) ServeWS(conn *fiberws.Conn, id uuid.UUID) {
	slog.Info("WebSocket connection established (producer)", "client_id", id.String())

	client := NewClient(id, conn, m)
	m.addClient(client)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		client.readMessages()
	}()
	go func() {
		defer wg.Done()
		client.writeMessages()
	}()
	go func() {
		for message := range client.inbound {
			select {
			case m.read <- ReadFromWs{Payload: message, ProducerID: client.id}:
			default:
				slog.Warn("hub read channel full, dropping frame", "client_id", id)
			}
		}
	}()

	wg.Wait() // block until both read+write goroutines exit
}

// ServeLive registers a CONSUMER-only dashboard client that receives broadcast
// LocoUpdate frames but never writes to the read channel.
// NOTE: this function BLOCKS until the connection is closed — required by gofiber/websocket.
func (m *Manager) ServeLive(conn *fiberws.Conn, id uuid.UUID) {
	slog.Info("WebSocket connection established (live dashboard)", "client_id", id.String())

	client := NewClient(id, conn, m)
	m.addClient(client)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		// drain inbound so the connection stays alive for pings; discard frames
		client.readMessages()
	}()
	go func() {
		defer wg.Done()
		client.writeMessages()
	}()

	wg.Wait()
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
					slog.Warn("Client not found for message", "client_id", message.ConsumerID)
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
