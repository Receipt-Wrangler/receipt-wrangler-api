// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'delete_account_command.dart';

// **************************************************************************
// BuiltValueGenerator
// **************************************************************************

class _$DeleteAccountCommand extends DeleteAccountCommand {
  @override
  final String? password;
  @override
  final String? confirmationCode;

  factory _$DeleteAccountCommand(
          [void Function(DeleteAccountCommandBuilder)? updates]) =>
      (DeleteAccountCommandBuilder()..update(updates))._build();

  _$DeleteAccountCommand._({this.password, this.confirmationCode}) : super._();
  @override
  DeleteAccountCommand rebuild(
          void Function(DeleteAccountCommandBuilder) updates) =>
      (toBuilder()..update(updates)).build();

  @override
  DeleteAccountCommandBuilder toBuilder() =>
      DeleteAccountCommandBuilder()..replace(this);

  @override
  bool operator ==(Object other) {
    if (identical(other, this)) return true;
    return other is DeleteAccountCommand &&
        password == other.password &&
        confirmationCode == other.confirmationCode;
  }

  @override
  int get hashCode {
    var _$hash = 0;
    _$hash = $jc(_$hash, password.hashCode);
    _$hash = $jc(_$hash, confirmationCode.hashCode);
    _$hash = $jf(_$hash);
    return _$hash;
  }

  @override
  String toString() {
    return (newBuiltValueToStringHelper(r'DeleteAccountCommand')
          ..add('password', password)
          ..add('confirmationCode', confirmationCode))
        .toString();
  }
}

class DeleteAccountCommandBuilder
    implements Builder<DeleteAccountCommand, DeleteAccountCommandBuilder> {
  _$DeleteAccountCommand? _$v;

  String? _password;
  String? get password => _$this._password;
  set password(String? password) => _$this._password = password;

  String? _confirmationCode;
  String? get confirmationCode => _$this._confirmationCode;
  set confirmationCode(String? confirmationCode) =>
      _$this._confirmationCode = confirmationCode;

  DeleteAccountCommandBuilder() {
    DeleteAccountCommand._defaults(this);
  }

  DeleteAccountCommandBuilder get _$this {
    final $v = _$v;
    if ($v != null) {
      _password = $v.password;
      _confirmationCode = $v.confirmationCode;
      _$v = null;
    }
    return this;
  }

  @override
  void replace(DeleteAccountCommand other) {
    _$v = other as _$DeleteAccountCommand;
  }

  @override
  void update(void Function(DeleteAccountCommandBuilder)? updates) {
    if (updates != null) updates(this);
  }

  @override
  DeleteAccountCommand build() => _build();

  _$DeleteAccountCommand _build() {
    final _$result = _$v ??
        _$DeleteAccountCommand._(
          password: password,
          confirmationCode: confirmationCode,
        );
    replace(_$result);
    return _$result;
  }
}

// ignore_for_file: deprecated_member_use_from_same_package,type=lint
