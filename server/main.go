package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClientManager struct {
	clients    	map[string]*Client
	broadcast  	chan []byte
	register   	chan *Client
	unregister 	chan *Client
	dbClient	*mongo.Client
}

func (manager *ClientManager) send(message []byte, recipient string) {
	if recipient != "" {
		conn, ok := manager.clients[recipient]
		if ok {
			conn.send <- message
		}
	}
}

type Client struct {
	id     			string           
	socket 			*websocket.Conn 
	send   			chan []byte
	Username		string
}

func (c *Client) read(manager *ClientManager) {
	defer func() { 
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
		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			fmt.Println("Error decoding message: ", err)
			continue
		}
		msg.Sender = c.Username
		jsonMessage, _ := json.Marshal(&msg)
		manager.broadcast <- jsonMessage

		err = sendMessageToDB(manager.dbClient, msg)
		if err != nil {
			log.Println("Error saving the message to the MongoDB: ", err)
		}
	}
}

// 

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{}) 
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

type Message struct {
	ID        	primitive.ObjectID 	`json:"_id,omitempty" bson:"_id,omitempty"`
	Sender    	string             	`json:"sender,omitempty" bson:"sender,omitempty"`
	Recipient 	string             	`json:"recipient,omitempty" bson:"recipient,omitempty"`
	Content   	string             	`json:"content,omitempty" bson:"content,omitempty"`
	Timestamp 	time.Time          	`json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

type User struct {
	Username   string            `json:"username,omitempty" bson:"username,omitempty"` 
}

func connectToMongoDB(ctx context.Context, mongoURI string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	log.Println("Connected to MongoDB!")
	return client, nil
}

func sendMessageToDB(client *mongo.Client, message Message) error {
	collection := client.Database("chat").Collection("messages")
	message.Timestamp = time.Now()
	_, err := collection.InsertOne(context.Background(), message)
	if err != nil {
		return err
	}
	return nil
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[string]*Client),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn.Username] = conn
			log.Println("User ", conn.Username, " connected")
		case conn := <-manager.unregister: 
			if _, ok := manager.clients[conn.Username]; ok { 
				close(conn.send)              
				delete(manager.clients, conn.Username)
			}
		case message := <-manager.broadcast:
			var msg Message
			err := json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("Error decoding message: ", err)
				continue
			}
			manager.send(message, msg.Recipient)
		}
	}
}

func getMessagesfromDB(client *mongo.Client, username, receiver string) ([]Message, error) {
	collection := client.Database("chat").Collection("messages")

	filter := bson.M{
		"$or": []bson.M{
			{"sender": username, "recipient": receiver},
			{"sender": receiver, "recipient": username},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []Message
	for cursor.Next(context.Background()) {
		var msg Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func getChatsFromFB(client *mongo.Client, username string) ([]map[string]string, error) {
	collection := client.Database("chat").Collection("messages")

	filter := bson.M{
		"$or": []bson.M{
			{"sender": username},
			{"recipient": username},
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	usersMap := make(map[string]bool)
	for cursor.Next(context.Background()) {
		var msg Message
		if err := cursor.Decode(&msg); err != nil {
			return nil, err
		}
		if msg.Sender != username {
			usersMap[msg.Sender] = true
		}
		if msg.Recipient != username {
			usersMap[msg.Recipient] = true
		}
	}

	var users []map[string]string
	for user:= range usersMap {
		chat := map[string]string{"username":user}
		users = append(users, chat)
	}
	
	return users, nil
}

func history(res http.ResponseWriter, req *http.Request) {

	username := req.URL.Query().Get("username")
	receiver := req.URL.Query().Get("receiver")
	messages, err := getMessagesfromDB(manager.dbClient, username, receiver)
	if err != nil {
		log.Println("Error fetching the data fomr the database", err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(messages)
}

func chats(res http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	chats, err := getChatsFromFB(manager.dbClient, username)
	if err != nil {
		log.Println("Error fetching the data fomr the database", err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(chats)
}

func wsPage(res http.ResponseWriter, req *http.Request) {
	
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}

	username := req.URL.Query().Get("username")
	if username == "" {
		username = "Anonymous"
	}
	
	client := &Client{id: uuid.NewV4().String(), socket: conn, send: make(chan []byte), Username: username}

	manager.register <- client
	
	go client.read(&manager)
	go client.write()
}

func main() {
	fmt.Println("Starting application...")
	ctx := context.Background()
	dbClient, err := connectToMongoDB(ctx, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB", err)
	}
	defer dbClient.Disconnect(ctx)
	manager.dbClient= dbClient
	
	go manager.start()

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})
	router := http.NewServeMux()

	router.HandleFunc("/ws", wsPage)
	router.HandleFunc("/history", history)
	router.HandleFunc("/chats", chats)

	handler := corsHandler.Handler(router)
	http.ListenAndServe(":12345", handler)		// A websocket connection can be stablished via: ws://localhost:12345/ws
}
