package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/a-h/templ"
	"github.com/lesismal/nbio/nbhttp"
	"github.com/lesismal/nbio/nbhttp/websocket"
)

var upgrader = newUpgrader()

type color struct {
	r uint8
	g uint8
	b uint8
}

const gridWidth = 50
const gridHeight = 50

const gridFile = "/data/grid"

var grid [gridWidth * gridHeight]color

type Session struct {
	color color

	// TODO add lastFullFetch and rate limit
}

func newUpgrader() *websocket.Upgrader {
	u := websocket.NewUpgrader()
	u.OnOpen(func(c *websocket.Conn) {
		buf := new(bytes.Buffer)

		// byte 1 -> message type -> initialize
		err := binary.Write(buf, binary.BigEndian, (uint8)(1))
		if err != nil {
			log.Println("binary.Write failed:", err)
		}

		// byte 2,3 -> grid width
		err = binary.Write(buf, binary.BigEndian, (uint16)(gridWidth))
		if err != nil {
			log.Println("binary.Write failed:", err)
		}

		// byte 4,5 -> grid width
		err = binary.Write(buf, binary.BigEndian, (uint16)(gridHeight))
		if err != nil {
			log.Println("binary.Write failed:", err)
		}

		// byte 6, 7, 8 -> RGB for the client
		session := Session{}
		session.color.r = (uint8)(rand.UintN(256))
		session.color.g = (uint8)(rand.UintN(256))
		session.color.b = (uint8)(rand.UintN(256))

    err = writeColorToBinary(buf, binary.BigEndian, session.color)
    if err != nil {
      log.Println("Error writing color to binary:", err)
    }

		c.WriteMessage(websocket.BinaryMessage, buf.Bytes())
		c.SetSession(session)
	})
	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		if data[0] == 2 {
			// command 2 -> set tile in grid

			offset := binary.LittleEndian.Uint32(data[1:5])
			session := c.Session().(Session)

			grid[offset] = session.color

      // TODO buffer
      saveGridToFile()
		} else if data[0] == 3 {
			// command 3 -> fetch whole grid

			buf := new(bytes.Buffer)

			// byte 1 -> message type -> whole grid
			err := binary.Write(buf, binary.BigEndian, (uint8)(3))
			if err != nil {
				log.Println("binary.Write failed:", err)
			}

			// rest bytes -> grid
      for i := 0; i < gridWidth*gridHeight; i++ {
        writeColorToBinary(buf, binary.BigEndian, grid[i])
      }

			c.WriteMessage(websocket.BinaryMessage, buf.Bytes())
		}
	})
	u.OnClose(func(c *websocket.Conn, err error) {
		log.Printf("OnClose: %+v %s", c.RemoteAddr().String(), err)
	})
	return u
}

func onWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%+v\n", err)
		return
	}
	log.Print("Upgraded: ", conn.RemoteAddr().String())
}

func onIndex(w http.ResponseWriter, r *http.Request) {
}

func main() {
  readGridFromFile()

	mux := &http.ServeMux{}

	mux.Handle("/", templ.Handler(indexComponent()))
	mux.Handle("GET /static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))
	mux.HandleFunc("/ws", onWebsocket)

	engine := nbhttp.NewEngine(nbhttp.Config{
		Network:                 "tcp",
		Addrs:                   []string{"0.0.0.0:8080"},
		MaxLoad:                 1_000_000,
		ReleaseWebsocketPayload: true,
		Handler:                 mux,
	})

	err := engine.Start()
	if err != nil {
		log.Printf("nbio.Start failed: %v\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	engine.Shutdown(ctx)
}

func writeColorToBinary(w io.Writer, order binary.ByteOrder, color color) error {
	err := binary.Write(w, order, color.r)
	if err != nil {
		return err
	}

	err = binary.Write(w, order, color.g)
	if err != nil {
		return err
	}

	err = binary.Write(w, order, color.b)
	if err != nil {
		return err
	}
  
  return nil
}

func readGridFromFile() {
  bytes, err := os.ReadFile(gridFile)
  if err != nil {
    log.Println("Couldn't read grid file", err)
    return
  }

  for i := range min(gridWidth*gridHeight, len(bytes)) {
    grid[i].r = bytes[i*3]
    grid[i].g = bytes[i*3+1]
    grid[i].b = bytes[i*3+2]
  }
}

func saveGridToFile() {
  f, err := os.Create(gridFile)
  if err != nil {
    log.Println("Couldn't create grid file", err)
    return
  }
  defer f.Close()

  var bytes [gridWidth*gridHeight*3]byte
  for i := range(gridWidth*gridHeight) {
    bytes[i*3] = grid[i].r
    bytes[i*3+1] = grid[i].g
    bytes[i*3+2] = grid[i].b
  }

  f.Write(bytes[:])
}
