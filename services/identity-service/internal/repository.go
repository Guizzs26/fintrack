package identity

import (
	"context"
	"fmt"
	"log/slog"

	ctxlogger "github.com/Guizzs26/fintrack/pkg/logger/context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBUserRepository is a DynamoDB implementation of the UserRepository interface
type DynamoDBUserRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBUserRepository(c *dynamodb.Client, tn string) *DynamoDBUserRepository {
	return &DynamoDBUserRepository{
		client:    c,
		tableName: tn,
	}
}

// Save persists a new or updated user to DynamoD
func (r *DynamoDBUserRepository) Save(ctx context.Context, user *User) error {
	log := ctxlogger.GetLogger(ctx)

	// marshal Go struct into a map of DynamoDB attribute values
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user for dynamodb: %v", err)
	}

	// create the input for the PutItem operation
	input := &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      item,
	}

	log.Debug("saving user to dynamodb", slog.Any("item", item))
	if _, err := r.client.PutItem(ctx, input); err != nil {
		return fmt.Errorf("failed to save user to dynamodb: %v", err)
	}

	return nil
}

// FindByEmail finds a user by their email using a Global Secondary Index (GSI)
func (r *DynamoDBUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	log := ctxlogger.GetLogger(ctx)

	// define the query input
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              aws.String("EmailIndex"),
		KeyConditionExpression: aws.String("#email = :email"),
		ExpressionAttributeNames: map[string]string{
			"#email": "Email",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	}

	log.Debug("finding user by email in dynamodb", slog.String("email", email))
	output, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email from dynamodb: %v", err)
	}

	if len(output.Items) == 0 {
		return nil, ErrUserNotFound
	}

	if len(output.Items) > 1 {
		log.Warn("found multiple users with the same email", slog.String("email", email))
	}

	var user User
	// unmarshal the first found item back into our Go struct
	if err := attributevalue.UnmarshalMap(output.Items[0], &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user from dynamodb: %v", err)
	}

	return &user, nil
}
