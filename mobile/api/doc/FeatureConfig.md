# openapi.model.FeatureConfig

## Load the model package
```dart
import 'package:openapi/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**aiPoweredReceipts** | **bool** | Whether AI powered receipts are enabled | 
**enableLocalSignUp** | **bool** | Whether local sign up is enabled | 
**loginQrUrl** | **String** | Composed deep link the desktop login page encodes as a QR; empty unless the login QR is enabled with a mobile server URL | [optional] 
**oidcProviders** | [**BuiltList&lt;OidcProviderSummary&gt;**](OidcProviderSummary.md) | Enabled OIDC identity providers, so a login screen can render one button per provider. Always present as an array (never null) and carries only the slug and display name -- never the issuer, client id, or any secret. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


