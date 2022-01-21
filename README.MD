# Overview
The source code implements a simple concurrent tcp client/server functionality

The code for the server is in the /server folder, respectively.
The code for the client is in the /client folder.  Client instances must be run only after a server instance is up. The client supports user input from the shell.

# How to run
```
\server> go run main.go
\client> go run main.go
```

# Message format
According to the task description, the message packet format is as follows:
- the first two bytes encode the message length
- the other bytes encode the message itself

In order to distinguish broadcast messages from private messages to the specified client,
I expanded on the message packet format as follows:

```go
type MessagePacket struct {
	Length uint16 //two bytes for the message length
	Body   struct {
		Address uint32 //address for a private message (only one address is supported)
		Command uint16 //command id (the description is below)
		Message string //a text message
	}
}
```

Basically, the client sends to the server a command they want to invoke (send message, get online client list,etc) along with a text message if applicable.

# Command list
The following commands are supported:

```go
const (
	PING                   = 0
	GET_CLIENT_LIST        = 1
	SEND_PRIVATE_MESSAGE   = 2
	SEND_BROADCAST_MESSAGE = 3
	DISCONNECT             = 4
)
```