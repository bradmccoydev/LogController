package logger

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// Constants
const (
	// Name of the "Application" table in DynamoDB
	applicationTableName string = "application"
	// Name of the Log Controller SQS queue
	logControllerQueueName string = "logging_queue.fifo"
	// MessageAttribAppName - SQS message attribute that stores the application
	MessageAttribAppName string = "APPLICATION_NAME"
	// MessageAttribAppVers - SQS message attribute that stores the application version
	MessageAttribAppVers string = "APPLICATION_VERS"
)

// Handler handles incoming logger requests.
type Handler struct {
	ddb DynamoDBAPI
	sqs SQSAPI
}

// Item - represents the application table
type Item struct {
	application string
	version     string
	loghandler  string
}

// NewHandler initializes and returns a new Handler.
func NewHandler(ddb DynamoDBAPI, sqs SQSAPI) *Handler {
	return &Handler{ddb: ddb, sqs: sqs}
}

// Handle handles the logger request.
func (h *Handler) Handle(ctx context.Context, sqsEvent events.SQSEvent) error {

	// Do we have anything to process?
	if len(sqsEvent.Records) == 0 {
		return nil
	}

	// Iterate through the messages
	for _, msg := range sqsEvent.Records {

		// Determine Log Processor
		procname, err := getQueueName(h, msg)
		if err != nil {

			// Log details
			fmt.Printf("Unable to determine queue for message ID: %v\n", msg.MessageId)
			fmt.Println("Error reported: ", err)
			fmt.Println("Skipping further processing of this message")

			// Skip from processing
			continue
		}

		// Send the message to the processor
		err = submitMessage(h, msg, procname)
		if err != nil {

			// Log details
			fmt.Printf("Unable to submit message to queue: %v for message ID: %v\n", procname, msg.MessageId)
			fmt.Println("Error reported: ", err)
			fmt.Println("Skipping further processing of this message")

			// Skip from processing
			continue
		}

		// Delete the message
		err = deleteMessage(h, msg)
		if err != nil {

			// Log details
			fmt.Printf("Unable to delete message ID: %v\n", msg.MessageId)
			fmt.Println("Error reported: ", err)
		}
	}

	// Return
	return nil
}

// getQueueName - queries DynamoDB to retrieve the downstream queue name to use
func getQueueName(h *Handler, msg events.SQSMessage) (string, error) {

	// Grab the message attributes
	msgattribs := msg.MessageAttributes
	if msgattribs == nil {
		return "", newErrorMessageAttributesNil()
	}

	// Get the application version details
	logapp := *msgattribs[MessageAttribAppName].StringValue
	logvers := *msgattribs[MessageAttribAppVers].StringValue
	if logapp == "" {
		return "", newErrorMessageAttributesAppNameEmpty()
	}
	if logvers == "" {
		return "", newErrorMessageAttributesAppVersionEmpty()
	}

	// Fetch the application record from DynamoDB
	db := NewDB(h.ddb)
	item, err := db.performGet(logapp, logvers)
	if err != nil {
		return "", err
	}

	// Return the processor queue name
	qname := item.loghandler
	if qname == "" {
		return "", newErrorUnableToFetchProcQueueName()
	}
	return qname, nil
}

// submitMessage - publishes a message to the log processor queue
func submitMessage(h *Handler, msg events.SQSMessage, queue string) error {

	// Get the queue url
	sqsc := NewSQS(h.sqs)
	url, err := sqsc.lookupURL(queue)
	if err != nil {
		return err
	}

	// Submit the message
	err = sqsc.performSend(msg, url)
	return err
}

// deleteMessage - deletes the message from the log controler queue
func deleteMessage(h *Handler, msg events.SQSMessage) error {

	// Get the queue url
	sqsc := NewSQS(h.sqs)
	url, err := sqsc.lookupURL(logControllerQueueName)
	if err != nil {
		return err
	}

	// Delete the message
	err = sqsc.performDelete(url, msg.ReceiptHandle)
	return err
}
