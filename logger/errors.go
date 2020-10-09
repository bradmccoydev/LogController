package logger

import "errors"

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
