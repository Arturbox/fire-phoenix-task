package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

	server_address := os.Getenv("server_address")

	log := logrus.WithField("server_address", server_address)

	log.Infof("starting Simple Concurrent TCP-Client")

	conn, err := net.Dial("tcp4", server_address)

	if err != nil {
		fmt.Printf("Error connection to server %s!\n", server_address)
		log.WithError(err).Error("error connecting to server")
		return
	}

	defer conn.Close()

	manageClientConnection(conn)
}

//listening for new messages in background
func receiveMessage(conn net.Conn) {
	for {
		var res = &simplepacket.MessagePacket{}

		received, err := res.GetMessage(conn)

		if !received {
			continue
		}

		if err != nil {
			logrus.WithError(err).Info("error getting message")
			continue
		}

		body := res.Body
		var sender string

		if body.Address == simplepacket.SERVER_ID {
			sender = "Simple Server"
		} else {
			sender = fmt.Sprintf("Client #%d", body.Address)
		}

		fmt.Printf("\n>>>> message from %s: %s\n", sender, string(res.Body.Message))
	}
}

func manageClientConnection(conn net.Conn) {
	log := logrus.WithField("server_address", conn.RemoteAddr())

	fmt.Printf("Hello! You are connected to %s\n", conn.RemoteAddr())
	go receiveMessage(conn)

	//send empty packet with default command (PING)
	var res = &simplepacket.MessagePacket{}
	m := res.Marshal()
	_, err := conn.Write(m)

	if err != nil {
		log.WithError(err).Info("error sending ping request")
	}

	commands := []string{"Get client list", "Send message to another client", "Send a broadcast message", "Disconnect"}

	connected := true

	for connected {
		commandNumber, err := getClientCommand(commands)

		if err != nil {
			fmt.Println("You should enter only digits, without other characters")
			//logrus.WithError(err).Error("wrong input from client")
			continue
		}

		if commandNumber < 0 || commandNumber > len(commands) {
			fmt.Printf("You should enter command numbers from 1 to %d\n", len(commands))
			//logrus.Error("wrong command number from client: %d", commandNumber)
			continue
		}

		var message string

		req := &simplepacket.MessagePacket{}
		req.Body.Command = uint16(commandNumber)

		switch commandNumber {
		case simplepacket.SEND_BROADCAST_MESSAGE:
			message = getMessage()
		case simplepacket.SEND_PRIVATE_MESSAGE:
			clientId, err := getClientID()

			if err != nil {
				fmt.Println("You should enter only digits, without other characters")
				//logrus.Error("error getting client id: %s", err)
				continue
			}

			message = getMessage()
			req.Body.Address = uint32(clientId)
		case simplepacket.DISCONNECT:
			connected = false
		case simplepacket.GET_CLIENT_LIST:

		default:
			fmt.Println("Invalid command!")
		}

		req.Body.Message = message
		m := req.Marshal()

		_, err = conn.Write(m)

		if err != nil {
			log.WithError(err).Info("error sending requst")
		}
	}
}

func getMessage() string {
	fmt.Print("Enter message: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimRight(input, "\r\n")

	return input
}

func getClientID() (uint32, error) {
	fmt.Print("Enter the id of the client (id should be from 0 to 255): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimRight(input, "\r\n")

	clientId, err := strconv.ParseUint(input, 10, 32)

	if err != nil {
		return 0, fmt.Errorf("error reading client id: %s", err)
	}

	return uint32(clientId), nil
}

func getClientCommand(commands []string) (int, error) {
	fmt.Println("Choose command:")

	for i, command := range commands {
		fmt.Printf("%d. %s\n", i+1, command)
	}

	fmt.Print("Enter the number of command: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\r')
	input = strings.TrimRight(input, "\r\n")

	fmt.Println("----")

	commandNumber, err := strconv.Atoi(input)

	if err != nil {
		return 0, fmt.Errorf("error reading command: %s", err)
	}

	return commandNumber, nil
}
