# openapi.api.UserApi

## Load the API package
```dart
import 'package:openapi/api.dart';
```

All URIs are relative to */api*

Method | HTTP request | Description
------------- | ------------- | -------------
[**bulkDeleteUsers**](UserApi.md#bulkdeleteusers) | **DELETE** /user/bulk | Bulk delete users
[**convertDummyUserById**](UserApi.md#convertdummyuserbyid) | **POST** /user/{userId}/convertDummyUserToNormalUser | Converts dummy user
[**createUser**](UserApi.md#createuser) | **POST** /user | Create user
[**deleteAccount**](UserApi.md#deleteaccount) | **POST** /user/deleteAccount | Delete own account
[**deleteUserById**](UserApi.md#deleteuserbyid) | **DELETE** /user/{userId} | Delete user
[**getAmountOwedForUser**](UserApi.md#getamountowedforuser) | **GET** /user/amountOwedForUser | Get amount owed for user
[**getAppData**](UserApi.md#getappdata) | **GET** /user/appData | Get app data
[**getPagedUsers**](UserApi.md#getpagedusers) | **POST** /user/getPagedUsers | Gets paged users
[**getUserClaims**](UserApi.md#getuserclaims) | **GET** /user/getUserClaims | Get claims for logged in user
[**getUsernameCount**](UserApi.md#getusernamecount) | **GET** /user/{username} | Get username count
[**getUsers**](UserApi.md#getusers) | **GET** /user | Get users
[**resetPasswordById**](UserApi.md#resetpasswordbyid) | **POST** /user/{userId}/resetPassword | Reset password
[**updateUserById**](UserApi.md#updateuserbyid) | **PUT** /user/{userId} | Update user by id
[**updateUserProfile**](UserApi.md#updateuserprofile) | **PUT** /user/updateUserProfile | Update user profile


# **bulkDeleteUsers**
> bulkDeleteUsers(bulkUserDeleteCommand)

Bulk delete users

This will delete multiple users by their IDs [SYSTEM ADMIN]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final BulkUserDeleteCommand bulkUserDeleteCommand = ; // BulkUserDeleteCommand | User IDs to delete

try {
    api.bulkDeleteUsers(bulkUserDeleteCommand);
} catch on DioException (e) {
    print('Exception when calling UserApi->bulkDeleteUsers: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **bulkUserDeleteCommand** | [**BulkUserDeleteCommand**](BulkUserDeleteCommand.md)| User IDs to delete | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **convertDummyUserById**
> convertDummyUserById(userId, resetPasswordCommand)

Converts dummy user

This will convert a dummy user to a normal system user, [SYSTEM ADMIN]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final int userId = 56; // int | Id of user to convert to normal system user
final ResetPasswordCommand resetPasswordCommand = ; // ResetPasswordCommand | Login credentials for new user

try {
    api.convertDummyUserById(userId, resetPasswordCommand);
} catch on DioException (e) {
    print('Exception when calling UserApi->convertDummyUserById: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **int**| Id of user to convert to normal system user | 
 **resetPasswordCommand** | [**ResetPasswordCommand**](ResetPasswordCommand.md)| Login credentials for new user | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createUser**
> createUser(user)

Create user

This will to create a user, [SYSTEM ADMIN]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final User user = ; // User | User to create

try {
    api.createUser(user);
} catch on DioException (e) {
    print('Exception when calling UserApi->createUser: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **user** | [**User**](User.md)| User to create | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteAccount**
> deleteAccount(deleteAccountCommand)

Delete own account

Deletes the signed-in user's own account once it is confirmed. An account with a password of its own confirms with that password (401 when it is wrong). An account without one - created by signing in with an identity provider - confirms with the confirmationCode that oidcDeleteConfirmationStart hands back (400 when it is missing, spent, expired or not this account's). Which applies is decided by the account, not by which field is sent. Requires an interactive session; an API key is refused with 403 whatever its scope.

### Example
```dart
import 'package:openapi/api.dart';

final api = Openapi().getUserApi();
final DeleteAccountCommand deleteAccountCommand = ; // DeleteAccountCommand | Confirmation for account deletion

try {
    api.deleteAccount(deleteAccountCommand);
} catch on DioException (e) {
    print('Exception when calling UserApi->deleteAccount: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **deleteAccountCommand** | [**DeleteAccountCommand**](DeleteAccountCommand.md)| Confirmation for account deletion | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUserById**
> deleteUserById(userId)

Delete user

This will delete a system user by id [SYSTEM ADMIN]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final int userId = 56; // int | Id of user to update

try {
    api.deleteUserById(userId);
} catch on DioException (e) {
    print('Exception when calling UserApi->deleteUserById: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **int**| Id of user to update | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAmountOwedForUser**
> BuiltMap<String, String> getAmountOwedForUser(groupId, receiptIds)

Get amount owed for user

This will return the amount owed for the logged in user, in the specified group, [SYSTEM USER]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final int groupId = 56; // int | The Id of the group to get amount owed for
final BuiltList<int> receiptIds = ; // BuiltList<int> | The Id of the receipts to get amount owed for

try {
    final response = api.getAmountOwedForUser(groupId, receiptIds);
    print(response);
} catch on DioException (e) {
    print('Exception when calling UserApi->getAmountOwedForUser: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **groupId** | **int**| The Id of the group to get amount owed for | [optional] 
 **receiptIds** | [**BuiltList&lt;int&gt;**](int.md)| The Id of the receipts to get amount owed for | [optional] 

### Return type

**BuiltMap&lt;String, String&gt;**

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAppData**
> AppData getAppData()

Get app data

This will return the user's app data for the currently logged in user [SYSTEM USER]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();

try {
    final response = api.getAppData();
    print(response);
} catch on DioException (e) {
    print('Exception when calling UserApi->getAppData: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**AppData**](AppData.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getPagedUsers**
> PagedData getPagedUsers(pagedRequestCommand)

Gets paged users

This will return paged users

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final PagedRequestCommand pagedRequestCommand = ; // PagedRequestCommand | Paging and sorting data

try {
    final response = api.getPagedUsers(pagedRequestCommand);
    print(response);
} catch on DioException (e) {
    print('Exception when calling UserApi->getPagedUsers: $e\n');
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

# **getUserClaims**
> Claims getUserClaims()

Get claims for logged in user

This will return the user's token claims for the currently logged in user [SYSTEM USER]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();

try {
    final response = api.getUserClaims();
    print(response);
} catch on DioException (e) {
    print('Exception when calling UserApi->getUserClaims: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**Claims**](Claims.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUsernameCount**
> int getUsernameCount(username)

Get username count

This will return the number of users in the system with the same username

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final String username = username_example; // String | Username to get the count of

try {
    final response = api.getUsernameCount(username);
    print(response);
} catch on DioException (e) {
    print('Exception when calling UserApi->getUsernameCount: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **username** | **String**| Username to get the count of | 

### Return type

**int**

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUsers**
> BuiltList<UserView> getUsers()

Get users

This will get all the users in the system and return a view without sensative information

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();

try {
    final response = api.getUsers();
    print(response);
} catch on DioException (e) {
    print('Exception when calling UserApi->getUsers: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**BuiltList&lt;UserView&gt;**](UserView.md)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **resetPasswordById**
> resetPasswordById(userId, resetPasswordCommand)

Reset password

This will reset a password for a user, [SYSTEM ADMIN]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final int userId = 56; // int | Id of user to reset password
final ResetPasswordCommand resetPasswordCommand = ; // ResetPasswordCommand | Login credentials for new user

try {
    api.resetPasswordById(userId, resetPasswordCommand);
} catch on DioException (e) {
    print('Exception when calling UserApi->resetPasswordById: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **int**| Id of user to reset password | 
 **resetPasswordCommand** | [**ResetPasswordCommand**](ResetPasswordCommand.md)| Login credentials for new user | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUserById**
> updateUserById(userId, user)

Update user by id

This will update a user by id, [SYSTEM ADMIN]

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final int userId = 56; // int | Id of user to update
final User user = ; // User | User to update

try {
    api.updateUserById(userId, user);
} catch on DioException (e) {
    print('Exception when calling UserApi->updateUserById: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **int**| Id of user to update | 
 **user** | [**User**](User.md)| User to update | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUserProfile**
> updateUserProfile(updateProfileCommand)

Update user profile

This will update the logged in user's user profile

### Example
```dart
import 'package:openapi/api.dart';
// TODO Configure API key authorization: apiKeyAuth
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKey = 'YOUR_API_KEY';
// uncomment below to setup prefix (e.g. Bearer) for API key, if needed
//defaultApiClient.getAuthentication<ApiKeyAuth>('apiKeyAuth').apiKeyPrefix = 'Bearer';

final api = Openapi().getUserApi();
final UpdateProfileCommand updateProfileCommand = ; // UpdateProfileCommand | User profile to update

try {
    api.updateUserProfile(updateProfileCommand);
} catch on DioException (e) {
    print('Exception when calling UserApi->updateUserProfile: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **updateProfileCommand** | [**UpdateProfileCommand**](UpdateProfileCommand.md)| User profile to update | 

### Return type

void (empty response body)

### Authorization

[apiKeyAuth](../README.md#apiKeyAuth), [bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

