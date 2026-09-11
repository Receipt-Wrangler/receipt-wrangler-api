import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:image_picker_android/image_picker_android.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_native_splash/flutter_native_splash.dart';
import 'package:go_router/go_router.dart';
import 'package:sentry_flutter/sentry_flutter.dart';
import 'package:provider/provider.dart';
import 'package:receipt_wrangler_mobile/auth/login/screens/auth_screen.dart';
import 'package:receipt_wrangler_mobile/groups/nav/group/group_app_bar.dart';
import 'package:receipt_wrangler_mobile/groups/nav/group/group_bottom_nav.dart';
import 'package:receipt_wrangler_mobile/groups/nav/group_select/group_select_app_bar.dart';
import 'package:receipt_wrangler_mobile/groups/nav/group_select/group_select_bottom_nav.dart';
import 'package:receipt_wrangler_mobile/groups/screens/group_dashboards.dart';
import 'package:receipt_wrangler_mobile/groups/screens/group_receipts_screen.dart';
import 'package:receipt_wrangler_mobile/groups/screens/group_select.dart';
import 'package:receipt_wrangler_mobile/guards/auth-guard.dart';
import 'package:receipt_wrangler_mobile/guards/permission-guard.dart';
import 'package:receipt_wrangler_mobile/home/screens/home.dart';
import 'package:receipt_wrangler_mobile/models/auth_model.dart';
import 'package:receipt_wrangler_mobile/models/category_model.dart';
import 'package:receipt_wrangler_mobile/models/group_model.dart';
import 'package:receipt_wrangler_mobile/models/loading_model.dart';
import 'package:receipt_wrangler_mobile/models/permissions_model.dart';
import 'package:receipt_wrangler_mobile/models/receipt-list-model.dart';
import 'package:receipt_wrangler_mobile/models/receipt_model.dart';
import 'package:receipt_wrangler_mobile/models/search_model.dart';
import 'package:receipt_wrangler_mobile/models/tag_model.dart';
import 'package:receipt_wrangler_mobile/models/user_model.dart';
import 'package:receipt_wrangler_mobile/models/user_preferences_model.dart';
import 'package:receipt_wrangler_mobile/persistence/global_shared_preferences.dart';
import 'package:receipt_wrangler_mobile/receipts/screens/receipt_form_screen.dart';
import 'package:receipt_wrangler_mobile/reports/screens/report_list_screen.dart';
import 'package:receipt_wrangler_mobile/search/nav/search_app_bar.dart';
import 'package:receipt_wrangler_mobile/search/screens/search_screen.dart';
import 'package:receipt_wrangler_mobile/search/widgets/searchbar.dart';
import 'package:receipt_wrangler_mobile/services/token_refresh_service.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/circular_loading_progress.dart';
import 'package:receipt_wrangler_mobile/service/crash_reporting.dart';
import 'package:receipt_wrangler_mobile/shared/widgets/screen_wrapper.dart';
import 'package:receipt_wrangler_mobile/utils/url.dart';

import 'package:receipt_wrangler_mobile/profile/screens/user_profile_screen.dart';

import 'constants/search.dart';
import 'models/context_model.dart';
import 'models/custom_field_model.dart';
import 'models/system_settings_model.dart';

void main() async {
  final widgetsBinding = WidgetsFlutterBinding.ensureInitialized();
  FlutterNativeSplash.preserve(widgetsBinding: widgetsBinding);

  // Opt into the Android Photo Picker on API 33-35. It is already the default
  // on 36+, where this is a no-op, and below 33 the Play Services backport
  // covers it (see the ModuleDependencies service in AndroidManifest.xml).
  //
  // This is plugin *configuration* -- no permission request, no channel round
  // trip, no system dialog -- so it does not fall under the launch-time-work
  // ban documented in `_ReceiptWrangler.initState` for the iOS render-pause
  // freeze (GitHub #617). Do not move it there. It has to run before the first
  // pick, and `ensureInitialized()` above has just registered the plugin whose
  // instance it reads.
  final imagePicker = ImagePickerPlatform.instance;
  if (imagePicker is ImagePickerAndroid) {
    imagePicker.useAndroidPhotoPicker = true;
  }

  await GlobalSharedPreferences.initialize();

  // Crash/error reporting is opt-out (on by default). When disabled we don't
  // initialize Sentry at all, so nothing runs. SentryFlutter.init also installs
  // the FlutterError/PlatformDispatcher error handlers that report uncaught
  // exceptions.
  if (isCrashReportingEnabled()) {
    await SentryFlutter.init(configureSentry, appRunner: () => runApp(buildApp()));
  } else {
    runApp(buildApp());
  }
}

/// Builds the production widget tree (providers + root app widget).
/// Extracted from `main()` so integration tests can pump it directly via
/// `tester.pumpWidget(buildApp())` instead of calling `main()`. Each call
/// returns a fresh tree — the `GoRouter` lives inside `_ReceiptWrangler`
/// as a per-`State` `late final` field, so test #N never inherits test
/// #N-1's router location.
///
/// [initialDeepLink] / [deepLinkStream] are test seams standing in for the two
/// `app_links` sources (cold start / warm delivery). `main()` passes neither, so
/// production always reads the real plugin. See [ReceiptWrangler].
Widget buildApp({Uri? initialDeepLink, Stream<Uri>? deepLinkStream}) {
  return MultiProvider(
    providers: [
      ChangeNotifierProvider(create: (_) => AuthModel()),
      ChangeNotifierProvider(create: (_) => CategoryModel()),
      ChangeNotifierProvider(create: (_) => ContextModel()),
      ChangeNotifierProvider(create: (_) => CustomFieldModel()),
      ChangeNotifierProvider(create: (_) => GroupModel()),
      ChangeNotifierProvider(create: (_) => LoadingModel()),
      ChangeNotifierProvider(create: (_) => PermissionsModel()),
      ChangeNotifierProvider(create: (_) => ReceiptListModel()),
      ChangeNotifierProvider(create: (_) => ReceiptModel()),
      ChangeNotifierProvider(create: (_) => SearchModel()),
      ChangeNotifierProvider(create: (_) => SystemSettingsModel()),
      ChangeNotifierProvider(create: (_) => TagModel()),
      ChangeNotifierProvider(create: (_) => UserModel()),
      ChangeNotifierProvider(create: (_) => UserPreferencesModel()),
    ],
    child: ReceiptWrangler(
      initialDeepLink: initialDeepLink,
      deepLinkStream: deepLinkStream,
    ),
  );
}

GoRouter _buildAppRouter() {
  return GoRouter(
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const ScreenWrapper(child: Home()),
        redirect: (context, state) {
          return unprotectedRouteRedirect(context, "/groups");
        },
      ),
      GoRoute(
        path: '/login',
        builder: (context, state) => const ScreenWrapper(child: AuthScreen()),
        redirect: (context, state) {
          return unprotectedRouteRedirect(context, "/groups");
        },
      ),
      ShellRoute(
          builder: (context, state, child) {
            return ScreenWrapper(
              appBarWidget: const GroupSelectAppBar(),
              bottomNavigationBarWidget: const GroupSelectBottomNav(),
              child: child,
            );
          },
          routes: [
            GoRoute(
                path: "/groups",
                builder: (context, state) => const GroupSelect()),
          ]),
      ShellRoute(
          builder: (context, state, child) {
            EdgeInsets? padding;
            if (state.fullPath == '/groups/:groupId/receipts') {
              padding = const EdgeInsets.all(0);
            }
            return ScreenWrapper(
              appBarWidget: const GroupAppBar(),
              bottomNavigationBarWidget: const GroupBottomNav(),
              bodyPadding: padding,
              child: child,
            );
          },
          routes: [
            GoRoute(
              path: '/groups/:groupId/dashboards',
              redirect: groupDashboardReadRedirect,
              builder: (context, state) => const GroupDashboards(),
            ),
            GoRoute(
              path: '/groups/:groupId/receipts',
              builder: (context, state) => const GroupReceiptsScreen(),
            ),
          ]),
      // The three receipt routes all render ReceiptFormScreen, whose form uses
      // a single shared GlobalKey (ReceiptModel.receiptFormKey). A default
      // animated page transition keeps the outgoing and incoming screens
      // mounted simultaneously for the duration of the slide, so two
      // ReceiptForms briefly share that GlobalKey ("a GlobalKey was specified
      // multiple times"). NoTransitionPage swaps atomically (no overlap), which
      // also stops a just-departed /receipts/add screen from lingering and
      // re-rendering itself as a stale "view" (its form state is derived from
      // the live URL).
      GoRoute(
        path: '/receipts/add',
        redirect: (context, state) {
          Provider.of<ReceiptModel>(context, listen: false).resetModel();
          return null;
        },
        pageBuilder: (context, state) => NoTransitionPage(
            key: state.pageKey, child: const ReceiptFormScreen()),
      ),
      GoRoute(
        path: '/receipts/:receiptId/view',
        pageBuilder: (context, state) => NoTransitionPage(
            key: state.pageKey, child: const ReceiptFormScreen()),
      ),
      GoRoute(
        path: '/receipts/:receiptId/edit',
        pageBuilder: (context, state) => NoTransitionPage(
            key: state.pageKey, child: const ReceiptFormScreen()),
      ),
      GoRoute(
        path: '/profile',
        builder: (context, state) => const UserProfileScreen(),
      ),
      GoRoute(
        path: '/reports',
        redirect: reportsReadRedirect,
        builder: (context, state) => const ReportListScreen(),
      ),
      ShellRoute(
        builder: (context, state, child) {
          var searchModel = Provider.of<SearchModel>(context, listen: false);
          searchModel.searchTermBehaviorSubject.add("");
          searchModel.setSearchResults([], notify: false);

          var extra = state.extra as Map<String, dynamic>;
          var from = extra["from"];

          return ScreenWrapper(
            appBarWidget: SearchAppBar(),
            bodyPadding: const EdgeInsets.all(0),
            bottomNavigationBarWidget: from == fromGroupBottomNav
                ? const GroupBottomNav()
                : const GroupSelectBottomNav(),
            child: child,
            bottomSheetWidget: const WranglerSearchBar(),
          );
        },
        routes: [
          GoRoute(
            path: '/search',
            redirect: receiptsSearchRedirect,
            builder: (context, state) => const SearchScreen(),
          ),
        ],
      )
    ],
  );
}

class ReceiptWrangler extends StatefulWidget {
  const ReceiptWrangler({
    super.key,
    this.initialDeepLink,
    this.deepLinkStream,
  });

  /// Injectable for tests: stands in for [AppLinks.getInitialLink] (the link
  /// that cold-launched the app). Null in production.
  final Uri? initialDeepLink;

  /// Injectable for tests: stands in for [AppLinks.uriLinkStream] (links
  /// delivered while the app is already running). Null in production.
  final Stream<Uri>? deepLinkStream;

  @override
  State<ReceiptWrangler> createState() => _ReceiptWrangler();
}

class _ReceiptWrangler extends State<ReceiptWrangler>
    with WidgetsBindingObserver {
  late final AppLifecycleListener _lifecycleListener;
  Timer? _refreshTimer;
  Timer? _launchWindowTimer;
  bool _inLaunchWindow = true;
  late Future<bool> _initFuture;
  bool _initialized = false;

  // Deep-link (App Links / Universal Links) plumbing for
  // receiptwrangler.io/app/setup. We handle links ourselves via app_links
  // rather than letting go_router try (and fail) to route /app/setup.
  final AppLinks _appLinks = AppLinks();
  StreamSubscription<Uri>? _linkSubscription;

  // GoRouter held per-State instance so each `pumpWidget(buildApp())` in
  // tests gets a fresh router starting at '/'. As a top-level `final` it
  // would be initialized once per isolate and leak location across tests.
  late final GoRouter _router = _buildAppRouter();

  late final authModel = Provider.of<AuthModel>(context, listen: false);
  late final groupModel = Provider.of<GroupModel>(context, listen: false);
  late final userModel = Provider.of<UserModel>(context, listen: false);
  late final categoryModel =
      Provider.of<CategoryModel>(context, listen: false);
  late final tagModel = Provider.of<TagModel>(context, listen: false);
  late final systemSettingsModel =
      Provider.of<SystemSettingsModel>(context, listen: false);
  late final userPreferencesModel =
      Provider.of<UserPreferencesModel>(context, listen: false);
  late final permissionsModel =
      Provider.of<PermissionsModel>(context, listen: false);

  @override
  void initState() {
    super.initState();

    WidgetsBinding.instance.addObserver(this);
    _lifecycleListener = AppLifecycleListener(onStateChange: _onStateChanged);

    // NOTE: camera/photo permissions are intentionally NOT requested here.
    // Requesting at launch makes the app go `inactive` (system dialog), and on
    // iOS the Flutter engine pauses rendering while inactive — on iOS 26.x it
    // doesn't reliably resume, leaving the UI frozen (GitHub #617). Permissions
    // are now requested in-context: the scanner requests camera itself, and the
    // save-to-gallery flow requests photo access at the point of use.
    FlutterNativeSplash.remove();

    // Cold launch itself passes through `inactive`, and iOS can paint the first
    // frame at transient launch metrics (wrong size/orientation) and then stop
    // producing frames — the UI is stuck on that stale, "half-rotated/unpainted"
    // frame (GitHub #617, seen on 120Hz iPhone 17 / Air devices even with no
    // permission dialog). Force repaints from the very first frame so the
    // pipeline re-latches and the layout re-flows to the real geometry.
    // `didChangeMetrics` re-arms this while the window is still settling.
    WidgetsBinding.instance.addPostFrameCallback((_) => nudgeFrames());
    _launchWindowTimer =
        Timer(const Duration(seconds: 6), () => _inLaunchWindow = false);

    _initDeepLinks();
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    _launchWindowTimer?.cancel();
    _frameNudgeTimer?.cancel();
    _linkSubscription?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    _lifecycleListener.dispose();

    super.dispose();
  }

  /// Subscribes to receiptwrangler.io/app/setup deep links. Handles the cold
  /// start ([AppLinks.getInitialLink]) and warm/resumed ([AppLinks.uriLinkStream])
  /// cases. A matching link pre-fills the Connect screen's server URL via
  /// [AuthModel.pendingServerUrl]; it is never auto-connected.
  ///
  /// Both sources fall back to the real plugin unless a test supplied
  /// [ReceiptWrangler.initialDeepLink] / [ReceiptWrangler.deepLinkStream].
  Future<void> _initDeepLinks() async {
    // Cold start: the app-link that launched the app. Stash it on AuthModel
    // immediately so it survives the FutureBuilder first-paint gate and the
    // Connect screen reads it the moment it mounts.
    try {
      final initial = await _resolveInitialDeepLink();
      if (initial != null) {
        _handleDeepLink(initial);
      }
    } catch (_) {
      // Ignore an unavailable / malformed initial link.
    }

    // Warm / resumed: further links delivered while the app is running.
    _linkSubscription = (widget.deepLinkStream ?? _appLinks.uriLinkStream).listen(
      _handleDeepLink,
      onError: (_) {},
    );
  }

  /// The cold-start link, from the test seam if one was injected.
  ///
  /// Deliberately a separate awaited call rather than
  /// `widget.initialDeepLink ?? await _appLinks.getInitialLink()`: `??` would
  /// short-circuit the `await`, so an injected link would be handled
  /// SYNCHRONOUSLY inside `initState` — routing before the tree is attached, on
  /// a timing production never sees. Awaiting always yields a microtask, so
  /// injected and real links arrive at the same point in the lifecycle.
  Future<Uri?> _resolveInitialDeepLink() async {
    return widget.initialDeepLink ?? await _appLinks.getInitialLink();
  }

  void _handleDeepLink(Uri uri) {
    final serverUrl = extractDeepLinkServerUrl(uri.toString());
    if (serverUrl == null) {
      return;
    }

    // Stash the URL for the Connect screen to pre-fill, then route to it. A
    // logged-in user hitting '/' is bounced to '/groups' by the auth redirect,
    // so the pre-fill only surfaces for unauthenticated sessions (intended — a
    // logged-in user is already set up).
    authModel.setPendingServerUrl(serverUrl);
    _router.go('/');
  }

  @override
  void didChangeMetrics() {
    super.didChangeMetrics();
    // A metrics change during the launch window means the window geometry just
    // settled; force a repaint so a stale first frame re-flows to the real
    // size/orientation. Gated to the launch window so ordinary keyboard
    // show/hide later doesn't trigger a burst. (GitHub #617.)
    if (_inLaunchWindow) nudgeFrames();
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();

    if (!_initialized) {
      _initialized = true;

      authModel.initializeAuth();

      TokenRefreshService().initialize(
        authModel: authModel,
        groupModel: groupModel,
        userModel: userModel,
        userPreferencesModel: userPreferencesModel,
        categoryModel: categoryModel,
        tagModel: tagModel,
        systemSettingsModel: systemSettingsModel,
        permissionsModel: permissionsModel,
      );

      _initFuture = TokenRefreshService().refreshTokens();

      _refreshTimer =
          Timer.periodic(const Duration(minutes: 15), (timer) async {
        await TokenRefreshService().refreshTokens();
      });
    }
  }

  Widget _buildMaterialApp() {
    return MaterialApp.router(
      color: Colors.white,
      debugShowCheckedModeBanner: false,
      title: 'Receipt Wrangler',
      theme: ThemeData(
        fontFamily: "Raleway",
        inputDecorationTheme: const InputDecorationTheme(
          border: OutlineInputBorder(),
        ),
        chipTheme: ChipThemeData(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(50),
          ),
        ),
        bottomSheetTheme: const BottomSheetThemeData(
          backgroundColor: Colors.white,
          modalBackgroundColor: Colors.white,
          surfaceTintColor: Colors.white,
        ),
        colorScheme: const ColorScheme(
          primary: Color(0xFF27B1FF),
          secondary: Color(0xFF8EA1AC),
          surface: Color(0xFFFFFFFF),
          background: Color(0xFFFFFFFF),
          error: Color(0xFFd63333),
          onPrimary: Color(0xFFFFFFFF),
          onSecondary: Color(0xFF000000),
          onSurface: Color(0xFF000000),
          onBackground: Color(0xFF000000),
          onError: Color(0xFFFFFFFF),
          brightness: Brightness.light,
        ),
        useMaterial3: true,
      ),
      routerConfig: _router,
      // Hosts an invisible repaint pump (see [nudgeFrames]) so the app can
      // force frames after returning from an inactive state and recover from
      // the iOS render-pause freeze (GitHub #617).
      builder: (context, child) => Stack(
        textDirection: TextDirection.ltr,
        children: [
          child ?? const SizedBox.shrink(),
          ValueListenableBuilder<int>(
            valueListenable: _frameNudge,
            builder: (_, __, ___) => const SizedBox.shrink(),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder(
      future: _initFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.done) {
          return _buildMaterialApp();
        }

        return const CircularLoadingProgress();
      },
    );
  }

  // Listen to the app lifecycle state changes
  void _onStateChanged(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.detached:
        _onDetached();
      case AppLifecycleState.resumed:
        _onResumed();
      case AppLifecycleState.inactive:
        _onInactive();
      case AppLifecycleState.hidden:
        _onHidden();
      case AppLifecycleState.paused:
        _onPaused();
    }
  }

  void _onDetached() {}

  void _onResumed() async {
    // Recover from the iOS render-pause freeze after an inactive transition.
    nudgeFrames();
    await TokenRefreshService().refreshTokens(force: true);
  }

  void _onInactive() {}

  void _onHidden() {}

  void _onPaused() {}
}

/// Ticked by [nudgeFrames] to force the render pipeline to produce frames.
final ValueNotifier<int> _frameNudge = ValueNotifier<int>(0);
Timer? _frameNudgeTimer;

/// The iOS Flutter engine pauses rendering while the app is `inactive` — a cold
/// launch, a system permission dialog, the app switcher, an incoming call — and
/// on iOS 26.x it doesn't reliably resume, leaving the UI frozen on its last
/// (sometimes wrong-geometry) frame. For a short window this both bumps
/// [_frameNudge] (marking the tree dirty) and calls
/// [SchedulerBinding.scheduleForcedFrame] each tick: the forced call bypasses
/// the engine's "frames disabled while inactive" gate, which a plain notifier
/// bump does not — that re-latches the display link and unfreezes the UI.
/// (GitHub #617 / flutter/engine#17396.)
///
/// iOS-only: a no-op on Android/desktop (and under the test suites, where
/// `defaultTargetPlatform` is not iOS) so nothing else is affected.
void nudgeFrames() {
  if (defaultTargetPlatform != TargetPlatform.iOS) return;
  _frameNudgeTimer?.cancel();
  var ticks = 0;
  WidgetsBinding.instance.scheduleForcedFrame();
  _frameNudgeTimer = Timer.periodic(const Duration(milliseconds: 100), (t) {
    _frameNudge.value++;
    WidgetsBinding.instance.scheduleForcedFrame();
    if (++ticks >= 30) t.cancel(); // ~3s
  });
}
