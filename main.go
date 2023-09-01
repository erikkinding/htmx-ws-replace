package main

import (
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/init-ws", initWs)
	http.HandleFunc("/connect-ws", connectWs)

	http.ListenAndServe(":8081", nil)
}

func index(w http.ResponseWriter, req *http.Request) {
	indexTemplate := template.Must(template.ParseFiles("./index.html"))
	indexTemplate.Execute(w, nil)
}

func initWs(w http.ResponseWriter, req *http.Request) {
	var responseTemplate = `
		<div id="ws-output" hx-ext="ws" ws-connect="/connect-ws">	
			<div id="ws-output-lines"></div>
		</div>
		`

	tmpl := template.New("ws-output")
	tmpl.Parse(responseTemplate)

	tmpl.Execute(w, nil)
}

func connectWs(w http.ResponseWriter, req *http.Request) {
	connectionId := uuid.New()
	fmt.Printf("new connection: %s\n", connectionId)

	// try upgrade
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		fmt.Printf("Error trying to upgrade: %s\n", err.Error())
		return
	}
	defer c.Close()

	// continuously send a message with an incremented number
	message := 0
	for {
		var responseTemplate = `
		<div id="ws-output-lines" hx-swap-oob="beforeend">
			<div class="ws-output-line">{{ . }}</div>
		</div>
		`

		tmpl := template.New("log-line")
		tmpl.Parse(responseTemplate)

		socketWriter, err := c.NextWriter(websocket.TextMessage)

		if err != nil {
			fmt.Printf("%s: Error getting socket writer %s\n", connectionId, err.Error())
			break
		}

		err = tmpl.Execute(socketWriter, message)

		if err != nil {
			fmt.Printf("%s: write error: %s\n", connectionId, err.Error())
			break
		}

		socketWriter.Close()
		fmt.Printf("%s: message '%d' sent!\n", connectionId, message)
		message++

		time.Sleep(time.Second)
	}

	fmt.Printf("%s: closing connection\n", connectionId)
}
