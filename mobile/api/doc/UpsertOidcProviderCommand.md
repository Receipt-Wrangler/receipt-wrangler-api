# openapi.model.UpsertOidcProviderCommand

## Load the model package
```dart
import 'package:openapi/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **String** | URL slug. Lowercase letters, numbers and dashes. Cannot be changed after creation, and cannot be one of the reserved words login, callback, link, exchange or connections. | 
**displayName** | **String** |  | 
**issuerUrl** | **String** | Must use https unless the host is localhost | 
**clientId** | **String** |  | 
**clientSecret** | **String** | Omit on update to keep the stored secret. Required on create. | [optional] 
**scope** | **String** | Space-separated OIDC scopes; must include openid | 
**allowProvisioning** | **bool** |  | [optional] 
**linkByUsername** | **bool** |  | [optional] 
**enabled** | **bool** |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


