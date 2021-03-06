package db

import (
	"github.com/jinzhu/gorm"
	"github.com/notegio/openrelay/channels"
	"github.com/notegio/openrelay/types"
	"github.com/notegio/openrelay/common"
	"log"
)

type IndexConsumer struct {
	idx *Indexer
	s   common.Semaphore
}

func (consumer *IndexConsumer) Consume(msg channels.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Failed to index order: %v", r)
			msg.Reject()
		}
	}()
	consumer.s.Acquire()
	go func(){
		defer consumer.s.Release()
		orderBytes := [441]byte{}
		copy(orderBytes[:], []byte(msg.Payload()))
		order := types.OrderFromBytes(orderBytes)
		if err := consumer.idx.Index(order); err == nil {
			msg.Ack()
		} else {
			log.Printf("Failed to index order: '%v', '%v'", order.Hash(), err.Error())
			msg.Reject()
		}
	}()
}

func NewIndexConsumer(db *gorm.DB, status int64, concurrency int) *IndexConsumer {
	return &IndexConsumer{NewIndexer(db, status), make(common.Semaphore, concurrency)}
}
