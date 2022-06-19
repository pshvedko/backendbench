package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"reflect"
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

	n, e, x, z := 0, 0, 0, 0

	var tt sync.Map

	err = ws.On("response", func(c *gosocketio.Channel, reply Reply) {
		log.Printf("response %d %#v", n, reply)
		if reply.Error != nil {
			e++
		} else {
			v, ok := tt.LoadAndDelete(reply.Id)
			if !ok {
				x++
			} else if !equalUpdateCodecGroupParams(reply.Result, v) {
				z++
			}
		}
		n++
		wg.Done()
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

	gid1 := uuid.MustParse("818ed602-54bb-4a31-a70c-76558000ef6a")
	cid1 := uuid.MustParse("08763fc1-a644-4004-ad87-8fc69e3fc7c5")
	cid2 := uuid.MustParse("e9abcaa0-9cc9-4a70-a97e-bb8a7758917f")
	cid3 := uuid.MustParse("3e7ddb68-a3c4-46e2-a721-e0e4c228732e")

	{
		for i := 0; i < 999; i++ {
			wg.Add(1)

			codecs := func(uuids ...uuid.UUID) []uuid.UUID {
				sort.Slice(uuids, func(i, j int) bool {
					return rand.Int()&1 == 1
				})
				return uuids[:i%4]
			}(cid1, cid2, cid3)

			params := UpdateCodecGroupParams{
				Id:          gid1,
				Name:        "default-audio-codec-group",
				Type:        "audio",
				DtmfRfc2833: true,
				DtmfInband:  false,
				DtmfSipInfo: false,
				Codecs:      codecs,
			}

			id := uuid.New()

			tt.Store(id, params)

			err = ws.Emit("query", Query{
				Id:     id,
				Method: "updateCodecGroup",
				Params: params,
			})
			if err != nil {
				return
			}
		}

		wg.Wait()

		log.Println()
		log.Println("ERRORS", e, z, x)
		log.Println()
	}

	return
}

func equalUpdateCodecGroupParams(r, v any) bool {
	b, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}

	var u UpdateCodecGroupParams
	err = json.Unmarshal(b, &u)
	if err != nil {
		log.Fatal(err)
	}

	if len(u.Codecs) == 0 {
		u.Codecs = []uuid.UUID{}
	}

	ok := reflect.DeepEqual(u, v)
	if !ok {
		log.Printf("expext === %#v", v)
		log.Printf("actual === %#v", u)
	}

	return ok
}

func main() {
	rand.Seed(time.Now().Unix())
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
