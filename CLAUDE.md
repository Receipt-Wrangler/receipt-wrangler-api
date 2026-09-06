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

After a mobile regen, re-apply the two documented dart-dio patches (`mobile/CLAUDE.md` → "Known
dart-dio default-value regressions") and run `flutter analyze`.

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
pre-existing generator noise: ~73 in `mobile/api` of 107 across `mobile/`, the split recorded in
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
- **Both clients apply the set on the receipt form**, on create and whenever the group changes, with
  the same "smart swap" rule: an auto-added field that is still **empty** is dropped when you switch
  away, anything you typed into or added by hand is kept, and the new group's missing defaults are
  added. A field the user adds or removes by hand stops being auto-managed. See
  `desktop/CLAUDE.md` → "Per-group default custom fields" and `mobile/CLAUDE.md` → "Group default
  custom fields".
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
