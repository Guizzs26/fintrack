package identity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	ctxlogger "github.com/Guizzs26/fintrack/pkg/logger/context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var _ UserRepository = (*DynamoDBUserRepository)(nil)

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

// Save persists a new or updated user to DynamoDb (upsert-like)
func (r *DynamoDBUserRepository) Save(ctx context.Context, user *User) error {
	log := ctxlogger.GetLogger(ctx)

	isNewUser := user.CreatedAt.IsZero()
	if isNewUser {
		// marshal Go struct into a map of DynamoDB attribute values
		item, err := attributevalue.MarshalMap(user)
		if err != nil {
			return fmt.Errorf("failed to marshal user for dynamodb: %v", err)
		}

		// create the input for the PutItem operation
		input := &dynamodb.PutItemInput{
			TableName:           &r.tableName,
			Item:                item,
			ConditionExpression: aws.String("attribute_not_exists(Email)"),
		}

		log.Debug("creating new user in dynamodb", slog.Any("item", item))
		if _, err := r.client.PutItem(ctx, input); err != nil {
			var condErr *types.ConditionalCheckFailedException
			if errors.As(err, &condErr) {
				return ErrEmailAlreadyInUse
			}
			return fmt.Errorf("failed to create user to dynamodb: %v", err)
		}
	} else {
		log.Debug("updating existing user in dynamodb", slog.String("user_id", user.ID.String()))
		updateExpr := "SET #name = :name, #pwhash = :pwhash, #ua = :ua"
		exprAttrNames := map[string]string{
			"#name":   "Name",
			"#pwhash": "PasswordHash",
			"#ua":     "UpdatedAt",
		}
		exprAttrValues, err := attributevalue.MarshalMap(map[string]interface{}{
			":name":   user.Name,
			":pwhash": user.PasswordHash,
			":ua":     user.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal update values for dynamodb: %v", err)
		}

		input := &dynamodb.UpdateItemInput{
			TableName: &r.tableName,
			Key: map[string]types.AttributeValue{
				"ID": &types.AttributeValueMemberS{Value: user.ID.String()},
			},
			UpdateExpression:          aws.String(updateExpr),
			ExpressionAttributeNames:  exprAttrNames,
			ExpressionAttributeValues: exprAttrValues,
		}

		if _, err := r.client.UpdateItem(ctx, input); err != nil {
			return fmt.Errorf("failed to update user in dynamodb: %v", err)
		}
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
		KeyConditionExpression: aws.String("Email = :email"),
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

	// !!! critical !!!
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

func (r *DynamoDBUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	log := ctxlogger.GetLogger(ctx)

	input := &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id.String()},
		},
	}

	log.Debug("finding user by id in dynamodb", slog.String("user_id", id.String()))
	output, err := r.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id from dynamodb: %v", err)
	}

	if output.Item == nil {
		return nil, ErrUserNotFound
	}

	var user User
	if err := attributevalue.UnmarshalMap(output.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user from dynamodb: %v", err)
	}

	return &user, nil
}
