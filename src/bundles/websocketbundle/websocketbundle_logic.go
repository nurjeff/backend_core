package websocketbundle

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sc-js/backend_core/src/bundles/authbundle"
	"github.com/sc-js/backend_core/src/tools"
	"github.com/sc-js/pour"
	"gorm.io/gorm"
)

type wsclient struct {
	hub *hub

	conn *websocket.Conn

	mu sync.Mutex

	// Buffered channel of outbound messages.
	send chan []byte

	User authbundle.AuthUser
}

type hub struct {
	// Registered clients
	clients map[*wsclient]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *wsclient

	// Unregister requests from clients
	unregister chan *wsclient

	idClientMap sync.Map

	DataWrap *tools.DataWrap
}

func newHub(wrap *tools.DataWrap) *hub {
	return &hub{
		broadcast:  make(chan []byte),
		register:   make(chan *wsclient),
		unregister: make(chan *wsclient),
		clients:    make(map[*wsclient]bool),
		DataWrap:   wrap,
	}
}

func (h *hub) run() {

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if connectedClient, ok := h.idClientMap.Load(client.User.ID); ok {
				connectedClient.(*wsclient).conn.Close()
			}
			h.idClientMap.Store(client.User.ID, client)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				h.idClientMap.Delete(client.User.ID)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					h.idClientMap.Delete(client.User.ID)
				}
			}
		}
	}
}

func (c *wsclient) SendMessage(msg WSMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.WriteJSON(&msg)
	if err != nil {
		go CacheSendMessage(msg, c.User.ID)
	}
}

type WSMessage struct {
	Type    uint   `json:"type"`
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var wsSendingCache sync.Map

func serveWs(hub *hub, w http.ResponseWriter, r *http.Request, user authbundle.AuthUser) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pour.LogColor(false, "WS Error:", err)
		return
	}
	client := &wsclient{hub: hub, conn: conn, send: make(chan []byte, 256), User: user}
	pour.LogColor(false, pour.ColorYellow, "Registered WS Client Account:", client.User.ID)
	client.hub.register <- client

	cache, ok := wsSendingCache.Load(user.ID)
	if ok {
		wsSendingCache.LoadAndDelete(user.ID)
		for _, element := range cache.([]WSMessage) {
			go SpreadMessageToIds(element, []tools.ModelID{user.ID})
		}
	}

	go client.writePump()
	go client.readPump()
}

func SpreadMessageToIds(msg WSMessage, ids []tools.ModelID) {
	clients, missing := wshub.getClientsWithAccountIds(ids)
	for _, element := range clients {
		go element.SendMessage(msg)
	}
	for _, element := range missing {
		go CacheSendMessage(msg, element)
	}
}

func CacheSendMessage(msg WSMessage, id tools.ModelID) {
	cache, ok := wsSendingCache.Load(id)
	if ok {
		cachePort := cache.([]WSMessage)
		if len(cachePort) >= 50 {
			cachePort = cachePort[1:]
		}
		cachedMessages := cachePort
		cachedMessages = append(cachedMessages, msg)
		wsSendingCache.Store(id, cachedMessages)
		return
	}
	newCache := []WSMessage{}
	newCache = append(newCache, msg)
	wsSendingCache.Store(id, newCache)
}

func (h *hub) getClientsWithAccountIds(ids []tools.ModelID) ([]*wsclient, []tools.ModelID) {
	missingIds := ids
	clients := []*wsclient{}

	for _, element := range ids {
		client, ok := h.idClientMap.Load(element)
		if ok {
			clients = append(clients, client.(*wsclient))
			missingIds = removeId(missingIds, element)
		}
	}

	return clients, missingIds
}

func removeId(s []tools.ModelID, id tools.ModelID) []tools.ModelID {
	for index, element := range s {
		if element == id {
			return append(s[:index], s[index+1:]...)
		}
	}
	return s
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 100 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 120 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (c *wsclient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *wsclient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			pour.LogColor(false, pour.ColorRed, "WS Client", c.User.ID, "disconnected:", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				pour.LogColor(false, pour.ColorRed, "Error:", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		checkAndDistributeMessage(message, c)
		//c.hub.broadcast <- message
	}
}

var messageCallbacks map[uint]func(msg WSMessage, c *wsclient, db *gorm.DB) = make(map[uint]func(msg WSMessage, c *wsclient, db *gorm.DB))

func RegisterCallback(t uint, f func(msg WSMessage, c *wsclient, db *gorm.DB)) {
	messageCallbacks[t] = f
}

func checkAndDistributeMessage(message []byte, client *wsclient) {
	msg := WSMessage{}
	err := json.Unmarshal(message, &msg)
	if err != nil {
		pour.LogColor(false, pour.ColorRed, "Error unmarshalling ws message:", message)
		return
	}
	if messageCallbacks[msg.Type] != nil {
		go messageCallbacks[msg.Type](msg, client, wshub.DataWrap.DB)
		return
	}
	pour.LogColor(false, pour.ColorRed, "WS Message", msg, "has no callback")
}
