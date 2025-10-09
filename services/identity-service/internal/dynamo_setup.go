package identity

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

// localEndpointResolver is an implementation of the dynamodb.EndpointResolverV2 interface
// that always resolves to a static, local endpoint URL.
// This is used to direct SDK requests to a local DynamoDB instance
// (such as the Docker image amazon/dynamodb-local) during development.
type localEndpointResolver struct {
	url string
}

// ResolveEndpoint satisfies the dynamodb.EndpointResolverV2 interface.
// The SDK calls this method to determine the endpoint for a DynamoDB operation
// (pointing to a local instance instead of AWS).
func (r *localEndpointResolver) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (
	smithyendpoints.Endpoint,
	error,
) {
	u, err := url.Parse(r.url)
	if err != nil {
		return smithyendpoints.Endpoint{}, fmt.Errorf("failed to parse static endpoint URL: %v", err)
	}

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}

// NewDynamoDBClient creates and configures a new DynamoDB client.
//   - If the DYNAMODB_ENDPOINT environment variable is set, the function configures the client
//     to use that endpoint. This is ideal for connecting to a local DynamoDB instance.
//   - If the DYNAMODB_ENDPOINT variable is NOT set, the function creates a client
//     with the default AWS configuration, making it production-ready.
func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	var cfgOptions []func(*config.LoadOptions) error
	endpointURL, isEndpointSet := os.LookupEnv("DYNAMODB_ENDPOINT")

	if isEndpointSet {
		// For local DynamoDB, credentials and region are irrelevant,
		// but the SDK requires them to be set.
		cfgOptions = append(cfgOptions,
			config.WithRegion("us-east-1"),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider("DUMMY", "DUMMY", ""),
			),
		)
	}

	cfg, err := config.LoadDefaultConfig(ctx, cfgOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %v", err)
	}

	var clientOptions []func(*dynamodb.Options)

	if isEndpointSet {
		resolver := &localEndpointResolver{url: endpointURL}
		clientOptions = append(clientOptions, dynamodb.WithEndpointResolverV2(resolver))
	}

	client := dynamodb.NewFromConfig(cfg, clientOptions...)

	return client, nil
}

/*

The purpose of this code is to provide a single function, NewDynamoDBClient,
that is smart enough to behave in two different ways:

In Development: Connects to your local DynamoDB instance running in Docker.

In Production: Connects to the real AWS service using environment credentials (e.g., IAM Roles).

*/
