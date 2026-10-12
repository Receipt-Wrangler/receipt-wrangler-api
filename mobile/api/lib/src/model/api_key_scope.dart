//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:built_collection/built_collection.dart';
import 'package:built_value/built_value.dart';
import 'package:built_value/serializer.dart';

part 'api_key_scope.g.dart';

class ApiKeyScope extends EnumClass {

  /// What an API key may do, enforced on every request. The values are literal - \"r\" permits reads and denies writes, \"w\" permits writes and denies reads, \"rw\" permits both. \"Read\" means data is returned to the caller, not that a row was read internally, so a \"w\" key may update freely. Reads are GET/HEAD/OPTIONS plus the endpoints that answer a query over POST because their filter arrives in the request body (the paged lists, the report endpoints, the exports, the pie chart, the receipt summary). A few endpoints both return data and change it - duplicating a receipt or a report template answers with the source's contents - and those require \"rw\". A request outside the key's scope is refused with 403.
  @BuiltValueEnumConst(wireName: r'r')
  static const ApiKeyScope r = _$r;
  /// What an API key may do, enforced on every request. The values are literal - \"r\" permits reads and denies writes, \"w\" permits writes and denies reads, \"rw\" permits both. \"Read\" means data is returned to the caller, not that a row was read internally, so a \"w\" key may update freely. Reads are GET/HEAD/OPTIONS plus the endpoints that answer a query over POST because their filter arrives in the request body (the paged lists, the report endpoints, the exports, the pie chart, the receipt summary). A few endpoints both return data and change it - duplicating a receipt or a report template answers with the source's contents - and those require \"rw\". A request outside the key's scope is refused with 403.
  @BuiltValueEnumConst(wireName: r'w')
  static const ApiKeyScope w = _$w;
  /// What an API key may do, enforced on every request. The values are literal - \"r\" permits reads and denies writes, \"w\" permits writes and denies reads, \"rw\" permits both. \"Read\" means data is returned to the caller, not that a row was read internally, so a \"w\" key may update freely. Reads are GET/HEAD/OPTIONS plus the endpoints that answer a query over POST because their filter arrives in the request body (the paged lists, the report endpoints, the exports, the pie chart, the receipt summary). A few endpoints both return data and change it - duplicating a receipt or a report template answers with the source's contents - and those require \"rw\". A request outside the key's scope is refused with 403.
  @BuiltValueEnumConst(wireName: r'rw')
  static const ApiKeyScope rw = _$rw;

  static Serializer<ApiKeyScope> get serializer => _$apiKeyScopeSerializer;

  const ApiKeyScope._(String name): super(name);

  static BuiltSet<ApiKeyScope> get values => _$values;
  static ApiKeyScope valueOf(String name) => _$valueOf(name);
}

/// Optionally, enum_class can generate a mixin to go with your enum for use
/// with Angular. It exposes your enum constants as getters. So, if you mix it
/// in to your Dart component class, the values become available to the
/// corresponding Angular template.
///
/// Trigger mixin generation by writing a line like this one next to your enum.
abstract class ApiKeyScopeMixin = Object with _$ApiKeyScopeMixin;

