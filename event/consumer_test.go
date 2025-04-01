package event_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-image-importer/event"
	"github.com/ONSdigital/dp-image-importer/event/mock"
	"github.com/ONSdigital/dp-image-importer/schema"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-kafka/v2/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
)

var testCtx = context.Background()

var errHandler = errors.New("handler error")

var testEvent = event.ImageUploaded{
	ImageID:  "123",
	Filename: "Filename.png",
	Path:     "1234-uploadpng",
}

var numKafkaWorkers = 1

// kafkaStubConsumer mock which exposes Channels function returning empty channels
// to be used on tests that are not supposed to receive any kafka message
var kafkaStubConsumer = &kafkatest.IConsumerGroupMock{
	ChannelsFunc: func() *kafka.ConsumerGroupChannels {
		return &kafka.ConsumerGroupChannels{}
	},
}

func TestConsume(t *testing.T) {

	Convey("Given kafka consumer and event handler mocks", t, func() {
		// Create mock kafka consumer with upstream channel with 2 buffered messages
		cgChannels := &kafka.ConsumerGroupChannels{
			Upstream: make(chan kafka.Message, 2),
		}
		mockConsumer := kafkatest.NewMessageConsumerWithChannels(cgChannels, true)

		handlerWg := &sync.WaitGroup{}
		mockEventHandler := &mock.HandlerMock{
			HandleFunc: func(ctx context.Context, event *event.ImageUploaded) error {
				defer handlerWg.Done()
				return nil
			},
		}

		Convey("And a kafka message with the valid schema being sent to the Upstream channel", func() {

			message := kafkatest.NewMessage(marshal(testEvent), 0)
			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {
				handlerWg.Add(1)
				event.Consume(testCtx, mockConsumer, mockEventHandler, numKafkaWorkers)
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler", func() {
					So(mockEventHandler.HandleCalls(), ShouldHaveLength, 1)
					So(*mockEventHandler.HandleCalls()[0].ImageUploaded, ShouldResemble, testEvent)
				})

				Convey("Then the message is committed and the consumer is released", func() {
					<-message.UpstreamDone()
					So(message.CommitCalls(), ShouldHaveLength, 1)
					So(message.ReleaseCalls(), ShouldHaveLength, 1)
				})
			})
		})

		Convey("And two kafka messages, one with a valid schema and one with an invalid schema", func() {

			validMessage := kafkatest.NewMessage(marshal(testEvent), 1)
			invalidMessage := kafkatest.NewMessage([]byte("invalid schema"), 0)
			mockConsumer.Channels().Upstream <- invalidMessage
			mockConsumer.Channels().Upstream <- validMessage

			Convey("When consume messages is called", func() {
				handlerWg.Add(1)
				event.Consume(testCtx, mockConsumer, mockEventHandler, numKafkaWorkers)
				handlerWg.Wait()

				Convey("Then only the valid event is sent to the mockEventHandler ", func() {
					So(mockEventHandler.HandleCalls(), ShouldHaveLength, 1)
					So(*mockEventHandler.HandleCalls()[0].ImageUploaded, ShouldResemble, testEvent)
				})

				Convey("Then both valid and invalid messages are committed and the consumer is released for both messages", func() {
					<-validMessage.UpstreamDone()
					<-invalidMessage.UpstreamDone()
					So(validMessage.CommitCalls(), ShouldHaveLength, 1)
					So(invalidMessage.CommitCalls(), ShouldHaveLength, 1)
					So(validMessage.ReleaseCalls(), ShouldHaveLength, 1)
					So(invalidMessage.ReleaseCalls(), ShouldHaveLength, 1)
				})
			})
		})

		Convey("With a failing handler and a kafka message with the valid schema being sent to the Upstream channel", func() {
			mockEventHandler.HandleFunc = func(ctx context.Context, event *event.ImageUploaded) error {
				defer handlerWg.Done()
				return errHandler
			}
			message := kafkatest.NewMessage(marshal(testEvent), 0)
			mockConsumer.Channels().Upstream <- message

			Convey("When consume message is called", func() {
				handlerWg.Add(1)
				event.Consume(testCtx, mockConsumer, mockEventHandler, numKafkaWorkers)
				handlerWg.Wait()

				Convey("An event is sent to the mockEventHandler ", func() {
					So(mockEventHandler.HandleCalls(), ShouldHaveLength, 1)
					So(*mockEventHandler.HandleCalls()[0].ImageUploaded, ShouldResemble, testEvent)
				})

				Convey("The message is committed and the consumer is released", func() {
					<-message.UpstreamDone()
					So(message.CommitCalls(), ShouldHaveLength, 1)
					So(message.ReleaseCalls(), ShouldHaveLength, 1)
					// TODO in this case, once we have an error reported, we should validate that the error is correctly reported.
				})
			})
		})
	})
}

// marshal helper method to marshal a event into a []byte
func marshal(event event.ImageUploaded) []byte {
	bytes, err := schema.ImageUploadedEvent.Marshal(event)
	So(err, ShouldBeNil)
	return bytes
}
