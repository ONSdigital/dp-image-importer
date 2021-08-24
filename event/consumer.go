package event

import (
	"context"

	"github.com/ONSdigital/dp-image-importer/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out mock/handler.go -pkg mock . Handler

// Handler represents a handler for processing a single event.
type Handler interface {
	Handle(ctx context.Context, ImageUploaded *ImageUploaded) error
}

// Consume converts messages to event instances, and pass the event to the provided handler.
func Consume(ctx context.Context, cg kafka.IConsumerGroup, handler Handler, numWorkers int) {
	// func to be executed by each worker in a goroutine
	workerConsume := func(workerNum int) {
		for {
			select {
			case message, ok := <-cg.Channels().Upstream:
				logData := log.Data{"message_offset": message.Offset(), "worker_num": workerNum}
				if !ok {
					log.Event(ctx, "upstream channel closed - closing event consumer loop", log.INFO, logData)
					return
				}

				err := processMessage(ctx, message, handler)
				if err != nil {
					log.Event(ctx, "failed to process message", log.ERROR, log.Error(err), logData)
				}
				log.Event(ctx, "message committed", log.INFO, logData)

				message.Release()
				log.Event(ctx, "message released", log.INFO, logData)

			case <-cg.Channels().Closer:
				log.Event(ctx, "closing event consumer loop because closer channel is closed", log.Data{"worker_num": workerNum}, log.INFO)
				return
			}
		}
	}

	// workers to consume messages in parallel
	for w := 1; w <= numWorkers; w++ {
		go workerConsume(w)
	}
}

// processMessage unmarshals the provided kafka message into an event and calls the handler.
// After the message is successfully handled, it is committed.
func processMessage(ctx context.Context, message kafka.Message, handler Handler) error {
	defer message.Commit()

	logData := log.Data{"message_offset": message.Offset()}

	event, err := unmarshal(message)
	if err != nil {
		log.Event(ctx, "failed to unmarshal event", log.ERROR, logData, log.Error(err))
		return err
	}

	logData["event"] = event

	log.Event(ctx, "event received", log.INFO, logData)

	err = handler.Handle(ctx, event)
	if err != nil {
		log.Event(ctx, "failed to handle event", log.ERROR, log.Error(err))
		return err
	}

	return nil
}

// unmarshal converts a event instance to []byte.
func unmarshal(message kafka.Message) (*ImageUploaded, error) {
	var event ImageUploaded
	err := schema.ImageUploadedEvent.Unmarshal(message.GetData(), &event)
	return &event, err
}
