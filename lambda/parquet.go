package lambda

import (
	"bufio"
	"bytes"

	"github.com/aws/aws-lambda-go/events"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

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
func convertToParquet(msg events.SQSMessage) ([]byte, int64, error) {

	// Get the application name
	logapp, err := getMsgAttrib(msg, MessageAttribAppName)
	if err != nil {
		return nil, 0, err
	}

	// Get the application version
	logvers, err := getMsgAttrib(msg, MessageAttribAppVers)
	if err != nil {
		return nil, 0, err
	}

	// Get the log level
	loglevel, err := getMsgAttrib(msg, MessageAttribLogLevel)
	if err != nil {
		return nil, 0, err
	}

	// Get the timestamp
	logtstamp, err := getMsgAttrib(msg, MessageAttribTimestamp)
	if err != nil {
		return nil, 0, err
	}

	// Get the tracking id
	logtrackingid, err := getMsgAttrib(msg, MessageAttribTrackingID)
	if err != nil {
		return nil, 0, err
	}

	// Create a parquet writer
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	pw, err := writer.NewParquetWriterFromWriter(w, new(LogItem), 1)
	if err != nil {
		return nil, 0, err
	}

	// Set compression
	pw.CompressionType = parquet.CompressionCodec_GZIP

	// Write the data
	l := &LogItem{
		Time:        logtstamp,
		TrackingID:  logtrackingid,
		MessageID:   msg.MessageId,
		Level:       loglevel,
		Application: logapp,
		Version:     logvers,
		Message:     msg.Body,
	}
	err = pw.Write(l)
	if err != nil {
		return nil, 0, err
	}

	// Close the parquet writer
	err = pw.WriteStop()
	if err != nil {
		return nil, 0, err
	}

	// Return
	return nil, 0, nil
}
