package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

// ClientManager will keep track of:
type ClientManager struct {
	clients    map[*Client]bool // Connected clients
	broadcast  chan []byte      // Messages broadcasted from and to clients
	register   chan *Client     // Clients trying to register
	unregister chan *Client     // Clients trying to unregister
}

// A function to send a message to every client, except from the one sending it
func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

type Client struct {
	id     string          // Unique id
	socket *websocket.Conn // Socket connection
	send   chan []byte     // Message to be sent
}

// A function to read the data sont from the clients through the websocket, it then will be added to the manager broadcast
func (c *Client) read() {
	defer func() { // If there's any error reading the websocket, the client is assumed disconnected, then deleted from the pool
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message)})
		manager.broadcast <- jsonMessage
	}
}

// A function for clients to write data into the websocket for the other clients

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{}) // If the cannel is not alright, we'll send a message to disconnect the client
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register: // In case som client wants to connect to the service, the manager.register will get its data
			manager.clients[conn] = true // Then the client will be set to active in the manager's Client pool
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			manager.send(jsonMessage, conn) // A message is sent to all the other clients connected
		case conn := <-manager.unregister: // In case some client unregisters, the manager will get its data as well
			if _, ok := manager.clients[conn]; ok { // Check if the client does exist in the pool
				close(conn.send)              // The channel data is closed
				delete(manager.clients, conn) // The client is deleted from the service
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
				manager.send(jsonMessage, conn) // A message is sent to all the other clients connected
			}
		case message := <-manager.broadcast: // If the broadcast has data in it, some client is trying to send/receive data
			for conn := range manager.clients { // The we send the data to each connected client
				select {
				case conn.send <- message: // We insert the message into each client's "mailbox"
				default: // If the message can't be sent, the client is considered dead, so it'll be deleted
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func wsPage(res http.ResponseWriter, req *http.Request) {
	// The http request is upgraded to a websocket request, checkOrigin solves possible CORS errors
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	// When a connection is made, a client with a unique id is created (using uuid) and registered
	client := &Client{id: uuid.NewV4().String(), socket: conn, send: make(chan []byte)}

	manager.register <- client
	// Then the read an write goroutines are triggered
	go client.read()
	go client.write()
}

func main() {
	fmt.Println("Starting application...")
	go manager.start()
	http.HandleFunc("/ws", wsPage)
	http.ListenAndServe(":12345", nil)
}
