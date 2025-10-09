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
	"github.com/google/uuid"
)

// Single Table Design :)
type tokenItem struct {
	PK        string    `dynamodbav:"PK"` // Format: USER#<UserID>
	SK        string    `dynamodbav:"SK"` // Format: TOKEN#<TokenHash>
	UserID    uuid.UUID `dynamodbav:"UserID"`
	TokenHash string    `dynamodbav:"TokenHash"`
	ExpiresAt int64     `dynamodbav:"ExpiresAt"`
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

// revoke a refresh token - usign 'read-then-write' pattern
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

func (r *DynamoDBTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	log := ctxlogger.GetLogger(ctx)

	pk := fmt.Sprintf("USER#%s", userID)
	queryInput := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: pk},
			":sk_prefix": &types.AttributeValueMemberS{Value: "TOKEN#"},
		},
	}
	paginator := dynamodb.NewQueryPaginator(r.client, queryInput)

	var writeRequests []types.WriteRequest
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to query tokens for user: %v", err)
		}

		for _, item := range output.Items {
			writeRequests = append(writeRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"PK": item["PK"],
						"SK": item["SK"],
					},
				},
			})
		}
	}

	if len(writeRequests) == 0 {
		return nil
	}

	log.Debug("revoking all refresh tokens for user", slog.String("user_id", userID.String()), slog.Int("token_count", len(writeRequests)))
	const maxBatchSize = 25
	for i := 0; i < len(writeRequests); i += maxBatchSize {
		end := min(i+maxBatchSize, len(writeRequests))
		chunk := writeRequests[i:end]

		batchInput := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.tableName: chunk,
			},
		}

		output, err := r.client.BatchWriteItem(ctx, batchInput)
		if err != nil {
			return fmt.Errorf("failed to batch delete tokens: %v", err)
		}

		// Handle unprocessed items (simplified approach with logging).
		// In the future, we may have retry logic here.
		if len(output.UnprocessedItems) > 0 {
			unprocessedCount := len(output.UnprocessedItems[r.tableName])
			log.Warn("some tokens were not processed in batch delete and will be orphaned",
				slog.Int("unprocessed_count", unprocessedCount),
				slog.String("user_id", userID.String()),
			)
		}
	}

	return nil
}
