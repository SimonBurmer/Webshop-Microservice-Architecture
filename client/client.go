package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"gitlab.lrz.de/vss/semester/ob-21ws/blatt-2/blatt2-gruppe14/api"
	"google.golang.org/grpc"
)

type Client struct {
	Nats  *nats.EncodedConn
	Redis *redis.Client
}

func (c *Client) Scenarios(scenario string) {
	log.Printf("Run scenario: %s ", scenario)
	err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("run scenario: %s", scenario), Subject: "Client.Scenarios"})
	if err != nil {
		panic(err)
	}

	switch scenario {
	case "s1":
		c.scenario1()
	case "s2":
		c.scenario2()

	default:
		log.Fatalf("no scenario called: %s", scenario)
		err := c.Nats.Publish("log", api.Log{Message: fmt.Sprintf("no scenario called: %s", scenario), Subject: "Client.Scenarios"})
		if err != nil {
			panic(err)
		}
	}
}

// 1. Bestellung lagernder Produkte durch einen Neukunden bis zum Verschicken.
// 2. Bestellung von drei Produkten, von denen nur eines auf Lager ist, bis zum Verschicken.
//      Die beiden nicht lagernden Produkte werden zu verschiedenen Zeit- punkten, durch verschiedene Zulieferer, geliefert.
// 3.  Stornieren einer Bestellung, die noch nicht verschickt wurde.
// 4. Empfangen eines defekten Artikels aus einer Bestellung mit mehreren Artikeln als Retoure mit sofortiger Ersatzlieferung.
// 5. Empfangen eines defekten Artikels aus einer Bestellung mit mehreren Artikeln als Retoure mit Rückbuchung des entsprechenden Teilbetrages.

func (c *Client) scenario1() {
	// Testen der einzelnen Komponenten
	// 1. Customer service
	// 2. Payment service

	////////////////////////////
	// Kommunikation mit Customer:
	////////////////////////////
	// Mithilfe von Redis Verbindung zu Customer aufbauen
	customer_redisVal := c.Redis.Get(context.TODO(), "customer")
	if customer_redisVal == nil {
		log.Fatal("service not registered")
	}
	customer_address, err := customer_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	customer_conn, err := grpc.Dial(customer_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer customer_conn.Close()
	customer := api.NewCustomerClient(customer_conn)
	customer_ctx, customer_cancel := context.WithTimeout(context.Background(), time.Second)
	defer customer_cancel()

	// - Neuen Kunden erstellen
	customer_r, customer_err := customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Simon", Address: "Munich"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	customer_r, customer_err = customer.NewCustomer(customer_ctx, &api.NewCustomerRequest{Name: "Max", Address: "Berlin"})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Created customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	// - Kunde über ID anfordern
	customer_r, customer_err = customer.GetCustomer(customer_ctx, &api.GetCustomerRequest{Id: customer_r.GetId()})
	if customer_err != nil {
		log.Fatalf("Direct communication with customer failed: %v", customer_r)
	}
	log.Printf("Got customer: Name:%v, Address:%v, Id:%v", customer_r.GetName(), customer_r.GetAddress(), customer_r.GetId())

	////////////////////////////
	// Kommunikation mit Payment:
	////////////////////////////
	// Mithilfe von Redis Verbindung zu Payment aufbauen
	payment_redisVal := c.Redis.Get(context.TODO(), "payment")
	if payment_redisVal == nil {
		log.Fatal("service not registered")
	}
	payment_address, err := payment_redisVal.Result()
	if err != nil {
		log.Fatalf("error while trying to get the result %v", err)
	}
	payment_conn, err := grpc.Dial(payment_address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer payment_conn.Close()
	payment := api.NewPaymentClient(payment_conn)
	payment_ctx, payment_cancel := context.WithTimeout(context.Background(), time.Second)
	defer payment_cancel()

	// - Neues Payment erstellen
	payment_r, payment_err := payment.NewPayment(payment_ctx, &api.NewPaymentRequest{OrderId: 1, Value: 33.98})
	if payment_err != nil {
		log.Fatalf("direct communication with payment failed: %v", payment_r)
	}
	log.Printf("created payment: Id:%v, OrderId:%v, Value:%v", payment_r.GetId(), payment_r.GetOrderId(), payment_r.GetValue())

	// - Payment über ID anfordern
	payment_r, payment_err = payment.GetPayment(payment_ctx, &api.GetPaymentRequest{Id: payment_r.GetId()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("got payment: Id:%v, OrderId:%v, Value:%v", payment_r.GetId(), payment_r.GetOrderId(), payment_r.GetValue())

	// - Payment bezahlen
	payment_r, payment_err = payment.PayPayment(payment_ctx, &api.PayPaymentRequest{Id: payment_r.GetId(), Value: 33.98})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("payed payment: Id:%v, OrderId:%v, Value:%v", payment_r.GetId(), payment_r.GetOrderId(), payment_r.GetValue())

	// - Payment über ID Löschen
	payment_r, payment_err = payment.DeletePayment(payment_ctx, &api.DeletePaymentRequest{Id: payment_r.GetId()})
	if payment_err != nil {
		log.Fatalf("Direct communication with payment failed: %v", payment_r)
	}
	log.Printf("deleted payment: Id:%v, OrderId:%v, Value:%v", payment_r.GetId(), payment_r.GetOrderId(), payment_r.GetValue())

	time.Sleep(8 * time.Second)
}

func (c *Client) scenario2() {
	// Standartfall: Kund Bestellt lagernde Produkte
	// 1. Kunde erstellen
	// 2. Bestellung erstellen
	// 3. Auf Versandbestätigung warten

	log.Printf("Send xx to xx")

}
