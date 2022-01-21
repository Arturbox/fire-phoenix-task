package main

import (
	"net"
	"os"
	"tcptask/pkg/simplepacket"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	var err error

	err = godotenv.Load()

	if err != nil {
		logrus.Errorf("error loading environment variables: %s\n", err)
	}

	port := os.Getenv("app_port")

	logrus.Infof("starting Simple Concurrent TCP Server on port %s", port)

	listener, err := net.Listen("tcp4", "localhost:"+port)

	if err != nil {
		logrus.WithError(err).Errorf("error listening port %d", port)
		return
	}

	defer listener.Close()

	cm := simplepacket.NewClientManager(16)

	for {
		conn, err := listener.Accept()

		if err != nil {
			logrus.WithError(err).Errorf("error accepting connection")
			return
		}

		go handleConnection(conn, cm)
	}
}

func handleConnection(conn net.Conn, cm *simplepacket.ClientManager) {
	client := conn.RemoteAddr().String()
	log := logrus.WithField("client_info", client)

	log.Infof("starting connection")

	clientID := cm.AddClient(conn)

	connected := true

	for connected {
		var req = &simplepacket.MessagePacket{}

		received, err := req.GetMessage(conn)

		if !received {
			continue
		}

		if err != nil {
			log.WithError(err).Info("error getting message")
			continue
		}

		log.Infof("packet from client: %+v", req)

		var res = &simplepacket.MessagePacket{}

		switch req.Body.Command {
		case simplepacket.PING:
			res.Body.Address = clientID
			res.Body.Message = "Hello from Simple Server!"
			cm.SendPrivateMessage(simplepacket.SERVER_ID, res)
		case simplepacket.DISCONNECT:
			cm.DisconnectClient(clientID)
			connected = false
		case simplepacket.SEND_BROADCAST_MESSAGE:
			req.Body.Address = clientID
			cm.Broadcast(req)
		case simplepacket.SEND_PRIVATE_MESSAGE:
			if req.Body.Address == clientID {
				res.Body.Address = clientID
				res.Body.Message = "You're trying to send a message to yourself!"
				cm.SendPrivateMessage(simplepacket.SERVER_ID, res)
				break
			}

			cm.SendPrivateMessage(clientID, req)
		case simplepacket.GET_CLIENT_LIST:
			req.Body.Address = clientID
			cm.SendClientList(req)
		}
	}

	conn.Close()
}
