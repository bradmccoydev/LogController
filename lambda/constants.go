package lambda

// Constants
const (
	// Name of the "Application" table in DynamoDB
	applicationTableName string = "application"

	// Name of the Lambda environment variable for logging level
	envNameLogLevel string = "LOG_LEVEL"

	// Name of the Lambda environment variable for the S3 bucket to use
	envNameS3Bucket string = "S3_BUCKET"

	// Name of the Lambda environment variable for the path in the S3 bucket to use
	envNameS3BucketPath string = "S3_PATH"

	// Name of the Lambda environment variable for the region of the S3 bucket to use
	envNameS3BucketRegion string = "S3_REGION"

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

	// MessageAttribLogLevel - SQS message attribute that stores the log level
	MessageAttribLogLevel string = "LOG_LEVEL"

	// MessageAttribTimestamp - SQS message attribute that stores the timestamp
	MessageAttribTimestamp string = "TIMESTAMP"

	// MessageAttribTrackingID - SQS message attribute that stores the tracking id
	MessageAttribTrackingID string = "TRACKING_ID"

	// The S3 Bucket queue - used as default if nothing provided
	processQueueS3 string = "S3QUEUE"
)
