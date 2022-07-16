package main

import (
	"github.com/nats-io/nats.go"
	"log"
	"sync/atomic"
	"time"
)

const URL = nats.DefaultURL

// INFO[0000] Listening on [swl_backend], clientID=[swl-backend_undroid-home-lan_U5iqcr8h09]  broadcast=false
// INFO[0000] Listening on [swl_backend:undroid.home.lan], clientID=[swl-backend_undroid-home-lan_U5iqcr8h09]  broadcast=false
// INFO[0000] Listening on [swl_backend:undroid.home.lan:U5iqcr8h09], clientID=[swl-backend_undroid-home-lan_U5iqcr8h09]  broadcast=false
// INFO[0000] Listening on [@swl_backend], clientID=[swl-backend_undroid-home-lan_U5iqcr8h09]  broadcast=true
// INFO[0000] Listening on [@swl_backend:undroid.home.lan], clientID=[swl-backend_undroid-home-lan_U5iqcr8h09]  broadcast=true

func main() {
	nc, err := nats.Connect(URL)
	if err != nil {
		log.Fatal(1, err)
	}
	defer nc.Close()

	//err = nc.Publish("swl_backend:undroid.home.lan")
}

func main1() {
	nc, err := nats.Connect(URL)
	if err != nil {
		log.Fatal(1, err)
	}
	defer nc.Close()

	//// Create JetStream Context
	//js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	//if err != nil {
	//	log.Fatal(2, err)
	//}
	//_, err = js.AddStream(&nats.StreamConfig{
	//	Name: "IVA",
	//	//Subjects: []string{"ORDERS", "ORDERS:HOST", "ORDERS:HOST:PORT1"},
	//})
	//if err != nil {
	//	log.Fatal(22, err)
	//}

	var n int64
	sub1, err := nc.QueueSubscribe("ORDERS:HOST", "ORDERS:HOST", func(m *nats.Msg) {
		log.Printf("[%d] %s: %s", atomic.AddInt64(&n, 1), m.Subject, m.Data)
	})
	if err != nil {
		log.Fatal(4, err)
	}

	// Simple   Publisher
	for i := 0; i < 5; i++ {
		err = nc.Publish("ORDERS:HOST", []byte(time.Now().String()))
		if err != nil {
			log.Fatal(3, err)
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(22 * time.Second)

	err = sub1.Unsubscribe()
	if err != nil {
		log.Fatal(5, err)
	}
}
