module github.com/Guizzs26/fintrack/services/identity-service

go 1.25.0

require (
	github.com/Guizzs26/fintrack v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.51.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/golang/protobuf v1.5.4
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.42.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require github.com/aws/aws-sdk-go-v2/service/dynamodbstreams v1.31.0 // indirect

require (
	github.com/aws/aws-sdk-go-v2 v1.39.2
	github.com/aws/aws-sdk-go-v2/config v1.31.12
	github.com/aws/aws-sdk-go-v2/credentials v1.18.16
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.20.14
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.6 // indirect
	github.com/aws/smithy-go v1.23.0
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)

replace github.com/Guizzs26/fintrack => ../..
