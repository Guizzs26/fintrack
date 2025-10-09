package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// Single Table Design :)
type tokenItem struct {
	PK        string    `dynamodbav:"PK"` // Format: USER#<UserID>
	SK        string    `dynamodbav:"SK"` // Format: TOKEN#<TokenHash>
	UserID    uuid.UUID `dynamodbav:"UserID"`
	TokenHash string    `dynamodbav:"TokenHash"`
	ExpiresAt time.Time `dynamodbav:"ExpiresAt"`
}

var _ TokenRepository = (*DynamoDBTokenRepository)(nil)

type DynamoDBTokenRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBTokenRepository(c *dynamodb.Client, tn string) *DynamoDBTokenRepository {
	return &DynamoDBTokenRepository{
		client:    c,
		tableName: tn,
	}
}

func (r *DynamoDBTokenRepository) Save(ctx context.Context, token *RefreshToken) error {
	item := tokenItem{
		PK:        fmt.Sprintf("USER#%s", token.UserID),
		SK:        fmt.Sprintf("TOKEN#%s", token.TokenHash),
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal token for dynamodb: %v", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      av,
	}

	if _, err := r.client.PutItem(ctx, input); err != nil {
		return fmt.Errorf("failed to save token to dynamodb: %v", err)
	}

	return nil
}

func (r *DynamoDBTokenRepository) Revoke(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	// use GSI to find the full token item
	queryInput := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              aws.String("TokenHashIndex"), // GSI to be created
		KeyConditionExpression: aws.String("TokenHash = :hash"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hash": &types.AttributeValueMemberS{Value: tokenHash},
		},
	}

	output, err := r.client.Query(ctx, queryInput)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to query token by hash: %v", err)
	}
	if len(output.Items) == 0 {
		return uuid.Nil, fmt.Errorf("token not found")
	}

	var item tokenItem
	if err := attributevalue.UnmarshalMap(output.Items[0], &item); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal token item: %v", err)
	}

	// delet the item using its full primary key (PK and SK)
	deleteInput := &dynamodb.DeleteItemInput{
		TableName: &r.tableName,
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: item.PK},
			"SK": &types.AttributeValueMemberS{Value: item.SK},
		},
	}

	if _, err := r.client.DeleteItem(ctx, deleteInput); err != nil {
		return uuid.Nil, fmt.Errorf("failed to delete token: %v", err)
	}

	return item.UserID, nil
}
