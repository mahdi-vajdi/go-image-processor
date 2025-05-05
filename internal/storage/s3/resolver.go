package s3Storage

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

type resolverV2 struct{}

func (*resolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (smithyendpoints.Endpoint, error) {
	if params.Endpoint != nil {
		u, err := url.Parse(*params.Endpoint)
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return s3.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
