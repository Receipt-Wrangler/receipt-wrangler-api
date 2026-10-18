# openapi.model.GroupSettings

## Load the model package
```dart
import 'package:openapi/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **int** | Group settings id | 
**groupId** | **int** | Group foreign key | 
**emailIntegrationEnabled** | **bool** | Whether email integration is enabled | [optional] 
**emailBodyProcessingEnabled** | **bool** | Whether email body text processing is enabled (opt-in, default false) | [optional] 
**emailRequireSignedSender** | **bool** | Whether an email is accepted only when it carries a valid DKIM signature from its From domain (default true). Omitting it on update leaves the stored value unchanged. | [optional] 
**systemEmailId** | **int** | System email foreign key | [optional] 
**systemEmail** | [**SystemEmail**](SystemEmail.md) |  | [optional] 
**emailToRead** | **String** | Email to read | [optional] 
**subjectLineRegexes** | [**BuiltList&lt;SubjectLineRegex&gt;**](SubjectLineRegex.md) | Subject line regexes | [optional] 
**emailWhiteList** | [**BuiltList&lt;GroupSettingsWhiteListEmail&gt;**](GroupSettingsWhiteListEmail.md) | Email white list | [optional] 
**emailDefaultReceiptStatus** | [**ReceiptStatus**](ReceiptStatus.md) | Default receipt status | [optional] 
**emailDefaultReceiptPaidById** | **int** | User foreign key | [optional] 
**prompt** | [**Prompt**](Prompt.md) |  | [optional] 
**promptId** | **int** | Prompt foreign key | [optional] 
**fallbackPrompt** | [**Prompt**](Prompt.md) |  | [optional] 
**fallbackPromptId** | **int** | Fallback prompt foreign key | [optional] 
**createdAt** | **String** |  | [optional] 
**createdBy** | **int** |  | [optional] 
**updatedAt** | **String** |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


