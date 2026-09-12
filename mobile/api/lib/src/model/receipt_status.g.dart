// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'receipt_status.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

const ReceiptStatus _$OPEN = const ReceiptStatus._('OPEN');
const ReceiptStatus _$NEEDS_ATTENTION =
    const ReceiptStatus._('NEEDS_ATTENTION');
const ReceiptStatus _$RESOLVED = const ReceiptStatus._('RESOLVED');
const ReceiptStatus _$DRAFT = const ReceiptStatus._('DRAFT');
const ReceiptStatus _$DECLINED = const ReceiptStatus._('DECLINED');
const ReceiptStatus _$empty = const ReceiptStatus._('empty');

ReceiptStatus _$valueOf(String name) {
  switch (name) {
    case 'OPEN':
      return _$OPEN;
    case 'NEEDS_ATTENTION':
      return _$NEEDS_ATTENTION;
    case 'RESOLVED':
      return _$RESOLVED;
    case 'DRAFT':
      return _$DRAFT;
    case 'DECLINED':
      return _$DECLINED;
    case 'empty':
      return _$empty;
    default:
      return _$empty;
  }
}

final BuiltSet<ReceiptStatus> _$values =
    BuiltSet<ReceiptStatus>(const <ReceiptStatus>[
  _$OPEN,
  _$NEEDS_ATTENTION,
  _$RESOLVED,
  _$DRAFT,
  _$DECLINED,
  _$empty,
]);

class _$ReceiptStatusMeta {
  const _$ReceiptStatusMeta();
  ReceiptStatus get OPEN => _$OPEN;
  ReceiptStatus get NEEDS_ATTENTION => _$NEEDS_ATTENTION;
  ReceiptStatus get RESOLVED => _$RESOLVED;
  ReceiptStatus get DRAFT => _$DRAFT;
  ReceiptStatus get DECLINED => _$DECLINED;
  ReceiptStatus get empty => _$empty;
  ReceiptStatus valueOf(String name) => _$valueOf(name);
  BuiltSet<ReceiptStatus> get values => _$values;
}

abstract class _$ReceiptStatusMixin {
  // ignore: non_constant_identifier_names
  _$ReceiptStatusMeta get ReceiptStatus => const _$ReceiptStatusMeta();
}

Serializer<ReceiptStatus> _$receiptStatusSerializer =
    _$ReceiptStatusSerializer();

class _$ReceiptStatusSerializer implements PrimitiveSerializer<ReceiptStatus> {
  static const Map<String, Object> _toWire = const <String, Object>{
    'OPEN': 'OPEN',
    'NEEDS_ATTENTION': 'NEEDS_ATTENTION',
    'RESOLVED': 'RESOLVED',
    'DRAFT': 'DRAFT',
    'DECLINED': 'DECLINED',
    'empty': '',
  };
  static const Map<Object, String> _fromWire = const <Object, String>{
    'OPEN': 'OPEN',
    'NEEDS_ATTENTION': 'NEEDS_ATTENTION',
    'RESOLVED': 'RESOLVED',
    'DRAFT': 'DRAFT',
    'DECLINED': 'DECLINED',
    '': 'empty',
  };

  @override
  final Iterable<Type> types = const <Type>[ReceiptStatus];
  @override
  final String wireName = 'ReceiptStatus';

  @override
  Object serialize(Serializers serializers, ReceiptStatus object,
          {FullType specifiedType = FullType.unspecified}) =>
      _toWire[object.name] ?? object.name;

  @override
  ReceiptStatus deserialize(Serializers serializers, Object serialized,
          {FullType specifiedType = FullType.unspecified}) =>
      ReceiptStatus.valueOf(
          _fromWire[serialized] ?? (serialized is String ? serialized : ''));
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
