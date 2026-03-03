package collector

import (
	"bufio"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"

	"phant/internal/dump"
)

type Server struct {
	socketPath string
	buffer     *RingBuffer
	decode     Decoder

	mu          sync.RWMutex
	subscribers map[int]chan Event
	nextSubID   int

	listener net.Listener
	stopOnce sync.Once
	stopped  chan struct{}
	wg       sync.WaitGroup
}

func NewServer(socketPath string, bufferSize int) *Server {
	if bufferSize <= 0 {
		bufferSize = DefaultBufferSize
	}

	return &Server{
		socketPath:  socketPath,
		buffer:      NewRingBuffer(bufferSize),
		decode:      dump.DecodeNDJSONLine,
		subscribers: make(map[int]chan Event),
		stopped:     make(chan struct{}),
	}
}

func (s *Server) Start() error {
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0o755); err != nil {
		return err
	}

	if err := os.Remove(s.socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	s.listener = listener
	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

func (s *Server) Stop() error {
	var closeErr error

	s.stopOnce.Do(func() {
		if s.listener != nil {
			closeErr = s.listener.Close()
		}
		close(s.stopped)
		s.wg.Wait()

		s.mu.Lock()
		for id, ch := range s.subscribers {
			close(ch)
			delete(s.subscribers, id)
		}
		s.mu.Unlock()
	})

	if errors.Is(closeErr, net.ErrClosed) {
		return nil
	}
	return closeErr
}

func (s *Server) SocketPath() string {
	return s.socketPath
}

func (s *Server) Events() []Event {
	return s.buffer.Snapshot()
}

func (s *Server) DroppedCount() uint64 {
	return s.buffer.DroppedCount()
}

func (s *Server) Subscribe(channelSize int) (int, <-chan Event) {
	if channelSize < 1 {
		channelSize = 1
	}

	ch := make(chan Event, channelSize)

	s.mu.Lock()
	id := s.nextSubID
	s.nextSubID++
	s.subscribers[id] = ch
	s.mu.Unlock()

	return id, ch
}

func (s *Server) Unsubscribe(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch, ok := s.subscribers[id]
	if !ok {
		return
	}

	close(ch)
	delete(s.subscribers, id)
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopped:
				return
			default:
				if errors.Is(err, net.ErrClosed) {
					return
				}
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		event, err := s.decode(scanner.Text())
		if err != nil || event == nil {
			continue
		}

		s.buffer.Add(*event)
		s.broadcast(*event)
	}
}

func (s *Server) broadcast(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}
