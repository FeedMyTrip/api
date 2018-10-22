package db

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const (
	//AWSRegion defines the database region on AWS
	AWSRegion = "sa-east-1"
)

// DeleteListItem delete the defined key item on a nested list with the index
func DeleteListItem(table, keyLabel, keyValue, listName string, index int) error {
	db, err := connect(AWSRegion)
	if err != nil {
		return err
	}
	indexStr := strconv.Itoa(index)
	input := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
		TableName:        aws.String(table),
		UpdateExpression: aws.String("REMOVE " + listName + "[" + indexStr + "]"),
	}

	_, err = db.UpdateItem(input)
	if err != nil {
		return err
	}
	return nil
}

// DeleteItem delete the defined key item
func DeleteItem(table, keyLabel, keyValue string) error {
	db, err := connect(AWSRegion)
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
func PutListItem(table, keyLabel, keyValue, listName string, data interface{}) (*dynamodb.UpdateItemOutput, error) {
	db, err := connect(AWSRegion)
	if err != nil {
		return nil, err
	}

	update, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		return nil, err
	}

	emptyList := []*dynamodb.AttributeValue{}
	update[":empty_list"] = &dynamodb.AttributeValue{L: emptyList}

	updateExpression := "SET " + listName + " = list_append(if_not_exists(" + listName + ", :empty_list), :" + listName + ")"

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: update,
		Key: map[string]*dynamodb.AttributeValue{
			keyLabel: {
				S: aws.String(keyValue),
			},
		},
		ReturnValues:     aws.String("ALL_NEW"),
		TableName:        aws.String(table),
		UpdateExpression: aws.String(updateExpression),
	}

	result, err := db.UpdateItem(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//UpdateListItem update nested list item updating attributes value
func UpdateListItem(table, keyLabel, keyValue, listName string, index int, data interface{}) (*dynamodb.UpdateItemOutput, error) {
	db, err := connect(AWSRegion)
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
		valueKey := strings.Replace(key, ".", "", -1)
		updateValues[":"+valueKey] = value
		updateExpresion += " " + listName + "[" + indexStr + "]." + key + " = :" + valueKey + ","
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
func UpdateItem(table, keyLabel, keyValue string, data interface{}) (*dynamodb.UpdateItemOutput, error) {
	db, err := connect(AWSRegion)
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
		valueKey := strings.Replace(key, ".", "", -1)
		updateValues[":"+valueKey] = value
		updateExpresion += " " + key + " = :" + valueKey + ","
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
func GetAllItems(table, filterExpression string, expressionAttributeValues map[string]*dynamodb.AttributeValue) (*dynamodb.ScanOutput, error) {
	db, err := connect(AWSRegion)
	if err != nil {
		return nil, err
	}

	params := &dynamodb.ScanInput{
		TableName: aws.String(table),
	}

	if filterExpression != "" {
		params.FilterExpression = aws.String(filterExpression)
		params.ExpressionAttributeValues = expressionAttributeValues
	}

	if value, ok := expressionAttributeValues["limit"]; ok {
		limit, err := strconv.ParseInt(*value.S, 10, 64)
		if err != nil {
			params.Limit = &limit
		}
	}

	result, err := db.Scan(params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

//GetItem returns an object defined by the keyValue
func GetItem(table, keyLabel, keyValue string) (*dynamodb.GetItemOutput, error) {
	db, err := connect(AWSRegion)
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

// PutItem insert a new object to the defined table
func PutItem(object interface{}, table string) error {
	db, err := connect(AWSRegion)
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

//CreateTable creates a new database table
func CreateTable(tableName, keyLabel string, rcu, wcu int64) error {
	db, err := connect(AWSRegion)
	if err != nil {
		return err
	}

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(keyLabel),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(keyLabel),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(rcu),
			WriteCapacityUnits: aws.Int64(wcu),
		},
		TableName: aws.String(tableName),
	}

	_, err = db.CreateTable(input)
	if err != nil {
		return err
	}
	return nil
}

//DeleteTable delete database table
func DeleteTable(tableName string) error {
	db, err := connect(AWSRegion)
	if err != nil {
		return err
	}

	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}

	_, err = db.DeleteTable(input)
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
