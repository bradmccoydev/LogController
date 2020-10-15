package lambda

import "github.com/aws/aws-lambda-go/events"

// LogItem represents a log message
type LogItem struct {
	Time        string `parquet:"name=time, type=TIME_MICROS, encoding=PLAIN_DICTIONARY"`
	TrackingID  string `parquet:"name=trackingid, type=UTF8, encoding=PLAIN_DICTIONARY"`
	MessageID   string `parquet:"name=messageid, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Level       string `parquet:"name=level, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Application string `parquet:"name=application, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Version     string `parquet:"name=version, type=UTF8, encoding=PLAIN_DICTIONARY"`
	Message     string `parquet:"name=message, type=UTF8, encoding=PLAIN_DICTIONARY"`
}

// convertToParquet - converts the SQS message into parquet
func convertToParquet(msg events.SQSMessage) ([]byte, int, error) {

	// Grab the message attributes
	msgattribs := msg.MessageAttributes
	if msgattribs == nil {
		return nil, 0, newErrorMessageAttributesNil()
	}

	// Get the application name
	val, found := msgattribs[MessageAttribAppName]
	if !found {
		return nil, 0, newErrorMessageAttributesAppNameEmpty()
	}
	logapp := *val.StringValue
	if logapp == "" {
		return nil, 0, newErrorMessageAttributesAppNameEmpty()
	}

	// Get the application version
	val, found = msgattribs[MessageAttribAppVers]
	if !found {
		return nil, 0, newErrorMessageAttributesAppVersionEmpty()
	}
	logvers := *val.StringValue
	if logvers == "" {
		return nil, 0, newErrorMessageAttributesAppVersionEmpty()
	}

	// Get the log level
	val, found = msgattribs[MessageAttribLogLevel]
	if !found {
		return nil, 0, newErrorMessageAttributesLogLevelEmpty()
	}
	loglevel := *val.StringValue
	if loglevel == "" {
		return nil, 0, newErrorMessageAttributesLogLevelEmpty()
	}

	// Get the timestamp
	val, found = msgattribs[MessageAttribTimestamp]
	if !found {
		return nil, 0, newErrorMessageAttributesTimestampEmpty()
	}
	logtstamp := *val.StringValue
	if logtstamp == "" {
		return nil, 0, newErrorMessageAttributesTimestampEmpty()
	}

	// Get the tracking id
	val, found = msgattribs[MessageAttribTrackingID]
	if !found {
		return nil, 0, newErrorMessageAttributesTrackingIDEmpty()
	}
	logtrackingid := *val.StringValue
	if logtrackingid == "" {
		return nil, 0, newErrorMessageAttributesTrackingIDEmpty()
	}

	// Massage the message into the structure
	l := &LogItem{
		Time:        logtstamp,
		TrackingID:  logtrackingid,
		MessageID:   msg.MessageId,
		Level:       loglevel,
		Application: logapp,
		Version:     logvers,
		Message:     msg.Body,
	}

	// Return
	return nil, 0, nil
}
