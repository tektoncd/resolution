// Code generated by smithy-go-codegen DO NOT EDIT.

package ecr

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Notifies Amazon ECR that you intend to upload an image layer. When an image is
// pushed, the InitiateLayerUpload API is called once per image layer that has not
// already been uploaded. Whether or not an image layer has been uploaded is
// determined by the BatchCheckLayerAvailability API action. This operation is used
// by the Amazon ECR proxy and is not generally used by customers for pulling and
// pushing images. In most cases, you should use the docker CLI to pull, tag, and
// push images.
func (c *Client) InitiateLayerUpload(ctx context.Context, params *InitiateLayerUploadInput, optFns ...func(*Options)) (*InitiateLayerUploadOutput, error) {
	if params == nil {
		params = &InitiateLayerUploadInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "InitiateLayerUpload", params, optFns, c.addOperationInitiateLayerUploadMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*InitiateLayerUploadOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type InitiateLayerUploadInput struct {

	// The name of the repository to which you intend to upload layers.
	//
	// This member is required.
	RepositoryName *string

	// The Amazon Web Services account ID associated with the registry to which you
	// intend to upload layers. If you do not specify a registry, the default registry
	// is assumed.
	RegistryId *string

	noSmithyDocumentSerde
}

type InitiateLayerUploadOutput struct {

	// The size, in bytes, that Amazon ECR expects future layer part uploads to be.
	PartSize *int64

	// The upload ID for the layer upload. This parameter is passed to further
	// UploadLayerPart and CompleteLayerUpload operations.
	UploadId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationInitiateLayerUploadMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpInitiateLayerUpload{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpInitiateLayerUpload{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpInitiateLayerUploadValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opInitiateLayerUpload(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opInitiateLayerUpload(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "ecr",
		OperationName: "InitiateLayerUpload",
	}
}
