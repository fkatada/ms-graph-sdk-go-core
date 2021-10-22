package msgraphgocore

import (
	"errors"

	absauth "github.com/microsoft/kiota/abstractions/go/authentication"
	absser "github.com/microsoft/kiota/abstractions/go/serialization"
	khttp "github.com/microsoft/kiota/http/go/nethttp"
)

// GraphRequestAdapterBase is the core service used by GraphServiceClient to make requests to Microsoft Graph.
type GraphRequestAdapterBase struct {
	khttp.NetHttpRequestAdapter
}

// NewGraphRequestAdapterBase creates a new GraphRequestAdapterBase with the given parameters
// Parameters:
// authenticationProvider: the provider used to authenticate requests
// clientOptions: the options used to configure the client
// Returns:
// a new GraphRequestAdapterBase
func NewGraphRequestAdapterBase(authenticationProvider absauth.AuthenticationProvider, clientOptions GraphClientOptions) (*GraphRequestAdapterBase, error) {
	return NewGraphRequestAdapterBaseWithParseNodeFactory(authenticationProvider, clientOptions, nil)
}

// NewGraphRequestAdapterBaseWithParseNodeFactory creates a new GraphRequestAdapterBase with the given parameters
// Parameters:
// authenticationProvider: the provider used to authenticate requests
// clientOptions: the options used to configure the client
// parseNodeFactory: the factory used to create parse nodes
// Returns:
// a new GraphRequestAdapterBase
func NewGraphRequestAdapterBaseWithParseNodeFactory(authenticationProvider absauth.AuthenticationProvider, clientOptions GraphClientOptions, parseNodeFactory absser.ParseNodeFactory) (*GraphRequestAdapterBase, error) {
	return NewGraphRequestAdapterBaseWithParseNodeFactoryAndSerializationWriterFactory(authenticationProvider, clientOptions, parseNodeFactory, nil)
}

// NewGraphRequestAdapterBaseWithParseNodeFactoryAndSerializationWriterFactory creates a new GraphRequestAdapterBase with the given parameters
// Parameters:
// authenticationProvider: the provider used to authenticate requests
// clientOptions: the options used to configure the client
// parseNodeFactory: the factory used to create parse nodes
// serializationWriterFactory: the factory used to create serialization writers
// Returns:
// a new GraphRequestAdapterBase
func NewGraphRequestAdapterBaseWithParseNodeFactoryAndSerializationWriterFactory(authenticationProvider absauth.AuthenticationProvider, clientOptions GraphClientOptions, parseNodeFactory absser.ParseNodeFactory, serializationWriterFactory absser.SerializationWriterFactory) (*GraphRequestAdapterBase, error) {
	return NewGraphRequestAdapterBaseWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(authenticationProvider, clientOptions, parseNodeFactory, serializationWriterFactory, nil)
}

// NewGraphRequestAdapterBaseWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient creates a new GraphRequestAdapterBase with the given parameters
// Parameters:
// authenticationProvider: the provider used to authenticate requests
// clientOptions: the options used to configure the client
// parseNodeFactory: the factory used to create parse nodes
// serializationWriterFactory: the factory used to create serialization writers
// httpClient: the client used to send requests
// Returns:
// a new GraphRequestAdapterBase
func NewGraphRequestAdapterBaseWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(authenticationProvider absauth.AuthenticationProvider, clientOptions GraphClientOptions, parseNodeFactory absser.ParseNodeFactory, serializationWriterFactory absser.SerializationWriterFactory, httpClient *khttp.NetHttpMiddlewareClient) (*GraphRequestAdapterBase, error) {
	if authenticationProvider == nil {
		return nil, errors.New("authenticationProvider cannot be nil")
	}
	middlewares := GetDefaultMiddlewaresWithOptions(&clientOptions)
	cErr := khttp.ChainMiddlewares(middlewares)
	if cErr != nil {
		return nil, cErr
	}
	if httpClient == nil {
		defaultClient, err := khttp.NewNetHttpMiddlewareClientWithMiddlewares(middlewares)
		if err != nil {
			return nil, err
		}
		httpClient = defaultClient
	}
	if serializationWriterFactory == nil {
		serializationWriterFactory = absser.DefaultSerializationWriterFactoryInstance
	}
	if parseNodeFactory == nil {
		parseNodeFactory = absser.DefaultParseNodeFactoryInstance
	}
	baseAdapter, err := khttp.NewNetHttpRequestAdapterWithParseNodeFactoryAndSerializationWriterFactoryAndHttpClient(authenticationProvider, parseNodeFactory, serializationWriterFactory, httpClient)
	if err != nil {
		return nil, err
	}
	result := &GraphRequestAdapterBase{
		NetHttpRequestAdapter: *baseAdapter,
	}

	return result, nil
}
