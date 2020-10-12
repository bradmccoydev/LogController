package lambda

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// DynamoDBAPI - bits from the DynamoDB interface we need
type DynamoDBAPI interface {
	GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
}

// NewDB returns a new DynamoDB handler
func NewDB(c DynamoDBAPI) *DBHandler {
	return &DBHandler{ddb: c}
}

// DBHandler points back to the interface
type DBHandler struct {
	ddb DynamoDBAPI
}

// performGet wrappers the DynamoDB GetItem request
func (s *DBHandler) performGet(app string, vers string) (Item, error) {

	// Build the query params
	params := &dynamodb.GetItemInput{
		TableName: aws.String(applicationTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"application": {
				S: aws.String(app),
			},
			"version": {
				S: aws.String(vers),
			},
		},
	}

	// Fetch from DynamoDB
	item := Item{}
	result, err := s.ddb.GetItem(params)
	if err != nil {
		return item, err
	}
	if result.Item == nil {
		return item, newErrorUnableToFindAppVersion()
	}

	// Extract the table item
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	// Return
	return item, err
}
