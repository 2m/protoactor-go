package couchbase_persistence

import (
	"log"
	"sync"

	"github.com/couchbase/gocb"
	proto "github.com/golang/protobuf/proto"
)

type cbState struct {
	*Provider
	wg sync.WaitGroup
}

func (state *cbState) Restart() {
	//wait for any pending writes to complete
	state.wg.Wait()
}

func (state *cbState) GetEvents(actorName string, eventIndexStart int, callback func(event interface{})) {
	q := gocb.NewN1qlQuery("SELECT b.* FROM `" + state.bucketName + "` b WHERE meta(b).id >= $1 and meta(b).id <= $2")
	q.Consistency(gocb.RequestPlus)
	var p []interface{}
	p = append(p, formatEventKey(actorName, eventIndexStart))
	p = append(p, formatEventKey(actorName, 9999999999))

	rows, err := state.bucket.ExecuteN1qlQuery(q, p)
	if err != nil {
		log.Fatalf("Error executing N1ql: %v", err)
	}
	defer rows.Close()
	var row envelope

	for rows.Next(&row) {
		e := row.message()
		callback(e)
	}
}

func (state *cbState) GetSnapshot(actorName string) (snapshot interface{}, eventIndex int, ok bool) {

	q := gocb.NewN1qlQuery("SELECT b.* FROM `" + state.bucketName + "` b WHERE meta(b).id >= $1 and meta(b).id <= $2 order by b.eventIndex desc limit 1")
	var p []interface{}
	p = append(p, formatSnapshotKey(actorName, 0))
	p = append(p, formatSnapshotKey(actorName, 9999999999))

	rows, err := state.bucket.ExecuteN1qlQuery(q, p)
	if err != nil {
		log.Fatalf("Error executing N1ql: %v", err)
	}
	defer rows.Close()
	var row envelope
	if rows.Next(&row) {
		return row.message(), row.EventIndex, true
	}
	return nil, 0, false
}
func (provider *Provider) GetSnapshotInterval() int {
	return provider.snapshotInterval
}

func (state *cbState) PersistEvent(actorName string, eventIndex int, event proto.Message) {
	key := formatEventKey(actorName, eventIndex)
	envelope := newEnvelope(event, "event", eventIndex)
	state.wg.Add(1)
	persist := func() {
		_, err := state.bucket.Insert(key, envelope, 0)

		if err != nil {
			log.Fatal(err)
		}
		state.wg.Done()
	}
	if state.async {
		state.writer.Tell(&write{fun: persist})
	} else {
		persist()
	}
}

func (state *cbState) PersistSnapshot(actorName string, eventIndex int, snapshot proto.Message) {
	key := formatSnapshotKey(actorName, eventIndex)
	envelope := newEnvelope(snapshot, "snapshot", eventIndex)
	persist := func() {
		_, err := state.bucket.Insert(key, envelope, 0)
		if err != nil {
			log.Fatal(err)
		}
	}
	if state.async {
		state.writer.Tell(&write{fun: persist})
	} else {
		persist()
	}
}
