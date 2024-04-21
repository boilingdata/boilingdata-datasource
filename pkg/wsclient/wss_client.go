package wsclient

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/boilingdata/boilingdata/pkg/constants"
	"github.com/gorilla/websocket"
)

// WSSClient represents the WebSocket client.
type WSSClient struct {
	URL                string
	Conn               *websocket.Conn
	DialOpts           *websocket.Dialer
	idleTimeoutMinutes time.Duration
	idleTimer          *time.Timer
	Wg                 sync.WaitGroup
	ConnInit           sync.WaitGroup
	SignedHeader       http.Header
	Error              string
}

// NewWSSClient creates a new instance of WSSClient.
// Either fully signed url needs to be provided OR signedHeader
func NewWSSClient(url string, idleTimeoutMinutes time.Duration, signedHeader http.Header) *WSSClient {
	if signedHeader == nil {
		signedHeader = make(http.Header)
	}
	return &WSSClient{
		URL:                url,
		DialOpts:           &websocket.Dialer{},
		idleTimeoutMinutes: idleTimeoutMinutes,
		SignedHeader:       signedHeader,
	}
}

func (wsc *WSSClient) Connect() {
	if wsc.IsWebSocketClosed() {
		log.Println("Connecting to web socket..")
		wsc.ConnInit.Add(1)
		go wsc.connect()
		wsc.ConnInit.Wait()
	}
	if !wsc.IsWebSocketClosed() {
		log.Println("Websocket Connected!")
	}
}

func (wsc *WSSClient) connect() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wsc.Wg = sync.WaitGroup{}
	wsc.Wg.Add(1)

	go func() {
		defer wsc.Wg.Done()
		for {
			select {
			case <-interrupt:
				log.Println("Interrupt signal received, closing connection")
				wsc.Close()
				return
			}
		}
	}()
	// Connect to WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(wsc.URL, wsc.SignedHeader)
	if err != nil {
		wsc.Error = err.Error()
		log.Println("dial:", err)
		wsc.ConnInit.Done()
		return
	}
	wsc.Conn = conn // Assign the connection to the Conn field

	if wsc.idleTimeoutMinutes <= 0 {
		wsc.idleTimeoutMinutes = constants.IdleTimeoutMinutes
	} else {
		wsc.idleTimeoutMinutes = wsc.idleTimeoutMinutes * time.Minute
	}
	wsc.resetIdleTimer()
	wsc.ConnInit.Done()
}

// SendMessage sends a message over the WebSocket connection.
func (wsc *WSSClient) SendMessage(message string) error {
	if wsc.Conn == nil {
		wsc.Error = "not connected to WebSocket server"
		return fmt.Errorf("not connected to WebSocket server")
	}
	wsc.idleTimer.Reset(constants.IdleTimeoutMinutes)
	return wsc.Conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// Close closes the WebSocket connection.
func (wsc *WSSClient) Close() {
	if wsc.Conn != nil {
		wsc.Conn.Close()
		wsc.Conn = nil
		wsc.idleTimer = nil
	}
}

func (wsc *WSSClient) IsWebSocketClosed() bool {
	return wsc.Conn == nil
}

// resetIdleTimer resets the idle timer.
func (wsc *WSSClient) resetIdleTimer() {
	if wsc.idleTimer != nil {
		wsc.idleTimer.Stop()
	}
	wsc.idleTimer = time.AfterFunc(wsc.idleTimeoutMinutes, func() {
		log.Println("Idle timeout reached, closing connection")
		wsc.Close()
	})
}
