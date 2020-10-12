package lambda

import "errors"

func newErrorMessageAttributesNil() error {
	return errors.New("No message attributes provided with SQS message")
}

func newErrorMessageAttributesAppNameEmpty() error {
	return errors.New("No application name message attribute provided")
}

func newErrorMessageAttributesAppVersionEmpty() error {
	return errors.New("No application version message attribute provided")
}

func newErrorUnableToFindAppVersion() error {
	return errors.New("Unable to find a matching application version")
}

func newErrorUnableToFetchProcQueueName() error {
	return errors.New("Unable to fetch the log processor queue name from DynamoDB")
}
