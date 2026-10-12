//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

import 'package:dio/dio.dart';
import 'package:built_value/serializer.dart';
import 'package:openapi/src/serializers.dart';
import 'package:openapi/src/auth/api_key_auth.dart';
import 'package:openapi/src/auth/basic_auth.dart';
import 'package:openapi/src/auth/bearer_auth.dart';
import 'package:openapi/src/auth/oauth.dart';
import 'package:openapi/src/api/api_key_api.dart';
import 'package:openapi/src/api/auth_api.dart';
import 'package:openapi/src/api/category_api.dart';
import 'package:openapi/src/api/comment_api.dart';
import 'package:openapi/src/api/custom_field_api.dart';
import 'package:openapi/src/api/dashboard_api.dart';
import 'package:openapi/src/api/export_api.dart';
import 'package:openapi/src/api/feature_config_api.dart';
import 'package:openapi/src/api/groups_api.dart';
import 'package:openapi/src/api/import_api.dart';
import 'package:openapi/src/api/notifications_api.dart';
import 'package:openapi/src/api/oidc_api.dart';
import 'package:openapi/src/api/oidc_provider_api.dart';
import 'package:openapi/src/api/permission_api.dart';
import 'package:openapi/src/api/prompt_api.dart';
import 'package:openapi/src/api/receipt_api.dart';
import 'package:openapi/src/api/receipt_image_api.dart';
import 'package:openapi/src/api/receipt_processing_settings_api.dart';
import 'package:openapi/src/api/report_api.dart';
import 'package:openapi/src/api/role_api.dart';
import 'package:openapi/src/api/search_api.dart';
import 'package:openapi/src/api/system_email_api.dart';
import 'package:openapi/src/api/system_settings_api.dart';
import 'package:openapi/src/api/system_task_api.dart';
import 'package:openapi/src/api/tag_api.dart';
import 'package:openapi/src/api/user_api.dart';
import 'package:openapi/src/api/user_preferences_api.dart';
import 'package:openapi/src/api/widget_api.dart';

class Openapi {
  static const String basePath = r'/api';

  final Dio dio;
  final Serializers serializers;

  Openapi({
    Dio? dio,
    Serializers? serializers,
    String? basePathOverride,
    List<Interceptor>? interceptors,
  })  : this.serializers = serializers ?? standardSerializers,
        this.dio = dio ??
            Dio(BaseOptions(
              baseUrl: basePathOverride ?? basePath,
              connectTimeout: const Duration(milliseconds: 5000),
              receiveTimeout: const Duration(milliseconds: 3000),
            )) {
    if (interceptors == null) {
      this.dio.interceptors.addAll([
        OAuthInterceptor(),
        BasicAuthInterceptor(),
        BearerAuthInterceptor(),
        ApiKeyAuthInterceptor(),
      ]);
    } else {
      this.dio.interceptors.addAll(interceptors);
    }
  }

  void setOAuthToken(String name, String token) {
    if (this.dio.interceptors.any((i) => i is OAuthInterceptor)) {
      (this.dio.interceptors.firstWhere((i) => i is OAuthInterceptor) as OAuthInterceptor).tokens[name] = token;
    }
  }

  void setBearerAuth(String name, String token) {
    if (this.dio.interceptors.any((i) => i is BearerAuthInterceptor)) {
      (this.dio.interceptors.firstWhere((i) => i is BearerAuthInterceptor) as BearerAuthInterceptor).tokens[name] = token;
    }
  }

  void setBasicAuth(String name, String username, String password) {
    if (this.dio.interceptors.any((i) => i is BasicAuthInterceptor)) {
      (this.dio.interceptors.firstWhere((i) => i is BasicAuthInterceptor) as BasicAuthInterceptor).authInfo[name] = BasicAuthInfo(username, password);
    }
  }

  void setApiKey(String name, String apiKey) {
    if (this.dio.interceptors.any((i) => i is ApiKeyAuthInterceptor)) {
      (this.dio.interceptors.firstWhere((element) => element is ApiKeyAuthInterceptor) as ApiKeyAuthInterceptor).apiKeys[name] = apiKey;
    }
  }

  /// Get ApiKeyApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ApiKeyApi getApiKeyApi() {
    return ApiKeyApi(dio, serializers);
  }

  /// Get AuthApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  AuthApi getAuthApi() {
    return AuthApi(dio, serializers);
  }

  /// Get CategoryApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  CategoryApi getCategoryApi() {
    return CategoryApi(dio, serializers);
  }

  /// Get CommentApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  CommentApi getCommentApi() {
    return CommentApi(dio, serializers);
  }

  /// Get CustomFieldApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  CustomFieldApi getCustomFieldApi() {
    return CustomFieldApi(dio, serializers);
  }

  /// Get DashboardApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  DashboardApi getDashboardApi() {
    return DashboardApi(dio, serializers);
  }

  /// Get ExportApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ExportApi getExportApi() {
    return ExportApi(dio, serializers);
  }

  /// Get FeatureConfigApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  FeatureConfigApi getFeatureConfigApi() {
    return FeatureConfigApi(dio, serializers);
  }

  /// Get GroupsApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  GroupsApi getGroupsApi() {
    return GroupsApi(dio, serializers);
  }

  /// Get ImportApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ImportApi getImportApi() {
    return ImportApi(dio, serializers);
  }

  /// Get NotificationsApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  NotificationsApi getNotificationsApi() {
    return NotificationsApi(dio, serializers);
  }

  /// Get OidcApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  OidcApi getOidcApi() {
    return OidcApi(dio, serializers);
  }

  /// Get OidcProviderApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  OidcProviderApi getOidcProviderApi() {
    return OidcProviderApi(dio, serializers);
  }

  /// Get PermissionApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  PermissionApi getPermissionApi() {
    return PermissionApi(dio, serializers);
  }

  /// Get PromptApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  PromptApi getPromptApi() {
    return PromptApi(dio, serializers);
  }

  /// Get ReceiptApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ReceiptApi getReceiptApi() {
    return ReceiptApi(dio, serializers);
  }

  /// Get ReceiptImageApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ReceiptImageApi getReceiptImageApi() {
    return ReceiptImageApi(dio, serializers);
  }

  /// Get ReceiptProcessingSettingsApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ReceiptProcessingSettingsApi getReceiptProcessingSettingsApi() {
    return ReceiptProcessingSettingsApi(dio, serializers);
  }

  /// Get ReportApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  ReportApi getReportApi() {
    return ReportApi(dio, serializers);
  }

  /// Get RoleApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  RoleApi getRoleApi() {
    return RoleApi(dio, serializers);
  }

  /// Get SearchApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  SearchApi getSearchApi() {
    return SearchApi(dio, serializers);
  }

  /// Get SystemEmailApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  SystemEmailApi getSystemEmailApi() {
    return SystemEmailApi(dio, serializers);
  }

  /// Get SystemSettingsApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  SystemSettingsApi getSystemSettingsApi() {
    return SystemSettingsApi(dio, serializers);
  }

  /// Get SystemTaskApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  SystemTaskApi getSystemTaskApi() {
    return SystemTaskApi(dio, serializers);
  }

  /// Get TagApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  TagApi getTagApi() {
    return TagApi(dio, serializers);
  }

  /// Get UserApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  UserApi getUserApi() {
    return UserApi(dio, serializers);
  }

  /// Get UserPreferencesApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  UserPreferencesApi getUserPreferencesApi() {
    return UserPreferencesApi(dio, serializers);
  }

  /// Get WidgetApi instance, base route and serializer can be overridden by a given but be careful,
  /// by doing that all interceptors will not be executed
  WidgetApi getWidgetApi() {
    return WidgetApi(dio, serializers);
  }
}
