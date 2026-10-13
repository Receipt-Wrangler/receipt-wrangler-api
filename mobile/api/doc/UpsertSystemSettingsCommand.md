# openapi.model.UpsertSystemSettingsCommand

## Load the model package
```dart
import 'package:openapi/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**enableLocalSignUp** | **bool** | Whether local sign up is enabled | [optional] 
**currencyDisplay** | **String** | Currency display | [optional] 
**currencyThousandthsSeparator** | [**CurrencySeparator**](CurrencySeparator.md) |  | 
**currencyDecimalSeparator** | [**CurrencySeparator**](CurrencySeparator.md) |  | 
**currencySymbolPosition** | [**CurrencySymbolPosition**](CurrencySymbolPosition.md) |  | 
**currencyHideDecimalPlaces** | **bool** | Whether to hide decimal places | 
**debugOcr** | **bool** |  | [optional] 
**numWorkers** | **int** | Number of workers to use | [optional] [default to 1]
**emailPollingInterval** | **int** | Email polling interval | [optional] 
**receiptProcessingSettingsId** | **int** | Receipt processing settings foreign key | [optional] 
**fallbackReceiptProcessingSettingsId** | **int** | Fallback receipt processing settings foreign key | [optional] 
**taskConcurrency** | **int** | Concurrency for task worker | 
**pdfDpi** | **int** | DPI used when rasterizing PDFs for OCR/vision processing | [optional] 
**taskQueueConfigurations** | [**BuiltList&lt;UpsertTaskQueueConfiguration&gt;**](UpsertTaskQueueConfiguration.md) |  | [optional] 
**mcpEnabled** | **bool** | Whether the OAuth 2.1-protected MCP server is enabled | [optional] 
**mcpPublicUrl** | **String** | Externally reachable origin used for MCP OAuth/metadata/redirect URLs and token audience | [optional] 
**serverPublicUrl** | **String** | Externally reachable origin of this API, used to build each OIDC provider's redirect URI. Separate from mcpPublicUrl, which is bound into the MCP token audience. | [optional] 
**showLoginQr** | **bool** | Whether to show the mobile-setup QR code on the desktop login page | [optional] 
**mobileServerUrl** | **String** | Server/API URL mobile clients connect to; encoded into the login QR's deep link | [optional] 
**refreshTokenValidForHours** | **int** | How long a refresh token stays valid, in hours. Accepted values are 1-720 (30 days), or 0 meaning unset, which falls back to the default of 24. Omit the key entirely to leave the currently configured value unchanged. | [optional] 
**mcpRefreshTokenValidForHours** | **int** | How long an MCP/OAuth connector refresh token stays valid, in hours. Accepted values are 1-720 (30 days), or 0 meaning unset, which falls back to the default of 24. Omit the key entirely to leave the currently configured value unchanged. | [optional] 
**tempFileRetentionHours** | **int** | How long a temp file is kept once nothing can still act on it, in hours. Accepted values are 24-8760 (1 year), or 0 meaning unset, which falls back to the default of 720 (30 days). Omit the key entirely to leave the currently configured value unchanged. | [optional] 
**timeZone** | **String** | The app time zone, an IANA name such as America/New_York that the server can load (an unknown name is a 400 under the timeZone key). Omit the key entirely to leave the currently configured value unchanged. | [optional] 
**secureCookies** | **bool** | Marks every cookie the server sets as Secure. Only turn it on when the app is served over HTTPS: a browser drops a Secure cookie set over plain HTTP, which breaks login. Omit the key entirely to leave the currently configured value unchanged. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


