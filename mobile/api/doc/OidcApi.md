# openapi.api.OidcApi

## Load the API package
```dart
import 'package:openapi/api.dart';
```

All URIs are relative to */api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**deleteOidcConnection**](OidcApi.md#deleteoidcconnection) | **DELETE** /oidc/connections/{name} | Disconnect a provider from the caller&#39;s account
[**getOidcConnections**](OidcApi.md#getoidcconnections) | **GET** /oidc/connections | List the caller&#39;s connected accounts
[**oidcCallback**](OidcApi.md#oidccallback) | **GET** /oidc/{name}/callback | OIDC redirect URI
[**oidcExchange**](OidcApi.md#oidcexchange) | **POST** /oidc/exchange | Redeem a mobile sign-in code
[**oidcLinkLaunch**](OidcApi.md#oidclinklaunch) | **GET** /oidc/link/{name}/launch | Open a mobile \&quot;connect account\&quot; flow in the browser
[**oidcLinkStart**](OidcApi.md#oidclinkstart) | **GET** /oidc/link/{name} | Connect a provider to the signed-in account
[**oidcLogin**](OidcApi.md#oidclogin) | **GET** /oidc/{name}/login | Start an OIDC login


# **deleteOidcConnection**
> deleteOidcConnection(name)

Disconnect a provider from the caller's account

Refused when it is the caller's last connection and the account was created by that provider, because such an account has no password to fall back on.

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcApi();
final String name = name_example; // String | The provider's slug

try {
    api.deleteOidcConnection(name);
} catch on DioException (e) {
    print('Exception when calling OidcApi->deleteOidcConnection: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **String**| The provider's slug | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getOidcConnections**
> BuiltList<OidcConnectionView> getOidcConnections()

List the caller's connected accounts

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcApi();

try {
    final response = api.getOidcConnections();
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcApi->getOidcConnections: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**BuiltList&lt;OidcConnectionView&gt;**](OidcConnectionView.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **oidcCallback**
> oidcCallback(name, code, state)

OIDC redirect URI

Where the identity provider returns the user. This is the exact URL that must be registered with the provider. Not called by clients directly.

### Example
```dart
import 'package:openapi/api.dart';

final api = Openapi().getOidcApi();
final String name = name_example; // String | The provider's slug
final String code = code_example; // String | 
final String state = state_example; // String | 

try {
    api.oidcCallback(name, code, state);
} catch on DioException (e) {
    print('Exception when calling OidcApi->oidcCallback: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **String**| The provider's slug | 
 **code** | **String**|  | [optional] 
 **state** | **String**|  | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **oidcExchange**
> AppData oidcExchange(oidcExchangeCommand)

Redeem a mobile sign-in code

Trades the single-use code the app received on its private-use URL scheme for a session. The PKCE verifier is what proves this is the app that started the flow, so an app that intercepted the redirect cannot redeem the code. Returns the same payload as login with tokensInBody, and never sets a cookie.

### Example
```dart
import 'package:openapi/api.dart';

final api = Openapi().getOidcApi();
final OidcExchangeCommand oidcExchangeCommand = ; // OidcExchangeCommand | The code and the app's PKCE verifier

try {
    final response = api.oidcExchange(oidcExchangeCommand);
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcApi->oidcExchange: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **oidcExchangeCommand** | [**OidcExchangeCommand**](OidcExchangeCommand.md)| The code and the app's PKCE verifier | 

### Return type

[**AppData**](AppData.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **oidcLinkLaunch**
> oidcLinkLaunch(name, h)

Open a mobile \"connect account\" flow in the browser

The browser-facing half of a mobile link. A mobile client opens this URL, from oidcLinkStart's launchUrl, in the external user agent; it is never called as an API request.  Unauthenticated because it has to be reachable by that browser, which cannot carry a bearer token. The single-use handle in the query is the authorization - oidcLinkStart minted it moments earlier for the caller it had already authenticated and permission-checked. Letting the browser make this request is what allows the flow to be bound to it, which is the defense the callback relies on.

### Example
```dart
import 'package:openapi/api.dart';

final api = Openapi().getOidcApi();
final String name = name_example; // String | The provider's slug
final String h = h_example; // String | The single-use launch handle from oidcLinkStart

try {
    api.oidcLinkLaunch(name, h);
} catch on DioException (e) {
    print('Exception when calling OidcApi->oidcLinkLaunch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **String**| The provider's slug | 
 **h** | **String**| The single-use launch handle from oidcLinkStart | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **oidcLinkStart**
> OidcLinkStartView oidcLinkStart(name, client)

Connect a provider to the signed-in account

Starts the same flow as a login, but the session proves who the caller is, so the callback links the identity directly instead of matching or provisioning. Requires app.account.update, the same permission as disconnecting one.  A browser navigates to this URL and is redirected (302) to the identity provider. A native client passes client=mobile and gets a launch URL as JSON instead, because it authenticates with a bearer token and the external user agent it must use cannot carry one. It opens that launch URL in the external browser, which is what actually starts the flow - and what receives the binding cookie the callback checks. The app then waits on its private-use scheme, which the callback returns to with ?linked={name} or ?error={code}.

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getOidcApi();
final String name = name_example; // String | The provider's slug
final String client = client_example; // String | Which client is connecting. Mobile receives a launch URL as JSON rather than a redirect. No codeChallenge is needed here, unlike a mobile login - a link's callback hands back only ?linked={name}, which is not a credential, so there is nothing for the app to bind.

try {
    final response = api.oidcLinkStart(name, client);
    print(response);
} catch on DioException (e) {
    print('Exception when calling OidcApi->oidcLinkStart: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **String**| The provider's slug | 
 **client** | **String**| Which client is connecting. Mobile receives a launch URL as JSON rather than a redirect. No codeChallenge is needed here, unlike a mobile login - a link's callback hands back only ?linked={name}, which is not a credential, so there is nothing for the app to bind. | [optional] [default to 'desktop']

### Return type

[**OidcLinkStartView**](OidcLinkStartView.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **oidcLogin**
> oidcLogin(name, client, codeChallenge)

Start an OIDC login

Redirects the user agent to the configured identity provider. The API acts as the relying party, so the whole exchange (PKCE, state, nonce, code exchange, ID token verification) happens server-side and no client ever handles an identity provider token. Navigate to this URL; do not fetch it.

### Example
```dart
import 'package:openapi/api.dart';

final api = Openapi().getOidcApi();
final String name = name_example; // String | The provider's slug
final String client = client_example; // String | Which client is signing in. Mobile receives a single-use exchange code on the app's private-use URL scheme instead of session cookies.
final String codeChallenge = codeChallenge_example; // String | The mobile app's own PKCE S256 challenge. Required when client is mobile; it binds the resulting exchange code to the app that started the flow.

try {
    api.oidcLogin(name, client, codeChallenge);
} catch on DioException (e) {
    print('Exception when calling OidcApi->oidcLogin: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **String**| The provider's slug | 
 **client** | **String**| Which client is signing in. Mobile receives a single-use exchange code on the app's private-use URL scheme instead of session cookies. | [optional] [default to 'desktop']
 **codeChallenge** | **String**| The mobile app's own PKCE S256 challenge. Required when client is mobile; it binds the resulting exchange code to the app that started the flow. | [optional] 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

