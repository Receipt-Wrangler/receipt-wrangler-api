# openapi.model.OidcConnectionView

## Load the model package
```dart
import 'package:openapi/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**providerName** | **String** |  | 
**providerDisplayName** | **String** |  | 
**preferredUsername** | **String** |  | [optional] 
**email** | **String** |  | [optional] 
**provisionedUser** | **bool** | Whether this connection created the local account. Such an account has no usable password, so its last connection cannot be removed. | 
**linkedAt** | [**DateTime**](DateTime.md) |  | 
**lastLoginAt** | [**DateTime**](DateTime.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


