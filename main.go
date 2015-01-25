package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/bitly/go-nsq"
	"github.com/bitly/nsq/util"
	"github.com/boltdb/bolt"
)

var (
	showHelp    = flag.Bool("help", false, "print help")
	showVersion = flag.Bool("version", false, "print version")
	topic       = flag.String("topic", "", "NSQ topic")
	channel     = flag.String("channel", "", "NSQ channel")
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

// cmdline struct which will call Ansible playbook
type Worker struct {
	Command string
	Args    string
	Output  chan string
}

// the cmdline runner
func (cmd *Worker) Run() {
	out, err := exec.Command(cmd.Command, cmd.Args).Output()
	if err != nil {
		log.Fatalln("Err: command execution failed!", err)
	}

	cmd.Output <- string(out)
}

// shity output
func Collect(c chan string) {
	for {
		msg := <-c
		fmt.Printf("The command result is %s\n", msg)
	}
}

// not void but near... well, golang super newbies here...
// shit gets printed but never stops :)
// looks like I am missing a os.Interrupt somewhere...
func loop(inChan chan *nsq.Message) {

	// c := make(chan string)

	for msg := range inChan {

		fmt.Println(string(msg.Body) + "\n")

		// ansible := &Worker{Command: "echo", Args: "hi", Output: c}
		// go ansible.Run()

		msg.Finish()
	}
	// go Collect(c)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if *showHelp {
		fmt.Println(`
Usage:
    goloso --help
    goloso --version

    goloso --channel "orc.sys.events" --topic "ec2"
`)
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("Goloso v%s\n", util.BINARY_VERSION)
		os.Exit(0)
	}

	fmt.Println("Goloso.. starting")

	if *channel == "" {
		log.Fatalln("Err: missing channel")
	}

	if *topic == "" {
		log.Fatalln("Err: missing topic. \"--topic is required\"")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	var (
		consumer *nsq.Consumer
		err      error
	)

	// connect to database
	fmt.Print("Connecting to bolt...")

	// setup bolt db connection
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("done")

	// create buquete
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("Goloso"))
		if err != nil {
			return fmt.Errorf("Create bucket: %s", err)
			fmt.Println("Goloso exists")
		}

		fmt.Println("Goloso bucket created")

		return nil
	})

	// setup nsq config
	conf := nsq.NewConfig()
	conf.MaxInFlight = 1000

	// setup nsq consumer
	consumer, err = nsq.NewConsumer(*topic, *channel, conf)
	if err != nil {
		log.Fatalln("Err: can't consume", err)
	}

	inChan := make(chan *nsq.Message)

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		inChan <- message
		// log.Printf("Message: %s", message)

		// json_decode && store in KV

		// db.Update(func(tx *bolt.Tx) error {
		// 	b := tx.Bucket([]byte("Goloso"))
		// 	err := b.Put([]byte("answer"), []byte("42"))
		// 	return err
		// })

		return nil
	}))

	// someday this will be set with consumerOpts
	lookup := "localhost:4161"

	err = consumer.ConnectToNSQLookupd(lookup)
	if err != nil {
		log.Fatalln("Err: can't connect to lookupd", err)
	}

	// the code below actually works but I think we need to put goroutines here instead
	/*
		for {
			select {
			case <-consumer.StopChan:
				return
			case <-sigChan:
				consumer.Stop()
			}
		}
	*/

	go loop(inChan)
	<-consumer.StopChan
}
