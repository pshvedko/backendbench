package main

import (
	"encoding/json"
	gosocketio "github.com/ambelovsky/gosf-socketio"
	"log"
	"math/rand"
	"reflect"
	"sync"
	"time"

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
	Params interface{} `json:"params"`
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

type Clause struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Op    string      `json:"op"`
}

type ListParams struct {
	Filter []Clause `json:"filter"`
	Limit  int      `json:"limit,omitempty"`
}

func equalUpdateCodecGroupParams(r, v interface{}) bool {
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

type Client struct {
	*gosocketio.Client
	sync.WaitGroup
	n   int
	e   int
	fly chan struct{}
}

func (c *Client) onLogin(_ *gosocketio.Channel, reply Token) {
	log.Printf("login %#v", reply)
	c.Done()
}

func (c *Client) onResponse(_ *gosocketio.Channel, reply Reply) {
	log.Printf("response %d %#v", c.n, reply)
	if reply.Error != nil {
		c.e++
	}
	c.n++
	<-c.fly
	c.Done()
}

func (c *Client) Login(email, password string) error {
	c.Add(1)
	err := c.Emit("login", Login{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return err
	}
	c.Wait()
	return nil
}

func NewClient(url string, fly int) (*Client, error) {
	c, err := gosocketio.Dial(url,
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
		return nil, err
	}
	return &Client{
		Client: c,
		fly:    make(chan struct{}, fly),
	}, nil
}

func run() error {
	c, err := NewClient("ws://localhost:8080/socket.io/?EIO=3&transport=websocket", 99)
	if err != nil {
		return err
	}
	defer c.Close()

	err = c.On("login", c.onLogin)
	if err != nil {
		return err
	}
	err = c.On("response", c.onResponse)
	if err != nil {
		return err
	}
	err = c.On("event", c.onResponse)
	if err != nil {
		return err
	}

	err = c.Login("test@example.com", "12345")
	if err != nil {
		return err
	}

	//n, e, x, z := 0, 0, 0, 0
	//
	//var tt sync.Map
	//
	//c.fly = make(chan struct{}, 99)
	//
	//var gid1,
	//	cid1,
	//	cid2,
	//	cid3 uuid.UUID
	//
	//err = ws.On("response", func(c *gosocketio.Channel, reply Reply) {
	//	log.Printf("response %d %#v", n, reply)
	//	if reply.Error != nil {
	//		e++
	//	} else if reply.Method == "updateCodecGroup" {
	//		v, ok := tt.LoadAndDelete(reply.Id)
	//		if !ok {
	//			x++
	//		} else if !equalUpdateCodecGroupParams(reply.Result, v) {
	//			z++
	//		}
	//	} else if reply.Method == "listCodecGroups" {
	//		gid1 = reply.Result
	//	}
	//	n++
	//	wg.Done()
	//	<-fly
	//})
	//if err != nil {
	//	return
	//}
	//
	//wg.Add(1)
	//
	//
	//{
	//	wg.Add(1)
	//
	//	err = ws.Emit("query", Query{
	//		Id:     uuid.New(),
	//		Method: "listCodecGroups",
	//		Params: ListParams{
	//			Filter: []Clause{{
	//				Field: "type",
	//				Value: "audio",
	//				Op:    "EQUALS",
	//			}},
	//		},
	//	})
	//	if err != nil {
	//		return
	//	}
	//
	//	wg.Wait()
	//}
	//
	//{
	//	for i := 0; i < 1; i++ {
	//
	//		wg.Add(1)
	//
	//		codecs := func(uuids ...uuid.UUID) []uuid.UUID {
	//			sort.Slice(uuids, func(i, j int) bool {
	//				return rand.Int()&1 == 1
	//			})
	//			return uuids[:i%4]
	//		}(cid1, cid2, cid3)
	//
	//		params := UpdateCodecGroupParams{
	//			Id:          gid1,
	//			Name:        "default-audio-codec-group",
	//			Type:        "audio",
	//			DtmfRfc2833: true,
	//			DtmfInband:  false,
	//			DtmfSipInfo: false,
	//			Codecs:      codecs,
	//		}
	//
	//		id := uuid.New()
	//
	//		tt.Store(id, params)
	//
	//		fly <- struct{}{}
	//
	//		err = ws.Emit("query", Query{
	//			Id:     id,
	//			Method: "updateCodecGroup",
	//			Params: params,
	//		})
	//		if err != nil {
	//			return
	//		}
	//	}
	//
	//	wg.Wait()
	//}
	//
	//{
	//	for i := 0; i < 1; i++ {
	//
	//		wg.Add(1)
	//
	//		fly <- struct{}{}
	//
	//		err = ws.Emit("query", Query{
	//			Id:     uuid.New(),
	//			Method: "listVoicemails",
	//			Params: ListParams{
	//				Filter: []Clause{},
	//			},
	//		})
	//		if err != nil {
	//			return
	//		}
	//	}
	//
	//	wg.Wait()
	//}
	//
	//log.Println()
	//log.Println("ERRORS", e, z, x)
	//log.Println()

	return nil
}

func main() {
	rand.Seed(time.Now().Unix())
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
