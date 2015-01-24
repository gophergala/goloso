package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bitly/go-nsq"
)

var (
	showHelp = flag.Bool("help", false, "print help")
	topic    = flag.String("topic", "", "NSQ topic")
	channel  = flag.String("channel", "", "NSQ channel")
)

// bootstrap event struct
/*
type NSQMessage struct {
	Event      []string `json:event`      // event type
	Uuid       []string `json:uuid`       // event uuid
	InstanceId []string `json:instanceid` // instance id
	IpAddress  []string `json:ipaddress`  // ipaddess
	Os         []string `json:os`         // operaring system
}
*/

func main() {
	fmt.Println("Goloso")

	flag.Parse()

	if *showHelp {
		fmt.Println(`
Usage:
    goloso --help

    goloso --channel "orc.sys.events" --topic "ec2"
`)
		os.Exit(0)
	}

	if *channel == "" {
		log.Fatalln("Err: missing channel")
	}

	if *topic == "" {
		log.Fatalln("Err: missing topic. \"--topic is required\"")
	}

	var (
		consumer *nsq.Consumer
		err      error
	)

	lookup := "localhost:4161"

	// setup nsq config
	conf := nsq.NewConfig()
	conf.MaxInFlight = 1000

	// setup nsq consumer
	consumer, err = nsq.NewConsumer(*topic, *channel, conf)
	if err != nil {
		log.Fatalln("Err: can't consume", err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("Message; %v", message)
		return nil
	}))

	err = consumer.ConnectToNSQLookupd(lookup)
	if err != nil {
		log.Fatalln("Err: can't connect to lookupd", err)
	}

}
