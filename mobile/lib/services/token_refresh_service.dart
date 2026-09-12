import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:openapi/openapi.dart';
import 'package:receipt_wrangler_mobile/client/client.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/category_model.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/system_settings_model.dart';
import 'package:receipt_wrangler_mobile/models/tag_model.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/models/user_preferences_model.dart';
import 'package:receipt_wrangler_mobile/utils/auth.dart';

/// Serializes all token refresh calls so that only one HTTP request
/// is in-flight at a time. This prevents race conditions with the
/// backend's one-time-use refresh tokens.
///
/// Dart equivalent of the desktop's TokenRefreshService (shareReplay(1)).
class TokenRefreshService {
  static final TokenRefreshService _instance = TokenRefreshService._internal();

  factory TokenRefreshService() => _instance;

  TokenRefreshService._internal();

  Completer<bool>? _refreshCompleter;

  /// When AppData was last republished into the models. See
  /// [_loadAppData].
  DateTime? _appDataLoadedAt;

  /// Serializes AppData loads. See [_enqueueAppDataLoad].
  Future<void> _appDataQueue = Future<void>.value();

  late AuthModel _authModel;
  late GroupModel _groupModel;
  late UserModel _userModel;
  late UserPreferencesModel _userPreferencesModel;
  late CategoryModel _categoryModel;
  late TagModel _tagModel;
  late SystemSettingsModel _systemSettingsModel;
  late PermissionsModel _permissionsModel;

  bool _initialized = false;

  @visibleForTesting
  void resetForTesting() {
    _refreshCompleter = null;
    _initialized = false;
    _appDataLoadedAt = null;
    _appDataQueue = Future<void>.value();
  }

  void initialize({
    required AuthModel authModel,
    required GroupModel groupModel,
    required UserModel userModel,
    required UserPreferencesModel userPreferencesModel,
    required CategoryModel categoryModel,
    required TagModel tagModel,
    required SystemSettingsModel systemSettingsModel,
    required PermissionsModel permissionsModel,
  }) {
    _authModel = authModel;
    _groupModel = groupModel;
    _userModel = userModel;
    _userPreferencesModel = userPreferencesModel;
    _categoryModel = categoryModel;
    _tagModel = tagModel;
    _systemSettingsModel = systemSettingsModel;
    _permissionsModel = permissionsModel;
    _initialized = true;
  }

  /// Returns the current JWT for use by the auth interceptor.
  Future<String?> getCurrentJwt() => _authModel.getJwt();

  /// Serialized token refresh. If a refresh is already in-flight,
  /// all callers share the same Future (and thus the same HTTP request).
  Future<bool> refreshTokens({bool force = false}) async {
    if (!_initialized) return false;

    if (_refreshCompleter != null) {
      return _refreshCompleter!.future;
    }

    _refreshCompleter = Completer<bool>();

    try {
      final result = await _doRefresh(force: force);
      _refreshCompleter!.complete(result);
      return result;
    } catch (e) {
      _refreshCompleter!.complete(false);
      return false;
    } finally {
      _refreshCompleter = null;
    }
  }

  Future<bool> _doRefresh({bool force = false}) async {
    var jwt = await _authModel.getJwt();
    var refreshToken = await _authModel.getRefreshToken();

    bool needsRefresh = force || !isTokenValid(jwt);

    if (!needsRefresh) {
      await _tryLoadAppData();
      return true;
    }

    if (!isTokenValid(refreshToken)) {
      await _authModel.purgeTokens();
      return false;
    }

    try {
      await getAndSetTokens(_authModel);
    } catch (e) {
      print(e);
      await _authModel.purgeTokens();
      return false;
    }

    await _tryLoadAppData();
    return true;
  }

  /// Attempts to load app data, swallowing errors so that a transient
  /// getAppData failure never causes refreshTokens() to report false
  /// (which would trigger token purge and logout).
  Future<void> _tryLoadAppData() async {
    try {
      await _enqueueAppDataLoad(bypassDebounce: false);
    } catch (e) {
      print(e);
    }
  }

  /// Re-fetches AppData now, ignoring the debounce, for a caller that needs a
  /// guarantee rather than a best effort -- starting a Quick Scan, where the
  /// group's quick-scan field config decides which fields the form renders.
  ///
  /// Deliberately a separate method rather than a flag on [refreshTokens]:
  /// concurrent callers of that piggyback on `_refreshCompleter` and return
  /// *before* `_doRefresh` is entered, so a flag would be silently dropped for
  /// the piggybacking caller -- exactly the caller that most needs the reload.
  ///
  /// Never throws. The caller is on the path to a user action (the scanner is
  /// about to open), so a transient failure must fall through to whatever data
  /// the app already has rather than block the action.
  Future<void> reloadAppData() async {
    if (!_initialized) {
      return;
    }

    try {
      await _enqueueAppDataLoad(bypassDebounce: true);
    } catch (e) {
      print(e);
    }
  }

  /// Runs AppData loads one at a time.
  ///
  /// A load is a fetch *and* a `storeAppData`, and the two paths into it are
  /// independent: the scheduled one via [refreshTokens] -> [_tryLoadAppData],
  /// and the forced one via [reloadAppData]. Unserialized they can overlap, and
  /// then whichever response lands last wins -- so an older scheduled payload
  /// could overwrite the groups and permissions a Quick Scan just fetched, which
  /// is exactly the staleness this whole path exists to prevent. `storeAppData`
  /// also writes the token pair when the payload carries one, and interleaving
  /// that is what `_refreshCompleter` was introduced to avoid.
  ///
  /// Deliberately its own queue rather than `_refreshCompleter`: that completer
  /// makes concurrent callers *share one result*, which would defeat a forced
  /// reload by handing it an older fetch's outcome. Queuing serializes without
  /// collapsing two distinct requests into one.
  Future<void> _enqueueAppDataLoad({required bool bypassDebounce}) {
    final next =
        _appDataQueue.then((_) => _loadAppData(bypassDebounce: bypassDebounce));
    // Keep the chain alive: one failed load must not poison every later one. The
    // error still reaches `next`, which both callers already catch.
    _appDataQueue = next.catchError((_) {});
    return next;
  }

  /// Collapses the bursts, not the refreshes: app resume and the 15-minute timer
  /// both land here and can fire together, and re-requesting AppData twice in a
  /// row buys nothing. Deliberately short -- a longer window would turn back into
  /// a staleness budget, which is the thing being fixed.
  static const _appDataRefreshDebounce = Duration(seconds: 30);

  /// Re-fetches AppData and republishes it into the models.
  ///
  /// [storeAppData] is the only writer of `GroupModel`, `PermissionsModel`,
  /// `UserPreferencesModel` and the category/tag catalogs, and outside login this
  /// is the only path that reaches it. It used to run only when
  /// `_groupModel.groups.isEmpty` -- which is false for the whole of a logged-in
  /// session, so the 15-minute refresh timer and the on-resume refresh were both
  /// no-ops. A group setting changed elsewhere (a quick-scan field switched on,
  /// say) could then never reach a running app: you had to log out and back in.
  /// Desktop re-fetches AppData on every bootstrap, so the two clients disagreed
  /// about how the group was configured.
  ///
  /// The empty-groups case still forces a load regardless of the debounce: that
  /// is the first load, and nothing can render without it. [bypassDebounce]
  /// skips the window outright -- see [reloadAppData].
  Future<void> _loadAppData({required bool bypassDebounce}) async {
    final loadedAt = _appDataLoadedAt;
    final isFirstLoad = _groupModel.groups.isEmpty;

    if (!bypassDebounce &&
        !isFirstLoad &&
        loadedAt != null &&
        DateTime.now().difference(loadedAt) < _appDataRefreshDebounce) {
      return;
    }

    var appDataResponse = await OpenApiClient.client.getUserApi().getAppData();
    await storeAppData(
      _authModel,
      _groupModel,
      _userModel,
      _userPreferencesModel,
      _categoryModel,
      _tagModel,
      _systemSettingsModel,
      _permissionsModel,
      appDataResponse.data as AppData,
    );
    _appDataLoadedAt = DateTime.now();
  }
}
