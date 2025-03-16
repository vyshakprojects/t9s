package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"mytunnel/internal/config"
)

// Tunnel represents an active SSH tunnel
type Tunnel struct {
	LocalPort  int
	RemotePort int
	Bastion    *config.BastionConfig
	listener   net.Listener
	client     *ssh.Client
	done       chan struct{}
}

// TunnelManager manages multiple SSH tunnels
type TunnelManager struct {
	tunnels map[int]*Tunnel
	mu      sync.RWMutex
}

// NewTunnelManager creates a new tunnel manager
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		tunnels: make(map[int]*Tunnel),
	}
}

// CreateTunnel establishes a new SSH tunnel
func (tm *TunnelManager) CreateTunnel(localPort, remotePort int, bastion *config.BastionConfig) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.tunnels[localPort]; exists {
		return fmt.Errorf("tunnel already exists on local port %d", localPort)
	}

	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User:            bastion.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 10,
	}

	// Set up authentication
	if bastion.AuthType == "key" {
		key, err := ioutil.ReadFile(bastion.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		config.Auth = []ssh.AuthMethod{ssh.Password(bastion.Password)}
	}

	// Connect to bastion
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", bastion.Host, bastion.Port), config)
	if err != nil {
		return fmt.Errorf("failed to connect to bastion: %w", err)
	}

	// Start local listener
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localPort))
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to start local listener: %w", err)
	}

	tunnel := &Tunnel{
		LocalPort:  localPort,
		RemotePort: remotePort,
		Bastion:    bastion,
		listener:   listener,
		client:     client,
		done:       make(chan struct{}),
	}

	tm.tunnels[localPort] = tunnel

	// Start handling connections
	go tunnel.handleConnections()

	return nil
}

// handleConnections handles incoming connections to the tunnel
func (t *Tunnel) handleConnections() {
	for {
		select {
		case <-t.done:
			return
		default:
			conn, err := t.listener.Accept()
			if err != nil {
				select {
				case <-t.done:
					return
				default:
					fmt.Printf("Failed to accept connection: %v\n", err)
					continue
				}
			}
			go t.handleConnection(conn)
		}
	}
}

// handleConnection forwards a single connection through the tunnel
func (t *Tunnel) handleConnection(local net.Conn) {
	remote, err := t.client.Dial("tcp", fmt.Sprintf("localhost:%d", t.RemotePort))
	if err != nil {
		fmt.Printf("Failed to connect to remote port: %v\n", err)
		local.Close()
		return
	}

	// Copy data bidirectionally
	go func() {
		defer local.Close()
		defer remote.Close()
		copyData(local, remote)
	}()

	go func() {
		defer local.Close()
		defer remote.Close()
		copyData(remote, local)
	}()
}

// copyData copies data between connections
func copyData(dst, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	buffer := make([]byte, 32*1024)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			return
		}
		if _, err := dst.Write(buffer[:n]); err != nil {
			return
		}
	}
}

// CloseTunnel closes a specific tunnel
func (tm *TunnelManager) CloseTunnel(localPort int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tunnel, exists := tm.tunnels[localPort]
	if !exists {
		return fmt.Errorf("no tunnel exists on local port %d", localPort)
	}

	close(tunnel.done)
	tunnel.listener.Close()
	tunnel.client.Close()
	delete(tm.tunnels, localPort)

	return nil
}

// ListTunnels returns a list of active tunnels
func (tm *TunnelManager) ListTunnels() []*Tunnel {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tunnels := make([]*Tunnel, 0, len(tm.tunnels))
	for _, tunnel := range tm.tunnels {
		tunnels = append(tunnels, tunnel)
	}
	return tunnels
}

// CloseAll closes all active tunnels
func (tm *TunnelManager) CloseAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for port, tunnel := range tm.tunnels {
		close(tunnel.done)
		tunnel.listener.Close()
		tunnel.client.Close()
		delete(tm.tunnels, port)
	}
} 