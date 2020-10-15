package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	loglambda "github.com/bradmccoydev/LogController/lambda"
)

var sess *session.Session
var esqs *sqs.SQS
var ddb *dynamodb.DynamoDB
var s3c *s3.S3

func init() {
	sess = session.Must(session.NewSession())
	esqs = sqs.New(sess)
	ddb = dynamodb.New(sess)
	s3c = s3.New(sess)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	return loglambda.NewHandler(ddb, s3c, esqs).Handle(ctx, sqsEvent)
}

func main() {
	lambda.Start(handler)
}
