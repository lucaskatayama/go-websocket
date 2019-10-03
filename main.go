package main

import (
  "encoding/json"
  "log"
  "net/http"

  "github.com/gorilla/websocket"
  "github.com/husobee/vestigo"
)

type Message struct {
  Message string `json:"msg"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcastChannel = make(chan *Message)
var upgrader = websocket.Upgrader{
  CheckOrigin: func(r *http.Request) bool {
    return true
  },
}

func main() {
  // create a new vestigo router
  router := vestigo.NewRouter()

  // Define a GET /welcome route, and
  // specify the standard http.HandlerFunc
  // to run from that route
  router.Get("/home/*", HomeHandler)

  // Define a POST /welcome/:name where :name
  // is a URL parameter that will be available
  // anywhere you have access to the http.Request
  router.Post("/broadcast", MessageHandler)

  router.HandleFunc("/ws", wsHandler)

  go Broadcast()

  log.Fatal(http.ListenAndServe(":8080", router))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

  log.Println(r.URL)
  if r.URL.Path != "/" {
    http.Error(w, "Not found", http.StatusNotFound)
    return
  }
  if r.Method != "GET" {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }
  http.ServeFile(w, r, "./index.html")

}

func writer(msg *Message) {
  broadcastChannel <- msg
}

func MessageHandler(w http.ResponseWriter, r *http.Request) {
  var message Message
  if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
    log.Printf("ERROR: %s", err)
    http.Error(w, "Bad request", http.StatusTeapot)
    return
  }
  defer r.Body.Close()
  go writer(&message)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Fatal(err)
  }
  // register client
  clients[ws] = true
}

// 3
func Broadcast() {
  for {
    val := <-broadcastChannel
    // send to every client that is currently connected
    for client := range clients {
      err := client.WriteMessage(websocket.TextMessage, []byte(val.Message))
      if err != nil {
        log.Printf("Websocket error: %s", err)
        client.Close()
        delete(clients, client)
      }
    }
  }
}
