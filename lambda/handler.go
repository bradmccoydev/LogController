package lambda

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

// Constants
const (
	// Name of the "Application" table in DynamoDB
	applicationTableName string = "application"

	// Name of the Lambda environment variable for logging level
	envNameLogLevel string = "LOG_LEVEL"

	// Name of the Lambda environment variable for the S3 bucket to use
	envNameS3Bucket string = "BUCKET"

	// Name of the Lambda environment variable for the path in the S3 bucket to use
	envNameS3BucketPath string = "PATH"

	// Name of the Lambda environment variable for the region of the S3 bucket to use
	envNameS3BucketRegion string = "REGION"

	// Name of the Log Controller SQS queue
	logControllerQueueName string = "logging_queue.fifo"

	// LogLevelDebug defines the debug log level
	LogLevelDebug string = "DEBUG"

	// LogLevelInfo defines the info log level
	LogLevelInfo string = "INFO"

	// MessageAttribAppName - SQS message attribute that stores the application
	MessageAttribAppName string = "APPLICATION_NAME"

	// MessageAttribAppVers - SQS message attribute that stores the application version
	MessageAttribAppVers string = "APPLICATION_VERS"
)

// Handler handles incoming logger requests.
type Handler struct {
	ddb       DynamoDBAPI
	sqs       SQSAPI
	logger    *zerolog.Logger
	logreader *io.PipeReader
}

// Item - represents the application table
type Item struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Loghandler  string `json:"loghandler"`
}

// NewHandler initializes and returns a new Handler.
func NewHandler(ddb DynamoDBAPI, sqs SQSAPI) *Handler {

	// Initialise logging
	logreader, logwriter := io.Pipe()
	logger := initLogger(logwriter)

	// Return handler
	return &Handler{ddb: ddb, sqs: sqs, logger: &logger, logreader: logreader}
}

// Handle handles the logger request.
func (h *Handler) Handle(ctx context.Context, sqsEvent events.SQSEvent) error {

	// Check we have required environment variables
	err := checkVars()
	if err != nil {
		return err
	}

	// Debug log
	h.logger.Debug().Msg("Log controller starting processing")

	// What do we have to process?
	num := len(sqsEvent.Records)

	// Debug log
	h.logger.Debug().Int("Number of messages received", num).Msg("")

	// Bail if nothing
	if num == 0 {

		// Debug log
		h.logger.Debug().Msg("Completed processing")
		return nil
	}

	// Iterate through the messages
	for _, msg := range sqsEvent.Records {

		// Debug log
		h.logger.Debug().
			Str("Message ID", msg.MessageId).
			Msg("Processing message")

		// Determine Log Processor
		procname, err := getQueueName(h, msg)
		if err != nil {

			// Debug log
			h.logger.Debug().
				Err(err).
				Msg("")

			// Log details
			h.logger.Warn().
				Str("Message ID", msg.MessageId).
				Msg("Unable to determine queue. Skipping further processing of this message")

			// Skip from processing
			continue
		}

		// Send the message to the processor
		err = submitMessage(h, msg, procname)
		if err != nil {

			// Debug log
			h.logger.Debug().
				Err(err).
				Msg("Error reported")

			// Log details
			h.logger.Warn().
				Str("Message ID", msg.MessageId).
				Msg("Unable to submit message to queue. Skipping further processing of this message")

			// Skip from processing
			continue
		}

		// Delete the message
		err = deleteMessage(h, msg)
		if err != nil {

			// Debug log
			h.logger.Debug().
				Err(err).
				Msg("Error reported")

			// Log details
			h.logger.Warn().
				Str("Message ID", msg.MessageId).
				Msg("Unable to delete message.")
		}
	}

	// Return
	return nil
}

// checkVars - checks that required variables are available
func checkVars() error {

	// Grab the lambda environment variables
	s3bucket := os.Getenv(envNameS3Bucket)
	s3path := os.Getenv(envNameS3BucketPath)
	s3region := os.Getenv(envNameS3BucketRegion)

	// Sanity checks
	if s3bucket == "" {
		fmt.Println(newErrorEnvironmentVariableProvided(envNameS3Bucket))
		return newErrorEnvironmentVariableProvided(envNameS3Bucket)
	}
	if s3path == "" {
		fmt.Println(newErrorEnvironmentVariableProvided(envNameS3BucketPath))
		return newErrorEnvironmentVariableProvided(envNameS3BucketPath)
	}
	if s3region == "" {
		fmt.Println(newErrorEnvironmentVariableProvided(envNameS3BucketRegion))
		return newErrorEnvironmentVariableProvided(envNameS3BucketRegion)
	}

	// Return
	return nil
}

// initLogger - initializes the Zero log logger
func initLogger(pipe *io.PipeWriter) zerolog.Logger {

	// Grab the lambda environment variables
	loglevel := os.Getenv(envNameLogLevel)

	// Configure Zero log timestamp & log level
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	switch strings.ToUpper(loglevel) {
	case LogLevelDebug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Create slice for handling multiple writers if required
	var writers []io.Writer
	writers = append(writers, pipe)
	mw := io.MultiWriter(writers...)

	// Now create the logger
	logger := zerolog.New(mw).With().Timestamp().Logger()

	// Return
	return logger
}

// getQueueName - queries DynamoDB to retrieve the downstream queue name to use
func getQueueName(h *Handler, msg events.SQSMessage) (string, error) {

	// Grab the message attributes
	msgattribs := msg.MessageAttributes
	if msgattribs == nil {
		return "", newErrorMessageAttributesNil()
	}

	// Get the application name
	val, found := msgattribs[MessageAttribAppName]
	if !found {
		return "", newErrorMessageAttributesAppNameEmpty()
	}
	logapp := *val.StringValue
	if logapp == "" {
		return "", newErrorMessageAttributesAppNameEmpty()
	}

	// Get the application version
	val, found = msgattribs[MessageAttribAppVers]
	if !found {
		return "", newErrorMessageAttributesAppVersionEmpty()
	}
	logvers := *val.StringValue
	if logvers == "" {
		return "", newErrorMessageAttributesAppVersionEmpty()
	}

	// Debug log
	h.logger.Debug().
		Str("Message ID", msg.MessageId).
		Str("Application name", logapp).
		Str("Application version", logvers).
		Msg("Retrieved message attributes")

	// Fetch the application record from DynamoDB
	db := NewDB(h.ddb)
	item, err := db.performGet(logapp, logvers)
	if err != nil {
		return "", err
	}

	// Get the processor queue name
	qname := item.Loghandler
	if qname == "" {
		return "", newErrorUnableToFetchProcQueueName()
	}

	// Debug log
	h.logger.Debug().
		Str("Message ID", msg.MessageId).
		Str("Queue name", qname).
		Msg("Retrieved Log processor queue name")

	// Return
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
	if url == "" {
		return newErrorUnableToFetchProcQueueURL()
	}

	// Debug log
	h.logger.Debug().
		Str("Message ID", msg.MessageId).
		Str("Queue url", url).
		Msg("Retrieved Log processor queue url")

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
	if url == "" {
		return newErrorUnableToFetchControllerQueueURL()
	}

	// Debug log
	h.logger.Debug().
		Str("Message ID", msg.MessageId).
		Str("Queue url", url).
		Msg("Retrieved Log controller queue url")

	// Delete the message
	err = sqsc.performDelete(url, msg.ReceiptHandle)
	return err
}
