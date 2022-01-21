//Package simplepacket provides methods for concurrently managing clients and sending their messages

package simplepacket

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

//ClientManager holds every client connection information
type ClientManager struct {
	mu      sync.Mutex
	clients map[uint32]*Client
}

//NewClientManager
//client manager is supposed to be injected into every goroutine processing a new client connection
//size is how many clients you want to
func NewClientManager(size int) *ClientManager {
	cm := ClientManager{clients: make(map[uint32]*Client, size)}

	return &cm
}

//adding clients to map and returning their unique ID
func (cm *ClientManager) AddClient(conn net.Conn) uint32 {
	cm.mu.Lock()
	//clientID = 0 is reserved for the server
	clientID := uint32(len(cm.clients) + 1)

	cm.clients[clientID] = &Client{Connection: conn, ID: clientID, Connected: true}
	cm.mu.Unlock()

	return clientID
}

//broadcast message is delivered to everyone connected except the sender
func (cm *ClientManager) Broadcast(mp *MessagePacket) error {
	p := mp.Marshal()

	cm.mu.Lock()
	for _, client := range cm.clients {
		//excluding disconnected clients and the client who sent a broadcast message
		if !client.Connected || mp.Body.Address == client.ID {
			continue
		}

		_, err := client.Connection.Write(p)

		if err != nil {
			client.Connected = false
		}
	}
	cm.mu.Unlock()

	return nil
}

func (cm *ClientManager) SendClientList(mp *MessagePacket) error {
	var clients string

	cm.mu.Lock()
	for _, client := range cm.clients {

		if !client.Connected {
			continue
		}

		delimiter := ", "

		if mp.Body.Address == client.ID {
			delimiter = " (you)" + delimiter
		}

		clients = clients + strconv.FormatUint(uint64(client.ID), 10) + delimiter
	}

	cm.mu.Unlock()

	clients = strings.TrimRight(clients, ", ")

	mp.Body.Message = clients

	return cm.SendPrivateMessage(SERVER_ID, mp)
}

// sender id 0 is reserved to the server
func (cm *ClientManager) SendPrivateMessage(senderID uint32, mp *MessagePacket) error {
	cm.mu.Lock()
	client := cm.clients[mp.Body.Address]

	if client == nil {
		return fmt.Errorf("client %d not found", mp.Body.Address)
	}

	if !client.Connected {
		return fmt.Errorf("client %d disconnected", mp.Body.Address)
	}

	mp.Body.Address = senderID

	p := mp.Marshal()

	_, err := client.Connection.Write(p)

	if err != nil {
		client.Connected = false
		return fmt.Errorf("error sending message to client %d: %s", mp.Body.Address, err)
	}

	cm.mu.Unlock()

	return nil
}

//deleting clients is not available for simplicity sake
func (cm *ClientManager) DisconnectClient(clientID uint32) {
	cm.mu.Lock()
	client := cm.clients[clientID]
	client.Connected = false
	cm.mu.Unlock()
}
