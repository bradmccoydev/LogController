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

func newErrorMessageAttributesAppNameEmpty() error {
	return errors.New("No application name message attribute provided")
}

func newErrorMessageAttributesAppVersionEmpty() error {
	return errors.New("No application version message attribute provided")
}

func newErrorMessageAttributesLogLevelEmpty() error {
	return errors.New("No log level message attribute provided")
}

func newErrorMessageAttributesTimestampEmpty() error {
	return errors.New("No timestamp message attribute provided")
}

func newErrorMessageAttributesTrackingIDEmpty() error {
	return errors.New("No tracking id message attribute provided")
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
