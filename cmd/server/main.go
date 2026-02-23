package main

import (
	"log"
	"mmorpg/internal/game"
	"mmorpg/internal/network"
	"net/http"
)

func main() {
	g := game.NewGame()
	go g.Start()

	wsServer := network.NewWSServer(g)

	http.Handle("/", http.FileServer(http.Dir("client")))
	http.HandleFunc("/ws", wsServer.HandleWS)

	log.Println("Server starting on 0.0.0.0:9000 (Accessible from external IPs, e.g., 192.168.0.3:9000)")
	if err := http.ListenAndServe("0.0.0.0:9000", nil); err != nil {
		log.Fatal(err)
	}
}
