package main

import (
	"fmt"
	"flag"
	"os"
	"log"
	"gopkg.in/mgo.v2"
	"sync"
	"github.com/bitly/go-nsq"
	"gopkg.in/mgo.v2/bson"
)

var (
	fatalErr error
	counts map[string]int
	countsLock sync.Mutex
)



func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}


// doCount checks to see whether there are any values in the counts map.
// If there aren't it will log that it is skipping the update and wait
// for next time.
func doCount(countsLock *sync.Mutex, counts *map[string]int, pollData *mgo.Collection) {
	countsLock.Lock()
	defer countsLock.Unlock()

	if len(*counts) == 0 {
		log.Println("No new votes, skipping database update")
		return
	}

	log.Println("Updating database...")
	log.Println(*counts)
	ok := true
	for option, count := range *counts {
		sel := bson.M{"options": bson.M{"$in": []string{option}}}
		up := bson.M{"$inc": bson.M{"results." + option:count}}

		if _, err := pollData.UpdateAll(sel, up); err != nil {
			log.Println("faile to update:", err)
			ok = false
		}
	}

	if ok {
		log.Println("Finished updating database...")
		*counts = nil // reset counts
	}

}

func main() {
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	log.Println("Connecting to the database...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
		return
	}

	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()
	pollData := db.DB("ballots").C("polls")

	log.Println("Connection to NSQ...")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}

	// votes handler
	q.AddHandler(nsq.HandlerFunc(func (m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))

	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return
	}

}


