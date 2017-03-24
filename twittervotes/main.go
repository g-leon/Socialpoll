package main

import (
	"gopkg.in/mgo.v2"
	"log"
	"github.com/bitly/go-nsq"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	db *mgo.Session
)

type poll struct {
	Options []string
}

func loadOptions() ([]string, error) {
	var options []string
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	var p poll
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

func dialdb() error {
	var err error
	log.Println("dialing mongodb: localhost")
	// TODO read connection string from config
	db, err = mgo.Dial("localhost")
	return err
}

func closedb() {
	db.Close()
	log.Println("closed database connection")
}

func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)

	// TODO read connection string from config
	pub, err := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	if err != nil {
		log.Println("can't create NSQ producer")
		stopchan <- struct{}{}
		return stopchan
	}

	go func () {
		for vote := range votes {
			pub.Publish("votes", []byte(vote)) // publish vote
		}
		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
		stopchan <- struct{}{}
	} ()

	return stopchan
}


func main() {
	// graceful shutdown on system signals
	var stoplock sync.Mutex // protects stop
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	if err := dialdb(); err != nil {
		log.Fatalln("failed to dial MongoDb:", err)
	}
	defer closedb()

	// start the system
	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)

	// This goroutine will call closeConn every minute
	// causing the connection to die and cause
	// readFromTwitter to be called all over again.
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}
