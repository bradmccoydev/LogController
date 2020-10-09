package logger_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bradmccoydev/LogController/logger"
)

/*
*
*  Mock DynamoDB components
*
 */
type mockDynamoDB struct {
	getIn  *dynamodb.GetItemInput
	getOut *dynamodb.GetItemOutput
	err    error
}

func (m *mockDynamoDB) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return nil, m.err
}

/*
*
*  Mock SQS components
*
 */
type mockSQS struct {
	deleteIn  *sqs.DeleteMessageInput
	deleteOut *sqs.DeleteMessageOutput
	getUrlIn  *sqs.GetQueueUrlInput
	getUrlOut *sqs.GetQueueUrlOutput
	sendIn    *sqs.SendMessageInput
	sendOut   *sqs.SendMessageOutput
	err       error
}

func (m *mockSQS) DeleteMessage(*sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	return nil, m.err
}
func (m *mockSQS) GetQueueUrl(*sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	return nil, m.err
}
func (m *mockSQS) SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return nil, m.err
}

// TestMain routine for controlling setup/destruction for all tests in this package
func TestMain(m *testing.M) {

	// Run the various tests then exit
	exitVal := m.Run()
	os.Exit(exitVal)
}

// Test Handler
func TestHandler(t *testing.T) {

	// Setup simple log message
	msgNoAttribs := events.SQSMessage{
		MessageId:     "12345",
		ReceiptHandle: "Fred12345",
		Body:          "blah blah blah",
	}
	recordsNoAttribs := []events.SQSMessage{msgNoAttribs}

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
			request:       events.SQSEvent{Records: recordsNoAttribs},
			sqs:           &mockSQS{},
			ddb:           &mockDynamoDB{},
			errorExpected: false,
		},
	}

	// Iterate through the test data
	for _, test := range tests {

		t.Run(test.scenario, func(t *testing.T) {

			// Run the test
			h := logger.NewHandler(test.ddb, test.sqs)
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
