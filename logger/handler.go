package logger

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

// Constants
const (
	// Name of the "Application" table in DynamoDB
	applicationTableName string = "application"
	// Name of the Log Controller SQS queue
	logControllerQueueName string = "logging_queue.fifo"
	// Name of the SQS message attribute that stores the application
	messageAttribAppName string = "APPLICATION_NAME"
	// Name of the SQS message attribute that stores the application version
	messageAttribAppVers string = "APPLICATION_VERS"
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

			// Log to Cloudwatch then skip
		}

		// Send the message to the processor
		err = submitMessage(h, msg, procname)
		if err != nil {

			// Log to Cloudwatch then skip
		}

		// Delete the message
		err = deleteMessage(h, msg)
		if err != nil {

			// Log to Cloudwatch
		}
	}

	// Return
	return nil
}

// getQueueName - queries DynamoDB to retrieve the downstream queue name to use
func getQueueName(h *Handler, msg events.SQSMessage) (string, error) {

	// Get the application key details
	msgattribs := msg.MessageAttributes
	logapp := *msgattribs[messageAttribAppName].StringValue
	logvers := *msgattribs[messageAttribAppVers].StringValue

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
