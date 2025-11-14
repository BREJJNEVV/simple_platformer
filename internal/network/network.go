package network

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const (
	defaultSendBufferSize = 8
	defaultDialTimeout    = 5 * time.Second
	defaultListenAddress  = ":4000"
	defaultDialAddress    = "127.0.0.1:4000"
)

// PlayerState описывает состояние игрока, которое отправляется по сети.
type PlayerState struct {
	X, Y        float64
	VelocityX   float64
	VelocityY   float64
	OnGround    bool
	FacingRight bool
}

// BulletState описывает состояние пули, которое отправляется по сети.
type BulletState struct {
	X         float64
	Y         float64
	VelocityX float64
}

// StateMessage содержит состояние игрока и его пуль.
type StateMessage struct {
	Player  PlayerState
	Bullets []BulletState
}

// Manager управляет сетевым подключением.
type Manager struct {
	mu       sync.RWMutex
	peer     *peer
	listener net.Listener

	closeOnce sync.Once
	closed    chan struct{}

	errMu sync.Mutex
	err   error
}

func newManager(initialPeer *peer) *Manager {
	return &Manager{
		peer:   initialPeer,
		closed: make(chan struct{}),
	}
}

// Host запускает сервер и ожидает подключения клиента.
func Host(address string) (*Manager, error) {
	if address == "" {
		address = defaultListenAddress
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	manager := newManager(nil)
	manager.listener = listener

	go manager.acceptOnce()

	return manager, nil
}

// Join подключается к удаленному хосту.
func Join(address string) (*Manager, error) {
	if address == "" {
		address = defaultDialAddress
	}

	conn, err := net.DialTimeout("tcp", address, defaultDialTimeout)
	if err != nil {
		return nil, err
	}

	return newManager(newPeer(conn)), nil
}

// Send отправляет состояние игры удаленному игроку.
func (m *Manager) Send(state StateMessage) error {
	if m == nil {
		return nil
	}
	if peer := m.getPeer(); peer != nil {
		return peer.send(state)
	}
	return nil
}

// LatestState возвращает последнее состояние, полученное от удаленного игрока.
func (m *Manager) LatestState() (StateMessage, bool) {
	if m == nil {
		return StateMessage{}, false
	}
	if peer := m.getPeer(); peer != nil {
		return peer.latestState()
	}
	return StateMessage{}, false
}

// Err возвращает ошибку соединения, если она произошла.
func (m *Manager) Err() error {
	if m == nil {
		return nil
	}
	if err := m.getErr(); err != nil {
		return err
	}
	if peer := m.getPeer(); peer != nil {
		return peer.getErr()
	}
	return nil
}

// Close закрывает подключение.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}

	var result error
	m.closeOnce.Do(func() {
		close(m.closed)

		// Закрываем listener, если он еще активен.
		if listener := m.swapListener(nil); listener != nil {
			if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				result = err
			}
		}

		// Закрываем peer, если он уже подключен.
		if peer := m.getPeer(); peer != nil {
			if err := peer.close(); err != nil && result == nil && !errors.Is(err, net.ErrClosed) {
				result = err
			}
		}
	})

	return result
}

type peer struct {
	conn    net.Conn
	sendCh  chan StateMessage
	closed  chan struct{}
	closeFn sync.Once

	mu      sync.RWMutex
	latest  StateMessage
	hasData bool

	errMu sync.Mutex
	err   error
}

func newPeer(conn net.Conn) *peer {
	p := &peer{
		conn:   conn,
		sendCh: make(chan StateMessage, defaultSendBufferSize),
		closed: make(chan struct{}),
	}

	go p.readLoop()
	go p.writeLoop()

	return p
}

func (p *peer) readLoop() {
	decoder := json.NewDecoder(p.conn)

	for {
		var msg StateMessage
		if err := decoder.Decode(&msg); err != nil {
			if !errors.Is(err, io.EOF) {
				p.setErr(err)
			} else {
				p.setErr(io.EOF)
			}
			p.close()
			return
		}

		p.mu.Lock()
		p.latest = msg
		p.hasData = true
		p.mu.Unlock()
	}
}

func (p *peer) writeLoop() {
	encoder := json.NewEncoder(p.conn)

	for {
		select {
		case <-p.closed:
			return
		case msg, ok := <-p.sendCh:
			if !ok {
				return
			}
			if err := encoder.Encode(&msg); err != nil {
				p.setErr(err)
				p.close()
				return
			}
		}
	}
}

func (p *peer) send(state StateMessage) error {
	select {
	case <-p.closed:
		return p.getErr()
	case p.sendCh <- state:
		return nil
	default:
		// Канал переполнен — сбрасываем старые данные и отправляем новое состояние.
		select {
		case <-p.closed:
			return p.getErr()
		case <-p.sendCh:
		default:
		}
		select {
		case <-p.closed:
			return p.getErr()
		case p.sendCh <- state:
			return nil
		default:
			return nil
		}
	}
}

func (p *peer) latestState() (StateMessage, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.hasData {
		return StateMessage{}, false
	}

	return p.latest, true
}

func (p *peer) getErr() error {
	p.errMu.Lock()
	defer p.errMu.Unlock()
	return p.err
}

func (p *peer) setErr(err error) {
	if err == nil {
		return
	}

	p.errMu.Lock()
	if p.err == nil {
		p.err = err
	}
	p.errMu.Unlock()
}

func (p *peer) close() error {
	var result error

	p.closeFn.Do(func() {
		close(p.closed)
		close(p.sendCh)
		result = p.conn.Close()
	})

	return result
}

func (m *Manager) acceptOnce() {
	listener := m.getListener()
	if listener == nil {
		return
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		if !errors.Is(err, net.ErrClosed) {
			m.setErr(err)
		}
		return
	}

	if m.isClosed() {
		_ = conn.Close()
		return
	}

	newPeer := newPeer(conn)

	m.mu.Lock()
	if m.peer != nil {
		m.mu.Unlock()
		_ = newPeer.close()
		return
	}
	m.peer = newPeer
	m.mu.Unlock()
}

func (m *Manager) getPeer() *peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.peer
}

func (m *Manager) getListener() net.Listener {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listener
}

func (m *Manager) swapListener(next net.Listener) net.Listener {
	m.mu.Lock()
	defer m.mu.Unlock()
	prev := m.listener
	m.listener = next
	return prev
}

func (m *Manager) setErr(err error) {
	if err == nil {
		return
	}

	m.errMu.Lock()
	if m.err == nil {
		m.err = err
	}
	m.errMu.Unlock()
}

func (m *Manager) getErr() error {
	m.errMu.Lock()
	defer m.errMu.Unlock()
	return m.err
}

func (m *Manager) isClosed() bool {
	select {
	case <-m.closed:
		return true
	default:
		return false
	}
}
