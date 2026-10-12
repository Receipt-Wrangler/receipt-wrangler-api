# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Receipt Wrangler Mobile is a Flutter mobile application that provides a native interface for Receipt Wrangler, a receipt management and splitting system. The app enables users to manage receipts on the go with camera/gallery uploads, receipt scanning, group management, and receipt splitting capabilities.

## Development Commands

### Core Flutter Commands
- `flutter run` - Run the app on connected device/emulator
- `flutter build apk` - Build Android APK
- `flutter build ios` - Build iOS app
- `flutter test` - Run unit tests
- `flutter analyze` - Analyze Dart code for issues
- `dart format .` - Format Dart code
- `flutter clean` - Clean build artifacts
- `flutter pub get` - Install dependencies
- `flutter pub upgrade` - Upgrade dependencies

### API Client
The project uses a generated OpenAPI client located in the `api/` directory. The client is imported as a local package dependency in pubspec.yaml.

## Architecture Overview

### State Management
The app uses Provider pattern with ChangeNotifier models:
- **AuthModel**: Authentication state, JWT tokens, API client configuration
- **GroupModel**: Group management and selection
- **ReceiptModel**: Receipt data, form state, and image handling
- **UserModel**: User profile and preferences
- **CategoryModel**, **TagModel**: Metadata management
- **SearchModel**: Search functionality with RxDart streams
- **PermissionsModel**: The caller's effective permissions, used for UI gating (see "Permission-based UI gating")

### Navigation
Uses `go_router` with nested shell routes:
- **Group Selection Shell**: `/groups` with group selection UI
- **Group Context Shell**: `/groups/:groupId/*` with group-specific navigation
- **Search Shell**: `/search` with search interface
- Individual routes for receipt forms, viewing, and editing

The receipt form routes (`/receipts/:receiptId/{view,edit}` and the comments variants) use
`pageBuilder` with `NoTransitionPage(key: state.pageKey, ...)` instead of `builder`. During a
default transition the outgoing and incoming `ReceiptFormScreen`s coexist for the animation, and
both attach the same global `receiptFormKey` — a GlobalKey collision that corrupts the form. The
`NoTransitionPage` makes the swap atomic; `key: state.pageKey` forces a fresh element per
navigation so the new screen re-fetches rather than reusing the old element. Because the swap
removes the old page in the same frame, popup-menu items that navigate must capture the router
and defer the `go` until the menu finishes dismissing (see
`receipt_app_bar_action_builder.dart`'s `_goAfterMenu`).

### Core Directory Structure
- `lib/models/` - Provider-based state management models
- `lib/auth/` - Authentication screens and logic  
- `lib/groups/` - Group management, dashboards, receipts
- `lib/receipts/` - Receipt forms, viewing, image handling
- `lib/search/` - Search functionality
- `lib/reports/` - Reports slice: list / preview / generate / delete saved report templates
- `lib/shared/` - Reusable widgets and utilities
- `lib/client/` - OpenAPI client wrapper
- `lib/utils/` - Utility functions for auth, currency, dates, etc.

### Key Features
- **Receipt Management**: Create, edit, view receipts with items and images
- **Image Handling**: three sources — the document scanner (camera), the OS photo library and the OS file browser — plus scanning capabilities
- **Group Management**: Multi-user groups with role-based access
- **Search**: Full-text search across receipts
- **Offline Support**: Secure token storage with refresh token flow

### Form Handling
Uses `flutter_form_builder` for complex forms with validation. Receipt forms support:
- Dynamic item lists with custom fields
- Image carousel with infinite scroll
- Category and tag selection
- Currency formatting and validation

### API Integration
- Generated OpenAPI client from backend specification
- JWT-based authentication with automatic token refresh
- Centralized client configuration in `OpenApiClient` singleton
- Secure token storage using `flutter_secure_storage`

### AppData is re-fetched during a session, not just at login

`storeAppData` (`lib/utils/auth.dart`) is the **only** writer of `GroupModel`, `PermissionsModel`,
`UserPreferencesModel` and the category/tag catalogs, and outside the login response the only path
that reaches it is `TokenRefreshService._loadAppData`. That used to be guarded by
`if (_groupModel.groups.isEmpty)` — false for the whole of a logged-in session — so the 15-minute
refresh timer (`main.dart`) and the on-resume refresh were both **no-ops**. A group or permission
change made elsewhere could not reach a running app at all; you had to log out and back in. (On
Android, backgrounding does not re-run `main()`, so even "close and reopen" did not help.) Desktop
re-fetches AppData unconditionally on every bootstrap (`desktop/src/services/app-init.service.ts`),
so the two clients disagreed about how a group was configured.

It now loads every time, with a short **30-second debounce** (`_appDataRefreshDebounce`) that exists
only to collapse a burst — app resume and the periodic timer can fire together and both call through
here. It is deliberately not a staleness budget: a longer window would reinstate the bug. The
empty-groups case bypasses the debounce, since that is the first load and nothing renders without it.
Pinned by `test/services/token_refresh_service_test.dart` ("reloads app data even when groups already
exist" / "debounces a second refresh that follows immediately"), whose `buildMockAppData()` helper
stubs every field `storeAppData` touches — a mock missing one throws mid-store instead of failing the
assertion under test.

**Starting a Quick Scan forces a reload, ahead of the scanner.** `startScanEntry` and
`startPickerEntry` (`lib/shared/functions/receipt_entry.dart`) both `await
TokenRefreshService().reloadAppData()` before anything else, so the group's quick-scan field config,
the caller's permissions and the AI feature flag are all current for that tap.

- **It must bypass the debounce**, hence a dedicated public `reloadAppData()` rather than a flag on
  `refreshTokens`: concurrent callers of that piggyback on `_refreshCompleter` and return *before*
  `_doRefresh` is entered, so a flag would be silently dropped for the piggybacking caller — the one
  that most needs the reload. `_loadAppData({required bool bypassDebounce})` carries the split.
- **It runs before the scanner, not before the sheet.** By the time `showQuickScanBottomSheet` runs
  the images have already been captured *and* seeded from `userPreferences.quickScanDefaultGroupId` /
  `GroupModel.soleGroupId` (`buildQuickScanImages`), and `QuickScanForm` reads both providers with
  `listen: false`, so data landing after the sheet is open never reaches it. Refreshing first also
  makes the `resolveReceiptEntryAvailability` gate itself fresh. As a bonus,
  `showQuickScanBottomSheet` stays synchronous.
- **`openQuickScanFromGallery` deliberately does not refresh on its own behalf** — `fallBackToGallery`
  reaches it from inside `startScanEntry`, which already has, so the camera-denied path would
  double-fetch. The two picker *menu items* go through `startPickerEntry(context, method)` instead,
  because they are the initiations that bypass `startScanEntry`.
- **`reloadAppData()` never throws.** The scanner is about to open, so a transient failure falls
  through to the data already loaded rather than blocking the scan — Quick Scan stays usable offline.
  The nav tap is fire-and-forget (`BottomNav.onDestinationSelected` is `void Function(int)`), so a
  thrown error would be unobserved anyway.
- **Both load paths are serialized** through `_enqueueAppDataLoad`. A load is a fetch *and* a
  `storeAppData`, and the scheduled path (`refreshTokens` → `_tryLoadAppData`) and the forced one
  (`reloadAppData`) are independent — unserialized they overlap and the last response wins, so an
  older scheduled payload could overwrite the groups and permissions a Quick Scan just fetched. That
  is the very staleness this path exists to prevent, and `storeAppData` also writes the token pair
  when the payload carries one. It is deliberately **its own queue, not `_refreshCompleter`**: that
  completer makes concurrent callers share one result, which would defeat a forced reload by handing
  it an older fetch's outcome. One consequence: a second *queued* normal load is debounced away,
  where before both would have fetched (neither had stamped `_appDataLoadedAt` yet) — which is what
  the debounce is for. Another: the scan path may wait out an in-flight timer load before starting
  its own, bounded by one request.

**This changed what a quick-scan e2e has to do.** A refresh at scan time overwrites any client-side
`GroupModel` mutation with the server's copy, so `configureFirstGroup`
(`integration_test/helpers/quick_scan_actions.dart`) now **persists through the API** and then mirrors
the result locally. It sends `quickScanDefaultPaidByType: 'UPLOADER'` and `quickScanDefaultStatus:
'OPEN'` unconditionally, because the backend rejects a config where paid-by or status can be skipped
without a default to backfill it (`UpdateGroupReceiptSettingsCommand.Validate`) and most of those
specs relax one of the two. The one exception is the group-switch case, which fabricates a second
group that does not exist on the server and so sets `debugSkipQuickScanAppDataRefresh`
(`lib/shared/functions/receipt_entry.dart`, the `debugCameraAccessOverride` pattern) — use that seam
only for a spec driving an injected `GroupModel`, never to dodge persisting real config.

**`integration_test/quick_scan_app_data_refresh_test.dart` is the spec that proves the round trip**:
it persists the comment field OFF, logs in, flips it ON server-side, and opens Quick Scan without
restarting the app. Every other quick-scan spec arranges its config *before* login, so none of them
can tell a client that re-fetches from one that read the config once at login.

### `TopAppBar`'s loading bar is always mounted — never null

`TopAppBar` passes its `LinearProgressIndicator` as `AppBar.bottom`, and it must stay **non-null even
when idle** (zero-height `PreferredSize`), however tempting the conditional looks. `AppBar` re-parents
its whole toolbar into a `Column` *only* when `bottom != null`
(`flutter/lib/src/material/app_bar.dart`), so a null ↔ non-null swap changes that slot's widget from a
`ClipRect` to a `Column`, fails `Widget.canUpdate`, and **unmounts `leading`, `title` and every
`actions` child**. Anything in `actions` that captured its own `BuildContext` and uses it after an
`await` then fails its `mounted` guard and silently gives up.

That shipped: the receipts-screen overflow menu's **Quick Scan** and picker items did
nothing, because `_refreshBeforeQuickScan` — the very thing they await — is what raises this bar. Only
"Add Manual Receipt" worked, being synchronous, and the bottom-nav long-press was immune because its
context comes from the nav rather than the app bar. The avatar menu in the same bar survives because it
navigates from `this.context` (the `_TopAppBar` **State** context, *above* the `AppBar`) — keep it that
way.

The idle **child** is still swapped to `SizedBox.shrink()`: a permanently mounted
`LinearProgressIndicator` animates forever and `pumpAndSettle` would never return (its default timeout
is **ten minutes**, so that regression hangs CI rather than failing it). Height goes to zero instead of
the widget going away. Don't "fix" the 4px toolbar squeeze while the bar shows by making
`TopAppBar.preferredSize` dynamic — `Scaffold` pins 56 either way, and a dynamic size would relayout
every screen's body on every network call. Pinned by
`test/widgets/top_app_bar_loading_indicator_test.dart`.

### Theme & color roles — every neutral role is spelled out on purpose

The app's `ThemeData` lives in **`lib/theme/app_theme.dart`** (`buildAppTheme()`), extracted from
`main.dart` so it can be asserted on and so a harness can render a screen exactly as the app does.

**`ColorScheme` silently falls back for any role it is not given, and in this scheme every fallback
lands on pure black or pure white.** The getters are, verbatim from
`flutter/lib/src/material/color_scheme.dart`:

| role | falls back to | what that gave us |
|---|---|---|
| `outline`, `outlineVariant` | `onBackground` | **pure black** borders on every input, chip, divider and outlined button |
| `onSurfaceVariant` | `onSurface` | **black** hints, field labels, list subtitles and nav icons — no text hierarchy at all |
| every `surfaceContainer*` | `surface` | **white** — no tone to layer a raised card against |
| `secondaryContainer` | `secondary` | `#8EA1AC`, so an unstyled selected chip read as *disabled* |

So the scheme now names all of them, from the slate scale in `lib/constants/colors.dart`. **Do not
drop them** — nothing fails to compile and no widget test fails; the app just goes back to looking
like a wireframe. `test/theme/app_theme_test.dart` is the only guard, and it asserts each role is
both the expected value *and* not equal to its fallback.

Three consequences worth knowing:

- **`accentBlueDark` (`#0086D4`) is the accent for text on the accent *tints*, not a general
  "accent on white".** Accent icons and button labels on white are the theme's `primary` -- that is
  what `receipt_form.dart:374`/`:676` already do and what M3 gives every text and outlined button,
  so it is the convention already in place. `accentBlueDark` exists because `primary` over
  `primaryContainer` / `accentContainer` is barely 2:1.
- **A chip's selected label colour has to ride on `chipTheme.labelStyle` as a `WidgetStateColor`.**
  `RawChip` resolves `labelStyle.color` through `WidgetStateProperty.resolveAs` and, under Material
  3, **never consults `secondaryLabelStyle`** — that is a Material 2 field. Left to the M3 defaults
  the selected label is `onSecondaryContainer`, i.e. dark slate on the `#27B1FF` fill, which is the
  contrast bug the category / tag / status pickers shipped with. Note `labelStyle` **replaces** the
  defaults rather than merging, so the font family and size are restated there.
  Because the theme now owns this, `multi-select-field.dart` and `group_dashboard.dart` no longer
  set `selectedColor` themselves.
- **`appBarTheme` pins `iconTheme` and `actionsIconTheme` together.** M3 takes `leading` from
  `onSurface` and `actions` from `onSurfaceVariant`, so once `onSurfaceVariant` stopped being black
  a bar carrying both rendered two different greys.
- **`navigationBarTheme` states the bottom nav's selected treatment rather than riding a role.**
  `NavigationBar.indicatorColor` defaults to `secondaryContainer`, which *used* to fall through to
  the `#8EA1AC` secondary -- an accidentally serviceable solid pill. Naming that role `slate100`
  (correct for its other consumers) turned the selected destination into a near-white pill on a
  near-white bar, i.e. invisible, on every main screen. The bar now takes `accentContainer`
  (`#CCECFF`) with `accentBlueDark` icon and label, which is the design's own nav. **This is the
  shape of mistake to expect when filling in a role**: the fix is right for the role and wrong for
  one component that was quietly depending on the old fallback. `app_theme_test.dart` guards it.

**Contrast floors, and the one gap left open.** WCAG 2.1 SC 1.4.11 wants **3:1** for the visual
boundary that identifies a component -- which an outlined field's border *is* -- and SC 1.4.3 wants
**4.5:1** for normal text. `test/theme/app_theme_test.dart` measures these rather than pinning hex
values, so the assertions survive a palette tweak and fail only when one actually breaks access.
The first pass at this palette took `#CBD5E1` straight from the design, which draws its borders at
1.5-2px; Flutter renders 1px, so the app got the lightness without the weight and every field
boundary sat at **1.48:1**. Hence `borderSlate` (3.38:1) rather than `slate300`, and `slate400` is
now decorative-only -- a chevron or a dismiss X carries meaning and belongs at `onSurfaceVariant`.

The **known gap**: `accentBlueDark` is also small *text* in two places -- the nav's selected label
(3.58:1 on the bar) and the editor's selected operation chip (3.59:1 on its tint). Both clear the
3:1 non-text floor and fall short of 4.5:1. **It is closeable on-palette**: moving the accent from
`$primary-palette` 700 to **800 `#0072b4`** clears every floor (5.16 / 4.18 / 4.71 / 4.73). Staying
at 700 is a deliberate appearance call -- 800 reads noticeably navy -- not a limitation, and it is
the call to revisit if the gap ever matters. The focused floating label was a third such case and is
**not** accent-coloured for this reason: it keeps `onSurfaceVariant` (4.76:1) and focus rides on the
border instead.

**The palette is shared with desktop, and almost all of it is already canonical.** The source of
truth is the Angular Material M2 maps in `desktop/src/variables.scss`, which the web theme is
actually built from (`mat.m2-define-palette`, `desktop/src/styles.scss`). `lib/constants/colors.dart`
names the stop for each constant; the mapping is:

| mobile | brand palette |
|---|---|
| `primary` `#27B1FF` | `$primary-palette` **500** -- the brand blue, and the logo's own fill |
| `accentBlueDark` `#0086D4` | `$primary-palette` **700**, which `mat.m2-define-palette` makes the web theme's "darker" variant |
| `accentContainer` `#CCECFF` | `$primary-palette` **50** |
| `selectedChipBorder` `#BBE6FF` | `$primary-palette` **100** |
| `slate50` … `slate700` | `$accent-palette` **50 … 700**, exactly |
| `error` `#D63333` | `$warn-palette` **500** |
| `accentTint` `#EAF7FF` | *not a stop* -- the design's 8% wash on white |
| `borderSlate` `#7E8DA1` | *not a stop* -- a half-step, see below |

So **`#0086D4` is a brand colour, not an off-brand darkening** -- desktop uses that same stop for
the same job (accent text and icons on light) in about sixteen places, including
`roles/role-presets.ts`, which pairs `PRIMARY_TINT = "#ccecff"` with `PRIMARY_COLOR = "#0086d4"`:
the mobile nav pill's exact pairing. This has already been queried once; the answer is here so it
does not have to be re-derived.

**Source a new value from those maps before inventing one.** `borderSlate` is the one grey that
isn't a stop, and only because `$accent-palette` jumps 400 `#94A3B8` (2.56:1, under the border floor)
straight to 500 `#64748B` (4.76:1, as heavy as the label text) with nothing between. Nothing
mechanically ties `colors.dart` to the SCSS, so the two clients stay in step by hand.

**Two tints, not one.** `primaryContainer` is `#EAF7FF` and `accentContainer` is `#CCECFF`, both
from the design, and they are not interchangeable: `accentBlueDark` clears AA on the lighter one at
chip-label size but only clears the large-text / UI threshold on the stronger one, while the lighter
one behind a 24px nav icon is too faint to read as selected at all.

**An outlined field's border cannot be styled the obvious way.** `inputDecorationTheme.border` is a
`WidgetStateInputBorder`, and that is load-bearing. For a non-filled field
`InputDecorator._getDefaultBorder` **throws away whatever `borderSide` the theme's border carries**
and substitutes `_InputDecoratorDefaultsM3`'s, which is built fresh from the context and never
merged with this theme -- so a width set there is silently ignored, and
`InputDecorationTheme.outlineBorder` is ignored too (the *filled* branch consults the app theme; the
outlined one does not). Only the shape survives, which is why the 10px radius works. The single
escape hatch is an early return when the border **is** a `WidgetStateProperty<InputBorder>`, which
bypasses the resolution entirely.

The cost of taking that hatch is that this theme now owns **every** state, so all four are written
out. Dropping the error branch would leave every failed validator in the app without its red border
and **nothing would fail to compile** -- `app_theme_test.dart` resolves all four states for that
reason. Focused is `accentBlueDark` at 2px rather than M3's `primary`, which is 2.38:1.

**`BottomSubmitButton` is the app-wide bottom action bar** — the receipt form, Quick Scan, the
multi-select picker and the receipt filter all mount it. It is a white bar with a hairline top
border holding an inset 48px stadium `FilledButton`. Its total height is exported as
`bottomSubmitBarHeight`, and `submitButtonSpacing` (`lib/constants/spacing.dart`) is **derived from
it**: where the bar floats as a `Scaffold.bottomSheet` rather than a bottom bar, the scrollable
above has to reserve at least that much or its last field sits underneath and cannot be scrolled
clear. Keep the two tied together rather than restating a number.

A known rough edge left alone: Material's `Card` still renders its elevation-1 black shadow as a
hard ring. It has exactly one call site in the app (`receipt_form.dart`'s add-shares card), so it
was out of scope here; the filter's condition cards are hand-built containers instead.

### Permission-based UI gating

The mobile app gates UI on the caller's **effective permissions**, mirroring the desktop client
(`desktop/CLAUDE.md` → "Permission-based UI gating") and the backend's enforcement (`api/CLAUDE.md` →
"Roles & Permissions"). The JWT no longer carries any role field — effective permissions are
delivered on `AppData` and are a UI hint; the server re-checks real permissions on every request, so
a stale action at worst returns 403.

- **Delivery & hydration:** permissions arrive on `AppData` (`appPermissions`, and `groupPermissions`
  keyed by group id) and are stored in **`PermissionsModel`** (`lib/models/permissions_model.dart`)
  via `setPermissions`, called **only** from `storeAppData` (`lib/utils/auth.dart`) on login and
  app-init. Token-only refreshes never touch it. The `FutureBuilder` in `main.dart` blocks first
  paint until app-init completes, so permissions are present before any gated widget builds.
- **Per-group category/tag catalogs:** `AppData` also delivers `groupCategories` / `groupTags` (keyed
  by group id), each filtered to the caller's group-role grants (the full pool when unrestricted).
  `storeAppData` indexes them into `CategoryModel` / `TagModel` (`setGroupCategories` /
  `setGroupTags`); the receipt-form pickers (`category_select_field.dart` / `tag_select_field.dart`,
  passed the receipt's `groupId`) source options via `categoriesForGroup(groupId)` /
  `tagsForGroup(groupId)`. **Do not** source the pickers from the flat `categories` / `tags` arrays —
  those are admin-only (populated only for holders of `app.categories.read` / `app.tags.read`), so a
  normal user would see an empty picker. Mirrors desktop's per-group AppData catalog selectors.
- **Matcher:** `lib/utils/permission_matcher.dart` (`permissionMatches` / `hasAll` / `hasAny`) is a
  faithful port of the backend matcher (`api/internal/permissions/matcher.go`) and its desktop twin
  (`desktop/src/utils/permission.utils.ts`), wildcard semantics included, so UI gating === backend.
- **Effective permissions are plain STRINGS on the wire, deliberately not the `Permission` enum.**
  In `swagger.yml`, `AppData.appPermissions` is an **array of raw wire strings** and
  `AppData.groupPermissions` is a **map of group id → array of raw wire strings** (`BuiltList<String>`
  / `BuiltMap<String, BuiltList<String>>` in the generated client). `PermissionsModel` ingests and
  stores those raw strings; the **query** methods still take the `Permission` enum for call-site type
  safety, converting it to its wire string via `permissionWireName`. This split is load-bearing, not
  cosmetic:
  - `Permission` is a built_value `EnumClass` whose `_$valueOf` ends in
    `default: throw ArgumentError(name)`. An unknown value fails the **entire** `AppData`
    deserialization, and since permissions hydrate on login that **hard-fails login** — the request
    returns HTTP 200 and the user sees a generic red "An error occurred". It shipped **twice**:
    2026-07-24 (`group.members.create`) and 2026-08-06 (`group.members.grants.update`, PR #661).
    Both times the released binary simply predated a backend permission addition.
  - The enum is the **catalog** ("which permissions exist" — `Role.permissions`,
    `UpsertRoleCommand.permissions`, `PermissionDescriptor.key`); the effective list is **data**
    ("which permissions this user holds"), which is server-authoritative and open-ended.
  - A granted entry may be a **wildcard** (`group.receipts.*`), which the matcher supports and the
    enum could never represent.
  - Guarded by `api/internal/permissions/registry_test.go` →
    `TestAppDataEffectivePermissionsAreUntypedStrings` (fails if the `$ref` comes back) and
    `test/models/app_data_permission_ingest_test.dart` (deserializes a payload carrying an unknown
    permission key through the real generated serializer).
  - **General lesson:** adding a value to *any* enum on a response model is a breaking change for
    already-released mobile builds. Effective-permission-style payloads should not use closed enums.
- **Checks** (`PermissionsModel`): `hasAppPermission(p)`, `hasAnyAppPermission([..])`,
  `hasGroupPermission(groupId, p, {orApp})` — the group one applies the `orApp` app-scoped override
  first (the backend `OrAppPermissions` admin-not-a-member pattern) — and
  `hasGroupPermissionInAnyGroup(p)` for screens with no single current group.
- **Gating pattern:** read `Provider.of<PermissionsModel>(context, listen: false)` and conditionally
  render, referencing the generated `Permission` constants. Most gating is **widget-level**, but a
  couple of permission-scoped routes also have **go_router redirects** (see "Route redirects" below) —
  a `redirect:` callback reads the same `PermissionsModel` from the provider tree (the router is built
  under the app's `MultiProvider`, exactly like the existing `/receipts/add` redirect), so the check
  is identical to the widget-level one.
- **Current widget gates:**
  - `canEditReceipt(permissionsModel, groupId)` (`lib/shared/functions/permissions.dart`) →
    `group.receipts.update`, used by the receipt-edit popup (`receipt_edit_popup_menu.dart`) and the
    receipt swipe-to-edit (`receipt_list_item.dart`).
  - `canCommentCreate` / `canCommentDelete` (same file) → `group.comments.create` /
    `group.comments.delete`, scoped to the receipt's group. The comment-input bottom sheet
    (`receipt_comment_screen.dart`) is hidden without create; the swipe-to-delete
    (`receipt_comments.dart`) is disabled without delete (only in **edit** state — add-state comment
    removal is a local model edit with no API call, so it stays enabled).
  - Activity rerun (`group_activity_list_item.dart`) → `group.activities.rerun`. Covered by a widget
    test (`test/widgets/group_activity_list_item_test.dart`) only — there is **no e2e** for it because
    a failed activity's `canBeRestarted` backend state isn't deterministically seedable from the API.
  - The **receipt-entry affordances** — the scan/add bottom-nav slot, its long-press menu, the
    receipts-screen overflow menu, and the Quick Scan sheet — all read one resolver,
    `resolveReceiptEntryAvailability` (`lib/shared/functions/receipt_entry_availability.dart`). See
    "Receipt entry (Scan / Add)" below.
  - The **Search** bottom-nav destination (`group_bottom_nav.dart`, `group_select_bottom_nav.dart`) is
    shown only on `app.receipts.search`. It is the **trailing** destination in both navs; the scan
    slot ahead of it is also gated, so both navs resolve destinations by **id** rather than by index
    (`NavDestinationItem` in `bottom_nav.dart`).
  - The **Reports** avatar-menu entry (`top_app_bar.dart`'s `getUserAvatar`) is shown only when the
    caller holds `app.reports.read` **or** `app.reports.readAll` (`hasAnyAppPermission`), mirroring
    desktop's `canViewReports` sidebar gate. Base and `*All` are unrelated matcher keys, so the two are
    always OR-ed. The avatar is shared by both app bars, so this exposes Reports from every screen. See
    "Reporting (mobile slice)" below.
- **Route redirects** (`lib/guards/permission-guard.dart`, mirroring the desktop route guards; wired
  in `main.dart`'s `_buildAppRouter`):
  - `groupDashboardReadRedirect` on `/groups/:groupId/dashboards` → redirects to that group's
    `/receipts` when the caller lacks `group.dashboards.read` for the route's group (mirrors desktop's
    `groupDashboardReadGuard`). Group selection always lands on the dashboards route, so this covers
    both group entry and the Dashboards tab tap; the Dashboards tab itself stays visible (redirect-only,
    matching desktop).
  - `receiptsSearchRedirect` on `/search` → bounces deep links to the originating group's `/receipts`
    (or `/groups`) when the caller lacks `app.receipts.search`. Defense-in-depth behind the hidden
    search buttons; it also runs before the search shell's `state.extra as Map` cast.
  - `reportsReadRedirect` on `/reports` → bounces deep links to `/groups` when the caller lacks
    `app.reports.read` / `app.reports.readAll`. Defense-in-depth behind the hidden Reports menu entry.
- **403 handling (`lib/interceptors/auth_interceptor.dart`):** the backend returns **403 for both** an
  expired session and a permission denial (it never sends 401), so the interceptor distinguishes them
  by **token validity** (mirroring desktop's `http-interceptor.ts`): a 403 with a still-valid token is
  a permission denial and is surfaced untouched — **no token refresh, no logout** (refreshing can't
  grant a missing permission, and force-refreshing would burn the one-time refresh token / risk a
  logout). Only an expired/invalid token triggers a refresh + retry. Token freshness is otherwise kept
  current proactively (the 15-min timer in `main.dart` and the auth guard on navigation). **Paid-by
  visibility** rides on this: a group role limited to "their own receipts" gets a server-filtered
  receipts list, and any stray 403 on a hidden receipt is surfaced without disturbing the session.

### Reporting (mobile slice)

A **read/preview/generate/delete** slice of the desktop reporting feature, reachable from the avatar
menu (`lib/reports/`). Authoring is intentionally **out of scope** — there is no builder and no
create/edit/duplicate/config-view; the mobile app only lists saved report templates, previews one,
downloads (generates) one, and deletes one. The desktop client (`desktop/src/reports/`) and the Go
`/report/*` API are the source of truth; this mirrors desktop's behavior and — critically — its
permission model exactly.

- **Entry + route:** the gated avatar-menu "Reports" item (above) pushes `/reports`
  (`ReportListScreen`, a standalone `ScreenWrapper` route next to `/profile`), guarded by
  `reportsReadRedirect`.
- **List** (`report_list.dart` → `PagedDataList`): `POST /report/template/list` via
  `getReportApi().getReportTemplates`, sorted `updated_at desc` (matching desktop; the backend
  allow-lists `name`/`created_at`/`updated_at`). Rows unwrap `PagedDataDataInner.anyOf.values`
  (a `Map<int, Object?>`) by **type** — `.values.whereType<ReportTemplate>()` — not a brittle fixed
  index.
- **Per-row actions are gated ONLY on the server-computed `ReportTemplate.allowedActions`**
  (`read`→Preview eye, `generate`→Generate, `delete`→Delete), never AND-ed with a client permission
  check — `allowedActions` already bakes in the base/`*All` report permissions, the per-group ceiling,
  and the per-template grant matrix (`report_list_item.dart`, mirroring the desktop list). Deletion
  confirms via `report_delete_dialog.dart` then `DELETE /report/template/{id}`.
- **Preview = HTML, not PDF** (`report_preview_screen.dart`): desktop has no PDF preview — its preview
  is a live, row-capped **HTML** sample. `POST /report/preview` (with the template's stored
  `configuration`) returns `{ html, receiptCount }`, rendered in a **WebView**
  (`WebViewController.loadHtmlString`, JS disabled — the sample is self-contained, no network). Can
  403 if a covered group lacks `group.reports.read`; handled with a snackbar, no logout.
  - **`configuration.period.dateField` rides through untouched.** It names which receipt date the
    period covers (see `api/CLAUDE.md` → "The period's date field"). Mobile never reads it, but
    preview re-sends it, so it must survive the round trip.
  - **It is a plain `String?`, on purpose.** A closed enum would fail the whole template on the first
    date key added after a build ships.
  - **Builds released before the field existed drop the unknown key when re-sending**, so their
    *preview* of such a template runs on the receipt date. Generate and the dashboard widget load the
    configuration server-side, so they are correct on every build.
  - `test/models/report_period_date_field_ingest_test.dart` guards all of this.
  - **The split options ride through the same way** (`splitCategoriesEqually`, `splitTagsEqually`,
    `splitExcludedFields` — booleans and a list of plain strings, all optional). Mobile never reads
    them. Builds released before they existed drop them when re-sending, so their *preview* of a
    split template counts each receipt in full; generate and the dashboard widget are correct on
    every build. `test/models/report_split_ingest_test.dart` guards the round trip.
- **Generate = the template's saved formats** (`report_actions.dart`): `POST
  /report/template/{id}/generate` returns `Response<Uint8List>` (a single file, or a **ZIP** when the
  template has multiple formats). Bytes are written to `getTemporaryDirectory()` and handed to the OS
  **share / "Save to Files"** sheet via `SharePlus.instance.share(ShareParams(files: [XFile(...)]))`.
  The filename is a port of desktop's `reportFilename` (`report_filename.dart`): sanitized name +
  `.zip` for multi-format else `.<format>`.
- **Packages added for this slice:** `webview_flutter` (HTML preview — Android system WebView, needs
  `INTERNET` which is already declared, minSdk 24 which `flutter.minSdkVersion` already satisfies; iOS
  WKWebView, no `Info.plist`/`NSUsageDescription` for `loadHtmlString`), `path_provider` (temp file),
  `share_plus` (share sheet — auto-merges its own Android `FileProvider`, no manifest change). **None
  add a privacy-sensitive OS permission or purpose string**, and each ships its own iOS
  `PrivacyInfo.xcprivacy` (auto-processed under Flutter's dynamic framework linking). The app still has
  no app-level `PrivacyInfo.xcprivacy` — a pre-existing gap (`shared_preferences`/UserDefaults already
  qualifies), out of scope here.
- **Serialization contract (paged receipts + custom fields):** the receipts list
  (`getReceiptsForGroup` = `POST /receipt/group/{groupId}`, used by `group_receipts_list.dart` and
  `dashboard_widgets/filtered_receipts.dart`, both casting `anyOf.values[0] as api.Receipt`) now
  carries `customFields` populated. The API preloads each value **together with its `CustomField` and
  that field's options** precisely because of the rule below: `models.CustomFieldValue.CustomField` is
  a non-pointer Go struct with no `omitempty`, so loading the values alone would emit
  `"customField":{"type":""}`, the generated `CustomFieldType` enum would throw on the empty string,
  and the `AnyOfSerializer` would swallow it — blanking **every row** of the receipts list on
  already-released builds. The Go side has a guard test asserting the response never contains
  `"type":""`; keep the preload set intact in any future change to that handler. See
  `api/CLAUDE.md` → "Custom fields on the receipts list".
- **Serialization contract (`aggFunc` omitempty):** the mobile list unwraps each `PagedDataDataInner`
  by **type** (`item.anyOf.values.values.whereType<ReportTemplate>()`, `report_list.dart`), so a
  `ReportTemplate` that fails to deserialize silently collapses to a blank row (the `one_of`
  `AnyOfSerializer` swallows per-type errors). The generated `ReportColumnAggFuncEnum` **throws on an
  empty string**, so the Go `ReportColumn` fields (`aggFunc`/`label`/`field`/`measure`/`expr`) are
  `json:",omitempty"` (`api/internal/commands/report_request_command.go`) — a dimension column must not
  emit `"aggFunc":""`. **Keep this omitempty on the API side; any future generated-client regen must
  keep the enum tolerant / the field omitted**, or every real report (dimension/formula columns) shows
  invisible rows in the mobile list.
- **Dashboard report widget (view-only):** the fifth dashboard widget type, `REPORT`, ported from
  desktop's `desktop/src/dashboard/report-widget/`. `lib/groups/widgets/dashboard_widgets/report_widget.dart`
  pins a saved template (reads `configuration.reportTemplateId` from the widget's untyped config blob),
  asks the server to render the **full dataset** (`POST /report/template/{id}/render` →
  `renderReportTemplate` → `ReportPreviewResponse { html, receiptCount, allowedActions }`), and drops the
  self-contained HTML into a **WebView** — the same rendering path as `report_preview_screen.dart` (JS
  disabled, spinner cleared on `onPageFinished`). A revoked/deleted template returns restricted-notice HTML
  at 200 with empty `allowedActions`, so there is no client "restricted" branch. Wired into the widget-type
  `switch` in `group_dashboard.dart` (bounded by the shared `SizedBox(height: widgetHeight)` so the report
  scrolls inside the tile). The **Download** button is gated **only** on the render response's
  `allowedActions.contains('generate')` (never AND-ed with a client permission — same contract as
  `report_list_item.dart`); on tap it fetches the template then reuses the existing
  `generateAndSaveReport` share helper, mirroring desktop's `downloadTemplateById` (get template →
  generate). Authoring stays desktop-only. This required a **mobile client regen** to pick up
  `WidgetType.REPORT`, `renderReportTemplate`, and `allowedActions` on `ReportPreviewResponse` — **no
  swagger change** was needed (the spec already carried all three; the mobile client was simply stale).
- **Tests:** `test/widgets/report_list_item_test.dart` (the `allowedActions` row-gating contract),
  `test/widgets/report_widget_test.dart` (the dashboard widget's pure `reportTemplateIdFromConfig`
  extraction + `reportWidgetCanDownload` gate — the WebView render path is not widget-testable, matching
  `report_preview_screen.dart`), `test/widgets/top_app_bar_reports_menu_test.dart` (menu-entry gate),
  `test/reports/report_filename_test.dart` (filename derivation), and `reportsReadRedirect` cases in
  `test/guards/permission_guard_test.dart`.
  Shared builders in `test/helpers/report_test_helpers.dart`. **E2E:** `integration_test/reports_list_test.dart`
  drives the list view — seeds a template via the new `createReportTemplate` fixture and asserts the row
  renders (the omitempty regression guard), the empty state, and the avatar-menu gate; the
  `provisionUserWithAppPermissions` fixture (inverse of `provisionUserWithoutAppPermission`) grants
  `app.reports.read`/`readAll`. Preview/generate/delete are out of the e2e scope.
  `integration_test/report_dashboard_widget_test.dart` drives the **dashboard report widget** end-to-end:
  it provisions a user (`app.reports.read`/`readAll`/`generate`/`generateAll` + Legacy Owner group), seeds
  a template, then seeds a dashboard holding a `REPORT` widget via the new `createDashboard` fixture
  (**with the fixture user's own jwt** — `getDashboardsForUserByGroup` filters on `user_id`, so an
  admin-owned dashboard would be invisible to the viewer), enters the group, and asserts the `ReportWidget`
  mounts, the WebView renders (success branch, no error placeholder, spinner clears), and the server-gated
  Download button shows. Three `testWidgets` share a `seedAndOpenReportDashboard` helper — the positive path
  plus two **deny** cases that pin down the widget's rejection contract (the render endpoint never 403s; a
  denied caller gets restricted-notice HTML at 200 with empty `allowedActions`, so denial shows up as the
  Download button being withheld, never an error): (a) a user with `app.reports.read` but **not** generate →
  the report still renders but the Download button is `findsNothing`; (b) a Legacy Viewer with no
  `app.reports.*` and no `group.reports.read` → the restricted-notice HTML renders gracefully with no
  Download. **iOS/Android only** (`skip: Platform.isLinux`) — `webview_flutter` has no Linux desktop
  implementation, so the Linux `run-e2e.sh` runner can't mount the widget.

## Development Notes

### Flutter SDK Setup (Claude Code Environment)

When working in the Claude Code environment, Flutter may not be pre-installed. To install the latest Flutter SDK on Debian/Ubuntu:

```bash
# Prereqs. curl/git/pkg-config/xz-utils are usually already present; the rest
# are required for Linux desktop builds (needed for integration_test runs).
apt-get update && apt-get install -y --no-install-recommends \
  unzip zip clang lld llvm cmake ninja-build libgtk-3-dev

# Download and extract Flutter SDK. Check the current stable version at
# https://storage.googleapis.com/flutter_infra_release/releases/releases_linux.json
# (the `current_release.stable` field names the hash; find its `version`).
cd /tmp && rm -rf flutter && \
curl -fL https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.41.7-stable.tar.xz -o flutter.tar.xz && \
tar xf flutter.tar.xz && rm flutter.tar.xz

# Fix git "dubious ownership" warning, add to PATH persistently, disable analytics.
git config --global --add safe.directory /tmp/flutter
grep -q '/tmp/flutter/bin' /root/.bashrc || echo 'export PATH="/tmp/flutter/bin:$PATH"' >> /root/.bashrc
export PATH="/tmp/flutter/bin:$PATH"
flutter config --no-analytics

# Verify and enable Linux desktop target (needed for ./run-e2e.sh).
flutter --version
flutter config --enable-linux-desktop
flutter devices  # should list "Linux (desktop)"
```

After installing Flutter, standard commands work from `mobile/`:
```bash
cd /app/mobile
flutter pub get      # Install dependencies
flutter analyze      # Check for errors (recommended before building)
flutter test         # Run unit/widget tests
./run-e2e.sh         # Run integration tests on Linux desktop (see E2E Testing below)
flutter build apk    # Build Android APK (requires Android SDK — not installed by default)
```

**Note:** The base environment does not include the Android SDK or Chrome, so `flutter build apk` and web targets will not work without additional setup. Linux desktop + `flutter analyze` + `flutter test` + integration_test are fully supported.

### Regenerating API Client Models

After regenerating the API client with `generate-client.sh`, you need to run build_runner to generate the `.g.dart` files:

```bash
cd /home/user/receipt-wrangler/mobile/api
flutter pub run build_runner build --delete-conflicting-outputs
```

**Known dart-dio regressions — patched automatically, NOT by hand.**
`api/patches/apply-dart-dio-patches.sh` holds all four, and `api/generate-client.sh` runs it as the
last step of a `mobile` regen. Add a patch there, never to the generated file: a patch applied by
hand survives exactly until the next regen, which is how this list used to be a checklist item
someone had to remember.

The script has three outcomes per patch, and the third is the point: apply it, skip it if already
applied (so re-running is safe), or **exit 1** when neither the original nor the patched text is
present — meaning the generator's output changed shape and the patch must be re-derived. It fails
the regen rather than quietly handing back an unpatched client.

Two of the four are invalid Dart that `build_runner` can't compile (it deletes the matching
`.g.dart` first, so the package stops building); precedent commits `fad192a0`, `a2ec7479`:

- `model/user_preferences.dart` — `..quickScanDefaultStatus = 'OPEN'` →
  `..quickScanDefaultStatus = ReceiptStatus.OPEN`.
- `model/system_settings.dart` — `..currencyDisplay = '$'` → `..currencyDisplay = r'$'` (a bare `$`
  in a non-raw string is invalid Dart).

The other two are the `fallback: true` enum annotations below, which compile fine when absent and
so are caught only by `flutter test`.

`model/claims.dart` **no longer** needs a `userRole` patch — the role rework dropped `userRole` from
the swagger, so `Claims` carries only identity claims and the field is gone from the generated model.

Run `flutter analyze` **and `flutter test`** after a regen. (Hand-editing generated files is
otherwise forbidden — these four are the documented exception, and they are applied by the script.)

**A patch that does NOT surface as a compile error — `ReceiptStatus` enum tolerance.**
`model/receipt_status.dart` annotates the `empty` member `@BuiltValueEnumConst(wireName: r'',
fallback: true)`. `build_runner` turns that into `default: return _$empty;` in `_$valueOf`, replacing
the generated `default: throw ArgumentError(name)`. Without it, one unrecognized wire value fails the
**entire** enclosing payload — the whole paged receipts response, and **login** when the value rides
on `AppData` through a group's `emailDefaultReceiptStatus` or either `quickScanDefaultStatus`. That is
the same closed-enum mechanism behind the two `Permission` login outages (see "Permission-based UI
gating"); `ReceiptStatus` cannot take that feature's fix of becoming a plain wire string, because it
is a genuine closed domain backing the status dropdowns.

A bare regen does not emit `fallback` and unlike the two patches above nothing fails to compile, so
the only thing that catches the annotation going missing is
`test/models/receipt_status_ingest_test.dart`. Run `flutter test` after a regen, not just
`flutter analyze`.

**The generator CAN emit it, and the flag is a trap — do not reach for it.** `enumUnknownDefaultCase`
(via `--additional-properties`) turns on the stock template's `fallback: true`, but it is **global**,
so it would also open `Permission` — which this repo keeps deliberately closed, with
`test/models/app_data_permission_ingest_test.dart` asserting it still throws. It also anchors the
fallback to a **synthetic** `unknown_default_open_api` member rather than to `empty`, which changes
the write path (`empty` serializes to `""`, the synthetic member to a literal string the API rejects)
and breaks every assertion below. The patch script is the narrower instrument, on purpose.

**A new field on a response model needs a hand-written guard in `test/models/`, not the generated
stub.** The generator drops a `*_test.dart` per model in `mobile/api/test/`, but it **does not
overwrite one that already exists** — so those stubs still describe whatever the model looked like
the day they were first written, never gain a case when a field is added, and are all `// TODO`
bodies regardless. They also sit under `mobile/api/`, which is generated output nobody may
hand-edit. The six real API-boundary guards therefore live beside the app's own tests and are what
`flutter test` actually runs: `test/models/app_data_permission_ingest_test.dart`,
`test/models/receipt_status_ingest_test.dart`,
`test/models/receipt_summary_position_ingest_test.dart`,
`test/models/user_preferences_ingest_test.dart` (the `closeChipSelectOnSelect` default and wire
round trip — a desktop-only flag mobile carries and never reads, which is precisely why nothing in
the app would notice it disappearing), and `test/models/report_period_date_field_ingest_test.dart`
(`ReportPeriod.dateField` stays an open string and survives the preview round trip — an unknown key
must not fail the template), and `test/models/report_split_ingest_test.dart` (the report split
options survive the same round trip and are not invented when absent).

Two consequences worth knowing. An unknown status deserializes to `empty`, which
`receiptStatusLabel` / `receiptStatusColor` render as a blank neutral chip rather than crashing —
degraded, not broken. And on the **write** path `empty` serializes back to `""`, which
`UpsertReceiptCommand.Validate` rejects with a 400 "Status is required": an old build editing a
receipt whose status it cannot represent gets a visible error instead of silently downgrading it.
That is the intended failure — loud, and no data loss.

**`ReceiptSummaryPosition` carries the same patch — but on `BOTTOM`, not on an `empty` member.**
`model/receipt_summary_position.dart` annotates
`@BuiltValueEnumConst(wireName: r'BOTTOM', fallback: true)`, and
`test/models/receipt_summary_position_ingest_test.dart` is what catches it going missing. It rides on
`GroupReceiptSettings`, i.e. on **AppData**, i.e. on **login** — a third position added later would
brick every released build at the sign-in screen.

The **enum carries no `""` member at all**, which is the difference from `ReceiptStatus`. There
`empty` is a value the app actually uses (the overall summary row carries it, since the generated
`status` field is non-nullable), so it earns its place. Here nothing read it — it existed only to
host the fallback, while also making `""` expressible on
`UpdateGroupReceiptSettingsCommand`, where the server 400s it. Putting the fallback on `BOTTOM`
instead drops the member, makes the invalid write unrepresentable, and lands an unknown value on a
**real** placement: the same one `ReceiptSummaryPosition.OrDefault()` would have sent server-side, so
the two agree by construction. The server still normalizes `""` away at both emit points (see
`api/CLAUDE.md`); keep both.

The sibling closed enums (`ItemStatus`, `GroupStatus`, `SystemTaskStatus`, `Permission` as a
catalog) are **still intolerant**. Adding a value to any of them is a breaking change for released
builds until they get the same treatment.

### Role-required receipt fields & single-call create

A group role can require its members to keep at least one comment and/or one image on the group's
receipts. The server owns the rule and enforces it (see `api/CLAUDE.md` → "Role-required receipt
fields & single-call create"); mobile has **no role UI** and only reads the result, to fail early
instead of round-tripping a 400.

- **Delivery.** `AppData.groupReceiptRequirements` (`BuiltMap<String, ReceiptRequirements>?`) is
  already resolved server-side, waivers included (`hideComments` / `hideImages`, no
  `group.comments.create`, the All group). Only groups where something is required are present.
  `storeAppData` passes it to `PermissionsModel.setReceiptRequirements`;
  `receiptRequirements(groupId)` answers both-false for an absent group **and for a null map** — the
  field is deliberately outside AppData's swagger `required` list, so an older server's payload
  still parses. Pinned by `test/models/receipt_requirements_ingest_test.dart`.
- **Create is ONE multipart call.** `addReceipt` (`receipt_bottom_sheet_builder.dart`) calls
  `createReceiptWithImages` (`lib/shared/functions/receipt_upload.dart`) →
  `ReceiptApi.createReceiptWithFiles(receipt:, files:)`: the command carries the staged comments,
  `files` the staged images. It is **atomic** — a failure creates nothing — so the old JSON
  `createReceipt` + per-image `uploadReceiptImage` flow and its "Receipt added, but one or more images
  failed to upload" partial branch are gone. A failed create leaves the user on the add form. The
  deprecated JSON create is kept server-side for released builds only; do not call it.
  - **Each file is rebuilt from its bytes** (`MultipartFile.fromBytes`) rather than sending the
    staged `UploadMultipartFileData.multipartFile`: a dio `MultipartFile` can be finalized only
    once, so resending it after a failed create would throw before the retry left the device.
  - Edit is unchanged: images and comments still go through their own endpoints immediately.
- **Submit guard** (`buildReceiptSubmitButton`, after `saveAndValidate`): judged against the form's
  `groupId` (the destination on a move, as the server does) by `receiptSubmitRequirementsMessage`
  (`lib/shared/functions/receipt_requirements.dart`). Add checks the **staged** images and
  `receiptModel.comments`; edit checks the **saved** ones — `imageBehaviorSubject` once the images
  screen has loaded it, else `receipt.imageFiles`. Only non-blank comments count, matching the server.
- **Last-item deletes, edit state only:** the comment swipe (`receipt_comments.dart`) is disabled on
  the last non-blank comment, and "Delete Image" (`receipt_image_app_bar.dart`) on the last saved
  image, while that field is required in the receipt's group. To swap the only image, upload first.
  The image menu's entries are now built **inside** its `StreamBuilder`, so they follow the image
  list instead of being fixed at the app bar's build.
- **Quick Scan:** `resolveQuickScanFieldConfig` takes a required `commentRequiredByRole` and ORs it
  in — the comment is shown **and** required even when the group's quick-scan config leaves it off,
  mirroring the server's `ResolveQuickScanFields` (which would otherwise 400 the scan).
- **Known gap:** a requirement 400 the client did not predict (stale AppData) is a validator-shaped
  body (`{"comments": "..."}`) with no `errorMsg`, so `showApiErrorSnackbar` shows the generic
  "An error occurred".
- **`openReceiptEditForm` waits for the `name` field, not just the submit button.** The edit
  screen's `BottomSubmitButton` mounts while the form body is still loading, so a submit tapped
  on that signal finds no `FormBuilder` state and is a silent no-op (the handler's defensive null
  check) -- `receipt_role_requirements_test`'s edit case timed out on exactly that until the helper
  waited the way `receipt_edit_test` already did.
- **Tests:** `test/shared/functions/receipt_requirements_test.dart`,
  `test/shared/functions/receipt_create_with_images_test.dart` (one call, no follow-up upload, fresh
  files on retry), `test/widgets/receipt_submit_requirements_test.dart` (the real submit button on the
  real form, via `pumpReceiptForm(realSubmitButton: true, permissionsModel: ...)`),
  `receipt_image_app_bar_test.dart`, and new cases in `receipt_comments_test.dart`,
  `quick_scan_form_test.dart` and `quick_scan_field_config_test.dart`. **E2E:**
  `integration_test/receipt_role_requirements_test.dart` (add blocked then created in one call with
  both, edit save blocked, last comment locked, Quick Scan comment required) and
  `integration_test/receipt_add_partial_image_test.dart` (a server-rejected file creates nothing; a
  valid one is stored with the receipt).

### Receipt entry (Scan / Add)

The bottom-nav slot that used to open an "Add" menu is a **direct action**. A **tap** scans; a
**long-press** opens the menu. The whole feature is gated on three independent inputs, and the
backend enforces the two permissions **separately** (`handlers.QuickScan` → `group.receipts.quick-scan`,
`handlers.CreateReceipt` → `group.receipts.create`), so a role can hold either without the other.

- **One resolver.** `resolveReceiptEntryAvailability(context)`
  (`lib/shared/functions/receipt_entry_availability.dart`) returns `canQuickScan`
  (`aiPoweredReceipts` **and** `group.receipts.quick-scan`), `canCreateManual`
  (`group.receipts.create`), a `QuickScanBlockedReason?` and the group name. Scoping is the rule the
  old add menu had: per-group inside a group, "held in **any** group" on group-select / the
  all-groups view. **Every** affordance reads this — never re-derive the gates.
- **`blockedReason` prefers `aiDisabled`** when both are missing: it is the install-wide,
  administrator-level explanation, and pointing the user at a permission they still could not use
  would send them to the wrong person.
- **The slot is built once**, by `buildScanNavItem` (`lib/shared/widgets/scan_nav_item.dart`):
  `Icons.document_scanner` / "Scan" when `canQuickScan`, `Icons.add` / "Add" otherwise, and **`null`
  when neither permission is held** — the destination is omitted rather than shown and then refused.
  Both navs mount what it returns, so the behaviour is identical everywhere.
- **Index safety.** Because the slot can vanish from the middle, `BottomNav` takes
  `NavDestinationItem`s (id + destination) and both navs switch on the id. `BottomNav` also renders
  **nothing** below two destinations: Material's `NavigationBar` asserts there, and on group-select a
  user with neither entry permission nor `app.receipts.search` is left with "Groups" alone.
- **Long-press mechanics.** `NavigationBar` has no long-press API, so `BottomNav` overlays an
  equal-flex `Row` of `GestureDetector`s with `HitTestBehavior.translucent`. A tap loses the gesture
  arena to the bar's own `InkWell` (so selection still works); a hold is claimed by the long-press
  recognizer at 500ms before the tap can complete.
  **The action runs on `onLongPressUp`, not `onLongPress`.** `onLongPress` fires at the 500ms mark
  with the finger still down, and these actions open a route — pushing one mid-gesture leaves the
  recognizer holding a pointer it never saw released, after which that slice ignores every later tap
  *and* hold. The slot worked exactly once per launch. Waiting for the release costs nothing: the
  recognizer has already won the arena at 500ms, so the tap underneath stays suppressed either way.
  The guard is `integration_test/receipt_entry_menu_reopen_test.dart` and has to be an e2e — a
  minimal widget harness delivers the pointer-up cleanly enough that the recognizer recovers, so the
  equivalent widget test passes with the bug present.
  **The slot sets `tooltip: ""`** (empty = none). A `NavigationDestination` otherwise shows a tooltip
  **on long press**, falling back to its label when `tooltip` is null — a second long-press recognizer
  in the arena against the one that opens the menu. Rather than depend on which wins, Material's is
  taken out of the running, and the gesture hint rides on a `Semantics(hint: scanNavLongPressHint)`
  wrapper around the icon instead (`Semantics` installs no recognizer).
- **Actions** live in `lib/shared/functions/receipt_entry.dart`; copy lives in
  `lib/constants/receipt_entry.dart` (shared so the sheet's snackbar and the form's banner can't
  drift). `startScanEntry` is the tap: blocked → the manual form carrying the reason; otherwise
  `ensureCameraAccess()` decides camera vs. gallery fallback.
- **The banner only appears on the fall-through.** `openManualReceipt` attaches the reason to the
  route `extra` **only** from the tap path, so a deliberate "Add Manual Receipt" is never nagged.
  `ReceiptFormScreen` renders `QuickScanUnavailableBanner.fromRouteExtra(state.extra)`.
- **The gates are read with `listen: false`**, the house pattern — permissions and `featureConfig`
  hydrate before first paint (`main.dart`'s `FutureBuilder`) and don't change during a session. So
  the slot's **label** is fixed when the nav builds, while the **menu** is built fresh on each open.
  A test that mutates `AuthModel.featureConfig` live will see the menu move but not the label; assert
  the label where the flag is real (set server-side before login), as
  `permission_add_menu_test` and `quick_scan_entry_gated_test` do.
- **The overflow menu is the receipts screen's only.** `GroupAppBar` is shared with the dashboards
  route, so it gates on `GoRouterState.of(context).fullPath`. It is the accessible equivalent of the
  long-press, and carries the same items from `buildReceiptEntryMenuItems`. It is **not** on
  `GroupSelectAppBar`. Its items run from the `PopupMenuButton`'s **own in-app-bar context**, and
  `_refreshBeforeQuickScan` raises the loading bar in that same app bar — which is why that bar must
  never toggle `AppBar.bottom` between null and non-null (see "`TopAppBar`'s loading bar is always
  mounted"). It is also why `quick_scan_entry_test.dart` **taps** the overflow's Quick Scan rather than
  only asserting its label: a widget test cannot reproduce this, since `TokenRefreshService` is
  uninitialized there and `reloadAppData()` returns within a microtask, so the bar never gets a frame.
- **Both picker entries are gated on quick-scan, not create** — they feed the Quick Scan sheet, so
  offering them to a create-only user would produce a sheet they cannot submit.
- **A submitted sheet confirms itself** (`quick-scan-queued-confirmation`). Submitting disables every
  field and hides the submit button, so once the success snackbar fades the sheet would otherwise sit
  there greyed out with nothing saying why. Extraction is an async backend job, so the wording
  promises a result rather than showing one.
- **The sheet pages a multi-page scan with arrows as well as the swipe.** Under the image sits a nav
  row (`lib/receipts/widgets/quick_scan.dart`): a counter (`quick-scan-page-indicator`, `"2 of 3"`)
  flanked by `quick-scan-previous-page` / `quick-scan-next-page`, which call the carousel
  controller's `previousItem()` / `nextItem()`. The whole row is absent below two pages. A scan can
  carry up to 100 pages and the carousel gives no other hint that the ones after the first exist —
  without this a multi-page scan reads as a single-page one and the forms behind it go unfilled, and
  the counter alone left that to a line of text ("swipe for the next", now dropped) which is easy to
  skip past.
  - **An arrow at the end of the scan is dropped, not disabled**, so every arrow shown leads
    somewhere. It is replaced by a `SizedBox.square(dimension: kMinInteractiveDimension)` — the
    `IconButton` default constraint — so the counter stays put instead of sliding as the arrows come
    and go.
  - **The row reads the `itemBuilder`'s `realIndex`, not a tracked selection.** The carousel builds
    only the item on screen, so the page's own index *is* the visible one; deriving from it is what
    lets the delete icon (which does `images.removeAt(controller.selectedItem)`) shrink the scan
    without leaving a stale index behind an arrow that points off the end. That is why `QuickScan` is
    a `StatelessWidget` with no `onIndexChanged`.
  - Arrows stay live after submit, unlike the upload/delete icons: paging a queued scan is read-only
    browsing, and the swipe allows it either way.
  - Covered by `test/widgets/quick_scan_sheet_test.dart` ("the arrows only ever point at a page that
    exists"). There is **no e2e** for it: `installDocumentScannerMock()` returns a single fixed PNG,
    so no integration spec can reach a second page.
- **The sheet re-checks its own gate** and resolves `canCreateManual` **at the caller's context**.
  Its `bottomSheetWidget` is built inside the modal route, which is outside the GoRouter subtree, so
  `GoRouterState.of` throws there — resolve before opening and pass the result down.

### Picking receipt files

Three sources feed a receipt image, selected by `UploadMethod { camera, photos, files }` and
dispatched by the single `acquireReceiptFiles(context, method)`
(`lib/shared/functions/receipt_upload.dart`). Every entry point routes through it — the Quick Scan
sheet's two app-bar actions, the scan slot's long-press menu, the receipts overflow menu and the
receipt-image app bar — so which picker a source opens, how a failure is reported, and whether the
result is guarded before it touches the tree are decided once.

**Why the photo library is its own source.** `file_selector` opens a *document* picker on both
platforms — `ACTION_OPEN_DOCUMENT` (SAF) on Android, `UIDocumentPickerViewController` on iOS. On
Android photos are at least reachable through the file browser; **on iOS the Files app is not a view
onto the photo library and the camera roll is unreachable from it entirely**, so the button that used
to say "Upload from Gallery" could not open the gallery. `image_picker` opens the Android Photo
Picker / `PHPickerViewController` instead. Both are kept: the photo pickers are media-only by design,
and **only the file source can produce a PDF**, which the backend converts server-side
(`FileRepository.GetBytesFromImageBytes`) and quick scan OCRs end to end.

**Layering.** `lib/utils/media_picker.dart` is the raw half (`pickPhotos` / `pickDocuments`, no
Provider, no `BuildContext`, no snackbars); `lib/shared/functions/receipt_upload.dart` is the policy
half. That mirrors the existing `lib/utils/` vs `lib/shared/functions/` split. Top-level functions,
not a class — the operation is stateless, and the test seam already exists one layer down
(`FileSelectorPlatform.instance` / `ImagePickerPlatform.instance` are both settable), so a
class-level seam would be redundant *and* would stub out the bytes→`MultipartFile` conversion.

**No runtime permission is required and none must be added.** The Android Photo Picker and PHPicker
both work without one — that is their whole point, and Play Console requires a restricted-permission
declaration to justify `READ_MEDIA_IMAGES` when the system picker would do. Do **not** route these
through `ensureCameraAccess` or `Gal.requestAccess`: the latter is for *saving* to the library
(`saveReceiptImageToGallery`), which is what `WRITE_EXTERNAL_STORAGE` and
`NSPhotoLibraryAddUsageDescription` are for. `NSPhotoLibraryUsageDescription` already existed and
needed no change.

**No `imageQuality` / `maxWidth` / `maxHeight` on `pickMultiImage`.** Those make `image_picker`
re-encode on device, and the backend already normalises HEIC before OCR — re-encoding here would only
cost fidelity on the images we most need read accurately.

**One `XTypeGroup`, no platform switch.** `receiptFileTypeGroup` carries `mimeTypes`,
`uniformTypeIdentifiers` and `extensions` at once and each `file_selector` implementation reads the
field it understands. `extensions` is not decorative: `file_selector_windows` throws `ArgumentError`
without it and `file_selector_linux` (the e2e host) needs it because GTK's `image/*` wildcard
handling is unreliable. Deleting the old `Platform.operatingSystem` switch is what un-skipped three
Linux e2e specs.

**`useAndroidPhotoPicker` is set in `main()`**, after `WidgetsFlutterBinding.ensureInitialized()` and
before `runApp`. It is plugin *configuration* — no permission request, no channel round trip — so it
is exempt from the launch-time-work ban documented in `_ReceiptWrangler.initState` (GitHub #617);
the comment there says so, so it does not get "cleaned up" into that block later. Needed on API
33–35; a no-op on 36+, and below 33 the Play Services `ModuleDependencies` service in
`AndroidManifest.xml` pulls in the backported picker (without it those devices simply fall back to
`ACTION_OPEN_DOCUMENT`, i.e. the old behaviour). The branch itself lives in
`configureAndroidPhotoPicker([ImagePickerPlatform?])`, split out of `main()` **only** so it is
reachable from a test — the parameter defaults to the live instance, so the production call site is
unchanged, and it must stay in `main()` rather than moving into `initState`.
`test/main_photo_picker_test.dart` pins both branches (Android gets the opt-in, any other
implementation is untouched and does not throw). **e2e pumps `buildApp()`, not `main()`, so the
plugin *registration* is still uncovered** — verify that on a physical Android device.

**Undecodable bytes render a placeholder, not a broken-image glyph.**
`UnrenderableFilePlaceholder` / `unrenderableFileErrorBuilder`
(`lib/shared/widgets/unrenderable_file_placeholder.dart`) back the `errorBuilder` on every
`Image.memory` that shows picked or server bytes. A PDF picked from the file source *uploads and OCRs
fine* but cannot be decoded for preview; without this the user sees Flutter's grey broken-image glyph
and concludes the upload failed. It also replaces `Image.asset("assets/images/placeholder.png")` in
`receipt_image_carousel.dart`, which threw — there is no such asset. `ImageViewer.image` is typed
`Widget` rather than `Image` for this reason.

**`QuickScanImage` no longer redeclares `multipartFile` / `bytes`.** It used to, shadowing the
superclass fields. Both copies held the same object so it worked, but Dart fields are not virtual: a
getter on `UploadMultipartFileData` (such as the `filename` the placeholder uses) reads the *base*
field while subclass code reads the shadow. Do not reintroduce them.

**Known gap — `ImagePicker.retrieveLostData()` is not implemented.** On low-memory Android the
activity can be destroyed while the photo picker is foregrounded; the selection is then only
recoverable via `retrieveLostData()` on resume. Today it is silently lost and the user re-picks.
Deferred deliberately: recovering it means deciding *where* the file lands (the Quick Scan sheet is
gone by then), and the failure is rare and non-destructive.

### The receipt-image app bar had an unreachable twin

`ReceiptAppBarActionBuilder.buildAppBarMenu` used to branch on
`state.fullPath.contains("images") / ("comments")` and serve two further menus. **No such route
exists** — the full set is `/`, `/login`, `/groups`, `/groups/:groupId/{dashboards,receipts}`,
`/receipts/add`, `/receipts/:receiptId/{view,edit}`, `/profile`, `/reports`, `/search` — and
`ReceiptImageScreen` is pushed as a plain `MaterialPageRoute` from `receipt_form.dart`, so the path
never changes when it opens. Both branches were dead, and the images one was a stale copy of
`ReceiptImageAppBar` that navigated to `/receipts/:id/images/edit`, a URL the router does not define.
They were **deleted** (306 lines → 58), not merged; `ReceiptImageAppBar` is the live implementation
and now takes its upload plumbing from `receipt_upload.dart`. Do not recreate the copy.

Four latent bugs in that plumbing were fixed while it moved: no try/catch around the picker (now via
`acquireReceiptFiles`), no `context.mounted` guards after awaits, `showApiErrorSnackbar(context, e as
DioException)` throwing a `TypeError` *from inside the catch block* for any non-Dio error (the helper
now takes `Object` and reports non-Dio errors to Sentry, which fixes every call site at once), and a
multi-image upload that assigned a success message and then `return`ed before showing it. The loading
spinner is now raised and lowered in exactly one place, so a cancelled pick can no longer clear a
spinner another request raised.

**`acquireReceiptFiles` swallows every picker failure EXCEPT
`CunningDocumentScannerException`, which it rethrows** — that one means camera permission is
missing, and the right answer differs per call site, so turning it into a snackbar centrally would
delete the distinction. Every caller therefore has to handle it, and there are three:
`openQuickScanFromCamera` (`receipt_entry.dart`) answers it with `fallBackToPhotos`, because it is
*starting* a flow and can redirect the whole thing; `_addPickedImages` (`quick_scan.dart`) and
`addImagesToReceipt` (`receipt_upload.dart`) answer it with the `cameraDeniedFallbackMessage`
snackbar and nothing else, because both already sit inside a surface that offers a photo source one
tap away — opening a second Quick Scan flow over the top would be wrong. The latter two went
uncaught at first, so a denied camera threw out of an async `onTap` and the scan icon simply
appeared to do nothing. Pinned by `test/widgets/quick_scan_sheet_test.dart` ("a denied camera
explains itself instead of escaping"), which asserts on the **permission-request count** rather than
the snackbar count — `ScaffoldMessenger` renders one snackbar into *every* registered `Scaffold`, and
while the sheet is open that is both the sheet's and the route's underneath, so `findsOneWidget`
would fail for a reason that has nothing to do with the code under test.

**`uploadImagesToReceipt` publishes partial results.** Each image is its own API call, so a failure
on the third leaves the first two already persisted server-side. `uploaded` is declared **outside**
the `try` and emitted once in `finally` (guarded by `isNotEmpty`), so every exit carries it —
success, error, and the early return taken when the context is gone. Scoping it inside the `try`
instead meant a mid-batch failure discarded images the server had already stored: the user saw only
an error, and repeating the action uploaded them a second time, leaving the receipt with duplicates.
Guarded by `test/shared/functions/receipt_upload_publish_test.dart`, which also pins that a
first-image failure publishes nothing and that a successful batch emits exactly once.

#### Camera permission

`ensureCameraAccess()` (`lib/utils/permissions.dart`) maps the OS state to
`CameraAccess.granted | denied | permanentlyDenied`. `limited`/`provisional` count as granted;
`restricted` (parental controls / MDM) behaves as permanently denied, since the user cannot grant it
themselves. It requests **only** when the permission is undecided — a request in the permanently
denied state resolves instantly with no dialog, which reads to the user as the tap doing nothing —
and shares a single in-flight future for the same
`ERROR_ALREADY_REQUESTING_PERMISSIONS` reason `requestPermissions()` does.

Denied falls back to the **photo library** with a notice (`fallBackToPhotos`); permanently denied
adds an **Open Settings** snackbar action (`openAppSettings`). Photos rather than a source menu
because this is an automatic continuation, not a choice — the user just tried to *photograph* a
receipt. Someone whose receipt is a PDF in Files can still reach it from the menu. **Keep the request lazy** — a launch-time request was removed
deliberately for the iOS 26.x render-pause freeze (GitHub #617).

`debugCameraAccessOverride` is the test seam (the plugins are statics with no injectable seam),
mirroring the settable `OpenApiClient.client` and `QrScannerScreen`'s `debugForce*` flags.

Two latent bugs on this path were fixed alongside: `scanImagesMultiPart` dereferenced a null with
`!` (the scanner returns null on cancel, now an ordinary flow), and `getPictures` re-requests camera
permission itself and **throws** when it is missing (now caught → photo-library fallback). A third —
`getGalleryImages` throwing off android/ios — is gone with the picker rework below, which deleted the
`Platform.operatingSystem` switch entirely.

### Quick Scan field configuration

The Quick Scan per-image form (`lib/receipts/widgets/quick_scan_form.dart`) respects the selected
group's quick-scan config on `GroupReceiptSettings` — the
`quickScan{PaidBy,Status,Categories,Tags,Comment}{Enabled,Required}` fields (mirrored from the
API/desktop; see `api/CLAUDE.md` → "Quick Scan Field Configuration"). Every field renders per its
`*Enabled` flag, but only paid-by, status and comment get a `required()` **validator** (when shown
**and** `*Required`); Categories/Tags carry no field validator — their requiredness is checked at
submit in `quick_scan.dart` (see the `quick_scan_form_test.dart` note below). Categories/Tags
reuse the shared `CategorySelectField` / `TagSelectField`, sourced from the per-group catalogs, and the
comment is a multi-line `FormBuilderTextField` (`name: "comment"`, `maxLength: 500` matching the
backend's `models.MaxCommentLength`). Because the
backend **backfills** a default for a hidden/optional paid-by or status, `_submitQuickScan`
(`lib/shared/functions/quick_scan.dart`) sends the "unset" sentinel (`0` / `ReceiptStatus.empty`) for
those and per-file comma-joined `categoryIds` / `tagIds` plus a per-file `comments` string, building
**one aligned array entry per image** (never skipping, so `files` and the parallel arrays stay 1:1). It
requires a field only when that group's config marks it shown+required, mirroring the backend's
`resolveQuickScanFields`.

**Until a group is picked, ONLY the Group dropdown renders.** There is no configuration to honour
yet, and the old behaviour — falling back to the backend's column defaults, paid-by/status shown
and required — was a guess that flipped the field set the moment the user chose a group whose
config hides them. Common, not exotic: the group is seeded only from
`userPreferences.quickScanDefaultGroupId` or `soleGroupId`, so any user in ≥2 groups without a
default starts blank.

The show/require derivation for all five fields lives in **one** pure helper —
`resolveQuickScanFieldConfig(GroupReceiptSettings?, {required bool hasGroup, required bool
canCreateComments, required bool commentRequiredByRole})` → `QuickScanFieldConfig` (`lib/shared/functions/quick_scan_field_config.dart`)
— reused by both the form's `build()` and `_submitQuickScan`, so the two can't drift from each
other or from `resolveQuickScanFields`. `hasGroup: false` returns `noGroupQuickScanFieldConfig`
(all ten flags false). Covered by `test/shared/functions/quick_scan_field_config_test.dart`.

- **`hasGroup` is keyed off "a group id is chosen", deliberately NOT `settings == null`.**
  `getGroupReceiptSettings` returns null for *two* states — no group, and a group whose settings
  aren't in `GroupModel` (a stale `quickScanDefaultGroupId`, which
  `quick_scan_initial_values_test.dart` pins as reachable). Only the first collapses to Group-only;
  a selected-but-unresolvable group keeps the backend defaults, so this change moves exactly one
  behaviour. Desktop's resolver takes the same flag for the same reason.
- **Hiding a field must not destroy its value, so `onValueChange` MERGES rather than overwrites.**
  A hidden field is never built (`Visibility` defaults to `maintainState: false`), so it never
  registers with `FormBuilder` and its key is absent from `formKey.currentState.value`. `FormBuilder`
  runs with the default `clearValueOnUnregister: false`, so a key that *has* been registered
  survives the field being hidden later — **an absent key therefore means "never shown", not
  "cleared"**, and the image's own value is the right answer. Reporting `null` for an absent key
  made the carousel consumer (`quick_scan.dart`, which writes every record member onto the
  `QuickScanImage` unconditionally) erase the user's `quickScanDefault*` prefill on the very first
  group selection — for exactly the users who have to pick a group. A field that *is* mounted always
  reports its own value, so the explicit clears in the group dropdown's `onChanged` still apply.
- **The group dropdown's `onChanged` re-reports in a post-frame callback.** The new group's fields
  mount on the *next* frame, each seeding from the image — and a prefilled paid-by who is **not a
  member** of the group just picked seeds the dropdown **blank** (`valueExists` in
  `_buildUserDropDown`). Without the second report the image would keep that invisible id and
  `_submitQuickScan` would send a user the caller never chose. Safe to re-enter: `onValueChange`
  does not `setState`.
- Both regression paths are pinned by `test/widgets/quick_scan_form_test.dart` ("keeps the paid-by
  and status prefills through the first group selection" / "drops a prefilled paid-by who is not a
  member of the group just picked"). That harness's `onFormChangeCallback` **mirrors the real
  consumer** — with an inert `(_) {}` callback the image is never mutated and both tests pass with
  the bug present.
- **A value the user SELECTED never survives a group that hides its field — only an untouched
  prefill does.** The distinction is easy to get backwards, and was raised in review as a suspected
  bug. It holds for two reasons that are worth keeping together: the dropdown's `onChanged` clears
  paid-by/categories/tags *before* its own `setState`, so the clear lands while the **old** group's
  fields are still mounted and registered; and the merge above only falls back to the image for a key
  that is **absent**, which — with `clearValueOnUnregister: false` — means "never shown", not
  "hidden later". A field shown under group A and hidden under B keeps its (cleared) key, so a later
  group C that shows it again gets a blank field, while a prefill that was never rendered is still
  offered. Pinned by the A→B→C pair "a selected paid-by does not come back after a group that hides
  it" / "a prefilled paid-by DOES survive a group that hides it".

**The comment field is gated on `group.comments.create` as well as the group config.** The permission
is a **required named argument** to the resolver (not read from a provider) so the helper stays pure
and both call sites are forced at compile time to supply it; each passes the existing
`canCommentCreate(permissionsModel, groupId)` from `lib/shared/functions/permissions.dart`. Without
the permission the field is hidden **and not required** — a member who cannot comment must never be
locked out of quick scan by a comment they can't fill (the backend skips the check for them too, and
silently drops a comment sent anyway). `hideComments` on the group overrides the toggle the same way.
So **`QuickScanForm` now needs a `PermissionsModel` in its provider tree** — a widget test that omits
it throws `ProviderNotFoundException`.

`QuickScanForm.onFormChangeCallback` takes a **record** (`QuickScanFormValues`), not positional
arguments: the list opened with two `int?`s and only grows, so naming each value keeps the carousel
callback (`lib/receipts/widgets/quick_scan.dart`) from transposing them. Note the comment is
deliberately **not** cleared when the group changes (unlike paid-by/categories/tags, it isn't
group-scoped data, and the submit re-resolves visibility per group so a hidden one is never sent).

**Each slide's `QuickScanForm` is keyed by the image, not by its position.** `_QuickScanForm` seeds
its `groupId` State field once in `initState` and reads both `GroupModel` and `PermissionsModel` with
`listen: false`, so nothing re-derives it later. Deleting a page (`_getDeleteIcon` in
`lib/shared/functions/quick_scan.dart`) shifts every later image down an index — without a key Flutter
matches by position and hands the surviving form the **deleted** page's State, leaving `groupId`
disagreeing with the group its own dropdown displays, and the field set resolved against the wrong
group. The `FormBuilder` subtree does reset (its `key` is a per-image `GlobalKey`), so the mismatch is
invisible. Hence `key: ObjectKey(image)` on the form in `lib/receipts/widgets/quick_scan.dart`.
Desktop cannot have this bug: its `showComment(i)` re-reads `groupIds.at(i).value` on every
change-detection pass rather than caching a per-index copy.

**The sheet's layout is what makes a configured field usable — one vertical scrollable per
slide.** The Quick Scan sheet opens through `showFullscreenBottomSheet` with **`bodyFillsSheet:
true`** (`lib/utils/bottom_sheet.dart`), which does two things `QuickScan` depends on: it does **not**
wrap the body in a `SingleChildScrollView`, and it pins the submit button as the scaffold's
**bottom bar** rather than its `bottomSheet`. Both matter:

- The wrapper would make the body's height **unbounded**, and the horizontal `InfiniteCarousel` needs
  a bounded height. That is why `QuickScan.build()` used to hard-code
  `SizedBox(height: MediaQuery.of(context).size.height)` — the **screen** height, which is larger than
  the sheet's body by the app bar, drag handle and safe area. The slide overflowed the body, and
  because each slide's own `SingleChildScrollView` then believed it had a full-screen viewport it had
  nothing to scroll, so the overflow was **clipped and unreachable**. It is now `SizedBox.expand`,
  sized by the bounded constraints the scaffold body supplies.
- A `Scaffold.bottomSheet` **floats over** the body; a bottom bar reserves its space. Floating buried
  the form's last field under the submit button and the "enter details manually" link, which no amount
  of scrolling could clear (`submitButtonSpacing`'s fixed 70px is not the height of that Column).
- The bar slot is the one `Scaffold` does **not** lift over the keyboard, so that trade cost the
  submit button at exactly the moment the user reached for it — the form's last field is the
  multi-line comment, so the keyboard is up by then. Fixed in `ScreenWrapper`; see "A pinned bottom
  bar is not keyboard-safe" above.

Together these shipped a Quick Scan whose **configured comment field could never be seen** — the
comment is last in the form column and the image preview alone claims half the screen height
(`getImagePreviewHeight`). Note the preview is an `InteractiveViewer`, so it claims pans over itself:
a drag started on the image will never scroll the form.

**Testing this needs geometry, not finders.** Three things all quietly pass on a broken sheet:
`expect(..., findsOneWidget)` is true for a widget clipped outside its viewport; `Finder.hitTestable`
reports **0 under `IntegrationTestWidgetsFlutterBinding` on desktop even for plainly visible
widgets**, so it cannot be used to assert visibility here; and `tester.ensureVisible` walks *every*
ancestor scrollable, dragging a field into view even when the user could not (which is exactly how
this shipped — `quick_scan_submit_test.dart` reaches its comment that way). `binding.setSurfaceSize`
is additionally a **no-op on the Linux desktop target**, so a "phone-sized" spec is not what it looks
like; the runner's window is 1280x720 and that is already short enough to expose the bug.
`expectQuickScanFieldOnScreen` (`integration_test/helpers/quick_scan_actions.dart`) is the assertion
to use: it scrolls the field's **own** scroll view by the least amount that would bring the field
fully into the viewport, clamped to what that view can scroll, then checks the field's rect lies
inside both the window and that viewport. One rule covers the column — a mid-column field
(Categories, Tags) scrolls just far enough and is never pushed off the *top*, while the last field
needs more than `maxScrollExtent`, so the clamp leaves it as far down as a user could get it and the
window check still fails if it is unreachable there. Covered by the two
"actually on screen" cases in `quick_scan_config_response_test.dart`.

**Category/Tag picker needs a fallback context in Quick Scan.** `CategorySelectField` /
`TagSelectField` open their multi-select via `showMultiselectBottomSheet(...)`, which calls
`Navigator.of(context)`. They must resolve that context through
`ContextModel.resolveSheetContext(context)` (prefers the mounted shell context, else the field's own
context) — **not** the raw `ContextModel.shellContext`. The shell context is only set by
`receipt_form_screen.dart`; the Quick Scan flow opens straight from the bottom-nav Add menu and never
mounts it, so a raw `shellContext` is **null** and tapping Categories/Tags would crash with
`Navigator.of(null)`. (`receipt_form.dart` uses the same helper for its quick-actions sheet.) Guarded
by `test/widgets/quick_scan_form_test.dart` (tap-opens-picker, shellContext null) and on-device by
`quick_scan_submit_test.dart`.

### A pinned bottom bar is not keyboard-safe — `Scaffold` only lifts `bottomSheet`

`Scaffold` treats its two bottom slots differently, and the difference is invisible until a
software keyboard opens:

| slot | placed at | keyboard-safe? |
|---|---|---|
| `bottomSheet` | `contentBottom - height`, where `contentBottom = height - max(minInsets.bottom, bottomWidgetsHeight)` (`scaffold.dart:1088`, `:1154`) | **yes** — `minInsets.bottom` *is* `viewInsets.bottom` |
| `bottomNavigationBar` | `max(0, height - bottomWidgetsHeight)` — no inset term at all (`scaffold.dart:1054`) | **no** — pinned to the physical bottom |

So `showFullscreenBottomSheet`'s `bodyFillsSheet: true`, which moves the caller's button into the
bar slot precisely *because* a bar reserves its space rather than floating over the form's last
field, silently traded one bug for another: the Quick Scan submit button sat **under** the keyboard
raised by the comment field directly above it.

`ScreenWrapper._liftAboveKeyboard` closes it by padding the bar slot by
`MediaQuery.viewInsetsOf(context).bottom`. Four things about that are load-bearing:

- **The value is not computed.** `MediaQueryData.viewInsets`' own doc: *"When a mobile device's
  keyboard is visible `viewInsets.bottom` corresponds to the top of the keyboard."* The platform
  reports it (Android `WindowInsets` IME, iOS `UIKeyboardWillChangeFrame`) in logical pixels and
  updates it **every frame** while the keyboard slides, so the bar follows exactly — split,
  floating and hardware keyboards included. Never substitute a measured or hardcoded height.
- **The read must happen inside the bar slot,** which is why a `Builder` wraps it. Widgets are
  inflated where they are *mounted*, not where they are constructed, so the `Builder`'s context
  lands in the slot's own `MediaQuery`. `Scaffold._addIfNonNull` (`scaffold.dart:2917-2925`) passes
  `removeBottomInset: true` for the **body** slot (`:3036`) and not for this one (`:3166`), so a
  read from inside the body yields **0** and the padding silently does nothing.
- **It cannot double-count against the `SafeArea`** wrapping the scaffold. `SafeArea` consumes
  `MediaQuery.paddingOf` (`safe_area.dart:115`), and `padding` is `viewPadding - viewInsets` floored
  at 0 — so `padding.bottom` is 0 for exactly as long as the keyboard is up. The two are
  complementary, never additive.
- **Plain `Padding`, never `AnimatedPadding`.** The scaffold body resizes un-animated off the same
  value, so a second interpolation here would make the bar trail the body's edge for the whole
  slide, opening and closing a gap.

**The `bottomSheet` slot is deliberately left alone** — padding it too would double-count. That is
what keeps the receipt form (`receipt_form_screen.dart:133`), the quick-actions split sheet
(`receipt_form.dart:720`) and the comment screen (`receipt_comment_screen.dart:60`) unchanged; all
three were already safe, and the receipt form in particular is a common wrong guess.

**The trade runs both ways, and the other direction is its own bug.** `bottomSheet` is lifted
over the keyboard but **floats over** the body; `bottomNavigationBar` reserves its space but had to
be lifted by hand. So a screen using the floating slot has to reserve the button's height *itself*,
or its tail is unreachable: `SingleChildScrollView` at `maxScrollExtent` only brings content flush
with the viewport's bottom edge, and the button covers the last 50px at **every** scroll offset.
`submitButtonSpacing` (`constants/spacing.dart`, 70px) is what reserves it, and it belongs at the
end of any column sitting under a floating `BottomSubmitButton` — today
`receipt_form.dart`, `quick_actions.dart` and `quick_scan_form.dart`.

- **A `kDebugMode` widget hid this for as long as it existed.** The receipt form's column used to end
  with a debug-only "Check form value" `ElevatedButton`. It contributed ~48px in debug and collapsed
  to `SizedBox.shrink()` in release, so the debug tree cleared the button (by 18px) while shipped
  builds buried the last ~30px of the "Shared With" field — invisible to `flutter test` and every
  `integration_test/` spec. It was also drawn *entirely underneath* the submit button, so it could
  not be tapped. It was deleted rather than kept: the whole point is that debug and release lay out
  the same, which is what makes the bug testable at all. **Do not reintroduce a debug-only widget at
  the bottom of a scrollable form.**
- **Testing it needs a viewport short enough to overflow.** Below the overflow threshold
  `maxScrollExtent` is 0 and nothing can be buried, so the assertion passes on the broken tree for
  the wrong reason — the same vacuity trap as the keyboard-down control above. The quick-actions
  sheet's default content is under 200px, hence the deliberately tiny surface in its test.
- Covered by `test/widgets/receipt_form_submit_button_clearance_test.dart` and
  `test/widgets/quick_actions_submit_button_clearance_test.dart`, both of which drive
  `pumpReceiptForm(..., pinnedSubmitButton: true)` or the real sheet — a bare `Scaffold` has no
  floating button to be buried under. Both fail with the spacer removed; verify that when changing
  them.
- **Demo:** `tool/receipt-form-clearance.png` — the same form at max scroll with and without the
  spacer, with a marker at the button's top edge. Regenerate with `./tool/record_clearance_shot.sh`.

**`/search` restacks, on purpose.** It is the one screen that fills *both* slots
(`main.dart:234`/`:238`): the nav bar in `bottomNavigationBar`, `WranglerSearchBar` in `bottomSheet`.
Lifting the bar grows `bottomWidgetsHeight`, which lowers `contentBottom`, which moves the search
field **above the nav bar** instead of directly onto the keyboard. Before: `[body][search][keyboard]`
with the nav hidden. After: `[body][search][nav][keyboard]`. This was chosen over scoping the fix to
the two sheets, so that no future bottom bar has to rediscover the problem.

**Testing it needs geometry, and a control.** `findsOneWidget` is true for a button parked
off-screen under the keyboard, so only a rect distinguishes the fix from the bug. The recipe:

- `tester.view.viewInsets = FakeViewPadding(bottom: n)` — the same engine field the OS keyboard
  sets, so `Scaffold` **and** `EditableText` (which reads `View.of`, not `MediaQuery`) both react.
- `FakeViewPadding` is in **physical** pixels. Set `devicePixelRatio = 1.0` (as
  `filter_multiselect_test.dart` does) and they are the logical pixels the rects come back in.
- `addTearDown(tester.view.reset)` — `postTest` does **not** reset view values, so a leaked inset
  follows every later case in the file.
- Always assert a **keyboard-down control** too. Without it the guard also passes on a sheet that
  was never full height — a button at y=400 is "above the keyboard" for the wrong reason.

Covered by `test/widgets/screen_wrapper_keyboard_test.dart` (the slot contract, including the
both-slots `/search` shape and the bar returning on dismiss), plus one real-surface case in each of
`test/widgets/quick_scan_sheet_test.dart` and `test/widgets/filter_multiselect_test.dart`. Every
keyboard case fails pre-fix reading the full screen height instead of the keyboard top; verify that
when changing them. **Demo:** `tool/quick-scan-keyboard.gif`, `tool/multi-select-keyboard.gif` and
`tool/search-nav-keyboard.gif` — see "Recording a demo GIF" below.

### Receipt status presentation

`receiptStatusLabel` / `receiptStatusColor` (`lib/utils/receipts.dart`) are the **single owner** of
receipt-status presentation; `ReceiptListItem.getStatusText` / `getStatusColor` are thin delegates,
and `buildReceiptStatusDropDownMenuItems` (`lib/utils/forms.dart`) carries the same labels for the
form dropdown — the two label lists must stay in sync.

Every status is a **pale tint**, because `ListItemTrailingStatus` paints its label in the default
dark `onBackground` and `ListItemColorBlock` sits beside dark text. The tints match the desktop chip
(`desktop/CLAUDE.md` → "Receipt status colors"):

| Status | Tint |
|---|---|
| `NEEDS_ATTENTION` | `warningAmber` `#FFE0B2` |
| `DECLINED` | `errorRed` `#F2BFBF` — the red NEEDS_ATTENTION gave up |
| `OPEN` | `#FFFACD` |
| `RESOLVED` | `successGreen` |
| `DRAFT` (and the fallback) | `neutralStatusGrey` |

**`errorRed` / `successGreen` are shared with the activity list**
(`group_activity_list_item.dart`), where red still means a failed system task — so DECLINED reusing
`errorRed` keeps both meanings intact, and changing either constant moves both surfaces.

**Neither function throws on an unrecognized status.** Both used to end in
`throw Exception("Unknown status: …")`, which was already reachable via `ReceiptStatus.empty` (the
search screen builds a `Receipt` from a `SearchResult` whose `receiptStatus` is nullable) and would
take the whole receipt list down for any status the API adds after a build ships. The label falls
back to `titleCaseStatusName` (SNAKE_CASE → Title Case, matching desktop's `formatStatus`) and the
tint to the neutral grey.

These two are the **presentation** half of the tolerance; they are only reachable because the
**deserialization** half exists — `ReceiptStatus.empty` is annotated `fallback: true` so the
generated `_$valueOf` returns it instead of throwing (see "Regenerating API Client Models" → the
third hand-patch). Without that the payload fails at the serializer and neither function is ever
called. Neither half rescues a binary released *before* both landed; together they mean a build from
here on degrades to a blank neutral chip rather than losing the whole receipts list.

`titleCaseStatusName` is split out because a `ReceiptStatus` the generated enum does not declare
**cannot be constructed from a test** (the constructor is private to the openapi package), so the
fallback rule is otherwise unpinnable. Tests:
`test/utils/receipt_status_presentation_test.dart` (both maps, exhaustive over
`ReceiptStatus.values`, plus the fallback rule); the mapping is deliberately **not** asserted through
`ReceiptListItem`, whose trailing pill is a fixed 100px wide — "Needs Attention" reports a render
overflow that has nothing to do with the mapping. **E2E:**
`integration_test/receipt_status_lifecycle_test.dart` drives every status through the real API.

### Seeding the receipt group

The Group field is a default, not a lock. The add form seeds it from **the group being browsed**, and
failing that from **the user's only group**; Quick Scan seeds each picked image from the user's
`quickScanDefaultGroupId` and, failing that, from their only group. See the root `CLAUDE.md` →
"Seeding the Group Field" for the cross-client contract.

- **`GroupModel.soleGroupId`** is built off `groupsWithoutAllGroup`, and
  `buildGroupDropDownMenuItems` (`lib/utils/forms.dart`) now sources that same getter rather than
  re-filtering `groups` inline. That is not cosmetic: it is what guarantees the seed is always one of
  the dropdown's items, and `DropdownButton` asserts on an `initialValue` that is not.
- **`_resolveInitialGroupId()`** (`lib/receipts/widgets/receipt_form.dart`) is the single rule, pure
  and stable across rebuilds, called from both `buildGroupField()` (for `initialValue`, `0` → `null`)
  and the `initState` post-frame callback. It **must not** be called from `initState` itself —
  `getFormStateFromContext` and `getGroupId` both do `GoRouterState.of(context)`, an inherited-widget
  lookup that is illegal there. That is also why the post-frame callback is now registered
  **unconditionally**: the resolution can produce a group the receipt does not carry.
- **Seeding the dropdown is not enough.** Paid-by's item list, the category/tag `Visibility` gates and
  the add-share "+" all read the `groupId` **State** field, so the callback mirrors the resolved id
  into it with `setState` — after `_applyGroupDefaultCustomFields`, which writes to `ReceiptModel`
  and must not fire `notifyListeners()` from inside a `setState` callback. Cost: the group-derived
  parts render one frame late.
- **The route group comes from `extra`.** `openManualReceipt` (`lib/shared/functions/receipt_entry.dart`)
  already put the id there; the form simply discarded it before. The synthetic "All" group is rejected
  explicitly, mirroring desktop's `initForm`. `getGroupId` now pattern-matches `extra` instead of
  hard-casting it, since it is reached from a build method on every frame.
- **`_getInitialQuickScanValues`** (`lib/shared/functions/quick_scan.dart`) tests
  `quickScanDefaultGroupId > 0` rather than null-coalescing — the generated model **defaults it to
  `0`**, so "unset" never arrives as null (see "Regenerating API Client Models").
- **`selectDropdown` had to be hardened** (`integration_test/helpers/form_actions.dart`): its
  `pumpUntilFound(find.text(option))` wait assumed the option text only appears once the menu is
  open, which stops holding when the dropdown already *shows* that value -- the wait then returns on
  the first poll and `.last` can resolve the closed-state child. It now drains the open animation
  afterwards. (Waiting for the match count to *grow* instead looks tighter but is wrong: the same
  display name can already be on screen in another dropdown, and `receipt_add_share_test` deadlocks
  on it.)
- **Tests.** `test/models/group_model_test.dart` (the rule), `test/widgets/receipt_form_group_prefill_test.dart`
  (seeding *and* the State-field mirror — asserted through the category field, the add-share button
  and paid-by's items, since the form value alone would pass without it),
  `test/shared/functions/quick_scan_initial_values_test.dart` (preference-beats-sole-group). **E2E:**
  `integration_test/receipt_single_group_default_test.dart` — one case per rule, each on an account
  chosen so only that rule can explain the seed (`provisionPermUser()` with no role → one group;
  with `roleName` → two).

### Group default custom fields

A group can declare custom fields that are pre-added to its receipts
(`GroupReceiptSettings.defaultCustomFieldIds`, configured on desktop's Group Receipt Settings — see
the root `CLAUDE.md` → "Group Default Custom Fields" for the cross-component contract). The receipt
form (`lib/receipts/widgets/receipt_form.dart`) applies the selected group's set on **load** and on
every **active group change**, because each group is effectively its own receipt template.

- **The swap is conservative.** `_applyGroupDefaultCustomFields(newGroupId)` only takes back a field
  **this form added itself** (`_autoAppliedCustomFieldIds`) that is **still empty**. A default the
  user typed into is kept and handed over to them (dropped from the auto set); so is anything they
  added by hand. Adding or removing a field by hand (`_addCustomField` / `_removeCustomField`) removes
  its id from the auto set, so the swap never fights a decision the user made.
- **It runs on load in EVERY form state, and on every active group change.** A group's defaults are
  meant to read as its built-in receipt fields, so a receipt saved before the group was configured
  picks them up too — read-only in view (`CustomFieldWidget` gets `onRemove: null` and there is no
  Add button), and persisted as empty attached values once an edit is saved. `initState`'s post-frame
  callback runs `_applyGroupDefaultCustomFields` whenever `_resolveInitialGroupId` yields a group;
  the dropdown's `onChanged` runs it **before** its `setState`, because it writes to `ReceiptModel`
  and `notifyListeners()` must not fire from inside a setState callback.
- **Provenance lives on `ReceiptModel`, not the form's `State`.**
  `ReceiptModel.autoAppliedCustomFieldIds` is the set the form adds to when it attaches a default and
  removes from when the user touches one. It has to outlive the screen: the app bar menu's Edit entry
  is a plain `go` (`receipt_app_bar_action_builder.dart`) and `receipt_form_screen.dart` re-hydrates
  only for a *different* receipt id, so **view → edit remounts the form against the same
  `modifiedReceipt`** with a fresh `State`. Two sequences end with the same empty field attached and
  absent from the saved receipt, and only the set tells them apart: one the *form* added in the view
  mount (still the swap's, dropped on a later group change) and one the *user* removed and re-added
  by hand (theirs, kept). An earlier attempt inferred this on load from "every default the saved
  receipt does not carry", which reclaimed the second case and silently dropped it — see
  [receipt-wrangler#689](https://github.com/Receipt-Wrangler/receipt-wrangler/pull/689).
  - The set is cleared wherever the working copy is replaced — `setReceipt` (which also covers the
    post-save re-hydration, by which point the fields really are the receipt's own data) and
    `resetModel` (the back arrow, and the `/receipts/add` route redirect in `main.dart`) — so
    provenance and `modifiedReceipt` can never disagree. It deliberately does **not** notify:
    nothing renders from it.
  - The removal pass is inert on a *first* load, since the set is empty until something attaches a
    default.
- **One model write per swap.** `_customFieldValuesWith(add:, remove:)` is pure and every caller hands
  the whole result to a single `_setCustomFieldValues`, so a multi-field swap rebuilds the form once
  instead of once per field, with its fields half-mounted in between.
- **Each row is keyed `ValueKey("customFieldValue_<id>")`** so the element tree follows field identity,
  not list position. Without the key, removing a field re-uses its element for whatever shifted into the
  slot and — since `FormBuilderTextField` has no `didUpdateWidget` — the old `TextEditingController` text
  shows under the new field's label.
- **Removal clears the form value first.** FormBuilder's `clearValueOnUnregister` defaults to **false**,
  so an unregistered field's value stays in the form's value map and is handed straight back to any field
  that later re-registers under the same name; `_clearCustomFieldFormValue` (`didChange(null)`) is what
  stops a re-added default from resurrecting what was typed into it.
- **Ids missing from the `CustomFieldModel` catalog are skipped**, which is also the permission gate: the
  catalog is empty for a caller without `app.custom-fields.read` (the 403 is swallowed into an empty
  list), and auto-adding for them would make their receipts unsaveable — the backend's
  `enforceReceiptCustomFieldSelection` **403s** any save that changes the attached id set for such a
  caller. There is no separate `PermissionsModel` check; the empty catalog is the gate.
- **An empty attached value is still submitted.** `buildCustomFieldValueUpsertCommands` emits one entry
  per attached value with null columns for the empty ones — dropping them would fail that same 403,
  because the backend REPLACES the whole association on update.

**Tests.** `test/widgets/receipt_form_default_custom_fields_test.dart` covers the rules exhaustively
against injected models (18 cases: seeding, each keep/drop rule, A→B→A leaving no residue, an unchecked
BOOLEAN counting as empty, a missing/empty catalog, the load-time apply in view and edit, no duplicate
for a default the receipt already carries, and the two remount cases the provenance set exists for —
the view → edit carry-over staying swappable, and a hand-re-added default staying the user's). Those
last two pass `pumpReceiptForm`'s `receiptModel:` seam a model an earlier pump returned, so the
second pump is a **real remount** (fresh `State`, same working copy and provenance) rather than a
simulated one. **E2E:**
`integration_test/receipt_default_custom_fields_test.dart` — three specs proving the ids survive the
wire (persisted, hydrated onto `GroupModel` via AppData at login, applied by the real form, accepted on
save): the full swap matrix ending in an API read-back of the saved values, the stale-value guard (type
into a default → remove it by hand → switch away and back → it returns **blank**), and a receipt seeded
server-side with no custom fields showing its group's defaults in view and edit and still swapping on a
group change. Every step also asserts no UI error survived it.

Three things that spec encodes, all of which cost a debugging cycle:
- **Do not install a `FlutterError.onError` collector to catch UI errors.** It takes the handler away
  from `TestWidgetsFlutterBinding`, whose `_runTest` then trips its own
  `'_pendingExceptionDetails != null'` assertion on **any** failure — so a plain `expect` mismatch is
  reported as an unreadable framework assertion instead of its own message. The binding already
  hard-fails on an uncaught framework error; `expect(tester.takeException(), isNull)` plus
  `expect(find.byType(ErrorWidget), findsNothing)` per step adds the step name and the red-screen case.
- **`pumpAndSettle`'s first positional argument is the frame interval, not a timeout** (the timeout
  defaults to **10 minutes**). The receipt form hosts a `CircularLoadingProgress` that spins while
  `customFieldModel` reloads, so a tree that never settles hangs the whole run — use
  `pumpUntilFound(tester, find.byType(BottomSubmitButton).hitTestable())` instead.
- **Provision a user in exactly the groups the spec switches between**, rather than logging in as the
  shared admin. The group dropdown is a Material menu whose scrollable does not build off-screen items
  into the element tree, so the admin's accumulated groups can push the target below the viewport where
  `find.text` can't reach it (the same problem `keepOnlyGroup` works around for Quick Scan).

### Receipt filtering

The receipts list can be filtered on all ten fields the API supports -- the same set desktop's
advanced-filter dialog drives. **Client-only**: `swagger.yml`, the Go API and `mobile/api/` were
already capable, so nothing there changed.

Entry point is a badged filter action in the receipts app bar (`ReceiptFilterButton`, added to
`GroupAppBar`'s existing route-gated `actions`). It pushes `ReceiptFilterScreen`, which lists the
active conditions as cards, offers "Add filter", and commits with **"Apply Filter"**.

- **`lib/constants/receipt_filter_fields.dart` is the single field table** -- key, label, icon,
  hint and type for the ten fields, plus `filterOperationsByType` and `filterOperationLabels`.
  The add-list, the cards and the editor all read it, so a card can never disagree with the row
  that produced it. Mirrors desktop's `RECEIPT_FILTER_FIELDS` +
  `filter-operations-options.constant.ts`.
  - Labels match `receiptSortOptions` wherever the two overlap -- `date` is **"Receipt Date"**, not
    a bare "Date", because the list can also be filtered on `resolvedDate` and `createdAt`.
  - The operation lists are written out literally rather than derived from `FilterOperation.values`:
    the Dart enum carries an extra `empty` member desktop's does not, and a derived list would leak
    it into the chips.
- **`ReceiptListModel` owns the APPLIED filter; the screen edits a draft.** `ReceiptFilterScreen`
  copies `model.filter` into local `State`, and only "Apply Filter" writes back via `setFilter`.
  Backing out (the X, the system back gesture) therefore discards, and the list never refetches
  mid-authoring.
- **A notification from `ReceiptListModel` means the filter changed.** `GroupReceiptsList` listens
  for it and calls its `_refreshCallback` (plus a `setState`, because the empty-state text reads the
  applied filter). This works because **every sort setter is deliberately called with
  `notify: false`** and refreshes the list directly -- don't "tidy" those call sites into notifying,
  or every sort change refetches twice.
- **The filter is cleared on a group change**, in `didChangeDependencies` (which is where
  `getGroupId`'s `GoRouterState` read is legal). A filter holds the previous group's category, tag
  and user ids, which match nothing in the next one and would leave the badge counting conditions
  the user cannot see. It clears **everything**, not just the id-bearing fields: "switching groups
  shows that group's receipts" is the predictable rule. It clears **silently** (`notify: false`),
  because that runs during a build; the explicit `_refreshCallback` is the visible half.
  - **The scope is read off the filter, not off a group the list remembered** --
    `ReceiptListModel.filterGroupId`, recorded by `setFilter` (whose `groupId` is a *required* named
    argument for exactly this reason). The first version tracked "the last group I saw" in the
    widget, and **that guard never fired**: the app offers no lateral group switch, so every real
    group change leaves the group shell (the app-bar arrow goes to `/groups`, group cards go to
    `/groups/<id>/dashboards`, and `GroupBottomNav` only navigates within the current group), which
    destroys the `State`. The replacement mounted with a null "last group" while the filter sat
    untouched on the app-level model, so **group A's filter followed the user into group B**. It
    shipped that way and `integration_test/receipt_filter_lifecycle_test.dart` is what caught it.
  - **"Clear whenever the list mounts" is not the fix**, which is why the scope has to live on the
    filter: a round trip to a receipt tears the list down and rebuilds it just as a group change
    does (`/receipts/:id/view` is a top-level route). Only the filter knows the difference between
    the two, and both cases are pinned -- in `group_receipts_list_test.dart` for the mount, and
    end to end in the lifecycle spec.

**Encoding (`lib/utils/receipt_filter.dart`) is the highest-risk part, and its rules come from the
Go side (`api/internal/repositories/receipts.go`).**

- The generated `ReceiptPagedRequestFilter`'s ten properties are **`JsonObject?`**, not a typed
  `PagedRequestField`, so each condition is written as
  `JsonObject({"operation": ..., "value": ...})`. `setReceiptFilterField` owns the key-to-slot
  dispatch and is shared with `dashboardConfigurationToFilter` (`lib/utils/receipts.dart`), which
  was missing the `group` arm before it was extracted.
- **An EMPTY field is not an ABSENT one.** `initReceiptFilterValues` coerces a null date value to
  `""` and a null amount to `0`, and the query builder then runs `date = ''` (matches nothing) or
  `amount = 0` (matches the wrong rows) -- both silent. `buildReceiptPagedRequestFilter` therefore
  **skips any condition `isReceiptFilterConditionValid` rejects**. The editor also gates Save, but
  the encoder's contract has to hold on its own.
- **Every value is type-asserted with no comma-ok, so a wrong shape is a 500, not an ignored
  filter.** Amounts go as **numbers** (`AmountField`'s `valueTransformer` yields a *string*, so the
  editor parses it first); the five list fields always as arrays; `status` as wire strings, not
  labels; dates as zulu strings via `formatDate(zuluDateFormat, ...)`, the same call the receipt
  submit uses.
- **A BETWEEN date range spans start-of-day to end-of-day** (`startOfDay` / `endOfDay` in
  `lib/utils/date.dart`). These are datetime columns, so a bare `<= 2026-09-18T00:00:00Z` upper
  bound excludes everything recorded on the last day the user picked.
- **A zero amount is a real filter and is sent.** Nothing defaults to zero and the API applies
  `amount = 0`, matching desktop's `isFilterEntryActive`.
- Conditions hold the **display objects** (`Category`, `Tag`, `Group`, `UserView`, `ReceiptStatus`,
  `DateTime`, `double`, `String`), not ids -- so a captured option keeps rendering its own name even
  after the catalog changes, and the card, the chips and the encoder all read one thing.

**Shared components, not new ones.** `AmountField` for every amount (it gained an optional
`validator`, defaulting to today's required, because a filter amount is only authored when the user
asks for one); `CategorySelectField` / `TagSelectField` unchanged; `MultiSelectField` +
`showMultiselectBottomSheet` for Group, Paid By and Status. `FilterMultiSelect` keeps its **single
call site** inside that helper, so its `Expanded` assumption is untouched. Flutter's own
`showDatePicker` / `showDateRangePicker` cover dates -- the app had no date widget and no range
picker at all.

- `receiptStatusField` is deliberately **not** reused: it is single-select and hard-required, while
  the filter's `status` is `CONTAINS` over a list.
- The operation chips take the **tinted** selected treatment (`primaryContainer` fill,
  `onPrimaryContainer` label, a `#BBE6FF` border) via a local `ChipTheme`, deliberately *not* the
  app-wide filled accent chip. That one means "a value you picked" -- the categories and statuses in
  the field below it are drawn that way -- while these pick a *mode*, and the two rows share one
  sheet. The local override has to restate `labelStyle` for the same Material 3 reason the theme
  does (see "Theme & color roles").

**Two things the group context forces** (`lib/utils/receipt_filter_options.dart`):

- **`Group` is only offered on the synthetic "All" group**, mirroring desktop's `showGroupFilter`.
  Inside a real group the receipts endpoint already scopes every query to it.
- **Categories and tags need no special casing.** The All group is a *real row* the user belongs to
  (`GroupRepository.CreateAllGroup`), so `GetAppData` builds it a `groupCategories` / `groupTags`
  entry like any other group -- the wrappers work with the route's group id. **Paid-by is the
  exception**: the All group's roster is just the caller, so `filterPaidByOptions` unions the real
  groups' rosters there.

**Styling follows the design project's mobile panel** (`Quick Date Filtering.dc.html`, panel 4b),
minus its button gradient. The screen is a slate-50 canvas carrying white condition cards (radius
14, a 6%-black hairline, a 5%-black 1px shadow) under an uppercase condition count; the operation is
a filled `surfaceContainer` pill rather than an outlined chip; "Add filter" is a dashed 52px
placeholder, painted by a private `CustomPainter` because Flutter has no dashed border. Most of the
rest came from filling in the theme's color roles rather than from anything here -- see "Theme &
color roles".

- **The condition card carries no field icon.** The icons belong to the add sheet, where they help
  pick a field; repeating them on the card only competes with the label.
- **It is a hand-built container, not a `Card`.** Material's `Card` is the elevated one, whose
  tinted surface and shadow are far heavier than this near-flat row. Its fill has to be on the
  `BoxDecoration`, not only on the `Material` under it: a `BoxDecoration` paints its `boxShadow` as
  a silhouette of the whole shape, so without a colour there the shadow shows through the card's
  interior and greys it out.
- **The remove X sits inside the card's own `InkWell`**, so it has to win the gesture arena against
  it. `receipt_filter_condition_card_test.dart` pins that removing a condition does not also reopen
  the editor for the field just dropped.
- **The count header is rendered only when there is something to count.** The design shows it *and*
  the empty state at zero, which says the same thing twice.

**The filter screen is a pushed route, not a bottom sheet.** It keeps the modal stack at the depth
the app already ships (editor sheet -> picker sheet, the same as Quick Scan -> category picker), and
a real `Scaffold` puts "Apply Filter" in `bottomNavigationBar`, which reserves its space, rather than
`Scaffold.bottomSheet`, which floats over the last card.

**A latent paging bug this feature exposed.** `PagedDataList` never reset `_totalCount` on refresh,
and `getNextPageKey` stops paging once the loaded items reach it. Sorting cannot reach a zero total,
but filtering to no matches can -- after which `0 >= 0` stayed true and **no page was ever requested
again**, leaving the list permanently empty even once the filter was cleared. Fixed by nulling
`_totalCount` in the refresh callback; `test/widgets/paged_data_list_test.dart` was verified to fail
without it.

**Tests:** `test/constants/receipt_filter_fields_test.dart` (the table, including that every key
names a real wire field -- `setReceiptFilterField` dispatches on a string, so a typo would not fail
to compile), `test/utils/receipt_filter_test.dart` (a case per field type x operation, asserted
through the real serializer), `test/utils/receipt_filter_options_test.dart`,
`test/models/receipt_list_model_test.dart`, and widget tests for the button, the screen, the add
sheet, the editor, the condition card and the list wiring.

**E2e: `integration_test/receipt_filter_test.dart` and `receipt_filter_lifecycle_test.dart`.** The
unit suite proves the *serialized map* through the real serializer and the widget suite runs against
a mocked `ReceiptApi`, so neither can reach the thing most likely to break -- that the Go query
builder accepts the filter and narrows real rows. Five tests drive the wire (name, status via the
real multiselect sheet, categories from the real per-group catalog, a date range, an amount typed
into the currency field) and two cover the lifecycle (survives a receipt round trip; cleared on a
group change). Shared drivers live in `integration_test/helpers/receipt_filter_actions.dart`.

- **The seed set is three receipts in one fixture group**, differing in name, amount, date, status
  and category so each test gets one match and two non-matches. **R2's `18:30` timestamp is
  load-bearing**: a `BETWEEN` of 06/11-06/13 encodes to an upper bound of `06-13T23:59:59Z` because
  the client expands the range to end-of-day, and a receipt stored at *midnight* on the last day
  would match even a broken `<= 06-13T00:00:00Z` bound -- so the date test would pass against the
  bug it exists to catch.
- **The date picker is driven through its text-entry mode**, not the calendar. `firstDate` is 2000
  and `lastDate` five years out, so the grid is ~370 lazily-built months whose day cells are bare
  unkeyed `Text('11')` repeating every month. The app registers no `flutter_localizations`
  delegates, so `DefaultMaterialLocalizations` is in force: the toggle is tooltipped
  **"Switch to input"**, dates parse as US `mm/dd/yyyy`, and the confirm button is **"OK"** in text
  mode (it is "Save" in calendar mode). The spec asserts the picked range on the `_TapField` before
  applying, so a later row failure means the *server* disagreed rather than the picker misfiring.
- **Wait for a button to be *enabled*, not hittable.** A disabled `FilledButton` still hit-tests, so
  a tap on "Apply Filter" (gated on `LoadingModel.isLoading`) or "Save condition" (gated on the
  editor's validity) silently no-ops and fails ten seconds later pointing at the wrong widget.
- **Never `find.byKey` an amount field.** `AmountField` forwards its `widget.key` onto the
  `FormBuilderTextField` it builds, so the keyed finder matches two widgets and any tap throws. Use
  `formField("value")` -- and `CurrencyTextFieldController` reads keystrokes as cents, so type the
  full `50.00`, not `50`.

### Quick date filter (month stepper + date field picker)

Above the receipts list, a month stepper and a chip naming the date field it writes to -- the mobile
half of the design project's panel **4b**, and a port of desktop's `app-month-stepper` +
`receipts-quick-date-field` (`desktop/CLAUDE.md` -> "The month stepper targets one date field").
**Client-only**, like the filter screen below it: a month is `BETWEEN [first day, last day]`, which
the API and the encoder already handle.

```
◄   📅 September 2026   ✕   ►
[📅 On Receipt Date] [Added At] [Descending]
```

- **`lib/utils/receipt_date_filter.dart` owns the month <-> condition translation** --
  `monthFilterCondition`, `monthFromCondition`, `shiftMonth`, `monthOfDate`, `filterMonthLabel`.
  It hands the encoder plain month boundaries (`DateTime(y, m, 1)` / `DateTime(y, m + 1, 0)`) rather
  than day-bounded instants, because `buildReceiptPagedRequestFilter` already applies
  `startOfDay`/`endOfDay` to a `BETWEEN` pair -- duplicating that here would give the rule two homes.
  Day `0` of the next month and month `13` both normalize in Dart, so **February and December need no
  special case**.
- **`FilterMonth.month` is 1-based**, unlike desktop's zero-based `FilterMonth` (which is built on
  JavaScript's `Date`). `DateTime.month` is 1-based, so a faithful port of the off-by-one would fight
  every constructor.
- **`monthFromCondition` takes `DateTime`s only**, where desktop's `monthFromFilterEntry` also accepts
  ISO strings. Desktop persists its filter to localStorage and NGXS serializes through JSON;
  `ReceiptListModel` is in-memory, so the string branch would be dead code. It still matches on
  **calendar fields**, which is what lets it name a month the advanced filter's `showDateRangePicker`
  authored (both bounds at local midnight).
- **`ReceiptListModel` owns which field the control points at** (`quickDateField`,
  `setQuickDateField`, `quickDateCondition`), not `GroupReceiptsList`. That widget is torn down and
  rebuilt on almost every navigation, including a round trip to a receipt -- a widget-local field
  would snap back to `date` while the month it wrote still sat on `resolvedDate`, leaving the stepper
  describing a condition it does not own. Same reasoning as `filterGroupId`.
- **`setFilterField(key, condition, groupId:)`** is the single-field write path (the analogue of
  desktop's `SetReceiptFilterField`). It **notifies**, so the list refetches once through
  `_refreshForFilterChange`; the filter screen keeps committing whole drafts through `setFilter`.
- **Switching the date field must NOT refetch**, which is why the chip calls
  `setQuickDateField(key, false)` and repaints with a bare `setState`. It changes no condition, only
  which one the stepper describes, so the result set has not moved -- and a notification from this
  model means "the filter changed". The abandoned condition stays applied and stays on the filter
  screen. `group_receipts_list_test.dart` asserts **zero** requests for it; verified to fail with the
  flag flipped to `true`.
- **The chosen field is group-scoped exactly as the conditions are**, and it takes *both* halves to
  hold. `clearFilter`'s early return was widened to cover the field
  (`_filter.isEmpty && _quickDateField == defaultQuickDateFieldKey`), **and** `setQuickDateField`
  takes a **required `groupId`** and records it -- because `_filterGroupId` is what
  `didChangeDependencies` compares, and it is the only thing that gets `clearFilter` *called*.
  Widening the guard alone shipped a half-rule: with a condition applied the field reset, and with
  none applied `_filterGroupId` stayed null, the reset never ran, and the field followed the user
  into the next group (caught in review on #702). `_filterGroupId` is therefore recomputed on every
  write from one predicate, `_hasGroupScopedState` (`_filter.isNotEmpty || field != default`), so a
  field toggled away from the default and back leaves nothing scoped behind. Same shape as the
  filter leak `receipt_filter_lifecycle_test.dart` exists for, and
  `group_receipts_list_test.dart` pins the field-only navigation case separately from the
  with-a-condition one.
- **The label reads "September 2026" / "Custom" / "All time"**, and the stepper never hides. A
  condition it cannot describe (a partial range, a `GREATER_THAN`, `WITHIN_CURRENT_MONTH`) still has
  to read as a filter. **The clear ✕ follows the *condition*, not the month** -- "Custom" carries no
  month but is exactly the state a user most needs a way out of.
- **The arrows seed from today when no month is showing**, then apply the delta, so `◄` and `►` never
  do the same thing. Stepping off "Custom" overwrites that condition with a whole month.
- **The field chip is labelled "On Receipt Date", not "Receipt Date".** `receiptSortOptions` names
  the same three columns identically, so an unprefixed chip sits beside a sort chip reading exactly
  the same words while meaning something else. Desktop needs no prefix because its picker abuts the
  stepper and reads "September 2026 ... on Receipt Date" in one line; the two stacked rows here
  cannot. Pinned by a test that sorts by Receipt Date and asserts both chips.
- **`buildSortFilterBar()` is horizontally scrollable** now that it carries a third chip -- a bare
  `Row` overflows a narrow phone rather than truncating.
- **The month sheet is `isScrollControlled`.** A default `showModalBottomSheet` caps itself at 9/16 of
  the screen, which the year pager plus the grid plus the shortcuts exceeds on a short phone -- and a
  bottom sheet **overflows rather than scrolling**. It is also a plain compact sheet rather than
  `showFullscreenBottomSheet`, which mounts a `TopAppBar` over a full-height body for children that
  scroll themselves. Opened through `ContextModel.resolveSheetContext`, like every other sheet.
- **Paging the year selects nothing and leaves the sheet open** -- it is a view concern. Bounds are
  1970 through five years out, matching desktop's `MIN_YEAR` / `MAX_YEAR_OFFSET`.
- **`receiptDateFilterFields`** is derived from `receiptFilterFields` by `type`, so the picker, the
  condition card and the "Add filter" sheet cannot name a field differently. Desktop narrows on the
  *key* because TypeScript needs the runtime check to justify its narrower type; Dart does not, and
  filtering on the type means a new date field reaches the picker with no second edit.

**Demo:** `tool/quick-date-filter.gif`, regenerated by `tool/record_quick_date_demo.sh` (see
"Recording a demo GIF" below -- route 2, single panel).

**Tests:** `test/utils/receipt_date_filter_test.dart` (boundaries incl. leap February and the
December roll, the null cases, and the whole thing through the **real** encoder -- that last is what
proves a month reaches the server as `…-01T00:00:00Z` -> `…-<last>T23:59:59Z`),
`test/widgets/receipt_month_stepper_test.dart`, `test/widgets/receipt_month_picker_sheet_test.dart`,
plus new cases in `receipt_filter_fields_test.dart`, `receipt_list_model_test.dart` and
`group_receipts_list_test.dart`.

**E2e: `integration_test/receipt_quick_date_filter_test.dart`.** It **seeds relative to
`DateTime.now()`**, which `receipt_filter_test.dart` explicitly requires of any month-relative case --
that suite is pinned to June 2026 and would drift out of the stepper's reach. Two receipts: one dated
the 15th of last month with status **RESOLVED**, one dated the 15th of this month. The RESOLVED one
is what makes the field picker testable at all: the API stamps `resolved_date` with `time.Now()` on
create, so its two date columns fall in different months, and filtering "this month" returns the
*other* row on `resolvedDate` than on `date`. Nothing but a real column change on the wire produces
that.

### App time zone (server-side; mobile not yet zone-aware)

The server now has one admin-set **app time zone** (System Setting `timeZone`, IANA, default `UTC`)
that decides every calendar boundary -- "this month", report periods, which day a `createdAt` /
`resolvedDate` falls on. See the root `CLAUDE.md` -> "App time zone" for the rules. **On mobile this
is a regenerated client only**: `AppData` (and `SystemSettings`) gained an optional `String?
timeZone`, which nothing reads yet. Mobile never PUTs system settings, so it cannot clobber it.

- **Mobile keeps working unchanged.** It still builds filter values and receipt dates in
  **device-local wall time with a literal `Z`** (`formatDate(zuluDateFormat, ...)`, so a local
  midnight goes out as `…T00:00:00Z`). The server reads any RFC 3339 filter value by its **calendar
  day as written** and turns that day into a range in the right zone (app zone for instant columns,
  UTC for a receipt's `date`), and it normalizes a written receipt `date` to midnight UTC of its
  as-written day -- so the fake-`Z` values land on exactly the day the user picked.
- **What is still device-local** (the pending follow-up): the quick date filter's "today" seed and
  shortcuts, every displayed instant (Added At, Resolved Date, activity times), and the
  `WITHIN_CURRENT_MONTH` display. Desktop already does all of these through the app zone
  (`desktop/CLAUDE.md` -> "App time zone (dates)"); a device in a different zone from the install's
  setting can show a time, or a "this month", one day off from what the server resolves. Port it by
  reading `AppData.timeZone` -- never by changing what goes on the wire, which the server already
  interprets correctly.

### Receipt summary

A block of totals **pinned** above or below the receipts list, covering the whole current filter
result set rather than the visible page. `swagger.yml`, the Go API and `mobile/api/` already carried
`getReceiptSummaryForGroup` and all four models -- the original reason for skipping mobile (no
filter to describe) died with the receipt filter above -- so the only backend work was the new
per-group `receiptSummaryPosition`. See the root `CLAUDE.md` -> "Receipt Summary" for the
cross-client contract and `api/CLAUDE.md` for the wire one.

- **`ReceiptSummaryBar` (`lib/groups/widgets/receipt_summary_bar.dart`) is presentational**, the
  twin of the desktop's `app-receipt-totals`: it fetches nothing and knows nothing about the
  filter. Named *Bar* rather than *Summary* because `api.ReceiptSummary` is the model it takes and
  `dashboard_widgets/group_summary.dart` already owns the word -- the same collision desktop dodged
  by not calling its component `*summary*`. Pure helpers (labelling, the config-group fallback) live
  in `lib/utils/receipt_summary.dart`, ported from `desktop/src/utils/receipt-summary.ts`.
- **Both positions are pinned, and that comes for free.** `PagedDataList` returns an `Expanded`, so
  anything else in `GroupReceiptsList`'s `Column` takes its natural height and never scrolls with
  the list. No slivers, no `Scaffold` slot -- and the bottom position is reachable under infinite
  scroll, which a trailing sliver would not be.
- **`PagedDataList` carries a `GlobalKey`, and it is load-bearing** — this bit *cost a cycle*.
  The bar renders in one of two slots either side of the list, so the list's position among its
  siblings changes with the group's configured position. `Element.updateChildren` walks the old and
  new child lists inward from both ends while `Widget.canUpdate` holds, and in the middle it cannot
  match it reuses only **keyed** children — so unkeyed, `PagedDataList` is rebuilt from scratch,
  **discarding its State along with the paging controller, every loaded page and `_totalCount`, and
  silently refetching page 1**. Padding the `Column` with always-present placeholder slots also
  works, but the key travels with the widget it protects, so adding any other conditional child to
  this `Column` later cannot reintroduce the bug. Pinned by *"flipping the position does not reset
  the paged list"*, which asserts **State identity** — a request count does not catch it, because
  the reconstructed State refetches immediately and that looks exactly like the filter's own
  refetch.
- **The list owns the request; `ReceiptListModel` owns only the All-group pick.** The model has no
  Dio and no route access, and -- decisively -- its notify contract says a notification means *the
  filter changed*, so a model that fetched and notified on arrival would ping-pong with
  `_refreshForFilterChange`. Putting `_fetchSummary` beside the `_refreshCallback` call sites is
  what makes the refresh contract true **by construction**:

  | event | refetches the summary? |
  |---|---|
  | filter applied (the one notifying writer) | yes |
  | first mount / group change | yes, guarded on `_summaryGroupId` |
  | sort change (`notify: false`, calls `_refreshCallback` directly) | **no** -- order, not membership |
  | infinite-scroll page load | **no** -- the figures already cover every page |

- **`didChangeDependencies` must be guarded on the group id.** It fires for *any* inherited widget
  change and `getGroupId` makes `GoRouterState` a dependency, so an unguarded fetch there is a
  request storm on the app's most expensive request.
- **Out-of-order responses are dropped** via `_summaryRequestSeq`, the manual equivalent of the
  desktop's `switchMap`: two filter applies in quick succession can land backwards and paint figures
  for a filter the user has moved past.
- **A failed summary is swallowed but the figures are CLEARED** -- and this is the one place mobile
  deliberately diverges from desktop. Swallowing is shared: an unhandled `DioException` from a
  `setState`-driven fetch takes the whole receipts screen down, and a block of totals is not worth
  interrupting someone browsing receipts over. Desktop's `catchError` then *keeps* the last good
  figures, because its HTTP interceptor tells the user the refresh failed. **Mobile installs no
  interceptor** (`lib/client/client.dart` is four lines), so keeping them would render the previous
  filter's totals beside the new filter's list with nothing to say they are stale. "Covers the whole
  current filter result set" is the block's entire contract, so showing nothing is the honest
  degradation and showing wrong numbers is not. The clear sits **behind the same `_summaryRequestSeq`
  guard as the success path**, or a superseded failure would wipe a newer response's figures.
- **The cached group settings decide only WHETHER to ask.** Everything rendered -- `enabled`,
  `position`, the rows -- comes off the response, which stays authoritative. A group that never
  opted in therefore costs no request at all. The consequence: an admin who enables the summary
  while the app is running changes nothing on a screen already open. The list reads `GroupModel`
  with `listen: false` **deliberately** — a listener would re-fetch an unpaged aggregate on every
  15-minute AppData refresh for nothing — so the block appears on the next navigation into the
  group, not merely on the next refresh. Desktop behaves the same way.
- **ROWS WRAP; NOTHING IS A COLUMN -- and this is the one to not undo.** The first version was a
  frozen 148pt label column beside a horizontally scrolling grid of fixed 116pt figure columns. At
  390pt that leaves 210pt for figures and two columns want 232, so a group with a **single**
  currency field already had its second column chopped mid-number, and a long status name
  ("Needs Attention Receipts") wrapped to two lines and ate its own receipt count through the
  label's ellipsis. Fixed widths cannot fit an unknown number of configured fields into a phone.
  So a row is now a label line plus `name value` pairs in a `Wrap` -- the same flex-wrap the
  desktop's `receipt-totals` uses, and for the same reason. Anything that does not fit moves to the
  next line instead of disappearing.
  - **The label line and the figures line are separate, deliberately.** One `Wrap` over the label
    and every figure lets the first figure ride up whenever a label happens to be short, so "Total"
    starts at a different x on every row and the block reads as ragged. Two lines cost a little
    height and buy a left rail every figure starts from.
  - **Except for a lone `Total`**, which rides the label line pushed right: that is the shape most
    groups have, and giving it a line of its own wasted half the block's height on one number. It
    is a `Wrap` with `spaceBetween` rather than an `Expanded`, so at a large text scale it drops to
    the next line instead of becoming an overflow stripe.
  - Pinned by *"clips nothing horizontally, however much is configured"*, which asserts every
    figure's rect is inside the bar's. **Vertical** overflow is deliberately not asserted -- the
    35% cap scrolls, with a visible thumb, and that is a real affordance; there is no equivalent
    for a number past the right edge.
- **A full-bleed band carrying an inset card.** The receipts route zeroes `ScreenWrapper`'s body
  padding so rows sit edge to edge, and a block inset on all sides reads as a list row rather than
  as pinned chrome -- so the *band* still reaches the screen edges, tinted `surfaceDim` (the page
  canvas) with the divider on whichever edge faces the list, and the white rounded card sits 12pt
  inside it. The card is a hand-built `Container`, matching `ReceiptFilterConditionCard`: the app's
  `Card` is Material's elevated one, far heavier than this wants (see "Theme & color roles").
  Height is capped at 35% of the viewport with an inner vertical scroll, or five broken-out
  statuses on a short phone starve the `Expanded` list and overflow the `Column`.
- **The status dot is RINGED, and that is not decoration.** `receiptStatusColor` returns a pale
  *background* tint -- it is built to sit behind dark text, which is how `ListItemTrailingStatus`
  and `ListItemColorBlock` use it. Painted as a bare dot on white, OPEN (`#FFFACD`) lands near
  1.1:1 and is invisible; the 1px `outlineVariant` ring is what makes it read at every status while
  staying quieter than a chip. The overall row takes an empty gutter of the same width, so every
  label starts on the same x.
- **`_money` wraps `formatCurrency`**, which parses the wire string with `double.parse` and
  **throws**. From inside this bar that takes down the receipts screen, not just the block.
- **A zero row is muted with a COLOUR** (`onSurfaceVariant`), not an `Opacity` widget: the mute has
  to reach the row's label, its count and every figure, and keeping several `Opacity` layers in
  step is a drift waiting to happen.
- **The overall row arrives carrying `ReceiptStatus.empty`** -- the generated `status` is
  non-nullable, so there is no "no status" to express -- and `receiptStatusLabel` renders that as
  `""`. `receiptSummaryRowLabel` branches on it, or the row reads as a bare " Receipts".
- **Anything that is not `TOP` renders at the bottom**, so a value added later degrades to the
  server's own default rather than blanking the block. The generated enum agrees: its
  `fallback: true` sits on `BOTTOM` (see "Regenerating API Client Models"), so an unknown wire value
  is already `BOTTOM` by the time `isReceiptSummaryAtTop` sees it.
- **On the All group the pick lives in `ReceiptListModel`**, session-scoped: mobile has no persisted
  slice equivalent to the desktop's NGXS `receiptTable`. Written with `notify: false` and refreshed
  directly, like the sort setters -- a configuration pick changes the breakdown's shape, never the
  data. **But the write is still wrapped in `setState`**: silencing the model leaves nothing else to
  repaint, and `ReceiptSummaryBar` reads `selectedConfigGroupId` off its widget, so without it the
  tapped chip stays unhighlighted for the whole round trip and forever if the request fails. The two
  are not in tension -- `notify: false` keeps the write off `_refreshForFilterChange` (so the *list*
  is untouched), `setState` repaints this State alone. Desktop gets this for free: its NGXS dispatch
  updates a signal. It is reset on a **group change by the list**, in the same branch that re-fetches the
  summary — not from `clearFilter` (which early-returns on an already-empty filter) and not from
  the filter-clear branch beside it (which is itself gated on there having been a filter at all).
  Either would leave a user who never filtered carrying a pick out of the All group and back.
- **E2E: `integration_test/receipt_summary_test.dart`.** The widget suite injects a mocked
  `ReceiptApi` and a hand-seeded `GroupModel`, so it proves nothing about the wire; this covers the
  encoded filter reaching the summary endpoint and narrowing every row, and `position` surviving
  model -> command -> DB -> response -> render. `setGroupSummaryConfig` in
  `integration_test/helpers/permission_fixtures.dart` restores **all four** summary keys explicitly
  on teardown: they are pointers on the Go command, so replaying `_settingsToCommand` omits them and
  an omitted key means *leave unchanged* -- a spec turning the summary on for a shared group would
  otherwise leave it on for every later run. Do **not** add them to that helper's `?? false` loop;
  `receiptSummaryPosition: false` fails the enum decode.

### Category / Tag / Users pickers — the tap target lives in `MultiSelectField`

`CategorySelectField` and `TagSelectField` render no UI of their own: both are thin
wrappers that hand an `onTap` to **`lib/shared/widgets/multi-select-field.dart`**, which is
also used directly for the "Users" field in the quick-actions split sheet. So the field's
hit target is defined in exactly one place, for the receipt form, the Quick Scan sheet, the
per-item rows and the split sheet at once.

- **The `GestureDetector` must wrap the `InputDecorator`, with
  `behavior: HitTestBehavior.opaque`.** It originally sat *inside* the decorator around the
  inner `Wrap`, which made the field tappable only on its text: the label, the outline
  border and the content padding were outside the detector; `Wrap` shrink-wraps under
  `InputDecorator`'s loose width constraints, so the empty state's target was the width of
  the `"No <items> selected"` glyph run; and the default `deferToChild` behavior dropped
  even the gaps between chips, because `RenderWrap` and the childless `SizedBox` spacers
  both hit-test `false`. Adjacent `FormBuilderDropdown` fields are tappable edge to edge, so
  the difference was very visible.
- **View mode installs no tap surface at all.** When `onTap` is null (the wrappers pass null
  in `WranglerFormState.view`) the bare `InputDecorator` is returned rather than an opaque
  detector that would swallow pointers for a no-op.
- **`onSelected` is kept even though it is now redundant** — the chip wins the
  gesture arena and calls the same handler the outer detector would. It stays because
  `onSelected: null` renders the chip in its *disabled* style.
- **Each chip carries a remove X, so the chips are `InputChip`s, not `ChoiceChip`s.**
  `ChoiceChip` does not implement `DeletableChipAttributes` and so has no `onDeleted`;
  `InputChip` implements both it and `SelectableChipAttributes`, so every property the
  chip already set carries over and, under the app's M3 theme, the selected-state
  background, label color and border side are the same — the X is the only visual delta.
  `deleteIcon` **is** set explicitly (`Icon(Icons.cancel, size: 18)`): InputChip's M3
  default is `Icons.clear`, a bare X, where desktop's `matChipRemove` button renders
  `<mat-icon>cancel</mat-icon>`, the filled circle. `Wrap` gained `runSpacing: 5` because
  the wider chips reach a second run sooner and the runs would otherwise touch.
- **Removal goes out through `onRemove`, never through `field.didChange`.** The owning
  forms are the source of truth: each answers the picker's result with
  `setState(() => ...setValue(list))`, and Quick Scan additionally fires `onValueChange()`
  to mirror the form into its `QuickScanImage` (`quick_scan_form.dart`). A remove that
  mutated the `FormBuilderField` internally would leave that image holding the removed
  category and submit its id anyway. Routing it through the same callback the sheet result
  takes means every call site inherits its existing propagation — `CategorySelectField` /
  `TagSelectField` just forward to their `onCategoriesChanged` / `onTagsChanged`, so no
  call site of theirs changed. The new list is computed by **index**
  (`List.from(field.value)..removeAt(index)`), because the items carry no equality and a
  list can hold two equal entries; desktop's `removeOption(index)` does the same.
- **`onRemove == null` is what hides the X**, and it is what the wrappers pass in
  `WranglerFormState.view` alongside the null `onTap` — mirroring desktop gating its
  `matChipRemove` button on `*ngIf="!readonly"`. Note `setValue` does **not** call
  `setState` (only `didChange` does), so it is the owner's rebuild that re-runs the
  field's builder against the new value; a test harness has to mirror that, which is why
  `multi_select_field_test.dart` wraps the field in a `StatefulBuilder`.
- **No explicit min-height is set.** The app's global `InputDecorationTheme` uses
  `OutlineInputBorder` (`lib/main.dart`), whose default non-dense content padding already
  puts the decorator past the 48dp target even in the empty state.
- **The sheet itself opens with `bodyFillsSheet: true`** (`shared/functions/multi_select_bottom_sheet.dart`),
  and `FilterMultiSelect` pins the Filter bar with the chip grid under it in an `Expanded`. This is the
  same class of bug as the Quick Scan sheet above, and it needed both halves of that fix:
  - The `MasonryGridView` was `shrinkWrap: true` inside the sheet's `SingleChildScrollView`. A
    shrink-wrapped grid sizes itself to its content, so its **own scroll extent is zero** — but it
    still wins the vertical drag, which then reaches no scrollable at all. The chips cover the whole
    sheet, so on a **touch** device no drag a user could make scrolled anything and a catalog past one
    screenful was simply unreachable. A *mouse wheel* still bubbled to the outer scroll view, so the
    list did move on desktop — assert against the **grid's** `ScrollPosition`, not any ancestor's, or
    the desktop e2e proves nothing.
    Dropping `shrinkWrap` also restores lazy chip building; it was building every option eagerly.
  - `BottomSubmitButton` was pinned as `Scaffold.bottomSheet`, which **floats over** the body, so the
    last row of chips sat under the 50px "Select" button with no way to clear it. `bodyFillsSheet`
    moves it to `bottomNavigationBar`, which reserves the space instead.
  Both fixes land for Categories, Tags **and** the quick-actions Users picker at once — they share the
  one helper. `Expanded` is safe there because `FilterMultiSelect` has exactly one call site.
  - `bodyFillsSheet` moved the confirm button into the one slot `Scaffold` does not lift above the
    keyboard — and this sheet pins a `FormBuilderTextField` (the Filter bar) right above it, so
    keyboard-up is its **normal** state, not an edge case. `ScreenWrapper` lifts it; see "A pinned
    bottom bar is not keyboard-safe" above, and the keyboard case in
    `test/widgets/filter_multiselect_test.dart`.
- **Demo:** `tool/tap-target-demo.gif` shows the before/after hit regions under real clicks; it is
  regenerated by `tool/record_tap_target_demo.sh` (see "Recording a demo GIF" below).
  `tool/multi-select-keyboard.gif` shows the confirm button against an open keyboard.
- **Tests for the sheet's layout:** `test/widgets/filter_multiselect_test.dart` (phone-sized surface,
  90 options: a drag on the chips moves the **grid's own** `ScrollPosition`, the last chip ends up
  inside the window and clear of the confirm button, the filter bar does not move) and
  `integration_test/receipt_category_picker_scroll_test.dart` (the same contract against a 60-item
  catalog that arrived over the wire). Assert **geometry, not finders** here for the reason the Quick
  Scan section gives — `findsOneWidget` is true for a clipped chip and `ensureVisible` drags it into
  view even where a user could not. Three of the four widget cases and the e2e fail on the pre-fix
  sheet; verify that when changing them.
- **Tests:** `test/widgets/multi_select_field_test.dart` covers the widget end to end —
  edge/label/gutter/chip-gap taps (the regression guards; they assert the detector's rect
  equals the decorator's), chip and placeholder taps, rendering, form registration and
  `didChange`, the `required` validator, the inert view mode, and the remove button (present
  only with `onRemove`, reports the remaining values in order, drops one of two duplicates,
  reports `[]` rather than null for the last chip, and — the one most likely to regress
  silently — does **not** also fire `onTap`, since the button's `InkWell` has to win the
  gesture arena against the field's opaque ancestor detector).
  `test/widgets/quick_scan_form_test.dart` pins the propagation contract above: removing a
  chip reaches the `QuickScanImage` through `onValueChange`. The e2e specs still tap the
  `"No <items> selected"` placeholder as their locator, and the ones that drive chips drive
  the **sheet's** `ChoiceChip` toggles (`FilterMultiSelect`, unchanged), so both keep working.

### Recording a demo GIF

Two routes, and the choice is forced by what the demo has to show:

| route | use when | needs |
|---|---|---|
| **desktop capture** (below) | the demo needs real pointer input — hit targets, drags, hovers | ffmpeg, xdotool, Xvfb, a Linux desktop build |
| **widget-test capture** (further below) | the demo needs a platform state the Linux desktop target cannot produce — an open **software keyboard**, a device pixel ratio, a notch | nothing beyond `flutter test` |

#### Route 1 — desktop capture (headless Linux)

There is no emulator in the Claude Code sandbox and the Go API is expensive to bring up there
(ImageMagick 7 from source — see the root `CLAUDE.md`), so a demo of a **widget-level** change is
recorded by building a small harness against the **Linux desktop** target and driving real mouse
clicks at it on a headless X server. Worked example, all committed and re-runnable:

```bash
cd mobile
flutter build linux --release --target=tool/tap_target_demo.dart
./tool/record_tap_target_demo.sh            # -> tool/tap-target-demo.gif
```

`tool/tap_target_demo.dart` is the harness (mounts the real widget beside a verbatim copy of the
pre-fix tree and publishes each field's on-screen rect as JSON after first paint);
`tool/record_tap_target_demo.sh` runs Xvfb + the binary + `ffmpeg -f x11grab` and converts to GIF;
`tool/tap_target_click_plan.py` drives `xdotool`. Copy the trio for a new demo. ~20s at 10fps,
960px wide, lands around 100KB — a flat UI GIF compresses very well, but it needs the two-pass
`palettegen`/`paletteuse` filter or it bands badly.

Prerequisites beyond the Flutter SDK setup above (on Ubuntu 24.04):

```bash
apt-get install -y --no-install-recommends \
  ffmpeg xdotool xvfb libgtk-3-dev lld llvm libsecret-1-dev pkg-config libcurl4-openssl-dev
flutter config --enable-linux-desktop
```

Five things that each cost a cycle:

- **`libcurl4-openssl-dev` is required**, and the CMake error does not say so: `sentry_flutter`
  vendors sentry-native, whose `find_package(CURL)` fails with *"CURL: Required feature AsynchDNS is
  not found"* when only the runtime curl is installed.
- **A failed configure poisons the install prefix.** CMake caches `CMAKE_INSTALL_PREFIX=/usr/local`
  on the first (failed) configure, after which Flutter's
  `if(CMAKE_INSTALL_PREFIX_INITIALIZED_TO_DEFAULT)` redirect in `linux/CMakeLists.txt` no longer
  fires — the build reports `✓ Built build/linux/x64/release/bundle` while actually installing to
  `/usr/local`, and `cmake --install --prefix` won't override it. `rm -rf build/linux` and rebuild.
- **The harness must publish its own coordinates.** It writes each field's rect (in physical pixels,
  multiplied by `devicePixelRatio`) to `$TAP_DEMO_RECTS` in a post-frame callback. That file
  appearing is also the "app is up and painted" signal the recorder waits on — far more reliable
  than a fixed sleep, and it means the click plan never hardcodes a layout.
- **Don't mount the app's real bottom sheet in a bare harness** — *on this route*.
  `showMultiselectBottomSheet` → `showFullscreenBottomSheet` renders its `TopAppBar` /
  `BottomSubmitButton` as untextured grey blocks outside the app's full theme and provider tree,
  *and* it covers the whole 1280x720 window, hiding whatever is being compared. Demonstrate the
  outcome in-panel instead. Both halves are artefacts of a bare `runApp` harness and **do not apply
  to route 2**, where the test owns the surface: real fonts via `FontLoader`, the real provider
  tree, and one panel captured per pump so nothing can cover the comparison.
- **`pkill -f receipt_wrangler_mobile` kills the shell running it** — the pattern matches that
  shell's own command line, which shows up as a bare exit 144. Use `killall receipt_wrangler_mobile`.

`ffmpeg -f x11grab` draws the real X cursor (`-draw_mouse 1`, the default), so the pointer is
visible without faking one; crop to the content (`crop=1280:412:0:0`) rather than shipping 300px of
empty window.

Recording the **full app** (rather than a harness) against a live API adds three more:

- **A debug build records as a black screen.** Xvfb has no DRI3, and the debug embedder's EGL surface
  never lands in a drawable `XGetImage`/`x11grab` can read — the app runs and the test passes, but
  every captured frame is blank. A `--release` (or `--profile`) build renders normally. That rules out
  recording a `flutter test`/`run-e2e.sh` run directly, since integration tests need a debug build.
- **Unlock a keyring or the app cannot get past "Connect to Server".** `flutter_secure_storage` logs
  `libsecret_error: KeyringLocked` and the Connect button silently does nothing. Run under
  `dbus-run-session` with `printf 'x' | gnome-keyring-daemon --unlock --components=secrets`
  (`apt-get install gnome-keyring dbus-x11`). The session then persists across relaunches, which is
  what makes a before/after pair cheap: rebuild and the app comes back already logged in.
- **Scroll with the wheel (`xdotool click 5`), not a mouse drag.** Flutter's desktop `ScrollBehavior`
  excludes `PointerDeviceKind.mouse` from its drag devices, so a click-and-drag scrolls nothing even
  on a perfectly good list. To show a **touch**-specific failure, drive it in a widget test instead —
  route 2 below is the worked example.

#### Route 2 — widget-test frame capture

`tool/demo_capture/` holds the machinery and four demos — an animated GIF pair for the
keyboard-inset fix, a single before/after still for the floating-button clearance fix, the
receipt-summary placement pair, and a reference still of the summary bar itself:

```bash
cd mobile && ./tool/record_keyboard_demo.sh          # -> tool/*-keyboard.gif
cd mobile && ./tool/record_clearance_shot.sh         # -> tool/receipt-form-clearance.png
cd mobile && ./tool/record_summary_position_demo.sh  # -> tool/receipt-summary-position.gif
cd mobile && ./tool/record_summary_shot.sh           # -> tool/receipt-summary-bar.png
```

**Judge a LOOK from the still, not from a frame of the GIF.** `writeSideBySideGif` quantises to an
octree palette and bands badly on flat UI — the stripes through the status washes in the placement
GIF are encoding artifacts, not pixels any widget drew — so spacing, weight and colour decisions
made from it are made against noise. `record_summary_shot.sh` writes a PNG, in the two
configurations of the summary bar that look least alike: the one most groups have (a couple of
statuses, no currency fields) and the heavy one (every status broken out, two currency fields) that
is where a row-layout regression shows up first.

Route 1 **cannot** record that fix at all: there is no software keyboard on Linux desktop, so
`viewInsets.bottom` is permanently 0 and the bug is unreproducible on that target. This route also
needs no ffmpeg, ImageMagick, xdotool or Android SDK, none of which the Claude Code sandbox has —
frames are captured inside a widget test and encoded to GIF in pure Dart.

- **`capture.dart`** is the shared machinery — `loadDemoFonts`, `buildDemoSurface` (caption strip +
  phone-sized viewport), `grabFrame`, `stitchPanels` and the two writers. `fake_keyboard.dart` draws
  the keys and collapses to nothing at a zero inset, so a still that does not want a keyboard simply
  leaves `viewInsets` alone. `keyboard_demo_test.dart`, `clearance_shot_test.dart`,
  `summary_position_demo_test.dart` and `summary_shot_test.dart` are the demos.
- **A still does not always have a seam to flip.** The keyboard demo toggles
  `debugDisableKeyboardLift`, but there is no production flag for `submitButtonSpacing` and adding
  one for a screenshot is not worth it — so `record_clearance_shot.sh` captures one panel, strips the
  spacer from the source, captures the other, and restores under a `trap`. The trap is load-bearing:
  without it a crash between runs leaves the working tree patched. Verify `git diff lib/` is empty
  after running it.
- **Turn off the debug banner** (`showDebugBanner: false` on the pump helpers). `MaterialApp` draws
  its red ribbon across the top-right of every captured panel otherwise, and at panel scale it reads
  as a render-overflow stripe rather than a banner — which sends you hunting a layout bug that is not
  there.
- **The summary demo needed NO production seam at all**, and that is a property of the feature
  rather than luck: the position arrives as *data on the summary response*, so the two panels are
  the identical real tree fed two responses differing in one field. Better than a debug flag —
  nothing to leave switched on, no source patched and restored under a `trap` (verify `git diff
  lib/` is empty after a run; it should never have been non-empty), and if placement regresses the
  two panels simply become identical. It mocks only `OpenApiClient.client`; the theme, the
  `ScreenWrapper` with the receipts route's zeroed body padding, the providers and
  `GroupReceiptsList` are all the shipping tree.
- **Two things in `capture.dart` are parameterised for it, and both bite silently.**
  `writeSideBySideGif` hard-coded `demoInsetRamp()` for its per-frame durations, so any demo with a
  different frame count either `RangeError`s or gets keyboard timings applied to something that is
  not a keyboard — pass `durationsCentis`. And `numColors` was fixed at 128, which is right for flat
  UI art and wrong for a screen with a **gradient**: at 128 the octree banding across
  `ListItemTrailingStatus`'s status chips renders as hard black stripes that read as a rendering
  bug. Raise the palette (192 here) rather than switching the dither on, which breaks LZW's runs
  everywhere else and costs far more bytes.
- **Drive a scroll with an explicit gesture and fixed pumps, never `drag` + `pumpAndSettle`.**
  Ballistic physics give a frame count that is not reproducible run to run, and `pumpAndSettle` is
  independently unsafe here: `PagedDataList`'s first-load spinner animates, and `TopAppBar`'s
  progress bar is permanently mounted with a **ten-minute** settle timeout — that regression hangs
  CI rather than failing it.
- **Seed `UserModel` for any demo showing the receipts list.** `ReceiptListItem` resolves each row's
  payer with a `firstWhere` that has no `orElse`, so an unseeded model throws "Bad state: No
  element" once per row and the frames come out full of error widgets.
- **It lives in `tool/`, not `test/`, and that is what keeps it out of CI.** `flutter test` with no
  arguments scans **only** `test/`, while an explicit path is used verbatim. CI runs a bare
  `flutter test`, so a demo under `test/` would re-encode and rewrite committed binaries on every
  push. Keep the `_test.dart` suffix (name-based filters) and the `tool/` directory (discovery).
- **Before/after is the same real screen recorded twice**, differing only by the
  `debugDisableKeyboardLift` seam, then stitched side by side at encode time. That beats
  hand-copying the pre-fix tree (as `_LegacyMultiSelectField` did on route 1): the copy cannot
  drift, and if the fix regresses both panels simply become identical.
- **Everything with real asynchrony must be inside `tester.runAsync`** — `toImage`, `toByteData`,
  `File.readAsBytes`, `FontLoader.load`. `testWidgets` runs in fake-async, which never services the
  engine's futures, so a bare `await boundary.toImage()` **hangs until the test times out** rather
  than failing. `runAsync` also *swallows* errors and returns null, surfacing them via
  `takeException` — so null-check and `fail()`, or a real failure shows up as a confusing null.
- **Route 2 can drive a whole API-backed screen, not just a bare form.**
  `quick_date_demo_test.dart` records the real `GroupReceiptsList` -- real stepper, real sheet, real
  picker, real theme and provider tree -- with only `getReceiptsForGroup` stubbed, by a fake that
  **decodes the filter the client actually sent** and returns the seeded rows inside it. So the rows
  narrowing is evidence rather than choreography: break the encoder and the demo visibly stops
  filtering. That is also why route 1 was not an option for it -- driving the real control there
  needs the Go API behind it, which the sandbox cannot bring up (ImageMagick 7 from source).
- **A demo of a new feature is one panel, not a pair.** The other demos here record the same screen
  twice across a `debugDisable*` seam because they document *fixes*. A new control has no seam and
  nothing to compare against, so `writeGif` takes frames that each carry their own hold time (in
  1/100 s) instead of reading durations off `demoInsetRamp()`.
- **`numColors: 128` is for flat UI art; a real screen may need 256.** `ListItemTrailingStatus`
  paints a `LinearGradient` in every receipt row, which a short palette bands into visible vertical
  stripes with dithering off. 256 (the GIF maximum) clears it, and also cleans up the antialiased
  white caption text that ghosts at 128 -- visible in the committed keyboard GIFs.
  **Turning dithering on instead is not the fix**, and it is worth not rediscovering: Floyd-Steinberg
  at 256 colors dithered the screen's large *white* areas into visible noise and took the file from
  176KB to 445KB. The existing "no dither" guidance holds; raise the palette, not the dither.
- **Call `loadDemoFontsOrFail(tester)`, never `loadDemoFonts` directly.** Two failure modes sit on
  this one line and neither announces itself. A bare `await loadDemoFonts()` reads fonts off disk and
  awaits `FontLoader.load` in fake-async, so it **hangs** until the 10-minute test timeout rather
  than failing -- and the timeout names the test, not the call, so it reads as a mysterious stall.
  Wrapping it in `tester.runAsync` fixes that but introduces the opposite problem: `runAsync` returns
  null **both** when the callback throws and when it has nothing to return, and `loadDemoFonts`
  returns `Future<void>` -- so a genuine font-load failure is swallowed and the demo cheerfully
  records a whole clip in `--use-test-fonts`' stub font, where every glyph is a filled box. It
  passes, and writes a GIF nobody can read. `loadDemoFontsOrFail` returns a sentinel from inside
  `runAsync` and fails on anything else, the same shape `grabFrame` uses; all three demos call it.
- **`flutter test` always passes `--use-test-fonts`,** whose stub font draws every glyph as a filled
  box. Load Raleway from `fonts/` on disk, register it as `Roboto` too (Material's default
  `Typography` asks for that, and most text takes the theme default), and load `MaterialIcons` or
  the app bar's icons are boxes.
- **Set a `Timeout`** — `package:test`'s 30s default is wall-clock and unaffected by the fake clock,
  and pure-Dart quantize + LZW over a few dozen frames takes minutes.
- **Encoder settings are not the defaults.** `GifEncoder`'s default neural quantizer is a per-frame
  neural net (slow) and dithering destroys the pixel runs LZW needs; flat UI art wants
  `QuantizerType.octree`, `numColors: 64`, `DitherKernel.none`. Note `addFrame` encodes the
  *previous* image and `finish()` flushes the last, and `duration` is in **1/100 s**, not ms.
  There is no inter-frame diffing, so a held frame costs as much as a moving one — **buy hold time
  with long per-frame durations, not duplicate frames**.
- **Keep the captured subtree opaque.** `rawRgba` is premultiplied and the encoder switches on GIF
  transparency the moment it finds a zero-alpha palette entry, which makes the whole clip flicker.
- **`EditableText.debugDeterministicCursor = true`** or the caret blink makes frames differ run to
  run (and never call `pumpAndSettle` after focusing a field — the blink schedules frames forever).

### `integration_test` is a regular dependency on purpose (Android Studio signed-bundle builds)

`integration_test` is listed under **`dependencies`**, not `dev_dependencies`, in `pubspec.yaml`.
This looks wrong and is deliberate — moving it back breaks release builds driven from Android
Studio's **"Generate Signed Bundle / APK"** wizard.

Why: **Flutter 3.27+ strips `dev_dependency` plugins from release builds.** But
`android/app/src/main/java/io/flutter/plugins/GeneratedPluginRegistrant.java` (generated, gitignored)
is rewritten by the **flutter tool**, and `flutter pub get` — which Android Studio runs on every
**project sync** — re-injects `IntegrationTestPlugin` into it. The wizard then invokes the Gradle
task **directly**, so nothing regenerates the registrant for release mode, and
`:app:compileReleaseJavaWithJavac` fails with:

```
GeneratedPluginRegistrant.java:44: error: package dev.flutter.plugins.integration_test does not exist
```

`flutter clean` does **not** fix it — the registrant lives in `src/main/java`, not `build/`. The
failure is deterministic, not flaky: every Studio sync re-poisons the file. `flutter build
apk|appbundle --release` always works, because the flutter tool regenerates the registrant with dev
plugins stripped — which is why the CLI and the wizard disagree.

Keeping `integration_test` as a normal dependency restores the pre-3.27 behavior (the plugin stays on
the release classpath, so the reference compiles) and costs roughly **+0.6MB** of bundle plus an
inert `IntegrationTestPlugin` registration in production. The alternative — dropping the wizard for
`flutter build appbundle --release` with a real `signingConfigs.release` in `android/app/build.gradle`
— is the cleaner long-term fix but requires wiring up keystore signing, which the repo currently does
**not** have (release is pinned to `signingConfig signingConfigs.debug`; the wizard supplies signing).

### Android Gradle: cunning_document_scanner needs the Kotlin plugin + Kotlin 2.2.x

`cunning_document_scanner` **2.3.0** modernized its `android/build.gradle` to the
`kotlin { compilerOptions { ... } }` DSL but (a) **does not apply the Kotlin Gradle plugin
itself** and (b) requires **Kotlin 2.2.x** for that extension-level `compilerOptions` block.
This app historically applied `kotlin-android` only to `:app` and pinned Kotlin **2.1.0**, so a
clean Android build failed at the Gradle **configuration** step — first
`Could not find method kotlin()` then `Could not find method compilerOptions() ... on
KotlinAndroidProjectExtension` — for **every** build, blocking the whole Android e2e suite.

The fix (all in `mobile/android/`, committed app config — not generated):

- **`build.gradle` (root):** a `subprojects { sp -> sp.plugins.withId('com.android.library')
  { sp.apply plugin: 'org.jetbrains.kotlin.android' } }` block applies `kotlin-android` to any
  Android-library plugin subproject the moment it applies `com.android.library`, so the
  plugin's `kotlin {}` DSL resolves (re-applying an already-applied plugin is a no-op). The
  legacy `buildscript` Kotlin classpath was **removed** from this file — the version is now
  managed in `settings.gradle` (keeping it in both places triggers "plugin already on the
  classpath with an unknown version").
- **`settings.gradle`:** the `pluginManagement` `plugins {}` block declares
  `id "org.jetbrains.kotlin.android" version "2.2.21" apply false` (matching the modern Flutter
  template and the plugin's own example app). This is the lever that actually sets the Kotlin
  version the `kotlin-android` plugin resolves to — bumping the old `ext.kotlin_version` in the
  root `build.gradle` had **no effect**.

If a future dependency bump re-breaks the Android build with a Kotlin/AGP DSL error, the Kotlin
version to bump is in `settings.gradle`, not the root `build.gradle`. The simpler alternative
(if you don't need 2.3.0) is to pin `cunning_document_scanner: ^1.4.0`, whose `build.gradle`
applies `kotlin-android` itself and uses the legacy `kotlinOptions {}` DSL (works on Kotlin
2.1.0); its `getPictures(noOfPages:)` API — the only call the app uses (`lib/utils/scan.dart`) —
is identical.

**Related: `sentry_flutter` pins Kotlin `languageVersion 1.6`.** `sentry_flutter` **8.14.2**'s
`android/build.gradle` hardcodes `kotlinOptions { languageVersion = "1.6" }`, which the project's
Kotlin **2.2.x** toolchain (above) rejects at compile time — `flutter build apk --release` fails at
`:sentry_flutter:compileReleaseKotlin` with *"Language version 1.6 is no longer supported; please,
use version 1.8 or greater"*. iOS uses a different toolchain and never hits this. The fix (root
`android/build.gradle`) is a `subprojects { if (sp.name == 'sentry_flutter') { sp.afterEvaluate {
tasks.withType(KotlinCompile).configureEach { compilerOptions { languageVersion / apiVersion .set(
KOTLIN_1_8) } } } } }` block — scoped to that one module (via `afterEvaluate` so it overrides the
module's own `kotlinOptions`), so every other subproject keeps its own language version. The block
is a no-op / removable once `sentry_flutter` is upgraded past the 1.6 pin (9.x drops it, but 9.x
needs a newer Dart SDK than this app currently targets).

### Crash & error reporting (GlitchTip) + the iOS launch-freeze fix

**Crash reporting** is via `sentry_flutter` → a self-hosted **GlitchTip** (Sentry-compatible)
backend, wired in `lib/service/crash_reporting.dart`. It is **opt-out (on by default)** and
**privacy-hardened**: `sendDefaultPii=false`, no screenshots / view-hierarchy / print breadcrumbs,
`tracesSampleRate=0`, `enableAutoSessionTracking=false` (GlitchTip has no sessions), no `setUser`.
`main()` only calls `SentryFlutter.init` when `isCrashReportingEnabled()` (a `SharedPreferences`
flag, default true); the **opt-out toggle** ("Crash & error reporting") lives in
`UserProfileScreen` and flips it live via `setCrashReportingEnabled` (`SentryFlutter.init` /
`Sentry.close()`). The app has **zero** other analytics/tracking SDKs. `sentry_flutter` is pinned at
**8.14.2** (9.x needs a newer Dart SDK); its Android Kotlin pin needs the gradle override documented
above.

**iOS launch-freeze fix (GitHub #617).** On iOS the Flutter engine pauses rendering while the app is
`inactive`, and on **iOS 26.x with 120Hz ProMotion displays** (iPhone 17 / Air) it doesn't reliably
resume — the first frame is painted at transient launch window metrics (wrong size/orientation) and
then never repaints, so the UI freezes "half-rotated / unpainted." The fix (`lib/main.dart`):
- **No launch-time permission request.** The old fire-and-forget `requestPermissions()` in
  `initState` popped a camera dialog (→ `inactive` → freeze) *and* collided with in-context requests.
  Camera is now requested by the scanner itself; photo-library access is requested at the
  `Gal.putImageBytes` save sites (`receipt_app_bar_action_builder.dart`,
  `receipt_image_app_bar.dart`). `requestPermissions()` is kept (uncalled) because a unit test
  covers it.
- **A forced-frame pump** (`nudgeFrames()`): bumps an invisible `ValueNotifier` (hosted in
  `MaterialApp.builder`) **and** calls `WidgetsBinding.instance.scheduleForcedFrame()` each tick for
  ~3s — the forced call bypasses the engine's "frames disabled while inactive" gate. It fires from
  three points: a **launch** `addPostFrameCallback` (cold launch delivers no `resumed`), on
  **`didChangeMetrics`** while a 6s launch window is open (re-flows the stale frame the moment the
  window geometry settles), and on lifecycle **`resumed`** (app-switcher / scan-camera returns). It
  **early-returns unless `defaultTargetPlatform == TargetPlatform.iOS`**, so Android/desktop and the
  test suites are unaffected.

### QR-scan the server URL (login)

The "Connect to Server" screen (`lib/auth/set-homeserver-url/screens/set_homeserver_url.dart`, route
`/`) has a `qr_code_scanner` suffix-icon button on the URL field. **The server-URL field lives here,
not on the `/login` username/password screen** — `/login` only *shows* the already-set base path.
The button opens a full-screen scanner (`qr_scanner_screen.dart`) built on **`mobile_scanner`**
(`^7.2.0`, QR-only via `formats: [BarcodeFormat.qrCode]`), which pops the raw decoded string back.
The scanner draws a centered **targeting box** (dimmed scrim + four L-shaped corner brackets + a hint
line, via a small private `_CornerBracketPainter`) as a **visual aim guide only** — detection runs on
the **whole camera frame**. We deliberately do **not** use `MobileScanner.scanWindow` to gate
detection: it made scanning intermittent (a QR visibly inside the box would scan only sometimes).
Per mobile_scanner's own changelog the scan-window intersection test is unreliable — Android accuracy
was only just improved in 7.2.0 ("migrated boundingBox to cornerPoints") and 7.x lists "[Apple] scan
window does not work correctly" as a known issue — so whole-frame detection is the reliable choice.
The box Rect is computed inside the `overlayBuilder` from its `constraints` (the scanner sits below an
AppBar, where `MediaQuery.sizeOf` would misalign the box).
The screen validates the decoded string with `normalizeServerUrl` (`lib/utils/url.dart` — trim + require a well-formed
http/https URL with a non-empty host; **http is allowed** for LAN/self-hosted instances) and, on
success, `patchValue`s the `url` form field. It **never auto-connects** — the user reviews the
populated URL and taps Connect (phishing mitigation: a malicious QR can't silently point the app at
an attacker's server that would then harvest credentials). Invalid content shows an error snackbar and
leaves the field untouched.

Native / platform notes:
- **iOS: no deployment-target change.** mobile_scanner 7.x uses Apple's **Vision framework** on
  iOS/macOS (not GoogleMLKit — see its issue #1225), and its pod declares
  `s.ios.deployment_target = '12.0'`, so the repo's existing **13.0** is fine. Verified:
  `flutter build ios --simulator --no-codesign` builds clean at 13.0 (an earlier plan to bump to 15.5
  for a supposed GoogleMLKit requirement was wrong and was reverted — no iOS device support dropped).
  `NSCameraUsageDescription` already existed (shared with `cunning_document_scanner`); only its wording
  was broadened to mention QR scanning. After pulling this change, iOS needs `cd ios && pod install`.
- **Android** needs no gradle/manifest change: `CAMERA` is already declared and the inherited
  `flutter.minSdkVersion` (24) / `compileSdkVersion` (36) exceed mobile_scanner's floor (21 / 35). ML
  Kit is **bundled** by default; to shrink the APK set
  `dev.steenbakker.mobile_scanner.useUnbundled=true` in `android/gradle.properties` (needs Play
  Services).
- **`pubspec.yaml`** Dart SDK floor was raised `>=3.2.5` → `>=3.7.0` (mobile_scanner 7.2.0 requires it;
  the resolved toolchain is already 3.10+).
- **Permission double-request race:** the scanner `await`s the shared-in-flight `requestPermissions()`
  (`lib/utils/permissions.dart`) and uses `autoStart: false` before `controller.start()`, so it never
  races the app-init camera request (the `ERROR_ALREADY_REQUESTING_PERMISSIONS` case documented above
  for `cunning_document_scanner`).
- **Fallback states + recovery:** besides the live camera, `qr_scanner_screen.dart` renders three
  non-camera states via a shared `_buildMessage` helper — unsupported (Linux), **permission denied**
  ("Open Settings"), and **camera error** ("Retry"). `controller.start()` is wrapped in `_safeStart`
  (catches `MobileScannerException` → camera-error state; skips the start if already running). On
  `AppLifecycleState.resumed` the screen re-checks permission and clears the denied/error state if
  access was just granted (so "Open Settings" actually recovers). `_onDetect` is `!mounted`-guarded so
  a buffered detection can't pop a disposed context.
- **Linux / e2e + testing:** mobile_scanner has no Linux desktop implementation, so the screen guards
  on `Platform.isLinux` and renders the unsupported message instead of constructing `MobileScanner` —
  `run-e2e.sh` stays green. The controller is created **lazily** and the widget exposes small
  `@visibleForTesting` seams (`debugScannerSupported` / `debugForcePermissionDenied` /
  `debugForceCameraError`), so `test/widgets/qr_scanner_screen_test.dart` covers the three fallback
  states + the close button **without a camera or channel mocks**. Post-scan/validation logic is
  covered by `test/widgets/set_homeserver_url_test.dart` (injectable `scanQrCode` seam) and
  `test/utils/url_test.dart`. The live-camera path is exercised only on Android/iOS + manual runs.

### OIDC sign-in (external identity providers)

The login screen renders a "Log in with X" button per enabled provider, and the profile screen lets a
user connect/disconnect providers. The **whole OIDC exchange happens on the server** — this app opens
the system browser and redeems the one-time code that comes back. See `api/CLAUDE.md` → "OIDC
(OpenID Connect) login" for the flow and its security checks.

- **Dependency: `flutter_web_auth_2` `^5.1.0`.** Uses `ASWebAuthenticationSession` on iOS and a Chrome
  Auth Tab on Android — the external user agent RFC 8252 requires. An embedded `WebView` is **not** an
  option even though `webview_flutter` is already a dependency: Google actively rejects embedded
  user-agents with `disallowed_useragent`, and it would expose IdP credentials to the app.
  **Pinned to the 5.x line for the same reason `app_links` is pinned to 6.x**: 5.1.0 needs Dart
  `>=3.5.0` / Flutter `>=3.24.0`, both satisfied by this app's `>=3.7.0` floor, while 6.x requires
  Dart `^3.12.0` and is still alpha. Bumping it means raising the Dart floor — do that deliberately.
  `crypto` is declared explicitly because `lib/utils/pkce.dart` imports it directly.
- **PKCE padding is the one thing that only fails on a device.** `codeChallengeS256`
  (`lib/utils/pkce.dart`) **strips the padding**: Dart's `base64UrlEncode` pads with `=`, Go's
  `base64.RawURLEncoding` does not, and `utils.VerifyPkceS256` compares the two strings directly. A
  padded challenge passes every unit test on each side independently. Pinned with the RFC 7636
  Appendix B vector in `test/utils/pkce_test.dart`.
- **Scheme: `io.receiptwrangler`** — reverse-DNS of a domain the project owns (RFC 8252 §7.1),
  matching the `applicationId`. It only ever carries a single-use, PKCE-bound **code**, never tokens:
  a private-use scheme is unverifiable on Android, so any installed app can claim it; the verifier
  this app kept is what makes an interception useless.
- **Android** (`android/app/src/main/AndroidManifest.xml`): the plugin's
  `com.linusu.flutter_web_auth_2.CallbackActivity` is declared as a **sibling of `MainActivity`**
  (`android:exported="true"`, `android:taskAffinity=""`) with an intent filter on the scheme.
  **Do not** hang the scheme off `MainActivity` — that routes the callback into the launcher activity,
  bypasses the plugin's completer, and competes with the `/app/setup` App Link filter. The class name
  is verified against the installed 5.1.0 source (its *own* manifest declares a different activity,
  `AuthenticationManagementActivity`, which is not the one the host app registers).
- **iOS needs NO `Info.plist` change.** `ASWebAuthenticationSession` registers the callback scheme
  with the OS **for the lifetime of the session**, so a `CFBundleURLTypes` entry is unnecessary — and
  strictly worse, since it would let the scheme be routed to the app outside a session. The plugin's
  README says the same ("nothing special to do"). `Runner.entitlements` is untouched; the
  associated-domains setup for `/app/setup` is unrelated and keeps working.
- **`desktop_webview_window` is pulled in transitively**, and its Linux CMakeLists
  `pkg_check_modules` on `webkit2gtk` and `libsoup`. Without those headers `flutter test -d linux`
  (`run-e2e.sh`) fails to build **every** plugin, not just that one — so `libwebkit2gtk-4.1-dev` and
  `libsoup-3.0-dev` are in `docker/dev/Dockerfile` beside `libsecret-1-dev`. **CI is unaffected**:
  mobile e2e runs on Android and iOS, and `ci.yml` runs plain `flutter test`, which builds no native
  plugins.
- **Service** (`lib/services/oidc_service.dart`): `signInWithOidc` generates the verifier, opens the
  browser at `{basePath}/oidc/{name}/login?client=mobile&codeChallenge=…` (`basePath` already ends in
  `/api`), parses the code out of the callback, and POSTs it with the verifier. The response is
  `AppData` with `jwt` + `refreshToken` — the same shape as `login?tokensInBody=true` — so it goes
  through **`storeAppData` unchanged**. A `PlatformException` with code `CANCELED` becomes
  `OidcSignInCancelled` and is swallowed; the backend's error codes map to user copy in
  `oidcErrorMessage` (it only ever sends a small fixed vocabulary, never an IdP's own text).
- **Buttons** render in `lib/auth/login/widgets/auth_form.dart` from
  `authModel.featureConfig.oidcProviders`, already populated — the Connect-to-Server screen fetches
  `/featureConfig` unauthenticated before `/login`, so there is no extra request.
- **Connected accounts** (`lib/profile/widgets/connected_accounts.dart`): Connect calls
  **`linkWithOidc`, never `signInWithOidc`** — and the difference is the whole feature. A login flow
  asks the server *"who is this?"*, so running one from the profile screen provisions a brand-new
  account (or, with `linkByUsername` on, attaches the identity to whichever account holds that
  username) instead of connecting this one — and the button is only ever rendered for providers this
  account has **not** connected, which is exactly the case where that happens. It shipped wired to
  the login flow and reported "Account connected" while doing it.
  - **The start is an API call, not a navigation.** `GET /oidc/link/{name}?client=mobile` returns
    `{launchUrl}` as JSON. It cannot be a redirect: this app authenticates with a **bearer token**
    (`auth_interceptor.dart`), which the browser RFC 8252 requires has no way to carry, so a 302 into
    it arrives unauthenticated.
  - **What comes back points at OUR API, not the identity provider, and that matters.** It first
    shipped returning the provider's authorization URL for the app to open directly, which meant the
    browser never touched our server before the callback — so the server had nothing to bind the flow
    to. Providers leak that URL by Referer from their own login pages, and anyone holding it could
    authenticate as themselves and have their identity attached to this account. Opening a launch URL
    instead lets the server set a binding cookie on the browser that will return with the callback.
    Do not "simplify" this back into one hop. See `api/CLAUDE.md` → "Linking an existing account" for
    why PKCE cannot substitute (the challenge would be the victim's own, and Android's `exported`
    `CallbackActivity` lets an attacker deep-link a result into the waiting flow).
  - The callback returns to `io.receiptwrangler://oidc?linked={name}`, parsed by
    `extractLinkedProviderFromCallback` — a **link carries no `code`**, because it mints no session.
  - **No `codeChallenge` is sent**, unlike a login — but not because the bearer token on the start
    "already proved who the caller is", which was the original and wrong reasoning. It is because a
    link hands the app back nothing redeemable; the flow is bound by cookie instead.
  - Disconnect is hidden for a provisioned account's **last** connection, matching the server's
    lockout guard; the server refuses it either way. The flag is an **account** fact replicated onto
    every identity, so this per-connection predicate is also the account-level one — see
    `api/CLAUDE.md`. A failed load yields no section rather than an
    error, so an older server or a caller without `app.account.read` degrades quietly.
- **`app_links` is not involved** — `flutter_web_auth_2` captures the callback itself, so the existing
  deep-link handler in `lib/main.dart` and its test seams are untouched.
- **Tests:** `test/utils/pkce_test.dart` (the cross-language contract),
  `test/services/oidc_service_test.dart` (URL composition, escaping, callback parsing for **both** the
  login and link shapes — including that a bare `code` is not a successful link — and error mapping),
  `test/widgets/auth_form_test.dart` (a button per provider; none when the list is empty **or
  absent**), and `test/models/feature_config_oidc_ingest_test.dart` — which deserializes
  `featureConfig` with `oidcProviders` **absent, empty and populated**. That last one matters because
  the payload is fetched unauthenticated before login: a shape this client cannot read does not fail
  one screen, it reports itself as "Failed to connect to server", the same blast radius as the two
  documented login outages.

### App Links / Universal Links — server-URL pre-fill (login)

The desktop login page shows a QR encoding a **deep link** to
`https://receiptwrangler.io/app/setup`, with the server URL carried in the **URL fragment** as
`#url=<percent-encoded server url>` (e.g.
`https://receiptwrangler.io/app/setup#url=https%3A%2F%2Fdemo.receiptwrangler.io%2Fapi`). Scanning it
(or tapping it on device) opens this app and **pre-fills** the "Server URL" field on the Connect-to-Server
screen (`/`, `SetHomeserverUrl`). The **same** link also works when scanned by the in-app QR scanner.
We deliberately **do not auto-connect** — the user reviews the URL and taps Connect (phishing
mitigation, matching the plain-QR path above).

- **Dependency:** `app_links` (`^6.4.1`) handles the links ourselves rather than letting go_router try
  (and fail) to route `/app/setup`. We intentionally do **not** set `flutter_deeplinking_enabled`.
  Pinned to the **6.x** line on purpose: `app_links` 7.x requires Dart `>=3.12.0`, but this app's
  `environment.sdk` floor is `>=3.7.0` (6.4.1 needs only `>=3.5.0`), and the 6.x API
  (`getInitialLink()` / `uriLinkStream`) is identical to 7.x. Bumping to 7.x would require raising the
  Dart floor — do that deliberately if/when the app moves to Flutter 3.44+.
- **Extractor** (`lib/utils/url.dart`): `extractDeepLinkServerUrl(String raw)` requires
  `host == receiptwrangler.io` and `path == /app/setup`, reads `Uri.splitQueryString(uri.fragment)['url']`,
  and passes that value back through the existing `normalizeServerUrl` gate (so http/https +
  non-empty-host validation is reused). Returns null for anything that isn't this deep link. Covered by
  `test/utils/url_test.dart`.
- **Deep-link handler** (`lib/main.dart`, `_ReceiptWrangler`): subscribes via `app_links` in `initState`
  — `getInitialLink()` for cold start (stashed immediately so it survives the `FutureBuilder` first-paint
  gate) and `uriLinkStream` for warm/resumed. For each URI it runs `extractDeepLinkServerUrl`; on a match
  it sets `AuthModel.pendingServerUrl` and routes to `/`. The stream subscription is cancelled in
  `dispose`. Both sources are **injectable for tests** — `buildApp({initialDeepLink, deepLinkStream})`
  → `ReceiptWrangler`, and `_initDeepLinks` falls back to the real plugin
  (`widget.initialDeepLink ?? await _appLinks.getInitialLink()`, `widget.deepLinkStream ??
  _appLinks.uriLinkStream`). `main()` passes neither. The seam exists because the e2e process runs **on**
  the device and can't ask the OS to open a URL; mocking `app_links`' private channels instead would
  couple the spec to plugin internals.
- **Pre-fill** (`AuthModel.pendingServerUrl` + `SetHomeserverUrl`): the handler stashes the URL on
  `AuthModel.pendingServerUrl` (a nullable field with `setPendingServerUrl` / `clearPendingServerUrl`,
  both `notifyListeners`). `SetHomeserverUrl` consumes it in `build`: when non-null it `patchValue`s the
  `url` field in a post-frame callback (the FormBuilder state isn't attached during the first build) and
  clears it. This covers both the **cold-start** case (value present when the widget first mounts) and
  the **warm** case (value arrives later → listener fires → rebuild). Covered by
  `test/widgets/set_homeserver_url_test.dart`.
- **In-app scanner reuse** (`set_homeserver_url.dart` `_onScanPressed`): resolves the scanned value with
  `extractDeepLinkServerUrl(raw) ?? normalizeServerUrl(raw)`, so the deep-link QR and a plain server-URL
  QR both fill the field.
- **Redirect interaction:** the existing `/`→`/groups` auth redirect
  (`guards/auth-guard.dart` `unprotectedRouteRedirect`) means the pre-fill only surfaces for
  **unauthenticated** sessions — a logged-in user is already set up, so bouncing them to `/groups` is
  intended. Do not try to defeat the redirect.
- **Native config:**
  - **Android** (`android/app/src/main/AndroidManifest.xml`): a second `<intent-filter>` on
    `MainActivity` with `android:autoVerify="true"` (VIEW/DEFAULT/BROWSABLE) and
    `<data android:scheme="https" android:host="receiptwrangler.io" android:pathPrefix="/app/setup"/>`.
    App Links verify against the installed `applicationId` (`io.receiptwrangler`), so no namespace change
    is needed. No new permissions.
  - **iOS** (`ios/Runner/Runner.entitlements`): `com.apple.developer.associated-domains` =
    `[applinks:receiptwrangler.io]`, wired via `CODE_SIGN_ENTITLEMENTS = Runner/Runner.entitlements;` into
    **all three** Runner build configs in `project.pbxproj` — Debug, Release **and Profile** (not
    RunnerTests). Profile matters because `flutter run/build --profile` binaries would otherwise ship
    without the associated-domains entitlement and silently fail to verify Universal Links.
    Universal Links don't need an Info.plist change. Bundle id `io.receiptwrangler`, Apple Team ID
    `3VD3YNZ3KA` (already set).
  - The hosted `.well-known/assetlinks.json` / `apple-app-site-association` files live on the
    **receiptwrangler.io host** (nginx), not in this repo — nothing here can verify them, so a broken
    association shows up only on a real device.
- **E2E** (`integration_test/login_qr_deep_link_test.dart`): the cross-component contract test. It
  enables the login QR for the run's own backend (`enableLoginQrForTest` in
  `helpers/login_qr_fixtures.dart` — a **global** system-settings mutation, restored on teardown), reads
  the **real** `loginQrUrl` the Go API composed off the unauthenticated `GET /featureConfig`, and feeds
  that exact string to the app through the `buildApp` seam. Three cases: cold start (prefill → assert the
  Go-escaped fragment round-trips to the exact server URL → **no auto-connect** → Connect → log in →
  `GroupSelect`), warm stream delivery, and wrong-host / wrong-path links being ignored (with a positive
  control so a dead subscription can't fake the pass). Go tests and the Dart unit tests each assert
  against their own hardcoded strings — this spec is the only thing pinning the two together. The deep-link
  stream fixture is a **single-subscription** `StreamController` on purpose: it buffers events until
  `_initDeepLinks` attaches its listener (which happens after an `await`), where a broadcast controller
  would silently drop them.

### Testing

Run tests with `flutter test`. Run a single file with `flutter test test/path/to/file_test.dart`.

**All new code must have accompanying tests.** When adding a new widget, utility, model, or service, add a corresponding test in `test/` that exercises:
- The happy path
- Sign / boundary cases (negative, zero, empty) where applicable
- Wiring contracts (validators, keyboard types, transformers) that downstream code depends on

Existing reference tests:
- `test/services/token_refresh_service_test.dart` — service unit tests with mocktail
- `test/widgets/amount_field_test.dart` — widget tests with FormBuilder + Provider
- `test/utils/currency_test.dart` — pure utility tests
- `test/helpers/widget_test_helpers.dart` — shared widget-test setup helpers
- `test/helpers/auth_test_helpers.dart` — shared mocks and JWT builders

#### Directory layout
Mirror the `lib/` tree: `test/widgets/` for widget tests of `lib/shared/widgets/...`, `test/utils/` for `lib/utils/...`, `test/services/` for `lib/service[s]/...`, `test/interceptors/` for interceptors. Shared helpers go in `test/helpers/`.

#### Flutter widget-test best practices

These patterns are followed by the existing tests; new tests should keep to them:

- **Use `testWidgets` (not `test`) for widget tests.** It supplies the `WidgetTester` and binds the framework.
- **Locate by `Key`, not by widget type.** Pass a `ValueKey` to the widget under test and use `find.byKey(...)`. When you need a specific descendant (e.g. the inner `FormBuilderTextField` of an `AmountField`), use `find.descendant(of: find.byKey(...), matching: find.byType(...))`. `find.byType(...)` alone breaks as soon as another instance lands in the tree.
- **Prefer `pump()` over `pumpAndSettle()`.** `pumpAndSettle` waits for *all* frames to drain and will time out against any continuous animation or formatting-on-change controller (e.g. `currency_textfield`). Reach for `pumpAndSettle` only when a specific test introduces an animation that has to flush.
- **Inject ChangeNotifier dependencies with `ChangeNotifierProvider`.** Use the `create:` constructor when the test owns the instance (auto-disposes); use `.value(value: existing)` only when the test reuses a model created elsewhere.
- **Prefer real model instances over mocks** when the model has no I/O and reasonable defaults (e.g. `SystemSettingsModel`). Mocking via mocktail is for models with I/O or where you need to verify interactions.
- **Only call `registerFallbackValue` when stubs use `any()` matchers.** Concrete `when(() => mock.x()).thenReturn(...)` does not need fallback registration.
- **Don't `tester.enterText` against `currency_textfield` (or any input with a controller that intercepts/reformats keystrokes).** It's fragile across package versions. Test the read path via `initialAmount` round-tripped through `valueTransformer`, and test the write path by inspecting the widget's `keyboardType`.
- **Register the custom currency in `setUpAll`** before any test that calls `exchangeCustomToUSD` / `exchangeUSDToCustom`. The shared helper `registerCustomCurrencyForTests()` in `test/helpers/widget_test_helpers.dart` is idempotent — call it once per test file.
- **Skip golden tests** unless the component is visually critical and the team is set up to maintain reference images.

#### Workflow

1. Write the test alongside the change.
2. `flutter analyze` — must be clean on the new files (the codebase has pre-existing warnings; only check the files you touched).
3. `flutter test` — must be all green.
4. If a test surfaces a real production bug (it happens — e.g. `Money.parse` of a leading `-` against the USD pattern), fix the bug as part of the same change rather than skipping the test.

### E2E Testing

End-to-end tests live in `integration_test/` (sibling of `test/`) and use Flutter's first-party **`integration_test`** package. They drive the real app against a running Go API, mirroring the desktop Playwright suite under `desktop/e2e/`.

**Stack choice:** `integration_test` SDK package. Not Patrol (we don't need native permission dialogs yet). Not the deprecated `flutter_driver`.

**Supported targets:**
- **Local Android emulator** via `./run-e2e-android.sh` (macOS, auto-boots an AVD).
- **Local iOS Simulator** via `./run-e2e-ios.sh` (macOS, auto-boots a sim).
- **Local Linux desktop** via `./run-e2e.sh` (containers/CI Linux). Originally the primary target; kept for the dev container's headless flow.
- **CI Android + iOS** via `.github/workflows/mobile-e2e.yml`, currently **advisory** (`continue-on-error: true`). Triggers: `push` to `main` (post-merge), `push` to `tech/mobile-e2e` (iteration on the e2e setup itself), and `workflow_dispatch`. Deliberately **not** `pull_request` — the suite is advisory and slow, so it runs post-merge / on demand rather than gating every PR. The formerly skipped specs (`receipt_comments_test`, `receipt_cost_split_test`) are un-skipped and green — the product bugs they tracked are fixed — so nothing blocks flipping `continue-on-error` once CI demonstrates stability.

Screenshot/video capture on failure is still deferred — see the "Out of scope" note at the bottom of this section.

#### Prerequisites

**Per target:**

| Target          | Prereqs                                                                                                              |
| --------------- | -------------------------------------------------------------------------------------------------------------------- |
| Android (mac)   | Android SDK (Studio or `cmdline-tools`); at least one AVD; `coreutils` on PATH (`brew install coreutils`, for `gtimeout`). |
| iOS (mac)       | Xcode + iOS Simulator (`xcrun simctl` on PATH); `coreutils` on PATH. The iOS Simulator runtime matching your installed Xcode is required — see "iOS 26.x runtime gotcha" below. |
| Linux desktop   | `apt-get install -y --no-install-recommends libsecret-1-dev xvfb` (`libsecret-1-dev` to *build* `flutter_secure_storage_linux`, `xvfb` to *run* headlessly — `run-e2e.sh` auto-wraps in `xvfb-run` when `$DISPLAY` is unset). Plus `flutter config --enable-linux-desktop`. |

**Shared, one-time:**

1. `cd mobile && flutter pub get`
2. Seed the two e2e users by running `./api/dev/seed-e2e-users.sh` (idempotent — safe to re-run). It logs in as the default `admin/admin` user that `MakeMigrations()` auto-creates on a fresh DB, then calls the admin-protected `POST /user/` to create:
   - `e2e-admin` with role `ADMIN`
   - `e2e-user` with role `USER`

   The script uses creds matching `api/dev/switch-to-sqlite.sh` (and the mariadb/postgresql variants), so a later `source` of those scripts gives the runners credentials that line up with what's seeded. Override `ADMIN_USERNAME` / `ADMIN_PASSWORD` if the default admin's password has been changed; override `API_BASE_URL` to seed a non-local backend.

   Why API-seed instead of the UI: `enableLocalSignUp` is `false` locally so the UI signup 404s, and even if enabled, the auto-`admin/admin` already holds the "first user = ADMIN" slot — a subsequent UI signup for `e2e-admin` would land as USER. `POST /user/` requires admin auth but accepts an explicit `userRole`, so it bypasses both gotchas.

**Every run:** start the Go API separately (`cd api && go run main.go`). None of the runners start the API — same pattern as Playwright.

#### Running locally

Three runners, one per target. Each accepts optional spec paths; with no args it iterates every `*_test.dart` under `integration_test/`.

```bash
# Android emulator (mac dev primary target):
cd mobile && ./run-e2e-android.sh
cd mobile && ./run-e2e-android.sh integration_test/smoke_login_test.dart

# iOS Simulator:
cd mobile && ./run-e2e-ios.sh
cd mobile && ./run-e2e-ios.sh integration_test/smoke_login_test.dart

# Linux desktop (containers, CI Linux):
cd mobile && ./run-e2e.sh
```

All three runners source `api/dev/switch-to-sqlite.sh` for the four `E2E_*` credentials, write a temp dart-define JSON, and invoke flutter against the suite. **What differs:**

| Runner             | Target          | Invocation                       | Default `E2E_BASE_URL`           |
| ------------------ | --------------- | -------------------------------- | -------------------------------- |
| `run-e2e-android.sh` | Android emulator | `flutter drive` (one per spec) | `http://10.0.2.2:8081/api` (host loopback alias) |
| `run-e2e-ios.sh`   | iOS Simulator   | `flutter drive` (one per spec)   | `http://localhost:8081/api`      |
| `run-e2e.sh`       | Linux desktop   | `flutter test` (one per spec)    | `http://localhost:8081/api`      |

**Mobile-runner-specific env overrides:**
- `E2E_ANDROID_AVD` — AVD name (default `Pixel_3a_API_34_extension_level_7_arm64-v8a`). Auto-booted with `-no-snapshot-save -no-boot-anim` if no emulator is attached; the script waits for `sys.boot_completed=1` (5 min ceiling).
- `E2E_IOS_DEVICE` — simulator device name (default `iPhone 15`). Auto-booted via `xcrun simctl boot` + `bootstatus`. The grep anchor (`^ *<name> (`) excludes "iPhone 15 Pro" / "Pro Max" siblings, so set the exact name.
- `E2E_MOBILE_BASE_URL` — overrides the table defaults. Useful for pointing at a remote backend, e.g. `E2E_MOBILE_BASE_URL=https://demo.receiptwrangler.io/api ./run-e2e-android.sh`.

**Why `flutter drive` on mobile, `flutter test` on Linux:** on the Android/iOS path, `flutter test integration_test/...` repeatedly hit "Integration tests and unit tests cannot be run in a single invocation" even on a single file. Every working android-emulator-runner CI example uses `flutter drive` with an explicit driver + target (`test_driver/integration_test.dart` calls `integrationDriver()`), so we follow that pattern on mobile. Linux desktop works fine with `flutter test`.

**Why one-`flutter drive`-per-spec on mobile:** the top-level `GoRouter` in `lib/main.dart` is a final global — its location persists across `testWidgets` calls within the same flutter process. Spec N+1 inherits spec N's last URL and 403s on bootstrap. Looping one `flutter drive` per spec gives each its own process. Between specs, the mobile scripts force-stop and uninstall by bundle id `io.receiptwrangler` (flutter drive's own cleanup uninstalls by the namespace `com.example.receipt_wrangler_mobile`, which doesn't match the real package), `pkill -f dartvm` to free leaked dart isolate hosts that own the vmservice port, and wrap each spec in `gtimeout 600` so a hung run doesn't eat the whole job.

**Auto-discovery:** the mobile scripts walk a small list of candidate Flutter install dirs (`~/Documents/flutter/bin`, `~/flutter/bin`, `/opt/flutter/bin`, `/usr/local/flutter/bin`) if `flutter` is not on `$PATH`, and resolve the Android SDK from `$ANDROID_HOME` → `$ANDROID_SDK_ROOT` → `~/Library/Android/sdk`. They prepend `platform-tools` and `emulator` to PATH for the run.

**Devices are left running on exit.** The scripts don't shut down the emulator/simulator — reruns reuse the booted device for speed. To force a cold boot next time: `adb emu kill` (Android) or `xcrun simctl shutdown <udid>` (iOS).

#### How env vars reach the tests

`String.fromEnvironment` is a `const` constructor — the **key has to be a literal**, so you cannot build it dynamically per role. `integration_test/helpers/env.dart` declares all five `E2E_*` reads as `static const` fields and exposes `E2eEnv.assertAdmin()` / `assertUser()` to fail fast when vars are unset.

**Never use `Platform.environment`** — it returns an empty map on Android/iOS. `--dart-define` is the only portable mechanism.

**Base URL gotcha:** the desktop suite's `E2E_BASE_URL=http://localhost:4200` points at the Angular dev server, whose proxy forwards `/api` to the Go backend. The mobile app has no proxy — it hits the API directly. `run-e2e.sh` therefore reads `E2E_MOBILE_BASE_URL` (defaults to `http://localhost:8081/api`) and maps it into the `E2E_BASE_URL` dart-define the test sees. Override for remote targets: `E2E_MOBILE_BASE_URL=https://demo.receiptwrangler.io/api ./run-e2e.sh`.

#### Writing tests

- **Bootstrap:** call `await tester.pumpWidget(buildApp())` (imported as `import 'package:receipt_wrangler_mobile/main.dart' show buildApp;`). `buildApp()` returns a fresh `MultiProvider` + `ReceiptWrangler` widget tree, with a per-`State` `late final GoRouter` so router location does not leak across `testWidgets`. Do NOT call `app.main()` from a test — `main()` triggers `runApp()` and `FlutterNativeSplash.preserve`, which conflicts with the test binding.
- **`IntegrationTestWidgetsFlutterBinding.ensureInitialized()`** at the top of `main()` in every spec file. Required — `testWidgets` without it runs as a unit test and fails to reach native channels.
- **Gate `installLinuxDesktopMocks()` on `Platform.isLinux`** (from `integration_test/helpers/platform_mocks.dart`), right after the binding. It stubs three mobile-only plugins whose method channels are unimplemented on Linux desktop and would otherwise throw `MissingPluginException` during app bootstrap:
  - `permission_handler` (channel `flutter.baseflow.com/permissions/methods`) — camera permission request in `lib/utils/permissions.dart`.
  - `gal` (channel `gal`) — image-gallery access in the same helper.
  - `flutter_secure_storage` (channel `plugins.it_nomads.com/flutter_secure_storage`) — backed by an in-memory map; real libsecret would need an unlocked gnome-keyring + dbus session, which is fragile in containers/CI.

  On Android/iOS these plugins have real native implementations and must be hit directly, so the `if (Platform.isLinux) { installLinuxDesktopMocks(); }` gate in `smoke_login_test.dart` is the template — copy it into every new spec.
- **Never use `pumpAndSettle` on the bootstrap frame.** `main.dart` renders a `CircularProgressIndicator` inside a `FutureBuilder` during auth init; the indicator's animation means `pumpAndSettle` never returns. Use `pumpUntilFound` (from `integration_test/helpers/pump.dart`) instead — it polls until a target finder hits, with a timeout.
- **Locators:**
  - `FormBuilderTextField` has no Key; match by its `name` field:
    ```dart
    find.byWidgetPredicate((w) => w is FormBuilderTextField && w.name == 'username')
    ```
  - `CupertinoButton.filled` with a `Text` child is `find.widgetWithText(CupertinoButton, 'Log In')`.
- **Assert navigation by widget presence**, not URL. After login, `pumpUntilFound(find.byType(GroupSelect))` is stronger than reading the go_router state — the widget is present iff the `/groups` shell has mounted.
- **Each test cold-boots.** There is no Flutter equivalent of Playwright's `storageState`. When the suite grows past a handful of specs, either accept the per-test login cost or introduce a non-UI setup step. Don't hand-write state sharing between tests.

#### Caveats / things that will bite

- **Three tap-flake patterns** (each cost a debugging session on the Android emulator; recognize them by a
  "derived an Offset ... that would not hit test on the specified widget" warning followed by a downstream timeout).
  iOS makes pattern 2 **deterministic** where Android only flaked — Cupertino page/popup transitions are slower
  (~400ms), so any tap site without the hitTestable-wait + drain fails every run on the simulator:
  1. **`tester.ensureVisible` does not pump.** It jumps the scroll position but the widget's global offset only
     updates after a relayout, so an immediate `tap()` computes the stale (off-screen) center. Always
     `await tester.pump(...)` (or `pumpAndSettle`) between `ensureVisible` and `tap`.
  2. **Modal sheets and popup menus mount content on the animation's first frame.** `pumpUntilFound(find.text(...))`
     returns while the sheet is still sliding in / the popup is still scaling, and the tap lands where the item
     *was*. Wait on `finder.hitTestable()` and then drain a few frames (`for (...) pump(100ms)`) before tapping —
     see `addManualReceiptViaUI` and `receipt_cost_split_test._navigateToEdit` for the canonical shape.
  3. **Snackbars absorb taps on bottom-of-screen buttons.** The root ScaffoldMessenger renders snackbars ABOVE
     modal bottom sheets, so e.g. "Receipt added successfully" covers a sheet's bottom submit button for ~4s.
     `pumpUntilFound(tester, button.hitTestable())` resumes exactly when the snackbar departs.
- **Destination markers must be unique to the destination.** `find.text('Name')` matches on BOTH `/view` and
  `/edit` receipt forms, so it cannot prove an Edit navigation happened — use `find.byType(BottomSubmitButton)`
  (only mounted on edit/add paths) instead.
- **Quick Scan image input on Linux:** the old `"Unsupported platform"` throw is **gone** — `getGalleryImages` and its `Platform.operatingSystem` switch were deleted (see "Picking receipt files"), so `installFileSelectorMock()` now reaches the file source on desktop and `quick_scan_test.dart` / `receipt_add_gallery_test.dart` / `receipt_add_partial_image_test.dart` run un-skipped — all three verified green on the Linux runner (`receipt_add_partial_image_test.dart` was since rewritten for the atomic create — see "Role-required receipt fields & single-call create" — and is green on Linux again). `receipt_add_gallery_test.dart` needed the `hitTestable()` + frame-drain hardening before its popup-menu tap (tap-flake pattern 2 above); it had never run on Linux, so it had never needed it. **Assert the file source, not the photo one:** `image_picker_linux` is implemented on top of `file_selector_linux`, so on Linux the file-selector mock intercepts *both* and no Linux e2e can tell the two apart. The **document scanner** (`installDocumentScannerMock()`, `openQuickScanImageForm` in `helpers/quick_scan_actions.dart`) is still the way in when a spec wants a source that is neither picker.
- **Reaching the receipt entry points:** a **tap** on the scan slot is a direct action, so manual entry is reached by **holding** it — `openManualReceiptForm(tester)` (`helpers/receipt_test_helpers.dart`) is the shared path, and it works on every screen and in every flag state (the receipts-screen overflow menu does not). Use `scanNavSlot()` rather than `find.text('Add')` when you only need to *reach* the slot: its label is "Scan" or "Add" depending on the caller's gates.
- **Driving camera permission states:** set `debugCameraAccessOverride` (`lib/utils/permissions.dart`) rather than swapping the permission channel mock — login bootstrap also touches permission_handler, and the override pins only the branch under test. See `quick_scan_camera_denied_test.dart`. The suite uses the first-party `integration_test` package, which **cannot** drive native OS permission dialogs; that would need Patrol.
- **Shared channel mocks live in `test/helpers/channel_mocks.dart`,** not here: the widget suite is the gating one and must not import from `integration_test/`. `helpers/platform_mocks.dart` re-exports them, along with `test/helpers/image_picker_mock.dart` (`installImagePickerMock` / `installFailingImagePickerMock`) which lives there for the same reason. `installPermissionMocks(status:, requestStatus:)` is the parameterised variant (`PermissionStatusWire` names the wire ints); `installCameraGalleryPermissionMocks()` is the always-granted one the scanner path needs.
- **Document-scanner mock must grant camera permission on every platform:** `CunningDocumentScanner.getPictures` requests `Permission.camera` **itself** (Dart-side, via `permission_handler`) before invoking its native channel, so `installDocumentScannerMock()` also calls `installCameraGalleryPermissionMocks()` (extracted from `installLinuxDesktopMocks`) on **all** platforms — not just Linux. Without it, iOS/Android hit the real permission_handler: the fire-and-forget `requestPermissions()` in `main.dart` leaves an app-init camera request **pending** (its native dialog is never dismissed in a headless test), and `getPictures`' own camera request then collides with it → `PlatformException(ERROR_ALREADY_REQUESTING_PERMISSIONS)`. Granting up front makes both requests resolve instantly with no dialog. Only this scan path needs the mobile grant; every other spec hits permission_handler natively (one fire-and-forget request, never a second) and stays green. `flutter_secure_storage` stays **real** on iOS/Android (only Linux mocks it).
- **Linux build linker/ar:** the desktop build resolves its toolchain from the installed clang's dir (e.g. `/usr/lib/llvm-19/bin`). With only `clang` installed you get `Failed to find any of [ld.lld, ld]` then `[llvm-ar, ar]` — install `lld` **and** the matching `llvm` package (see the Flutter SDK Setup apt line above) so `ld.lld` / `llvm-ar` land in that dir.
- **Headless display:** Flutter Linux desktop apps render through GTK and exit immediately without a display. `run-e2e.sh` auto-wraps in `xvfb-run` when `$DISPLAY` is unset. If you see "The log reader stopped unexpectedly, or never started," your display setup isn't working — check `xvfb-run --help` or set `DISPLAY` to a real X server.
- **`libsecret-1-dev` at build time:** the `flutter_secure_storage_linux` plugin's CMakeLists.txt does a `pkg_check_modules(libsecret-1>=0.18.4)` — if the dev headers aren't installed, the build fails with "The following required packages were not found: libsecret-1". Installed as a prereq above.
- **`libsecret` at runtime is avoided via mocks.** We don't bring up gnome-keyring + dbus for tests. `installLinuxDesktopMocks()` intercepts the platform channel with an in-memory map. If you ever want to exercise the real storage path (e.g. to reproduce a token-persistence bug), start a dbus session + gnome-keyring-daemon before the test — but don't do that by default; it adds a lot of fragile state.
- **Install-wide settings are AMBIENT, not defaults — pin them, don't assume them.** `showLoginQr`,
  `mobileServerUrl` and `receiptProcessingSettingsId` (the `aiPoweredReceipts` flag) are single
  install-wide values shared by every spec, every runner and every CI job on a backend. A spec that
  needs one in a particular state must **declare it with a fixture**; a comment reasoning about "the
  default local backend" is not a precondition. This is not theoretical: a run aborted on 2026-07-06
  left the AI pointer aimed at a leaked `e2e-ai-flag-*` record, so the flag was stuck **on for eight
  weeks**, and the two specs that silently assumed it was off failed with
  `Timed out after 10s waiting for text "Name"` — which reads exactly like a product bug, since the
  scan slot really *was* opening a scanner instead of falling through to the manual form.
  Corollary for the fixture side: **any fixture that mutates a global must be able to tell its own
  leaked artifacts from real configuration**, and heal rather than replay the former — otherwise the
  restore faithfully re-wedges the backend for the next spec. `feature_flags.dart` does this with the
  `e2e-ai-flag-` name prefix, and heals in **setup**, not teardown (the state being healed was itself
  produced by a teardown that never ran).
- **Go API rate-limiter:** login is rate-limited. Rerunning the same test in tight succession can 429 — give it a few seconds between runs. The desktop suite notes the same issue in `desktop/e2e/helpers/auth.ts`.
- **DB accumulation:** tests write real rows (sessions, refresh tokens). Fine for a smoke test; when specs start creating receipts/groups/etc., build per-test uniqueness (UUIDs) into the data, mirroring the Playwright conventions.
- **Never commit credentials or the generated JSON.** `.e2e-env.json` is gitignored as belt-and-suspenders — the script already uses `mktemp`.
- **iOS 26.x runtime gotcha:** Xcode 26.x ships with only its own SDK (e.g. Xcode 26.5 → iOS 26.5 SDK only). Until that exact-version *simulator runtime* is installed, `xcodebuild -showdestinations` returns **zero** eligible destinations for this project — no sim on any iOS version is buildable, even though older runtimes (17.x / 18.x) show up under `xcrun simctl list runtimes`. Symptom: `run-e2e-ios.sh` reports `Unable to find a destination matching the provided destination specifier: { id:<udid> }` for whatever sim it picks. Fix: `xcodebuild -downloadPlatform iOS` (about 8GB, ~10–15 min). Once installed, the older simulators become buildable too.

#### Reference files

- `integration_test/smoke_login_test.dart` — canonical smoke test.
- `integration_test/quick_scan_entry_test.dart` — the scan slot's tap (captures → seeded sheet) and hold (menu contents), against real roles, plus the receipts-screen overflow: one case asserting its labels and one **tapping** its Quick Scan through to the sheet (the app-bar teardown regression — see "`TopAppBar`'s loading bar is always mounted").
- `integration_test/quick_scan_app_data_refresh_test.dart` — a group's comment field switched on **after** login reaches the running app when Quick Scan is next started. The only spec that proves the pre-scan AppData reload; every other quick-scan spec arranges its config before login, so none can distinguish a client that re-fetches from one that read the config once at login. Persists via the API and never touches `GroupModel`.
- `integration_test/quick_scan_entry_gated_test.dart` — the tap falling through to the manual form, with the right banner for each of the two blocked reasons. Exercises **both** feature-flag fixtures: the permission case enables the flag (to isolate the permission from it), the ai-disabled case **pins it off** with `disableAiPoweredReceiptsForTest()` rather than assuming — which also removes its dependency on the preceding test's teardown having succeeded, since they share the one global.
- `integration_test/receipt_entry_hidden_test.dart` — a Legacy Viewer gets no scan slot and no overflow at all. Flag-independent: the slot is hidden when the caller holds *neither* entry permission, whatever `aiPoweredReceipts` says.
- `integration_test/receipt_entry_menu_reopen_test.dart` — the long-press regression guard (`onLongPressUp`, not `onLongPress`): three open/dismiss cycles then a plain tap, proving the slot isn't left dead. **Pins `aiPoweredReceipts` off** — it logs in as the admin (who holds `group.receipts.quick-scan`) and installs no document-scanner mock, so the closing tap only reaches the manual form while Quick Scan is unavailable.
- `integration_test/quick_scan_camera_denied_test.dart` — the denied / permanently-denied camera fallbacks (the Settings action appears only in the latter).
- `integration_test/permission_add_menu_test.dart` / `permission_receipt_edit_test.dart` — permission-gating coverage (add-menu gate, edit-popup gate, swipe-to-edit gate) using per-spec provisioned users/groups. The add-menu spec also gates **"Quick Scan"** on `group.receipts.quick-scan` (flag on via `enableAiPoweredReceiptsForTest`, isolating the permission from the `aiPoweredReceipts` flag): a Legacy Editor **minus** quick-scan hides it while "Add Manual Receipt" stays; a full Legacy Editor shows it.
- `integration_test/permission_search_test.dart` — search bottom-nav destination gated on `app.receipts.search` (deny via a custom app role minus that permission; allow via a Legacy User).
- `integration_test/permission_dashboard_redirect_test.dart` — group dashboards route gated on `group.dashboards.read` (deny → redirected to the receipts list via a custom group role minus that permission; allow via a Legacy Viewer). Landing is told apart by `GroupReceiptsList` vs `GroupDashboardWrapper`.
- `integration_test/permission_comments_test.dart` — comment **deny** paths on the edit-state comment screen: `group.comments.create` hidden → no input; `group.comments.delete` hidden → swipe-to-delete disabled. Members are provisioned from the **Legacy Editor** baseline (holds `group.receipts.update`, needed to reach edit state) minus the permission under test, via `provisionGroupMemberWithoutPermission(..., baselineRole: 'Legacy Editor')`. Its navigation lives in `helpers/nav.dart` (`openReceiptEditForm` / `openReceiptCommentsInEditMode`).
- `integration_test/receipt_role_requirements_test.dart` — role-required receipt fields against a real role (`provisionMemberWithReceiptRequirements` → `createRole(..., requireReceiptComment:, requireReceiptImage:)`). See "Role-required receipt fields & single-call create".
- `integration_test/permission_paid_by_visibility_test.dart` — group-role paid-by visibility: a member restricted to "their own receipts" (via `provisionPaidByOwnMember` → `createRole(..., includeOwnPaidReceipts: true)`) sees only their own receipt in the group list; the admin-paid receipt is filtered out server-side. Mirrors desktop `paid-by-visibility.spec.ts` (list axis).
- `integration_test/permission_receipt_category_visibility_test.dart` — non-admin sees the per-group **category and tag** catalogs in the receipt-form pickers (sourced from `groupCategories` / `groupTags`, not the admin-only flat lists).
- `integration_test/reports_list_test.dart` — the Reports **list** view: seeds a template via `createReportTemplate` and asserts the row renders (regression guard for the `aggFunc` omitempty deserialization fix), the "No reports found" empty state, and the avatar-menu gate (`app.reports.read`/`readAll` shown, Legacy User hidden). Uses `provisionUserWithAppPermissions` in `permission_fixtures.dart`.
- `integration_test/receipt_default_custom_fields_test.dart` — **group default custom fields** end to
  end: the swap matrix across a group change (with an API read-back of the saved values), the
  stale-value guard, and view-never / edit-only-on-change. See "Group default custom fields" above for
  the three gotchas it encodes. Seeds via `setGroupDefaultCustomFields` (`permission_fixtures.dart`) and
  `createCustomField` / `deleteCustomField` (`api.dart`) — the field teardown is registered **before**
  the user/group ones so LIFO runs it last, since deleting a field destroys every value stored against
  it and the receipts holding those values have to be cascaded away first.
- `integration_test/helpers/env.dart` — dart-define consumption + guards.
- `integration_test/helpers/pump.dart` — `pumpUntilFound` polling helper.
- `integration_test/helpers/platform_mocks.dart` — Linux-desktop platform-channel stubs for `permission_handler`, `gal`, `flutter_secure_storage`.
- `integration_test/login_qr_deep_link_test.dart` — the login-QR **deep link** end to end: reads the real `loginQrUrl` off `GET /featureConfig` and injects it via the `buildApp` seam. See "App Links / Universal Links" above.
- `integration_test/helpers/login.dart` / `api.dart` — UI + API login as admin, the shared e2e-user, or arbitrary credentials (`loginAs` / `apiLoginAs`). `login.dart` also exposes the two halves `loginAs` is built from — `resetPersistedAppState()` (wipe secure storage + `basePath` so the app boots to the Connect screen) and `loginFromLoginScreen()` (credentials + `GroupSelect` landing) — for specs that reach the login screen some other way. `api.dart` carries the shared `getSystemSettings` / `putSystemSettings` (the PUT is an **upsert**: patch a fetched object, never send a partial body) plus `overrideSystemSettingsForTest(jwt, settings, overrides:, restoreTo:)` — the one save/restore choreography behind **both** global-settings fixtures (`feature_flags.dart`, `login_qr_fixtures.dart`). It patches `overrides` onto the captured object on the way in and `restoreTo` onto a **fresh read** on the way out, so a concurrent change to an unrelated setting isn't clobbered. `restoreTo` is a required argument rather than defaulting to the captured values, because *what to put back* is a real decision — `feature_flags.dart` deliberately declines to replay a leaked pointer. Teardowns run **LIFO**, so anything that must run *after* the restore (e.g. deleting a record the settings still point at) has to be registered **before** the call.
- `integration_test/helpers/permission_fixtures.dart` — admin-API provisioning for permission specs. `PermFixture`
  carries the provisioned user's **`displayName`**, which the paid-by / charged-to dropdowns render —
  `users.dart`'s lookup helpers only cover the two fixed `E2E_*` accounts, so a spec that submits a form
  as a fixture user has no other way to name it. **`_settingsToCommand` carries a hardcoded key list** — a new `GroupReceiptSettings` flag must be added there or every persisted e2e config silently resets it to `false`. Also: fresh user + group with a chosen system group role ("Legacy Viewer"/"Legacy Editor"), optional seeded receipt, `addTearDown` cleanup. Also mints **custom roles** for negative specs: `createRole`/`deleteRole`, `rolePermissionsByName`, and the convenience `provisionUserWithoutAppPermission` / `provisionGroupMemberWithoutPermission` (build a role = a Legacy role **minus one permission**). The backend won't delete an assigned role, so the role-delete teardown is registered **before** the user/group ones — LIFO makes it run last, after the assignments are gone.
- `integration_test/helpers/feature_flags.dart` — `enableAiPoweredReceiptsForTest()` / `disableAiPoweredReceiptsForTest()`: the Quick Scan flag is `systemSettings.receiptProcessingSettingsId != null`, so **enable** points it at a junk `receiptProcessingSettings` record named `e2e-ai-flag-<micros>`. Both restore on teardown, and both must be called **before login** (`featureConfig` hydrates from AppData there). The flag is **install-wide, so OFF is ambient state, not a default** — a spec that needs it off must call the disable fixture rather than assume (see the caveat above). **The restore is self-healing:** if the captured pointer names an `e2e-ai-flag-` record, that capture is a leak from a run that aborted before its teardown, so **both** fixtures restore `null` instead of faithfully replaying it — replaying would re-wedge the backend for the very next spec. A pointer at any *other* record is a real AI configuration and is restored exactly. The name prefix is the whole distinction, which is why it lives in one `_fixtureRpsNamePrefix` constant. Healing runs in **setup**, not teardown: the state being healed was itself produced by a teardown that never ran. **Only `disable` also deletes the orphaned record**, and only when it was the live pointer at setup; `enable` heals the pointer but leaves the row, because under a concurrent run it cannot tell a leaked orphan from *another job's live fixture row*, and deleting the latter would leave a dangling pointer (id set, row gone) that the name check can never heal. An orphaned row is inert — the flag depends only on the pointer — and gets swept the next time a `disable` finds it live. `disable` **no-ops entirely** (no write, no teardown) when the flag is already off, so it costs nothing on a healthy backend. Both pointers are always written as a **pair**: the backend rejects a fallback without a primary, so nulling only the primary 400s on a backend that has one.
- `integration_test/helpers/login_qr_fixtures.dart` — `enableLoginQrForTest(serverUrl)`: flips the **global** `showLoginQr` / `mobileServerUrl` system settings, restores them on teardown (via the shared `overrideSystemSettingsForTest`, which patches onto a fresh read rather than replaying a stale snapshot — same helper `feature_flags.dart` uses), and returns the backend-composed `loginQrUrl`. Plus `getLoginQrUrl()` — the unauthenticated `GET /featureConfig` read. **These are the same two settings the desktop suite's `login-qr.spec.ts` drives on the same shared backend** (both workflows read `secrets.E2E_BASE_URL`), so the `android-e2e` job and `e2e.yml`'s `e2e` job share a job-level concurrency group (`e2e-shared-backend`, `cancel-in-progress: false`) to keep the two suites from racing. The group is **ref-independent on purpose** — this workflow also fires on `tech/mobile-e2e` / `workflow_dispatch` while `e2e.yml` fires only on `main`, so a `${{ github.ref }}`-scoped group would separate them and defeat the lock; keep both groups byte-identical. **`ios-e2e` deliberately carries no lock and runs concurrently against the same backend — that is safe only because its spec list contains nothing that mutates or depends on a global system setting** (`showLoginQr` / `mobileServerUrl`, and `receiptProcessingSettingsId` a.k.a. the `aiPoweredReceipts` flag). Adding such a spec to `ios-e2e` requires serializing it against `android-e2e` first (`needs: [android-e2e]` + `if: always()`) — **not** a second `e2e-shared-backend` member: GitHub keeps only one *pending* job per group, so a third contender cancels the older pending job outright and silently drops a whole suite's coverage.
- `integration_test/quick_scan_config_response_test.dart` — the Quick Scan per-image **form** shows/hides/requires fields per the selected group's `GroupReceiptSettings.quickScan*`, **without hitting the backend** (visibility assertions + client-side required-empty blocking, which short-circuits before the API call). Green on **all** targets (Linux/iOS/Android): it feeds an image via the mocked document-scanner channel (not the desktop-blocked gallery path) and injects the group config by mutating the live `GroupModel` via Provider (the same technique `quick_scan_disabled_test` uses for the AI flag — deterministic, no reliance on the local API persisting the fields). Six `testWidgets`: (1) single-group visibility + a categories-required fix-error snackbar; (2) **group switch** — two groups with opposite configs, asserts the field set **flips** when the dropdown changes; (3)+(4) **paid-by** / **tags** required+empty each block submit; (5) a required **comment** blocks submit; (6) `hideComments` hides the comment field even when the quick-scan toggle enables and requires it. Shared actions live in `helpers/quick_scan_actions.dart`. **Client injection only works for these no-backend tests** — successful submits must persist the config (see the submit spec). When injecting a synthetic group for a dropdown assertion, **trim `GroupModel.groups` to the all-group placeholder(s) + your target group(s)** (`keepOnlyGroup`) — the seeded admin accumulates many groups and an appended entry lands below the dropdown menu's viewport, where `find.text` can't reach it (a Material dropdown's scrollable doesn't build off-screen items into the element tree).
- `integration_test/quick_scan_submit_test.dart` — Quick Scan **submit** outcomes that actually POST to the API. The backend's `resolveQuickScanFields` validates against the group's **persisted** config, so these tests **persist** the config via `PUT /group/{id}/groupReceiptSettings` (`setGroupQuickScanConfig`, restored on teardown) instead of client-injecting — client (via AppData at login) and server then agree, as in production. Three `testWidgets`, all keeping paid-by/status **required** (making them optional needs a persisted default the backend enforces): (1) required paid-by+status filled + an **optional** category left empty still queues; (2) a **required** category **selected via the picker** queues — this is the on-device regression guard for the shellContext crash fix (see caveat below); (3) a **required comment** blocks submit while empty and queues once typed, which is the end-to-end proof the client sends the aligned `comments` array (an omitted or misaligned one would 400 against the persisted config). Uses `firstNonAllGroup` + `keepOnlyGroup`, seeds the category via `createCategory`, and drives the multi-select via the filter field + `ChoiceChip` (mirroring `receipt_cost_split_test`).
- `integration_test/quick_scan_prefill_test.dart` — Quick Scan per-image **prefill** from user preferences vs group config. Each added image seeds group/paid-by/status from `userPreferences.quickScanDefault*` (`_getInitialQuickScanValues`, delivered via AppData). Persists a group that **hides** paid-by (real prefs→AppData→form path via `setUserQuickScanPrefs` + `setGroupQuickScanConfig`, both restored) and asserts a **preset paid-by falls off** (field absent) while a **preset status is kept** (shown), then that submit **queues** (the hidden paid-by is backfilled from the group's `UPLOADER` default). Hiding paid-by/status in a persisted config **requires a group default** (`quickScanDefaultPaidByType: 'UPLOADER'` needs no id; `quickScanDefaultStatus` for status) — the backend rejects hiding/optional without one.
- `test/widgets/quick_scan_form_test.dart` — fast widget-level coverage of the same form logic: per-config visibility, paid-by/status required-vs-optional validators, null + non-null default configs, the group-switch behaviors (fields re-render per the new group's config; paid-by/categories/tags values clear on switch; a required validator is re-evaluated after switching), the picker tap-opens-with-null-shellContext crash guards, the **comment** cases (optional vs required validator, hidden by default / by `hideComments` / without `group.comments.create`, and surviving a group change), and **prefill vs config** (a prefilled paid-by/status shows when the group shows the field, and **falls off** — field absent — when the group hides it). Categories/tags "required" is **not** a field validator (enforced at submit in `quick_scan.dart`), so it's asserted in the integration spec, not here.
- `integration_test/helpers/document_scanner_mock.dart` — `installDocumentScannerMock()`: stubs the `cunning_document_scanner` channel's `getPictures` to return a fixed on-disk PNG **and** grants camera/gallery permission via `installCameraGalleryPermissionMocks()` on every platform (see the "Document-scanner mock" caveat above). This is the **only** way to add an image to the Quick Scan sheet on Linux desktop.
- `integration_test/helpers/platform_mocks.dart` — `installLinuxDesktopMocks()` (Linux-only: permission_handler + gal + flutter_secure_storage) and the extracted `installCameraGalleryPermissionMocks()` (camera + gallery grant, shared with the document-scanner mock on all platforms). Also re-exports the photo-picker fakes.
- `test/helpers/image_picker_mock.dart` — `installImagePickerMock()` swaps `ImagePickerPlatform.instance` for a fake returning a real on-disk file (same empty-filename rationale as the file-selector mock); `installFailingImagePickerMock()` is the throwing variant the failure-message specs use. **Widget tests must pass `bytes:` explicitly** (`quickScanTestPngBytes`) — `rootBundle` is not serviced there — **and must wrap the pick in `tester.runAsync`**, since `testWidgets` runs in fake-async where the mock's temp-file I/O never completes and the test hangs rather than failing. `integration_test/helpers/file_selector_mock.dart` gained a matching `installFailingFileSelectorMock()`.
- `integration_test/helpers/receipt_test_helpers.dart` — `addManualReceiptViaUI` (group-selectable), URL/id extraction, receipt cleanup.
- `integration_test/helpers/nav.dart` — group-entry and group-receipt navigation helpers.
- `test_driver/integration_test.dart` — `integrationDriver()` entrypoint that `flutter drive` uses.
- `run-e2e-android.sh` — Android runner; auto-boots an AVD, runs per-spec `flutter drive` loop with cleanup between specs.
- `run-e2e-ios.sh` — iOS runner; resolves a simulator UDID by name and boots it, runs per-spec `flutter drive` loop.
- `run-e2e.sh` — Linux desktop runner; wraps in `xvfb-run` when headless, invokes `flutter test`.
- `.github/workflows/mobile-e2e.yml` — CI counterpart; same command shapes the local mobile scripts mirror.
- `desktop/e2e/helpers/auth.ts` — Playwright counterpart; follow its conventions when adding new flows.

#### Out of scope (future work)

- Flipping the workflow from advisory (`continue-on-error: true`) to required — unblocked now that every spec runs un-skipped; do it after CI demonstrates stability.
- Screenshot / video artifact capture on failure.
- `storageState`-style auth warmup across a multi-spec suite.
- Additional specs (receipt CRUD, group management, logout).

### Build Configuration
- Android configuration in `android/` directory
- iOS configuration in `ios/` directory  
- Web configuration in `web/` directory
- Custom fonts (Raleway) configured in pubspec.yaml
- Native splash screen and launcher icons configured