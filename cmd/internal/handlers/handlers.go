package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	jet "github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)
var clients = make(map[WebsocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("/Users/aniket/Desktop/project/chatapp/cmd/html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebsocketConnection struct {
	*websocket.Conn
}

//WsJsonResponse struct defines the response send back from ws
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"messagetype"`
	ConnectedUsers []string `json:"connectedUsers"`
}

type WsPayload struct {
	Username string `json:"username"`
	Action   string `json:"action"`
	Message  string `json:"message"`
	Conn     WebsocketConnection
}

//WsEndPoint upgrade connections to endpoint
func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println("err", err)
	}
	log.Println("Client connected to endpoint")
	var response WsJsonResponse
	response.Message = `<em>connected to server</em>`

	conn := WebsocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println("err", err)
	}

	go ListenForWs(&conn)
}

func ListenForWs(conn *WebsocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}

}

func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			clients[e.Conn] = e.Username
			response.Action = "list_users"
			response.Message = fmt.Sprintf("some message and aciton was %s", e.Action)
			response.ConnectedUsers = getUserList()
			BroadCastToAll(response)

		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			BroadCastToAll(response)

		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong> : %s", e.Username, e.Message)
			BroadCastToAll(response)
		}

		// response.Action = "Got here"
		// response.Message = fmt.Sprintf("some message and aciton was %s", e.Action)
		// BroadCastToAll(response)
	}
}

func getUserList() []string {

	var userList []string

	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList
}

func BroadCastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	log.Println("tmpl", tmpl, views)
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		return err
	}
	return nil
}
