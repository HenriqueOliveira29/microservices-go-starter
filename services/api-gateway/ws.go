package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"ride-sharing/shared/contracts"

	"github.com/gorilla/websocket"
)

var (
	riderConnsMu sync.Mutex
	riderConns   = map[string]*websocket.Conn{}
	wsUpgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type riderConnection struct {
	userID string
	conn   *websocket.Conn
}

func handleRiderWs(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userID")
	if userID == "" {
		http.Error(w, "userID query parameter required", http.StatusBadRequest)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade websocket: %v", err)
		return
	}

	riderConnsMu.Lock()
	if old, ok := riderConns[userID]; ok {
		old.Close()
	}
	riderConns[userID] = conn
	riderConnsMu.Unlock()

	log.Printf("Rider websocket connected: %s", userID)
	defer func() {
		riderConnsMu.Lock()
		delete(riderConns, userID)
		riderConnsMu.Unlock()
		conn.Close()
	}()

	for {
		var message contracts.WSMessage
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("rider websocket read error for %s: %v", userID, err)
			break
		}

		log.Printf("rider websocket received message for %s: %s", userID, message.Type)
	}
}

func sendToRider(userID string, message interface{}) error {
	riderConnsMu.Lock()
	conn, ok := riderConns[userID]
	riderConnsMu.Unlock()
	if !ok {
		return fmt.Errorf("no websocket connected for rider %s", userID)
	}

	if err := conn.WriteJSON(message); err != nil {
		log.Printf("failed to write websocket message to rider %s: %v", userID, err)
		conn.Close()
		riderConnsMu.Lock()
		delete(riderConns, userID)
		riderConnsMu.Unlock()
		return err
	}

	return nil
}
