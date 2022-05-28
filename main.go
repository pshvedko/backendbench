package main

import (
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/ambelovsky/gosf-socketio"
	"github.com/ambelovsky/gosf-socketio/transport"
	"github.com/google/uuid"
)

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type Query struct {
	Id     uuid.UUID   `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type Reply struct {
	Id     uuid.UUID   `json:"id"`
	Method string      `json:"method"`
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

type UpdateCodecGroupParams struct {
	Id          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	DtmfRfc2833 bool        `json:"dtmf_rfc_2833"`
	DtmfInband  bool        `json:"dtmf_inband"`
	DtmfSipInfo bool        `json:"dtmf_sip_info"`
	Codecs      []uuid.UUID `json:"codecs"`
}

func run() (err error) {
	var ws *gosocketio.Client
	ws, err = gosocketio.Dial("ws://localhost:8080/socket.io/?EIO=3&transport=websocket",
		&transport.WebsocketTransport{
			PingInterval:   transport.WsDefaultPingInterval,
			PingTimeout:    transport.WsDefaultPingTimeout,
			ReceiveTimeout: transport.WsDefaultReceiveTimeout,
			SendTimeout:    transport.WsDefaultSendTimeout,
			BufferSize:     transport.WsDefaultBufferSize,
			UnsecureTLS:    false,
		},
	)
	if err != nil {
		return
	}
	defer ws.Close()

	var wg sync.WaitGroup

	err = ws.On("login", func(c *gosocketio.Channel, reply Token) {
		log.Printf("login %#v", reply)
		wg.Done()
	})
	if err != nil {
		return
	}

	n := 0

	err = ws.On("response", func(c *gosocketio.Channel, reply Reply) {
		log.Printf("response %d %#v", n, reply)
		wg.Done()
		n++
	})
	if err != nil {
		return
	}

	wg.Add(1)

	err = ws.Emit("login", Login{
		Email:    "test@example.com",
		Password: "12345",
	})
	if err != nil {
		return
	}

	wg.Wait()

	gid1 := uuid.MustParse("80d5bde7-827f-42fb-bd8a-8f8a7835d9bc")
	cid1 := uuid.MustParse("08763fc1-a644-4004-ad87-8fc69e3fc7c5")
	cid2 := uuid.MustParse("e9abcaa0-9cc9-4a70-a97e-bb8a7758917f")
	cid3 := uuid.MustParse("3e7ddb68-a3c4-46e2-a721-e0e4c228732e")

	for i := 0; i < 66; i++ {
		wg.Add(1)

		err = ws.Emit("query", Query{
			Id:     uuid.New(),
			Method: "updateCodecGroup",
			Params: UpdateCodecGroupParams{
				Id:          gid1,
				Name:        "default-audio-codec-group",
				Type:        "audio",
				DtmfRfc2833: true,
				DtmfInband:  false,
				DtmfSipInfo: false,
				Codecs: func(uuids ...uuid.UUID) []uuid.UUID {
					sort.Slice(uuids, func(i, j int) bool {
						return rand.Int()&1 == 1
					})
					return uuids
				}(cid1, cid2, cid3),
			},
		})
		if err != nil {
			return
		}
	}

	wg.Wait()

	return
}

func main() {
	rand.Seed(time.Now().Unix())
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
