package event

import (
	"context"

	"github.com/ONSdigital/dp-image-importer/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/log.go/v2/log"
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
					log.Info(ctx, "upstream channel closed - closing event consumer loop", logData)
					return
				}

				err := processMessage(ctx, message, handler)
				if err != nil {
					log.Error(ctx, "failed to process message", err, logData)
				}
				log.Info(ctx, "message committed", logData)

				message.Release()
				log.Info(ctx, "message released", logData)

			case <-cg.Channels().Closer:
				log.Info(ctx, "closing event consumer loop because closer channel is closed", log.Data{"worker_num": workerNum})
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
		log.Error(ctx, "failed to unmarshal event", err, logData)
		return err
	}

	logData["event"] = event

	log.Info(ctx, "event received", logData)

	err = handler.Handle(ctx, event)
	if err != nil {
		log.Error(ctx, "failed to handle event", err)
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
