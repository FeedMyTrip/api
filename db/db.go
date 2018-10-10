package db

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	awsRegion = "sa-east-1"
)

// DeleteItem delete the defined key item
func DeleteItem(table, keyLabel, keyValue string) error {
	db, err := connect(awsRegion)
	if err != nil {
		return err
	}

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
		TableName: aws.String(table),
	}

	_, err = db.DeleteItem(input)
	if err != nil {
		return err
	}
	return nil
}

//PutListItem update nested list insertinig new object
func PutListItem(table, keyLabel, keyValue, listName string, data map[string]interface{}) (*dynamodb.UpdateItemOutput, error) {
	db, err := connect(awsRegion)
	if err != nil {
		return nil, err
	}

	update, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: update,
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(table),
		UpdateExpression: aws.String("SET " + listName + " = list_append(" + listName + ", :" + listName + ")"),
	}

	result, err := db.UpdateItem(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//UpdateListItem update nested list item updating attributes value
func UpdateListItem(table, keyLabel, keyValue, listName string, index int, data map[string]interface{}) (*dynamodb.UpdateItemOutput, error) {
	db, err := connect(awsRegion)
	if err != nil {
		return nil, err
	}

	update, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return nil, err
	}

	updateValues := map[string]*dynamodb.AttributeValue{}

	indexStr := strconv.Itoa(index)
	updateExpresion := "SET"

	for key, value := range update {
		updateValues[":"+key] = value
		updateExpresion += " " + listName + "[" + indexStr + "]." + key + " = :" + key + ","
	}

	updateExpresion = updateExpresion[:len(updateExpresion)-1]

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: updateValues,
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(table),
		UpdateExpression: aws.String(updateExpresion),
	}

	result, err := db.UpdateItem(input)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateItem update item attributes
func UpdateItem(table, keyLabel, keyValue string, data map[string]interface{}) (*dynamodb.UpdateItemOutput, error) {
	db, err := connect(awsRegion)
	if err != nil {
		return nil, err
	}

	update, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return nil, err
	}

	updateValues := map[string]*dynamodb.AttributeValue{}

	updateExpresion := "SET"
	for key, value := range update {
		updateValues[":"+key] = value
		updateExpresion += " " + key + " = :" + key + ","
	}

	updateExpresion = updateExpresion[:len(updateExpresion)-1]

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: updateValues,
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(table),
		UpdateExpression: aws.String(updateExpresion),
	}

	result, err := db.UpdateItem(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetAllItems return an array with all items from a difined table
func GetAllItems(table string) (*dynamodb.ScanOutput, error) {
	db, err := connect(awsRegion)
	if err != nil {
		return nil, err
	}

	params := &dynamodb.ScanInput{
		TableName: aws.String(table),
	}

	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//GetItem returns an object defined by the keyValue
func GetItem(table, keyLabel, keyValue string) (*dynamodb.GetItemOutput, error) {
	db, err := connect(awsRegion)
	if err != nil {
		return nil, err
	}

	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// PutItem insert a new icon to the defined table
func PutItem(object interface{}, table string) error {
	db, err := connect(awsRegion)
	if err != nil {
		return err
	}

	av, err := dynamodbattribute.MarshalMap(object)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}

	_, err = db.PutItem(input)

	if err != nil {
		return err
	}
	return nil
}

func connect(region string) (*dynamodb.DynamoDB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	if err != nil {
		return nil, err
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)
	return svc, nil
}
