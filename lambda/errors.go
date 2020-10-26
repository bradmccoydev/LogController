package lambda

import (
	"errors"
	"fmt"
)

func newErrorEnvironmentVariableProvided(varname string) error {
	return fmt.Errorf("The environment variable %s was not provided. Aborting", varname)
}

func newErrorMessageAttributesNil() error {
	return errors.New("No message attributes provided with SQS message")
}

func newErrorMessageAttributeMissing(attribname string) error {
	return fmt.Errorf("The message attribute %s was not provided", attribname)
}

func newErrorMessageAttributeEmpty(attribname string) error {
	return fmt.Errorf("The message attribute %s did not have a value", attribname)
}

func newErrorUnableToFindAppVersion() error {
	return errors.New("Unable to find a matching application version")
}

func newErrorUnableToFetchControllerQueueURL() error {
	return errors.New("Unable to fetch the log controller queue url")
}

func newErrorUnableToFetchProcQueueName() error {
	return errors.New("Unable to fetch the log processor queue name from DynamoDB")
}

func newErrorUnableToFetchProcQueueURL() error {
	return errors.New("Unable to fetch the log processor queue url")
}
