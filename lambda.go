package main

import (
	"context"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sqs"
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

// Message - refers to SQS message
type Message events.SQSMessage

// LogController interface
type LogController interface {
	GetQueueName(sess *session.Session) (string, error)
	SubmitMessage(sess *session.Session, queue *string) error
	DeleteMessage(sess *session.Session) error
}

// This is the entrypoint for the lambda
func main() {
	lambda.Start(Handler)
}

// Handler - the actual logic
func Handler(ctx context.Context, sqsEvent events.SQSEvent) error {

	// Do we have anything to process?
	if len(sqsEvent.Records) == 0 {
		return nil
	}

	// Get a session
	sess, err := session.NewSession()
	if err != nil {
		return newErrorCreatingSession()
	}

	// Iterate through the messages
	for _, rec := range sqsEvent.Records {

		// Cast sqs message to our local alias
		var msg Message = Message(rec)

		// Determine Log Processor
		procname, err := msg.GetQueueName(sess)
		if err != nil {

			// Log to Cloudwatch then skip
		}

		// Send the message to the processor
		err = msg.SubmitMessage(sess, &procname)
		if err != nil {

			// Log to Cloudwatch then skip
		}

		// Delete the message
		err = msg.DeleteMessage(sess)
		if err != nil {

			// Log to Cloudwatch
		}
	}

	// Return
	return nil
}

// GetQueueName - this function queries DynamoDB to retrieve
// the log processor queue name to use
func (msg Message) GetQueueName(sess *session.Session) (string, error) {

	// Application table item
	type Item struct {
		application string
		version     string
		loghandler  string
	}

	// Get the application details
	msgattribs := msg.MessageAttributes
	logapp := *msgattribs[messageAttribAppName].StringValue
	logvers := *msgattribs[messageAttribAppVers].StringValue

	// Build the query params
	params := &dynamodb.GetItemInput{
		TableName: aws.String(applicationTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"application": {
				S: aws.String(logapp),
			},
			"version": {
				S: aws.String(logvers),
			},
		},
	}

	// Query DynamoDB
	svc := dynamodb.New(sess)
	result, err := svc.GetItem(params)
	if err != nil {
		return "", err
	}
	if result.Item == nil {
		return "", newErrorUnableToFindAppVersion()
	}

	// Extract the table item
	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return "", newErrorUnableToUnmarshalDBItem()
	}

	// Return the processor queue name
	qname := item.loghandler
	if qname == "" {
		return "", newErrorUnableToFetchProcQueueName()
	}
	return qname, nil
}

// SubmitMessage - this function publishes a message to the log processor queue
func (msg Message) SubmitMessage(sess *session.Session, queue *string) error {

	// Create the SQS client
	svc := sqs.New(sess)

	// Get a the SQS queue url
	procQURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: queue})
	if err != nil {
		return err
	}
	procQueueURL := *procQURL.QueueUrl

	// Get the message attribute values
	var msgAttribVals map[string]*sqs.MessageAttributeValue
	msgAttribVals = make(map[string]*sqs.MessageAttributeValue)
	for key, val := range msg.MessageAttributes {
		var attribVal *sqs.MessageAttributeValue
		attribVal.SetDataType("String")
		attribVal.SetStringValue(*val.StringValue)
		msgAttribVals[key] = attribVal
	}

	// Build the params
	params := &sqs.SendMessageInput{}
	params = params.SetQueueUrl(procQueueURL)
	params = params.SetMessageBody(msg.Body)
	params = params.SetMessageAttributes(msgAttribVals)

	// Submit to SQS
	_, err = svc.SendMessage(params)
	return err
}

// DeleteMessage - this function deletes the message to the log controler queue
func (msg Message) DeleteMessage(sess *session.Session) error {

	// Create the SQS client
	svc := sqs.New(sess)

	// Get a the SQS queue url
	qParams := &sqs.GetQueueUrlInput{}
	qParams = qParams.SetQueueName(logControllerQueueName)
	qURL, err := svc.GetQueueUrl(qParams)
	if err != nil {
		return err
	}
	queueURL := *qURL.QueueUrl

	// Get the message receipt handle
	msghdl := msg.ReceiptHandle

	// Build the params
	dParams := &sqs.DeleteMessageInput{}
	dParams = dParams.SetQueueUrl(queueURL)
	dParams = dParams.SetReceiptHandle(msghdl)

	// Submit to SQS
	_, err = svc.DeleteMessage(dParams)
	return err
}

/*
Error functions
*/
func newErrorCreatingSession() error {
	return errors.New("Unable to create a session")
}

func newErrorUnableToFindAppVersion() error {
	return errors.New("Unable to find a matching application version")
}

func newErrorUnableToFetchProcQueueName() error {
	return errors.New("Unable to fetch the log processor queue name from DynamoDB")
}

func newErrorUnableToUnmarshalDBItem() error {
	return errors.New("Unable to unmarshal the application item from DynamoDB")
}
