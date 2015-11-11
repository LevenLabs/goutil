package main

import (
	"fmt"
	"log"
	"sort"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Member struct {
	ID    int    `bson:"_id"`
	Name  string `bson:"name"`
	State string `bson:"stateStr"`
	Self  bool   `bson:"self"`
}

type Members []Member

func (m Members) Len() int {
	return len(m)
}

func (m Members) Less(i, j int) bool {
	return m[i].ID < m[j].ID
}

func (m Members) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

type ReplSetStatus struct {
	Members Members `bson:"members"`
}

func main() {
	// Both connect=direct and monotonic must be set in order to actually only
	// talk to the localhost instance, otherwise mgo will try to discover the
	// rest of the cluster and talk to that
	db, err := mgo.Dial("127.0.0.1:27017?connect=direct")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMode(mgo.Monotonic, true)

	var res ReplSetStatus
	if err := db.Run(bson.D{{"replSetGetStatus", 1}}, &res); err != nil {
		log.Fatal(err)
	}
	members := res.Members

	// We sort by id so all servers running this will behave exactly the same
	sort.Sort(members)

	for _, m := range members {
		if m.State != "SECONDARY" {
			continue
		}
		// This will be the first secondary we run into in the list
		if m.Self {
			fmt.Println("first secondary")
		}
		return
	}
}
