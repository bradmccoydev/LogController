package lambda_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	loglambda "github.com/bradmccoydev/LogController/lambda"
)

// Constants
const (
	testapp     string = "fred"
	testappvers string = "1"
)

// Mock DynamoDB structure
type mockDynamoDB struct {
	getIn  *dynamodb.GetItemInput
	getOut *dynamodb.GetItemOutput
	err    error
}

// Mock DynamoDB GetItem
func (m *mockDynamoDB) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.getOut, m.err
}

// Mock SQS structure
type mockSQS struct {
	deleteIn  *sqs.DeleteMessageInput
	deleteOut *sqs.DeleteMessageOutput
	getUrlIn  *sqs.GetQueueUrlInput
	getUrlOut *sqs.GetQueueUrlOutput
	sendIn    *sqs.SendMessageInput
	sendOut   *sqs.SendMessageOutput
	err       error
}

// Mock SQS DeleteMessage
func (m *mockSQS) DeleteMessage(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return m.deleteOut, m.err
}

// Mock SQS GetQueueUrl
func (m *mockSQS) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return m.getUrlOut, m.err
}

// Mock SQS SendMessage
func (m *mockSQS) SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return m.sendOut, m.err
}

// TestMain routine for controlling setup/destruction for all tests in this package
func TestMain(m *testing.M) {

	// Run the various tests then exit
	exitVal := m.Run()
	os.Exit(exitVal)
}

// Test Handler
func TestHandler(t *testing.T) {

	// Message without attributes
	recordNoAttribs := []events.SQSMessage{
		events.SQSMessage{
			MessageId:     "12345",
			ReceiptHandle: "Fred12345",
			Body:          "blah blah blah",
		},
	}

	// Message with attributes
	var attribs map[string]events.SQSMessageAttribute
	attribs = make(map[string]events.SQSMessageAttribute)
	appAttrib := events.SQSMessageAttribute{StringValue: aws.String(testapp), DataType: "string"}
	versAttrib := events.SQSMessageAttribute{StringValue: aws.String(testappvers), DataType: "string"}
	attribs[loglambda.MessageAttribAppName] = appAttrib
	attribs[loglambda.MessageAttribAppVers] = versAttrib
	recordWithAttribs := []events.SQSMessage{
		events.SQSMessage{
			MessageId:         "12345",
			ReceiptHandle:     "Fred12345",
			Body:              "blah blah blah",
			MessageAttributes: attribs,
		},
	}

	// Setup test cases
	tests := []struct {
		scenario      string
		request       events.SQSEvent
		sqs           *mockSQS
		ddb           *mockDynamoDB
		errorExpected bool
	}{
		{
			scenario:      "No data",
			request:       events.SQSEvent{},
			sqs:           &mockSQS{},
			ddb:           &mockDynamoDB{},
			errorExpected: false,
		},
		{
			scenario:      "No message attributes",
			request:       events.SQSEvent{Records: recordNoAttribs},
			sqs:           &mockSQS{},
			ddb:           &mockDynamoDB{},
			errorExpected: false,
		},
		{
			scenario:      "With message attributes",
			request:       events.SQSEvent{Records: recordWithAttribs},
			sqs:           &mockSQS{},
			ddb:           &mockDynamoDB{getOut: &dynamodb.GetItemOutput{}},
			errorExpected: false,
		},
	}

	// Iterate through the test data
	for _, test := range tests {

		t.Run(test.scenario, func(t *testing.T) {

			// Run the test
			h := loglambda.NewHandler(test.ddb, test.sqs)
			err := h.Handle(context.Background(), test.request)
			if test.errorExpected {
				hasError(t, err)
			} else {
				noError(t, err)
			}
		})
	}
}

// Assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// Equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// HasError fails the test if an err is nil.
func hasError(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: expected an error!", filepath.Base(file), line)
		tb.FailNow()
	}
}

// NoError fails the test if an err is not nil.
func noError(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}
