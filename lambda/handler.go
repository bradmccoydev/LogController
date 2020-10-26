package lambda

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/rs/zerolog"
)

// Handler handles incoming logger requests.
type Handler struct {
	ddb    DynamoDBAPI
	sqs    SQSAPI
	s3c    S3API
	logger *zerolog.Logger
}

// Item - represents the application table
type Item struct {
	Application string `json:"application"`
	Version     string `json:"version"`
	Loghandler  string `json:"loghandler"`
}

// NewHandler initializes and returns a new Handler.
func NewHandler(ddb DynamoDBAPI, s3c S3API, sqs SQSAPI) *Handler {

	// Initialise logging
	logger := initLogger()

	// Return handler
	return &Handler{ddb: ddb, s3c: s3c, sqs: sqs, logger: &logger}
}

// Handle handles the logger request.
func (h *Handler) Handle(ctx context.Context, sqsEvent events.SQSEvent) error {

	// Debug log
	h.logger.Debug().Msg("Starting processing")

	// Check we have required environment variables
	err := checkVars(h)
	if err != nil {

		// Log error
		h.logger.Error().
			Err(err).
			Msg("")

		// Return
		return err
	}

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
			Msg("Starting message processing")

		// Determine Log Processor
		procname, err := getProcessorName(h, msg)
		if err != nil {

			// Log details
			h.logger.Warn().
				Str("Message ID", msg.MessageId).
				Err(err).
				Msg("Unable to determine processor")

			// Ignore further processing
			continue
		}

		// Send the message to the processor
		err = processMessage(h, msg, procname)
		if err != nil {

			// Log details
			h.logger.Warn().
				Str("Message ID", msg.MessageId).
				Err(err).
				Msg("Unable to process message")

			// Ignore further processing
			continue
		}

		// Delete the message
		err = deleteMessage(h, msg)
		if err != nil {

			// Log details
			h.logger.Warn().
				Str("Message ID", msg.MessageId).
				Err(err).
				Msg("Unable to delete message")

			// Ignore further processing
			continue
		}

		// Debug log
		h.logger.Debug().
			Str("Message ID", msg.MessageId).
			Msg("Completed message processing")
	}

	// Debug log
	h.logger.Debug().Msg("Completed processing")
	return nil
}

// checkVars - checks that required variables are available
func checkVars(h *Handler) error {

	// Grab the lambda environment variables
	s3bucket := os.Getenv(envNameS3Bucket)
	s3path := os.Getenv(envNameS3BucketPath)
	s3region := os.Getenv(envNameS3BucketRegion)

	// Sanity checks
	if s3bucket == "" {
		return newErrorEnvironmentVariableProvided(envNameS3Bucket)
	}
	if s3path == "" {
		return newErrorEnvironmentVariableProvided(envNameS3BucketPath)
	}
	if s3region == "" {
		return newErrorEnvironmentVariableProvided(envNameS3BucketRegion)
	}

	// Debug log
	h.logger.Debug().
		Str("S3 Region", s3region).
		Str("S3 Bucket", s3bucket).
		Str("S3 Path", s3path).
		Msg("")

	// Return
	return nil
}

// initLogger - initializes the Zero log logger
func initLogger() zerolog.Logger {

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

	// Now create the logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Return
	return logger
}

// getMsgAttribs - extracts the specified message attribute value
func getMsgAttrib(msg events.SQSMessage, key string) (string, error) {

	// Grab the message attributes
	msgattribs := msg.MessageAttributes
	if msgattribs == nil {
		return "", newErrorMessageAttributesNil()
	}

	// Get the attribute
	val, found := msgattribs[MessageAttribAppName]
	if !found {
		return "", newErrorMessageAttributeMissing(key)
	}
	value := *val.StringValue
	if value == "" {
		return "", newErrorMessageAttributeEmpty(key)
	}

	// Return
	return value, nil
}

// getProcessorName - queries DynamoDB to retrieve the name of the processor to use
func getProcessorName(h *Handler, msg events.SQSMessage) (string, error) {

	// Grab the app name
	logapp, err := getMsgAttrib(msg, MessageAttribAppName)
	if err != nil {

		// Debug log
		h.logger.Debug().
			Str("Message ID", msg.MessageId).
			Err(err).
			Msg("Defaulting the S3 queue")

		return processQueueS3, nil
	}

	// Grab the app version
	logvers, err := getMsgAttrib(msg, MessageAttribAppVers)
	if err != nil {

		// Debug log
		h.logger.Debug().
			Str("Message ID", msg.MessageId).
			Err(err).
			Msg("Defaulting the S3 queue")

		return processQueueS3, nil
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

		// Debug log
		h.logger.Debug().
			Str("Message ID", msg.MessageId).
			Err(newErrorUnableToFetchProcQueueName()).
			Msg("Defaulting the S3 queue")

		return processQueueS3, nil
	}

	// Debug log
	h.logger.Debug().
		Str("Message ID", msg.MessageId).
		Str("Queue name", qname).
		Msg("Retrieved Log processor queue name from application table")

	// Return
	return qname, nil
}

// processMessage - processes the log message
func processMessage(h *Handler, msg events.SQSMessage, queue string) error {

	// Submit to S3 if using that for processing
	if queue == processQueueS3 {
		err := s3Processor(h, msg)
		return err
	}

	// Otherwise submit to SQS
	err := sqsProcessor(h, msg, queue)
	return err
}

// s3Processor - saves the log message to S3
func s3Processor(h *Handler, msg events.SQSMessage) error {

	// Get the S3 bucket details
	s3bucket := os.Getenv(envNameS3Bucket)
	s3path := os.Getenv(envNameS3BucketPath)
	s3region := os.Getenv(envNameS3BucketRegion)

	// Build S3 Key
	year, month, day := time.Now().Date()
	yyyy := strconv.Itoa(year)
	mm := strconv.Itoa(int(month))
	dd := strconv.Itoa(day)
	s3key := s3path + "/year=" + yyyy + "/month=" + mm + "/day=" + dd

	// Convert the SQS message to parquet format
	buffer, size, err := convertToParquet(msg)
	if err != nil {
		return err
	}

	// Upload to S3
	s3c := NewS3(h.s3c)
	err = s3c.performPut(s3region, s3bucket, s3key, buffer, size)
	return err
}

// sqsProcessor - submits the log message to the processor SQS queue
func sqsProcessor(h *Handler, msg events.SQSMessage, queue string) error {

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
