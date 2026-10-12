# openapi.api.OidcProviderApi

## Load the API package
```dart
import 'package:openapi/api.dart';
```

All URIs are relative to */api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**createOidcProvider**](OidcProviderApi.md#createoidcprovider) | **POST** /oidcProvider/ | Create an OIDC provider
[**deleteOidcProvider**](OidcProviderApi.md#deleteoidcprovider) | **DELETE** /oidcProvider/{oidcProviderId} | Delete an OIDC provider
[**getOidcProviderById**](OidcProviderApi.md#getoidcproviderbyid) | **GET** /oidcProvider/{oidcProviderId} | Get an OIDC provider
[**getPagedOidcProviders**](OidcProviderApi.md#getpagedoidcproviders) | **POST** /oidcProvider/getPagedOidcProviders | Get paged OIDC providers
[**updateOidcProvider**](OidcProviderApi.md#updateoidcprovider) | **PUT** /oidcProvider/{oidcProviderId} | Update an OIDC provider


# **createOidcProvider**
> OidcProviderView createOidcProvider(upsertOidcProviderCommand)

Create an OIDC provider

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcProviderApi();
final UpsertOidcProviderCommand upsertOidcProviderCommand = ; // UpsertOidcProviderCommand | The provider configuration

try {
    final response = api.createOidcProvider(upsertOidcProviderCommand);
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcProviderApi->createOidcProvider: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **upsertOidcProviderCommand** | [**UpsertOidcProviderCommand**](UpsertOidcProviderCommand.md)| The provider configuration | 

### Return type

[**OidcProviderView**](OidcProviderView.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteOidcProvider**
> deleteOidcProvider(oidcProviderId)

Delete an OIDC provider

Accounts linked to it lose the ability to sign in with it.

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcProviderApi();
final String oidcProviderId = oidcProviderId_example; // String | 

try {
    api.deleteOidcProvider(oidcProviderId);
} catch on DioException (e) {
    print('Exception when calling OidcProviderApi->deleteOidcProvider: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **oidcProviderId** | **String**|  | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getOidcProviderById**
> OidcProviderView getOidcProviderById(oidcProviderId)

Get an OIDC provider

The client secret is never returned; hasClientSecret reports whether one is stored.

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcProviderApi();
final String oidcProviderId = oidcProviderId_example; // String | 

try {
    final response = api.getOidcProviderById(oidcProviderId);
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcProviderApi->getOidcProviderById: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **oidcProviderId** | **String**|  | 

### Return type

[**OidcProviderView**](OidcProviderView.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPagedOidcProviders**
> PagedData getPagedOidcProviders(pagedRequestCommand)

Get paged OIDC providers

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcProviderApi();
final PagedRequestCommand pagedRequestCommand = ; // PagedRequestCommand | Paging and sorting data

try {
    final response = api.getPagedOidcProviders(pagedRequestCommand);
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcProviderApi->getPagedOidcProviders: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **pagedRequestCommand** | [**PagedRequestCommand**](PagedRequestCommand.md)| Paging and sorting data | 

### Return type

[**PagedData**](PagedData.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateOidcProvider**
> OidcProviderView updateOidcProvider(oidcProviderId, upsertOidcProviderCommand)

Update an OIDC provider

Omit clientSecret to keep the stored one. The name cannot be changed -- it is part of the redirect URI already registered with the identity provider.

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcProviderApi();
final String oidcProviderId = oidcProviderId_example; // String | 
final UpsertOidcProviderCommand upsertOidcProviderCommand = ; // UpsertOidcProviderCommand | The provider configuration

try {
    final response = api.updateOidcProvider(oidcProviderId, upsertOidcProviderCommand);
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcProviderApi->updateOidcProvider: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **oidcProviderId** | **String**|  | 
 **upsertOidcProviderCommand** | [**UpsertOidcProviderCommand**](UpsertOidcProviderCommand.md)| The provider configuration | 

### Return type

[**OidcProviderView**](OidcProviderView.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

