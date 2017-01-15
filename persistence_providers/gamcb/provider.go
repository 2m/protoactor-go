package gamcb

import (
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/persistence"
	"github.com/AsynkronIT/protoactor-go/process"
	"github.com/couchbase/gocb"
)

type Provider struct {
	async            bool
	bucket           *gocb.Bucket
	bucketName       string
	snapshotInterval int
	writer           *process.PID
}

func (provider *Provider) GetState() persistence.ProviderState {
	return &cbState{
		Provider: provider,
	}
}

func New(bucketName string, baseU string, options ...CouchbaseOption) *Provider {
	c, err := gocb.Connect(baseU)
	if err != nil {
		log.Fatalf("Error connecting:  %v", err)
	}
	bucket, err := c.OpenBucketWithMt(bucketName, "")
	if err != nil {
		log.Fatalf("Error getting bucket:  %v", err)
	}
	bucket.SetTranscoder(transcoder{})

	config := &couchbaseConfig{}
	for _, option := range options {
		option(config)
	}

	provider := &Provider{
		snapshotInterval: config.snapshotInterval,
		async:            config.async,
		bucket:           bucket,
		bucketName:       bucketName,
	}

	if config.async {
		pid := actor.Spawn(actor.FromFunc(newWriter(time.Second / 10000)))
		provider.writer = pid
	}

	return provider
}
