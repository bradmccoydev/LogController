package logger

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQSAPI - bits from the SQS interface we need
type SQSAPI interface {
	DeleteMessage(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
	GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

// NewSQS returns a new SQS handler
func NewSQS(c SQSAPI) *SQSHandler {
	return &SQSHandler{sqs: c}
}

// SQSHandler points back to the interface
type SQSHandler struct {
	sqs SQSAPI
}

// performDelete removes a message from the queue
func (s *SQSHandler) performDelete(url string, id string) error {

	// Build the params
	params := &sqs.DeleteMessageInput{}
	params = params.SetQueueUrl(url)
	params = params.SetReceiptHandle(id)

	// Submit to SQS
	_, err := s.sqs.DeleteMessage(params)

	// Return
	return err
}

// lookupURL handles retrieving the sqs queue url
func (s *SQSHandler) lookupURL(qname string) (string, error) {

	// Build the query params
	params := &sqs.GetQueueUrlInput{}
	params = params.SetQueueName(qname)

	// Fetch the SQS queue url
	procQURL, err := s.sqs.GetQueueUrl(params)
	if err != nil {
		return "", err
	}
	url := *procQURL.QueueUrl

	// Return
	return url, nil
}

// performSend submits a message onto the queue
func (s *SQSHandler) performSend(msg events.SQSMessage, url string) error {

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
	params = params.SetQueueUrl(url)
	params = params.SetMessageBody(msg.Body)
	params = params.SetMessageAttributes(msgAttribVals)

	// Submit to SQS
	_, err := s.sqs.SendMessage(params)

	// Return
	return err
}
