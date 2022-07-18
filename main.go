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
	sync.Map
	n int
	e int
}

func (c *Client) onLogin(_ *gosocketio.Channel, reply Token) {
	log.Printf("login %v", reply)
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

type Callback func(map[string]interface{}, map[string]interface{}, map[string]interface{})

func (c *Client) onReply(_ *gosocketio.Channel, reply Reply) {
	log.Printf("[%08d] %v", c.n, reply)
	if reply.Error != nil {
		c.e++
	}
	c.n++
	v, ok := c.LoadAndDelete(reply.Id)
	if ok {
		switch v := v.(type) {
		case Callback:
			// TODO
			v(nil, nil, nil)
		default:
			panic(v)
		}
	}
	c.Done()
}

func (c *Client) Query(method string, params interface{}, callback Callback) error {
	c.Add(1)
	id := uuid.New()
	c.Store(id, callback)
	return c.Emit("query", Query{
		Id:     id,
		Method: method,
		Params: params,
	})
}

func NewClient(url string) (*Client, error) {
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
	}, nil
}

func run() error {
	c, err := NewClient("ws://localhost:8080/socket.io/?EIO=3&transport=websocket")
	if err != nil {
		return err
	}
	defer c.Close()

	err = c.On("login", c.onLogin)
	if err != nil {
		return err
	}
	err = c.On("response", c.onReply)
	if err != nil {
		return err
	}
	err = c.On("event", c.onReply)
	if err != nil {
		return err
	}

	err = c.Login("test@example.com", "12345")
	if err != nil {
		return err
	}

	fly := make(chan struct{}, 99)

	for i := 0; i < 9999; i++ {
		fly <- struct{}{}
		err = c.Query("listCodecGroups", ListParams{
			Filter: []Clause{{
				Field: "type",
				Value: "audio",
				Op:    "EQUALS",
			}},
		}, func(result, params, err map[string]interface{}) {
			<-fly
		})
		if err != nil {
			return err
		}
	}

	c.Wait()

	log.Println()
	log.Println("ERRORS", c.e, c.n)
	log.Println()

	return nil
}

func main() {
	rand.Seed(time.Now().Unix())
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
