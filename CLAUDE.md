# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Receipt Wrangler is a full-stack receipt management and splitting application with OCR-powered scanning, AI-assisted data extraction, and multi-user group management. This is a **monorepo** containing three main components:

- **api/** - Go backend service (port 8081)
- **desktop/** - Angular 19 web interface (port 4200 dev, port 80 production)
- **mobile/** - Flutter cross-platform mobile app
- **docker/** - Monolith Docker build configuration

Each component has its own CLAUDE.md with detailed component-specific guidance. This file covers monorepo-level architecture and workflows.

## Monorepo Architecture

### Component Communication
- **API Contract**: OpenAPI 3.1 specification in `api/swagger.yml` defines the API contract
- **Client Generation**: API clients are auto-generated from swagger.yml using `api/generate-client.sh`
  - Desktop: TypeScript Angular client → `desktop/src/open-api/`
  - Mobile: Dart Dio client → `mobile/api/`
  - MCP: TypeScript client for MCP integration
- **Development Flow**: Changes to API → update swagger.yml → regenerate clients → update frontend
- **MCP Server**: The Go API also hosts a native, OAuth 2.1-protected **MCP server** (off by
  default; enabled and configured at runtime via **System Settings** — `mcpEnabled` /
  `mcpPublicUrl`, no env var) so clients like Claude can read receipts/groups/etc. It starts and
  stops live without a restart. See `api/CLAUDE.md` → "MCP Server & OAuth 2.1". This is distinct
  from the generated MCP TypeScript client above.

### Technology Stack
- **Backend**: Go 1.26 with Chi router, GORM ORM, Asynq background jobs
- **Frontend**: Angular 19 with NGXS state management, Material + Bootstrap UI
- **Mobile**: Flutter with Provider state management, go_router navigation
- **Infrastructure**: Docker, nginx, PostgreSQL/MySQL/SQLite

## Docker Deployment

**All Docker build config lives in `docker/` — nothing else.** Those two Dockerfiles are what CI
builds images from (`.github/workflows/ci.yml` and `release.yml` both pass `file: ./docker/Dockerfile`
with `context: .`). Per-component `api/Dockerfile` and `desktop/Dockerfile` used to exist and were
removed as dead config; do not add them back. Because the build context is the repo root, a
`.dockerignore` is only ever consulted at the repo root too.

### Production Build (Monolith)
The `docker/Dockerfile` builds a single container with both API and web interface:
- Stage 1: Build Angular desktop app
- Stage 2: Build Go API and install dependencies (Tesseract, ImageMagick, Python)
- Final: nginx serves frontend, proxies `/api` to Go backend on port 80

### Development Build
The `docker/dev/Dockerfile` includes:
- All production components plus development tools
- SSH access for debugging (port 22, password: "development")
- Documentation site build from receipt-wrangler-doc repo
- Java runtime for OpenAPI generator
- Flutter SDK at `/opt/flutter` (on `PATH` via `ENV` and `/root/.bashrc`) with Linux desktop enabled and the `mobile/` pub cache warmed, plus `xvfb` + `libsecret-1-dev` so `mobile/run-e2e.sh` works out of the box

### Build Commands
```bash
# Production monolith
docker build -f docker/Dockerfile -t receipt-wrangler .

# Development container
docker build -f docker/dev/Dockerfile -t receipt-wrangler-dev .
```

## API Client Regeneration

When the API swagger.yml changes, regenerate clients:

```bash
# From api/ directory
./generate-client.sh desktop ../desktop/src/open-api
./generate-client.sh mobile ../mobile/api
./generate-client.sh mcp <output-path>
```

**IMPORTANT**: Never manually edit generated client code in `desktop/src/open-api/` or `mobile/api/`. Changes will be overwritten.

**Drift runs the other way too — the spec can fall behind the server.** `swagger.yml`'s `QueueName`
enum listed four of the Go side's five names (`system_clean_up` was missing) while
`UpsertSystemSettingsCommand.Validate` had always required a configuration for all five; the System
Settings form only worked because it builds its FormArray from the server's settings rather than from
the enum. Fixed when that queue took on the temp-file sweep. When you add a value to a Go enum the API
serializes, add it to `swagger.yml` in the same change — and note this direction is the *safe* one to
fix, since a client learning a value the server already sends can only stop failing on it.

**The same property cut the other way in the DATABASE.** Sourcing the FormArray from the server's
settings is what let the form survive a stale enum — and it is exactly what broke on an install whose
persisted `task_queue_configurations` predated the new queue: the server returned four rows, the form
submitted four, and `Validate` rejected the whole save with a 400. So a new queue name needs a
backfill on the *stored* side too, not just the spec. Both halves now live in
`repositories/system_settings.go`; see `api/CLAUDE.md` → "Temporary file retention & cleanup" →
"Upgrade path".

**Regenerate `mobile/api/` in the SAME change as any `swagger.yml` edit** — not "later". It is easy
to update the backend and desktop and forget mobile, because nothing fails: the Go tests pass, the
desktop compiles, and the drift is invisible until a released Android build hits the new payload.
That has caused **two production login outages** (2026-07-24, 2026-08-06). The Dart client is the
strict one — a value added to any closed enum on a response model fails the *whole* payload's
deserialization on every already-released binary. See `mobile/CLAUDE.md` → "Permission-based UI
gating" for the mechanism and the guard tests.

To check for drift at any time, compare *when* each was last touched — commit timestamps, since two
short hashes tell you nothing about ordering:

```bash
swagger=$(git log -1 --format=%ct -- api/swagger.yml)
client=$(git log -1 --format=%ct -- mobile/api)
[ "$swagger" -le "$client" ] && echo "mobile/api is current" || echo "mobile/api is STALE — regenerate"
```

**Generator on macOS:** `generate-client.sh` shells out to `npx @openapitools/openapi-generator-cli`,
which fails with `EACCES` when the global npm prefix is root-owned. Run the pinned jar directly
instead (version is in `api/openapitools.json`), which produces identical output:

```bash
curl -fL -o /tmp/openapi-generator-cli-7.10.0.jar \
  https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/7.10.0/openapi-generator-cli-7.10.0.jar
cd api
java -jar /tmp/openapi-generator-cli-7.10.0.jar generate -i swagger.yml -g dart-dio -o ../mobile/api
java -jar /tmp/openapi-generator-cli-7.10.0.jar generate -i swagger.yml -g typescript-angular -o ../desktop/src/open-api
cd ../mobile/api && flutter pub get && dart run build_runner build
```

`generate-client.sh` re-applies the four documented dart-dio patches itself, as the last step of a
`mobile` regen (`api/patches/apply-dart-dio-patches.sh`); it **fails the regen** rather than
returning an unpatched client if one no longer applies. Invoking the jar directly, as above, skips
that step — run the patch script by hand afterwards. Either way run `flutter analyze` **and
`flutter test`**: two of the four compile fine when missing (see `mobile/CLAUDE.md` → "Known
dart-dio regressions").

**Mobile regen without Flutter (e.g. the Claude Code web sandbox):** `mobile/api/pubspec.yaml` has
**no Flutter dependency**, so the standalone **Dart SDK** is enough to finish the regen — Flutter is
only needed for the app itself. Point `DART_SDK` at the copy inside an existing Flutter install, or
at a standalone SDK unpacked from the dart-archive, and confirm the version before generating
anything:

```bash
DART_SDK=/opt/flutter/bin/cache/dart-sdk   # Flutter's own Dart; or an unpacked dart-archive SDK
export PATH="$DART_SDK/bin:$PATH"
dart --version                             # confirm before regenerating (verified with 3.12.2)
cd mobile/api && dart pub get && dart run build_runner build && dart analyze
```

**Use the Dart SDK that ships inside the pinned Flutter** (`3.41.7` — `.github/workflows/ci.yml`,
`docker/dev/Dockerfile` `FLUTTER_VERSION`); a mismatched SDK is the main thing that widens the diff.
Under Dart 3.12.2 `dart run build_runner build` reproduced the committed `.g.dart` files exactly, so
the diff stayed limited to the actual swagger change — but that is **not** guaranteed across
versions: the package declares `build_runner: any` and its `pubspec.lock` is gitignored, so a
different resolution can reformat unrelated files. The pin can't be added to
`mobile/api/pubspec.yaml` either — that file is generator output (see `.openapi-generator/FILES`) and
is overwritten on the next regen. So **always read the diff and revert churn unrelated to the swagger
change**.

`dart analyze` substitutes for `flutter analyze` here (it reports the same errors) and stays scoped to
`mobile/api` — judge a regen by the **error** count, which must be **0**. The warnings are
pre-existing generator noise: ~78 in `mobile/api` of 111 across `mobile/`, the split recorded in
`.github/workflows/ci.yml` where the analyzer is deliberately not gated. Keep those two numbers in
sync with that comment.

## Component Development

### Backend Development (api/)
```bash
cd api
go run main.go                    # Run API server
go test -v ./...                  # Run tests
./set-up-dependencies.sh          # Install system deps (first time)
```

See `api/CLAUDE.md` for detailed backend architecture and testing requirements.

### Frontend Development (desktop/)
```bash
cd desktop
npm start                         # Dev server with API proxy (localhost:4200)
npm test                          # Run tests with coverage
npm run build                     # Production build
```

See `desktop/CLAUDE.md` for Angular architecture, NGXS state management, and component structure.

### Mobile Development (mobile/)
```bash
cd mobile
flutter run                       # Run on device/emulator
flutter test                      # Run tests
flutter build apk                 # Build Android APK
flutter build ios                 # Build iOS app
```

See `mobile/CLAUDE.md` for Flutter architecture, Provider state management, and navigation.

## Running in the Claude Code Web/Cloud Sandbox

When running this app inside the **Claude Code web (cloud) sandbox**, the stock "just run it" commands
above do **not** work out of the box, and the setup must be redone from scratch **every session**
(the container is ephemeral — the built ImageMagick, Redis, the DB, and node_modules are all lost
when the session ends). Two component-specific playbooks capture the exact, verified steps — read
them instead of rediscovering:

- **Backend:** `api/CLAUDE.md` → "Running in the Claude Code Web/Cloud Sandbox"
- **Frontend:** `desktop/CLAUDE.md` → "Running in the Claude Code Web/Cloud Sandbox"

**Root cause of the friction:** the sandbox base image is **Ubuntu 24.04 (Noble)**, whereas the
project's Docker images / setup scripts assume **Debian** (`golang:1.26-trixie`, `bullseye`). The big
one is ImageMagick: the Go API's `imagick.v3` CGO binding needs **ImageMagick 7**, but Ubuntu only
ships ImageMagick **6** and has no IM7 package — so `set-up-dependencies.sh` can't provide it and IM7
has to be **built from source**. Redis is installed but not started, and Tesseract/ImageMagick native
dev libs must be installed by hand.

**Run order & quick orientation:**
1. Backend (`:8081`) first — start Redis, install Tesseract libs, build IM7 from source, then
   `go run main.go` from `api/` with SQLite + local-Redis env vars.
2. Frontend (`:4200`) second — `npm install` then `npm start`; it only *proxies* to the backend, so
   the API must already be up.
3. Log in with the auto-created default admin **`admin` / `admin`**; login lands on
   `/dashboard/group/<id>`.
4. To drive the UI / take screenshots, use Playwright against the **pre-installed** Chromium at
   `/opt/pw-browsers/chromium` (do not `playwright install`) — details in `desktop/CLAUDE.md`.

## Critical Cross-Component Considerations

### API Changes Workflow
1. Modify backend code in `api/internal/`
2. Update `api/swagger.yml` to reflect API changes
3. Regenerate clients: `cd api && ./generate-client.sh desktop ../desktop/src/open-api`
4. Update frontend code to use new client methods
5. Test integration between components

### Authentication Flow
- JWT-based authentication with refresh tokens
- Backend issues tokens in `api/internal/handlers/auth.go`
- Desktop stores tokens via NGXS persistent storage
- Mobile uses `flutter_secure_storage` for secure token storage
- All API endpoints except `/api/auth/login` and `/api/auth/signup` require authentication
- **The refresh-token lifetime is a System Setting** (`refreshTokenValidForHours`, plus a separate
  `mcpRefreshTokenValidForHours` for connector tokens) so each install can pick its own security
  posture. Because refresh tokens rotate on every use, it is an **inactivity timeout, not an
  absolute session cap**. The **access token is deliberately fixed at 20 minutes** — both clients
  size their 15-minute proactive refresh timer against it, and neither client needed any change for
  this. See `api/CLAUDE.md` → "Session lifetime" and `desktop/CLAUDE.md` → "Session lifetime
  settings".
- **Whether cookies carry `Secure` is a System Setting too** (`secureCookies`), **off by default** so
  plain-HTTP LAN installs keep working, and recommended for any install served over HTTPS. It covers
  every cookie the API sets, through the one resolver `services.UseSecureCookies()`; `dev` forces it
  on regardless. Turning it on over plain HTTP locks the web app out, so the desktop warns first and
  recovery goes through the API with an admin's API key. Mobile is unaffected (tokens ride in the
  body). See `api/CLAUDE.md` → "Cookie `Secure` is a System Setting".

### OIDC Login (external identity providers)

Administrators can register one or more OIDC providers (Keycloak, Authentik, Google, Twitch, …)
in **System Settings → OIDC Providers**; both login screens then render a `Log in with
{{ displayName }}` button per enabled provider. This is a three-component feature, and the
pieces have to agree:

- **The Go API is the Relying Party — the clients are not.** Discovery, PKCE, `state`, `nonce`,
  the code exchange and ID-token verification all happen server-side. Neither client ever sees an
  IdP token or the client secret. Do not "simplify" this by moving the exchange into a client.
  This is the mirror image of the OAuth 2.1 **authorization server** the API already hosts for
  MCP (`internal/oauth/`); the two share no code paths beyond two primitives in `internal/utils`
  (`VerifyPkceS256`, `GetRandomUrlSafeString`). See `api/CLAUDE.md` → "OIDC (OpenID Connect)
  login".
- **The identity anchor is `(provider, sub)`**, in the `oidc_identities` link table. `sub` is the
  only claim OIDC guarantees stable and unique. `preferred_username` matching is a **first-login
  only** convenience behind a per-provider `linkByUsername` toggle that defaults **off** —
  usernames are user-changeable (and recycled) at most IdPs, so an always-on match is an
  account-takeover path for `admin`.
- **The redirect URI is derived from a System Setting**, `serverPublicUrl` →
  `{serverPublicUrl}/api/oidc/{name}/callback`. It is deliberately **not** `mcpPublicUrl` —
  changing that value invalidates live connector tokens. The provider form shows the computed URI
  read-only so it can be pasted into the IdP. A provider's `name` is **immutable after create**
  for the same reason.
- **Both clients gate on `app.oidc-providers.*`** for the admin CRUD. Legacy Admin picks the four
  new permissions up for free (its key set is dynamic over every APP key); Legacy User's list is
  fixed and must stay without them.
- **Pre-existing password accounts link from the profile page**, not by matching — *Connected
  accounts* on both `desktop/` and `mobile/` profiles runs the OIDC flow from an
  authenticated session, so the server is *told* the identity rather than inferring it. That is
  the escape hatch which makes `linkByUsername = false` a safe default, and it is why the feature
  is usable end-to-end from the phone alone. It is gated on **`app.account.update`**, the same
  permission as disconnecting one — adding a way to sign in is at least as privileged as removing
  one — and on an **interactive session**: neither linking nor unlinking may be done with an API
  key, whatever its scope, because a key that could add a sign-in method would be minting a
  credential wider and longer-lived than itself (the link outlives revoking the key). See
  `api/CLAUDE.md` → "Linking an existing account". **The two clients start it differently and must**: the desktop navigates and gets a 302,
  while mobile calls it as an authenticated API request, because a bearer token cannot ride into the
  external user agent RFC 8252 requires. Wiring mobile's Connect button to the *login* flow instead
  — which is how it first shipped — silently provisions a second account rather than linking.
- **Every OIDC flow is bound to the user agent that started it**, by a cookie, mobile included. The
  mobile link start is the one request an app rather than a browser makes, so it creates no session:
  it returns a single-use **launch URL** on this API, and the browser following that is what mints
  the session and receives the binding. Two earlier shortcuts here were both holes — skipping the
  binding for mobile (the callback check was `len(hash) > 0 && …`, a no-op for exactly those
  sessions) and handing the app the IdP's authorization URL directly (providers leak it by Referer,
  and the callback could not tell the app's browser from an attacker's). Neither is fixable with
  PKCE, which defends a different thing: binding stops an attacker **driving** your flow, PKCE stops
  a rogue app **stealing its result**. See `api/CLAUDE.md` → "Linking an existing account".
- **`FeatureConfig.oidcProviders` must serialize as `[]`, never `null`**, and
  `OidcProviderSummary.name` must stay `type: string` and **never become an enum** — the
  generated Dart deserializer has no null guard and closed enums throw on unknown values, the
  documented cause of two production login outages.
- **Mobile uses an external user agent, never a WebView.** `flutter_web_auth_2` on the private-use
  scheme `io.receiptwrangler://oidc` (RFC 8252 §7.1), carrying a one-time **PKCE-bound exchange
  code** — never a token. Any installed Android app can claim a private-use scheme, so the PKCE
  binding is what makes an intercepted code worthless. See `mobile/CLAUDE.md` → "OIDC sign-in".
- **Desktop lands on `/auth/callback`**, which bootstraps app data before routing — a first-ever
  OIDC login has valid cookies but an empty NGXS store, so landing anywhere else bounces back to
  the login screen. See `desktop/CLAUDE.md` → "OIDC sign-in".

**No e2e specs yet, deliberately.** The Go suite (with a local fake-IdP `httptest` server serving
discovery, JWKS and a real RS256 ID token) plus the Jest and Flutter widget specs are the safety
net; the flow has not been pinned against a real IdP or a real device. The two specs to add later
are a Playwright provider-CRUD spec and a Flutter `integration_test` login-button spec — both
mutate **global** System Settings, so they need `login-qr.spec.ts`'s capture-and-restore pattern
and the `e2e-shared-backend` concurrency group.

### Authorization (Roles & Permissions)
- A configurable role system layers on top of auth: administrators define **app-level** and
  **group-level** roles from granular permission strings (e.g. `app.users.create`,
  `group.receipts.read`) and assign them to users / group members.
- **Backend source of truth** is the hardcoded permission registry plus role CRUD in `api/` —
  exposed via `GET /api/permission` and `/api/role`, and mirrored in `swagger.yml` (so regenerated
  clients carry the `Permission` enum and role types). See `api/CLAUDE.md` → "Roles & Permissions".
- The server **never trusts the JWT for authorization** — it re-checks a user's current permissions
  from the database on every request. The JWT no longer carries any role field.
- **Role rollout is complete across backend and desktop.** Handlers fully enforce the permission
  system, and the legacy `UserRole`/`GroupRole` enums have been **removed from the backend** (Go
  types, model fields, JWT role claim, and the `userRole`/`groupRole` API fields are all gone; only
  the physical `user_role`/`group_role` DB columns are retained for the one-time upgrade migration).
  The **desktop** has likewise dropped every legacy-role consumer: the user-list and group-member
  tables now resolve `appRoleId`/`groupRoleId` to a role **name** (via a shared `RoleNamePipe`), the
  group-form "must have an owner" rule is gone (the backend no longer enforces an owner concept), and
  the `AuthState.userRole`/`hasRole` selectors plus the group-member legacy-enum bridge are removed.
  See `api/CLAUDE.md` → "Roles & Permissions" and `desktop/CLAUDE.md`.

### Group Default Custom Fields

A group can declare custom fields that are **always pre-added** to its receipts, configured on
**Group Receipt Settings**. This is a three-component feature; the pieces have to agree:

- **Backend** owns the config and the cleanup — `GroupReceiptSettings.defaultCustomFieldIds` (a
  `gorm:"-"` projection hydrated by an explicit batched loader, never a GORM hook) plus
  `applyDefaultCustomFieldsOnIngest`, which attaches the set to receipts the **server** creates
  (quick scan, email). Deleting a custom field removes it from every group's set. See
  `api/CLAUDE.md` → "Group Default Custom Fields".
- **Both clients apply the set on the receipt form**, on load in **every** mode — create, edit and
  read-only view — and whenever the group changes, with the same "smart swap" rule: an auto-added
  field that is still **empty** is dropped when you switch away, anything you typed into or added by
  hand is kept, and the new group's missing defaults are added. A field the user adds or removes by
  hand stops being auto-managed. See `desktop/CLAUDE.md` → "Per-group default custom fields" and
  `mobile/CLAUDE.md` → "Group default custom fields".
  - **Applying on load is what makes a default read as a built-in field.** A receipt saved before the
    group was configured shows the field too: blank and read-only in view, editable in edit. Two
    consequences worth knowing. **Saving an edited receipt persists the defaults as empty attached
    values** — an empty value is meaningful, it records that the field belongs on the receipt, and
    both clients already submit one per attached field because the backend replaces the whole
    association on update. And **a default applied on load stays auto-managed**, so changing the
    group on that form drops it again while it is still empty.
  - **Each client tracks that "the form added this, the user didn't" differently**, because their
    form lifecycles differ. Desktop re-derives it every `initForm()`, which rebuilds the custom-field
    `FormArray` from the saved receipt, so an unsaved manual edit never survives a navigation.
    Mobile's `ReceiptModel` *does* outlive its screen, so the provenance set lives on the model
    (`ReceiptModel.autoAppliedCustomFieldIds`) — see `mobile/CLAUDE.md`. Inferring it from the saved
    receipt instead reclaims a field the user removed and re-added by hand, and then drops it.
- **Both clients gate on `app.custom-fields.read`.** The server's
  `enforceReceiptCustomFieldSelection` **403s** any save that changes the set of attached custom
  field ids for a caller without it, so auto-adding fields for such a user would make their receipts
  unsaveable. Desktop checks the permission directly; mobile self-gates because its catalog fetch
  403s into an empty list. The seeded Legacy User role holds the permission, so default installs are
  unaffected — but this is why the feature is a no-op for a hand-built role that drops it.
- **An empty set must serialize as `[]`, never `null`** — the generated Dart deserializer has no null
  guard, so a null would fail the whole AppData payload on already-released Android builds.
- **The command fields are pointers** (`*[]uint` / `*bool`): omitting a key leaves the stored value
  alone. A client that hides the section must omit them rather than send zero values, or it wipes
  another admin's configuration.
- **Each client has its own e2e for the group switch**, because the unit/widget tests on both sides
  inject group settings into a mocked store and so prove nothing about the wire:
  `desktop/e2e/group-default-custom-fields.spec.ts` and
  `mobile/integration_test/receipt_default_custom_fields_test.dart`.

### Receipt Summary

A block of totals under the receipts table, covering the **whole current filter result set** rather
than the visible page: a receipt count and amount total overall, then the same figures per
configured status, plus a column per configured CURRENCY custom field. **All three components.**
It shipped backend + desktop first, while mobile had no filter to describe; mobile gained one (see
`mobile/CLAUDE.md` → "Receipt filtering") and the block followed.

- **Configuration is per-group and applies to everyone**, on Group Receipt Settings: a master
  toggle, which statuses break out, which currency fields are totalled, and **where the block
  renders** (top or bottom). Stored in two join tables plus two plain columns (deliberately not a
  discriminator on the existing defaults join — see `api/CLAUDE.md`).
- **The server owns the configuration, not the client.** `ReceiptSummaryCommand` carries the filter
  and an optional `configurationGroupId`, never the field or status list, so a client cannot add a
  column or opt out of one. A real group may omit `configurationGroupId` or send **its own** id —
  which is what the desktop does — but naming a **different** group's is a 400. Borrowing another
  group's configuration is the synthetic All group's privilege alone, since it has none of its own;
  allowing it anywhere else would be a way around the very invariant this bullet states.
- **`POST /api/receipt/group/{groupId}/summary` is gated on `group.receipts.read`** — the same
  permission as the table it sits under, and deliberately not `app.custom-fields.read`: that gates
  the catalog, and any receipt reader already sees these field names.
- **Aggregated in Go with `shopspring/decimal`, never a SQL `SUM`.** SQLite has no decimal type, so
  `SUM` over a `decimal(10,2)` returns a float there and an exact decimal on Postgres/MySQL — the
  same endpoint would report different cents on the three supported engines.
- **A configured status that matches nothing still renders, as a zero row**, so the block keeps its
  shape as the filter narrows. A receipt whose status is *not* configured still counts toward the
  overall row, or the total would disagree with the table's own count.
- **Neither client re-requests on paging or sorting** — neither changes which receipts the filter
  matches. That, plus skipping the request entirely for a group that has not opted in, is what keeps
  an unpaged aggregate affordable on the app's hottest screen. Mobile's split is structural rather
  than conventional: its sort setters already bypass the notification the summary listens to. See
  `mobile/CLAUDE.md` → "Receipt summary".
- **The synthetic "All" group picks a configuration via chips**, since it spans several groups and
  has none of its own; the data still spans every group. Desktop persists that pick to localStorage;
  mobile keeps it in `ReceiptListModel` for the session, having no persisted slice of its own.
- **`receiptSummaryPosition` rides on the summary RESPONSE, not just on the group settings.** A
  client's cached `groupReceiptSettings` is stale the moment an admin changes the configuration, and
  placement is configuration — so both clients render where the current 200 says, exactly as they
  already do for `enabled`. The server normalizes an empty position to `BOTTOM` at every emit point,
  because an empty enum fails a closed Dart `EnumClass` and with it the whole payload; the enum
  itself carries `TOP` and `BOTTOM` only, and the generated Dart client's unknown-value fallback
  lands on `BOTTOM`, so client and server agree on the degradation by construction.
- **Placement means different mechanics per client.** Desktop *moves* one `ng-template` between two
  anchors, and top means above the who-owes-whom settlement card as well as above the table. Mobile
  pins the block above or below its list — free, because `PagedDataList` is an `Expanded` — which is
  what makes the bottom position reachable at all under infinite scroll. Both slots must hold a
  **stable widget type**, or `PagedDataList` loses its State and silently refetches page 1.
- **E2e per client**, because both unit suites inject group settings into a mocked store and so
  prove nothing about the wire: `desktop/e2e/receipt-summary.spec.ts` and
  `mobile/integration_test/receipt_summary_test.dart`. Between them they cover the settings
  round-trip through the real resolver, the figures off a real decimal fold, a filter recomputing
  every row, the All-group chip pick surviving a reload, and the position surviving
  model → command → DB → response → render.

### Seeding the Group Field

Both clients pre-select the receipt's group instead of handing the user a picker they have no real
choice in. **Client-only** — no backend, swagger or generated-client involvement.

- **The rule is "the picker has exactly one option", not "the user has one group".** Every user also
  carries the synthetic **"All" group** (`Group.isAllGroup`), which the API even sorts *first*, so a
  single-group user has two entries in state and lands on "All" by default. The count must therefore
  come from the same filtered set the picker offers: `GroupState.soleGroupId` (desktop, built off
  `groupsWithoutAll`) and `GroupModel.soleGroupId` (mobile, off `groupsWithoutAllGroup` — which
  `buildGroupDropDownMenuItems` also sources, so the seed can never be an id the dropdown lacks and
  trip `DropdownButton`'s "exactly one item with value" assert).
- **Precedence.** Manual form: the receipt's own group → the group being browsed (never "All", and
  still resolvable) → sole group → blank. Quick Scan: `userPreferences.quickScanDefaultGroupId` →
  sole group → blank. The user's own quick-scan default always wins.
- **On desktop the permission gate follows the seed.** `setReceiptPermissions()` and the
  `/receipts/add` route guard resolve the same `GroupState.addTargetGroupId`, so a sole-group user is
  not turned away because the group they happen to be browsing is the synthetic "All" one. Both fall
  back to the selected group when there is no single add target, which keeps multi-group users
  exactly as they were.
- **Desktop's blank seed is `""`, never `0`.** `Validators.required` treats `0` as present, so a `0`
  sentinel would let a group-less receipt submit. See `desktop/CLAUDE.md`.
- **A seeded group is a picked group.** It applies that group's default custom fields on the add form
  and, in Quick Scan, its show/require field config — exactly as a manual pick does.
- **Mobile has to mirror the seed into `_ReceiptForm.groupId` too**, not just the dropdown: paid-by,
  the category/tag pickers and the add-share button all read that State field and stay dead at `0`.
  It happens in a post-frame callback because the resolution reads the route
  (`getFormStateFromContext` / `getGroupId`), which is illegal in `initState`.
- **The user-preferences "Quick Scan Default Group" control is deliberately left blank** — blank is
  the encoded "no default", and pre-filling it would silently persist a group the user never chose.
- **E2e per client**, because both unit suites inject a group list into a mocked store and so prove
  nothing about the wire: `desktop/e2e/single-group-default.spec.ts` and
  `mobile/integration_test/receipt_single_group_default_test.dart`. Both **provision their own
  account** — a freshly created user owns exactly "My Receipts" plus "All" — since the shared e2e
  accounts accumulate groups as other specs run.
- **E2e helpers must work with the field already filled.** A desktop single-select autocomplete goes
  `readonly` once it holds a value, so clicking it never opens the panel; mobile's dropdown opens
  either way, but "wait for the option text to appear" stops meaning "the menu is up". Both suites'
  shared pickers were hardened for this (`selectFirstOption`/`clearAutocomplete` in
  `desktop/e2e/receipts.spec.ts`, `selectDropdown` in
  `mobile/integration_test/helpers/form_actions.dart`).

### Failed Uploads Keep Their Image

A quick scan or email upload that fails keeps the file it was working from, and the user can preview
or download it to enter the receipt by hand. **Backend + desktop**; the swagger change regenerates
both clients, but mobile gets no new UI.

- **The old cleanup was inverted.** It released a temp file once every referencing task was
  `Completed` **or** `Archived` — and `Archived` is exactly what makes an activity rerunnable. At the
  time no queue set `asynq.Retention`, so a successful task left Redis immediately and never reached
  the completed set, meaning the only files it ever deleted were the ones it had to keep. It also
  scanned only the email queue, and was registered inside `StartEmailPolling`, so an install without
  email polling ran no temp cleanup at all. See `api/CLAUDE.md` → "Temporary file retention & cleanup"
  for the replacement's precedence table and the two invariants that keep its orphan branch safe.
- **`asynq.Retention` is now set — on the email queue only** (6h, `enqueueOptions` in
  `wranglerasynq/task_enqueue.go`), so a *succeeded* email upload's attachment and OCR copy are
  released on the next hourly sweep rather than after the retention window. Quick scan is excluded
  because it already deletes its own file on success and is 1:1 task-to-file; email cannot, because
  one attachment fans out to a sibling task per `groupSettingsId`. Two knock-on rules came with it,
  both because a succeeded task now lingers in Redis: `Inspector.RunTask` does **not** refuse a task
  by state, so the rerun endpoint enforces `Archived` itself (otherwise a rerun of a succeeded email
  would duplicate its receipt), and a `Completed` task stops advertising its source file (otherwise
  preview/download would appear on a succeeded activity and 404 an hour later).
- **How long a file is kept is a System Setting** (`tempFileRetentionHours`, default 720 = 30 days,
  bounds 24-8760), following the same pointer-and-omitted-column machinery as the refresh-token
  lifetimes — an omitted key leaves the stored value alone.
- **`canBeRestarted` now also requires the files a rerun reads**, so it stops advertising reruns that
  cannot work. A body-only email reads none and stays rerunnable, which is why "expects a file" is
  tracked separately from "has a file". **Mobile inherits this for free** —
  `group_activity_list_item.dart` already gates its rerun slidable on the flag, so the regen is the
  whole mobile change.
- **`hasSourceFile` is deliberately NOT gated on `Archived`**, unlike `canBeRestarted`: asynq's retry
  backoff means minutes pass before a task archives, and the user should not watch an activity sit at
  FAILED with no way to get their image.
- **Each client's gate keys on the activity's own group**, never the surface's. The desktop widget had
  this wrong and hid every control on the "All" dashboard; mobile had fixed it years earlier. See
  `desktop/CLAUDE.md` → "Activity source file".
- **E2E on the desktop only** (`desktop/e2e/failed-activity-source-file.spec.ts`), because it is the
  only test that can prove the image survived in `temp/` through a real failure. It forces the failure
  by pointing the global AI provider at an unreachable host, so it works against the shared demo
  backend too — which makes it a global-state-mutating, serial spec.

### Quick Date Filter (month stepper)

A month stepper — `‹ September 2026 ›` — plus a picker naming which date field it writes to, sitting
above the receipts list on **both clients**. Stepping left or right adds or replaces that field's
condition in the *same* filter the advanced filter dialog/screen drives, so the two can never
disagree. **Client-only on both sides**: no backend, `swagger.yml` or generated-client involvement.

- **A month is `BETWEEN [first day, last day]`** — the one operation that can express *any* month,
  which is why the feature needed no API change. `WITHIN_CURRENT_MONTH` is **not** equivalent: the
  server pins it to month-start through *today*, so it can only ever mean the current month.
- **It targets one of three date fields** — `date` / `resolvedDate` / `createdAt` — chosen by the
  user and held in the same state as the filter. **Switching is non-destructive and refetches
  nothing**: it changes no condition, only which one the stepper describes, so the abandoned
  condition stays applied and stays visible (a chip on desktop, a card on the filter screen on
  mobile).
- **The label degrades rather than hiding.** `"<Month> <Year>"` for a whole calendar month, `"Custom"`
  for a condition the stepper cannot describe, `"All time"` for none. The arrows seed from **today's**
  month when nothing is showing, then apply the delta, so `‹` and `›` never do the same thing.
- **Each client stores the chosen field where its filter already lives**, and that is the one place
  they differ meaningfully: desktop persists both to localStorage (so `monthFromFilterEntry` has to
  accept ISO strings, and the field's default has to be applied on *read*), while mobile's
  `ReceiptListModel` is in-memory and needs neither. Mobile still keeps the field on the **model**
  rather than the list widget, because that widget is rebuilt on almost every navigation.
- **E2e per client**, because both unit suites assert against a mocked API and so prove nothing about
  the wire: `desktop/e2e/receipt-quick-date-filter.spec.ts` and
  `mobile/integration_test/receipt_quick_date_filter_test.dart`. The mobile one **seeds relative to
  `DateTime.now()`** — the shared mobile filter fixture is pinned to June 2026, which a
  month-relative control drifts away from.
- **The same field list drives the Report Builder's "Date field" picker**, next to "Period covering",
  and that one **is** server-backed: `ReportPeriod.dateField` picks which receipt date the report
  period filters. Adding a date field to desktop's `RECEIPT_DATE_FILTER_FIELDS` therefore also needs
  `commands.ReceiptDateFilterKeys()` / `DateFilterField` and the swagger description, or the picker
  offers a value the API rejects with a 400. `TestReceiptDateFilterKeys` pins the Go side. The wire
  field is a plain string, not an enum, so mobile never breaks on a new key. See `api/CLAUDE.md` →
  "The period's date field".

See `desktop/CLAUDE.md` → "The month stepper targets one date field" and `mobile/CLAUDE.md` → "Quick
date filter" for the per-client details.

### Report Split Options

The Report Builder can divide a receipt's money **equally across its categories and/or tags** instead
of counting it in full in each (two checkboxes, plus a per-report list of currency custom fields to
leave whole). **Backend + desktop**; mobile only regenerates its client and carries the keys through
a template preview.

- **The server does all the arithmetic**, before the engine: it splits each receipt into one copy per
  category (× tag), with cents spread so the shares add back up exactly, and it divides by the
  receipt's **true** category count even for a viewer who sees some as `(Restricted)`. See
  `api/CLAUDE.md` → "Splitting a receipt equally across categories / tags".
- **A box only applies where the report cuts by that dimension** (a grouping level or the aggregate
  dimension). The desktop flags a ticked box that does nothing rather than letting it look applied.
- **All three keys are optional and omitted when unset**, so existing templates are unchanged and no
  configuration-version bump was needed. The excluded keys are plain strings on the contract, never
  an enum, so a released mobile build cannot fail a template payload on them.
- E2e: `desktop/e2e/report-split-equally.spec.ts`. Mobile has no builder, so its guard is the ingest
  test `mobile/test/models/report_split_ingest_test.dart`.

### Role-Required Receipt Fields & Single-Call Create

A group role can require its members to keep **at least one comment** and/or **at least one image**
on the group's receipts (`requireReceiptComment` / `requireReceiptImage`, group roles only, set on the
desktop role form — mobile has no role UI). **All three components.**

- **The server is the authority.** It enforces on create, on update (against the receipt's **stored**
  comments/images, since edit mode still adds those immediately through their own endpoints), on
  deleting the last comment/image (refused — replace by adding first, then deleting), and on the Quick
  Scan comment. Each failure is a 400 validator error keyed `comments` / `files`.
- **Waivers keep a user from being stuck:** the comment rule is off when the group hides comments or
  the caller lacks `group.comments.create`; the image rule is off when the group hides images. The
  resolved result rides AppData as `groupReceiptRequirements` (only groups with a requirement; absent
  = nothing required; never `null`), which both clients use for their pre-submit check.
- **Create is one call: `POST /receipt/withFiles`** (multipart: `receipt` + `files`), storing the
  receipt, its comments and its images in one transaction. That is what lets the server see images at
  create time — the old flow uploaded them after the receipt existed. The generated clients encode the
  `receipt` part differently (desktop: an `application/json` blob part; Dart: a JSON form value), and
  the server accepts both. Both clients use it; the "receipt added but images failed" state is gone.
- **The JSON `POST /receipt/` is deprecated but kept** for already-released mobile builds. It enforces
  the same rules, so for a role requiring an image it always 400s with an "update your app" message —
  deliberately: an old client can never attach an image at create time.
- **The synthetic "All" group is never a destination.** Both creates and Quick Scan 400 it, as a
  move on update already did; neither client's group picker offers it, so this only closes a crafted
  request. See `api/CLAUDE.md` → "The synthetic \"All\" group is never a destination".
- E2e per client: `desktop/e2e/receipt-role-requirements.spec.ts` and
  `mobile/integration_test/receipt_role_requirements_test.dart`. See `api/CLAUDE.md`,
  `desktop/CLAUDE.md` and `mobile/CLAUDE.md` → "Role-required receipt fields & single-call create".

### App Time Zone

One admin-editable **System Setting**, `timeZone` (an IANA name, default `UTC` so existing installs
behave exactly as before), decides which calendar day an instant falls on and what "today" and "this
month" are — for filters, report periods, report buckets and labels alike. Before it, the server's
process zone, hard-coded UTC and the browser's zone each answered part of that question, so a "Last
Month" report on a UTC container left out a receipt added at 9PM on Sep 30 Eastern.

| Value | Kind | Zone its day is read in |
|---|---|---|
| `receipts.date`, DATE custom field values | calendar day | UTC — stored as midnight UTC of the day |
| `created_at` (Added At), `resolved_date`, task `started_at` / `ended_at` | instant | app zone |
| "today", "this month", report presets, `@currentYear`, `{{generatedAt}}`, the PDF footer | clock | app zone |

Unchanged: log timestamps, token expiry, `@every` schedules, sort order.

- **Filters send calendar days, not instants.** A date filter value is a bare `yyyy-MM-dd`; the
  server turns it into a range in the right zone for that column (`BuildDayFilterQuery`). An RFC 3339
  value is still accepted and read as the day **as written** — which is what keeps already-released
  mobile builds, which send local wall time with a literal `Z`, landing on the day the user picked.
- **A receipt's Date never shifts with the zone.** The server stores it (and DATE custom fields) as
  midnight UTC of the as-written day on every create/update; existing rows are not migrated and are
  read by their UTC day. Clients send midnight UTC of the picked day and display it in UTC.
- **`timeZone` is a plain string on the contract, never an enum** (`SystemSettings`,
  `UpsertSystemSettingsCommand`, `AppData`; optional everywhere). The update command's field is a
  pointer, so a client that omits the key leaves the stored zone alone. `AppData.timeZone` is the
  *resolved* name — `"UTC"` when unset or invalid, never `""`.
- **Per client:** the backend is complete (see `api/CLAUDE.md` → "App time zone"). The desktop reads
  `AppData.timeZone`, edits the setting on System Settings, sends filter days as `yyyy-MM-dd` and
  displays instants in the app zone (see `desktop/CLAUDE.md`). Mobile only carries the regenerated
  client for now; building filters and displaying dates in the app zone is a pending follow-up.

### State Management Patterns
- **Backend**: Service layer handles business logic, repositories handle data access
- **Desktop**: NGXS store with actions/selectors, persistent storage for auth/preferences
- **Mobile**: Provider pattern with ChangeNotifier models, models own their state

### Background Processing
- Backend uses Asynq for async jobs (OCR processing, email polling, cleanup)
- Long-running operations (OCR, AI extraction) run as background jobs
- Frontend polls for completion or uses WebSocket-like patterns where implemented

## Version Management

Each component has version tagging scripts:
- `api/tag-version.sh` - Tag API version
- `desktop/tag-version.sh` - Tag desktop version
- `mobile/tag-version.sh` - Tag mobile version

Version is embedded in Docker builds via `VERSION` and `BUILD_DATE` build args.

## Data Persistence

### Development
- API defaults to SQLite in `api/sqlite/`
- Desktop proxy config in `desktop/proxy.conf.json` routes to localhost:8081
- Mobile configures API base URL in app settings

### Production (Docker)
- Volumes for persistent data:
  - `/app/receipt-wrangler-api/data` - Receipt images and uploads
  - `/app/receipt-wrangler-api/sqlite` - SQLite database
  - `/app/receipt-wrangler-api/logs` - Application logs
- nginx serves frontend from `/usr/share/nginx/html`
- API runs on same container, proxied via nginx

## Common Pitfalls

1. **Forgot to regenerate clients**: After API changes, clients are out of sync → regenerate!
2. **Editing generated code**: Changes to `desktop/src/open-api/` or `mobile/api/` will be lost
3. **Missing system dependencies**: API requires Tesseract, ImageMagick → run `api/set-up-dependencies.sh`
4. **Test database cleanup**: Failed Go tests leave `app.db` in test dirs → remove before rerunning
5. **Port conflicts**: API (8081), desktop dev (4200), docker prod (80) must be available
6. **CORS in development**: Desktop proxy handles CORS, but mobile needs proper API base URL

## Project Structure Summary

```
receipt-wrangler-api/          # Monorepo root
├── api/                       # Go backend
│   ├── internal/              # Core application code
│   │   ├── handlers/          # HTTP handlers
│   │   ├── services/          # Business logic
│   │   ├── repositories/      # Database access
│   │   ├── models/            # Data models
│   │   └── wranglerasynq/     # Background jobs
│   ├── swagger.yml            # API specification (source of truth)
│   └── CLAUDE.md              # Backend-specific guidance
├── desktop/                   # Angular web app
│   ├── src/
│   │   ├── app/               # Application modules
│   │   ├── store/             # NGXS state management
│   │   ├── shared-ui/         # Reusable components
│   │   └── open-api/          # Generated API client (DO NOT EDIT)
│   └── CLAUDE.md              # Frontend-specific guidance
├── mobile/                    # Flutter mobile app
│   ├── lib/
│   │   ├── models/            # Provider state models
│   │   ├── groups/            # Group features
│   │   ├── receipts/          # Receipt features
│   │   └── shared/            # Shared widgets
│   ├── api/                   # Generated API client (DO NOT EDIT)
│   └── CLAUDE.md              # Mobile-specific guidance
└── docker/                    # Docker build configs
    ├── Dockerfile             # Production monolith
    └── dev/Dockerfile         # Development container
```

## Code Changes Philosophy

- Prefer minimal, targeted changes. Do not refactor or restructure code beyond what was explicitly requested.
- A primary focus of yours is overall code quality. Your focus should be on producing code that is stable, flexible when
  needed, readable and maintainable. You should not be writing code that is difficult to read, confusing, insecure or
  too long.
- Follow **DRY (Don't Repeat Yourself) pragmatically**. If two or more places share nearly identical logic that would
  need to be updated together, extract it into a shared utility, function, or component. This is not a dogmatic rule —
  three similar lines in a single file or minor template repetition is fine. Apply DRY when it meaningfully reduces
  maintenance burden, not for every tiny duplication.
- When the first approach fails, stop and ask the user for direction rather than trying multiple speculative approaches
  in sequence.
- After you have completed the planning phase, and you have your plan, please iterate over your plan at a maximum of 3
  times. During these iterations, your goals are to verify that your code makes sense, and solves the requested things,
  that your code is sound, secure and consistent with style across the codebase, and that your code is clean, and not a
  hacked together solution.

## Parallel Agent Execution

When a task spans multiple components (e.g., backend `api/` and frontend `desktop/` or `mobile/`), follow these rules:

- **Run backend and frontend agents in parallel** whenever possible. Do not serialize work across components unless
  there is a hard dependency.
- **Frontend agents should order their work to defer backend-dependent tasks.** If the frontend needs something from the
  backend (generated client, models, API endpoints), schedule that work last so independent frontend work happens first.
- **If the frontend agent is blocked on the backend agent** (e.g., waiting for a generated client, new API models, or
  endpoint changes), the frontend agent should:
    1. Continue planning its backend-dependent work (design the component, write the template, stub the types).
    2. **Wait** for the backend agent to finish before executing backend-dependent code. Do not guess at API shapes or
       generate placeholder clients.
    3. Resume execution once the backend deliverables are available.
- **The backend agent should signal completion clearly** — after finishing its work, the orchestrating agent should
  trigger any required client regeneration (e.g., `./generate-client.sh desktop ../desktop/src/open-api`) before
  unblocking the frontend agent.
- **Mobile (`mobile/`) changes** follow the same pattern: if a backend change requires a mobile update, run the mobile
  agent in parallel with the desktop agent after the backend agent completes.

### Example Task Ordering

For a feature that adds a new API endpoint and a corresponding UI:

1. **Phase 1 (parallel):**
    - Backend agent: handler → service → repository → route → tests → swagger update
    - Frontend agent: independent UI work (layout, styling, routing, non-API components)
2. **Phase 2 (sequential, after backend completes):**
    - Regenerate client (`cd api && ./generate-client.sh desktop ../desktop/src/open-api`)
    - Frontend agent: wire up API calls, integrate generated types, write dependent components
3. **Phase 3 (parallel):**
    - Backend agent: any follow-up fixes
    - Frontend agent: integration tests, final UI polish

## Testing

- After ANY code change, run the full relevant test suite before considering the task complete.
- When tests fail, fix both the code AND the tests — don't assume tests are correct or code is correct without
  verifying.

## Workflow Rules

- Always complete implementation AND verify (build + tests pass) before committing. Do not commit code that hasn't been
  validated.
- During your planning sessions, explicitly check if your planned code introduces regressions. We want to make sure that
  we do not break existing code, especially things that may not show themselves through build errors like scss changes,
  conflicting styles, and so on.
- During your planning sessions, take a moment to think if there are any edge cases, or possible regressions or any
  additional things for the user to test before considering the task complete.
- After implementing any full feature, always commit/push.

### Code Review Feedback Disposition

When addressing review feedback from CodeRabbit, human reviewers, or any other source, follow this protocol every time:

1. **Read every comment** before acting. Don't fix-by-fix.
2. **Build a disposition table for the user.** One row per comment: file/line, issue summary, decision (`ACCEPT` / `REJECT` / `ACCEPT + EXTEND` / `DEFER`), and a one-line justification. **Present this table inside the plan you write for the user** — it is for the user to review your reasoning before approving changes. Do NOT post this table as a bulk PR comment; it is a planning artifact, not a review response.
3. **Verify each "ACCEPT" against current code** — reviewers (especially bots) sometimes flag false positives, stale code, or generator output they don't recognize as such. If a flag turns out to be invalid on inspection, flip the decision to `REJECT` with the reason ("verified — generator output, matches existing repo convention" / "verified — code already handles this case at line X").
4. **Default to rejecting** comments that target auto-generated files (`desktop/src/open-api/`, `mobile/api/`) unless the comment identifies a real type/compile error the generator introduced. Hand-edits to generated files require an explicit justification and should match an established project precedent (search `git log` for "Fix build errors" / similar prior hand-patches).
5. **Reply on each individual review-comment thread** with the per-comment decision and justification. Every CodeRabbit (or human) comment must get its own reply explaining whether you accepted or rejected and why — that is the audit trail the reviewer sees. Use `gh api -X POST /repos/<owner>/<repo>/pulls/<num>/comments/<comment_id>/replies` (review comments live on the pulls endpoint, not the issues one) with a JSON body of `{"body": "..."}`. Fetch the review-comment IDs via `gh api /repos/<owner>/<repo>/pulls/<num>/comments`.
6. **Commit + push** only after every comment has an individual reply posted. The per-comment replies are the audit trail; the commit is the action.

## CLAUDE.md Maintenance

- After modifying files in any component, check whether the corresponding `CLAUDE.md` needs updating.
- Each component has its own documentation: `api/CLAUDE.md`, `desktop/CLAUDE.md`, `mobile/CLAUDE.md`.
- If a change alters behavior, configuration, architecture, commands, or conventions documented in a `CLAUDE.md` file,
  update that file to stay accurate before considering the task complete.
