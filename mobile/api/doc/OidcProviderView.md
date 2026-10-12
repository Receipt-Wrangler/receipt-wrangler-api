# openapi.model.OidcProviderView

## Load the model package
```dart
import 'package:openapi/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** |  | 
**name** | **String** | URL slug for this provider; immutable after creation | 
**displayName** | **String** |  | 
**issuerUrl** | **String** | OIDC discovery base, e.g. https://accounts.google.com | 
**clientId** | **String** |  | 
**scope** | **String** | Space-separated OIDC scopes; must include openid | 
**allowProvisioning** | **bool** | Create a local account for an identity we have never seen | 
**linkByUsername** | **bool** | On a first login only, attach to an existing local account whose username equals the preferred_username claim. Off by default; that claim is neither stable nor unique, and some providers recycle released usernames. | 
**enabled** | **bool** |  | 
**hasClientSecret** | **bool** | Whether a secret is stored. The secret itself is never returned. | 
**redirectUri** | **String** | The exact redirect URI to register with the identity provider | 
**createdAt** | [**DateTime**](DateTime.md) |  | [optional] 
**updatedAt** | [**DateTime**](DateTime.md) |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


