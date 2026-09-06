# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Receipt Wrangler API is a Go-based backend service for a receipt management and splitting application. It provides OCR-powered receipt scanning, AI-assisted data extraction, email integration, and multi-user support with group management capabilities.

## Development Commands

### Building and Running
- `go build` - Build the application
- `go run main.go` - Run the application directly
- `./set-up-dependencies.sh` - Install system dependencies (tesseract, ImageMagick, Chromium, Python deps)

### Running in the Claude Code Web/Cloud Sandbox

> Playbook for booting the API in the Claude Code web (cloud) sandbox from a **fresh session**.
> Everything below is ephemeral — the container is rebuilt each session, so these steps must be
> re-run every time. Follow them top-to-bottom and the server comes up on `:8081` in ~2 min.

**Why the normal path doesn't just work.** The sandbox base image is **Ubuntu 24.04 (Noble)**, but
the project's Docker images (and `set-up-dependencies.sh`) assume **Debian** (`golang:1.26-trixie` /
`bullseye`). Two consequences:
- The CGO binding `gopkg.in/gographics/imagick.v3` needs **ImageMagick 7**'s MagickWand API. Ubuntu's
  repos only ship ImageMagick **6** (`libmagickwand-dev` = IM6) and have no IM7 package, so
  `set-up-dependencies.sh`'s `libmagickwand-7.q16-dev` isn't installable and the Go build fails to
  link. **IM7 must be built from source.**
- `set-up-dependencies.sh` also installs heavy Python (torch, easyocr) used only by the **IMAP email
  client**, not the Go server. **Skip it** — install just the native libs below.

**Step 1 — Start Redis** (installed but not running by default):
```bash
redis-server --daemonize yes --save "" --appendonly no
redis-cli ping   # -> PONG
```

**Step 2 — Install Tesseract + build tooling** (apt, these IM6-free packages are all in Ubuntu's repos):
```bash
apt-get update -qq
apt-get install -y build-essential pkg-config libtesseract-dev libleptonica-dev tesseract-ocr-eng
```

**Step 3 — Install ImageMagick 7 delegate dev libs** (needed before configuring IM7):
```bash
apt-get install -y libpng-dev libjpeg-turbo8-dev libtiff-dev libwebp-dev libheif-dev libde265-dev ghostscript libtool
```

**Step 4 — Build & install ImageMagick 7 from source** (the `imagemagick.org/archive` tarball 404s —
clone from GitHub instead):
```bash
cd <scratchpad>            # build outside the repo
git clone --depth 1 https://github.com/ImageMagick/ImageMagick.git
cd ImageMagick
./configure --prefix=/usr/local --with-heic=yes --with-jpeg=yes --with-png=yes \
            --with-tiff=yes --with-webp=yes --with-gslib=yes --disable-docs --without-magick-plus-plus
make -j"$(nproc)"          # ~1-2 min
make install
ldconfig                   # <-- the key step
```
`ldconfig` is what makes both the Go build and the running server find the new libs: `/usr/local/lib`
is already listed in `/etc/ld.so.conf.d/libc.conf`, so after `ldconfig`, `pkg-config --modversion
MagickWand` resolves `7.1.2` from the **default** path and `libMagickWand-7.Q16HDRI.so` is in the
loader cache. You do **not** need `PKG_CONFIG_PATH` or `LD_LIBRARY_PATH` once `ldconfig` has run
(fallbacks if it somehow didn't: `PKG_CONFIG_PATH=/usr/local/lib/pkgconfig` for the build,
`LD_LIBRARY_PATH=/usr/local/lib` for the run). Verify: `magick -version` shows `7.1.2 ... Q16-HDRI`.

**Step 4b — Copy the private headers `make install` leaves behind.** Recent ImageMagick *git master*
(e.g. `7.1.2-32 Beta`) has a public header that includes private ones — `MagickCore/pixel-accessor.h`
pulls in `MagickCore/image-private.h`, which pulls in `MagickCore/quantum-private.h`, and so on — but
`make install` installs only the public set. Every cgo compile of `gopkg.in/gographics/imagick.v3`
then dies with `fatal error: MagickCore/image-private.h: No such file or directory`. Backfill the
whole set from the source tree:

```bash
IM_SRC=$PWD                 # Step 4 left us in the ImageMagick source tree
INC=/usr/local/include/ImageMagick-7
for d in MagickCore MagickWand; do
  for f in "$IM_SRC/$d"/*.h; do
    b=$(basename "$f"); [ -f "$INC/$d/$b" ] || cp "$f" "$INC/$d/$b"
  done
done

cd /home/user/receipt-wrangler/api
go build ./...
```

Do **not** pipe that loop into `head` — the SIGPIPE kills it partway and you get a second, different
missing-header error. Verify with `go build ./...` (not `go build ... | tail`, whose exit status is
`tail`'s and so hides the failure).

**Step 5 — Create `api/data/`** (gitignored via `/data/*`, so a fresh clone doesn't have it):
```bash
mkdir -p /home/user/receipt-wrangler/api/data
```
The server starts fine without it and only fails later, when something writes there — creating a
group makes a per-group directory under it, so `POST /group/` returns **HTTP 500 "Error creating
group"** with `mkdir .../api/data/<id>-<name>: no such file or directory` in `api/logs/app.log`.
That takes down every e2e fixture that provisions a group (`provisionPermUser` and friends), which
looks like a broken test rather than a missing directory.

**Step 6 — Run the Go server from the `api/` directory** (paths like `logs/`, `data/` and `./sqlite/`
are relative to the working directory, so cwd **must** be `api/`):
```bash
cd /home/user/receipt-wrangler/api
DB_ENGINE=sqlite DB_FILENAME=wrangler.db \
ENCRYPTION_KEY=dev-encryption-key SECRET_KEY=dev-secret-key \
REDIS_HOST=127.0.0.1 REDIS_PORT=6379 \
go run main.go            # (or build once: `go build -o /tmp/rw-api . && /tmp/rw-api`)
```
- `ENCRYPTION_KEY` / `SECRET_KEY` just need to be **non-empty** (`config.CheckRequiredEnvironmentVariables`
  fatals on empty) — throwaway dev values are fine.
- Gotcha: `dev/switch-to-sqlite.sh` exports `REDIS_HOST=redis` (the docker-compose service name). For a
  **local** Redis you must override it to `127.0.0.1`, or the connection fails at startup.
- Ready when the log prints `Listening on port 8081`. Smoke test: `curl localhost:8081/api/featureConfig`
  → `200`.

**Logging in.** First startup auto-creates a default admin **`admin` / `admin`** (in any env except
`-env test`; see `CreateUserIfNoneExist`). The login endpoint is **`POST /api/login/`** (not
`/api/auth/login`); add `?tokensInBody=true` to get the JWT in the response body:
```bash
curl -X POST 'localhost:8081/api/login/?tokensInBody=true' \
  -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin"}'
```

**Not needed to boot the server (skip in the sandbox):** the Python/torch/easyocr install, and the
ImageMagick PDF-policy edit in `set-up-dependencies.sh` (it targets `/etc/ImageMagick-7/policy.xml`; a
source build's policy is at `/usr/local/etc/ImageMagick-7/policy.xml` and doesn't block PDF by
default). Those only matter for exercising email-OCR / PDF-receipt processing, not for a running API.

**Running the Go test suite in the sandbox — set `CHROMIUM_BINARY_PATH`.** Steps 1-4 above are enough
to *compile* the module, but two tests additionally need a browser:
`TestHtmlToPdfService_Render_BasicHtml` and `TestReportService_Generate_PdfDocument` drive the
HTML-to-PDF pipeline through headless Chromium. `services/html_to_pdf.go` defaults to
`/usr/bin/chromium`, which **does not exist** in this sandbox — the pre-installed Playwright build is
at `/opt/pw-browsers/chromium`. Without the override both fail and `internal/services` reports FAIL,
which looks like a code regression and is not one:
```bash
CHROMIUM_BINARY_PATH=/opt/pw-browsers/chromium go test -count=1 ./...
```
Do **not** run `playwright install` to get `/usr/bin/chromium` — the download is blocked, and the
pre-installed binary works.

### Testing
- `go test -v ./...` - Run all Go tests with verbose output
- `go test -coverprofile=coverage.out -covermode=atomic -v ./...` - Run tests with coverage
- `python3 -m unittest discover -s ./imap-client` - Run Python IMAP client tests

### Seeding test data
- `RW_API_KEY='key.1.<id>.<secret>' node dev/seed-reporting-data.mjs` - Seed a realistic, high-volume
  reporting dataset (a dedicated group + member users + categories/tags + ~2000 receipts spread over a
  date range) via the HTTP API. The admin API key is read from `RW_API_KEY` (never hardcoded). Setup is
  idempotent (reuses the group/users/categories/tags by name); receipts are additive. Tunable via env
  (`RECEIPT_COUNT`, `SEED_GROUP_NAME`, `START_DATE`/`END_DATE`, `CONCURRENCY`, …) — see the script header.

### API Client Generation
- `./generate-client.sh desktop <output-dir>` - Generate TypeScript Angular client
- `./generate-client.sh mobile <output-dir>` - Generate Dart Dio client

### Go Toolchain
- Requires **Go 1.26.8+** — `api/go.mod` declares `go 1.26.8`, which `golang.org/x/crypto v0.56.0`
  forces (it declares `go 1.26.0`). The MCP `github.com/modelcontextprotocol/go-sdk` sets a lower,
  1.25 minimum.
  Docker images and CI containers use `golang:1.26-trixie`, which currently resolves to the same
  digest as `1.26.8-trixie` and so satisfies the directive without a toolchain download.

## Architecture Overview

### Core Structure
- **main.go** - Application entry point, initializes logging, config, database, and starts HTTP server
- **internal/** - Core application code organized by domain
- **imap-client/** - Python-based email processing client

### Key Directories
- **internal/handlers/** - HTTP request handlers for each API endpoint
- **internal/repositories/** - Database access layer using GORM
- **internal/services/** - Business logic layer
- **internal/models/** - Database models and domain objects
- **internal/commands/** - Command objects for API requests/responses
- **internal/routers/** - Route definitions and middleware setup
- **internal/wranglerasynq/** - Background job processing using Asynq
- **internal/ai/** - AI client implementations (OpenAI, Gemini, Ollama)
- **internal/permissions/** - Pure permission registry + wildcard matcher (no DB)
- **internal/reporting/** - Pure report engine (no DB); `receiptsource/` maps receipts into it

### Database
- Uses GORM ORM with support for SQLite, MySQL, and PostgreSQL
- Migrations are handled automatically on startup via `repositories.MakeMigrations()`
- Test databases are set up in `repositories/main_test.go`

### Background Processing
- Uses Hibiken's Asynq library for background job processing
- Email processing, OCR, and cleanup tasks run as background jobs
- Queue configurations defined in `internal/wranglerasynq/`

### AI Integration
- Supports multiple AI providers: OpenAI, Google Gemini, and Ollama
- AI clients implement a common interface defined in `internal/ai/base_client.go`
- Used for receipt data extraction and processing
- **The configured URL is used verbatim.** `internal/ai/open_ai.go` treats
  `ReceiptProcessingSettings.Url` as an OpenAI-compatible **base** url and appends only
  `/chat/completions`, authenticating with `Authorization: Bearer <key>`. There is no
  provider-specific rewriting: the client does **not** inspect the url, inject a deployment path, or
  add an `api-version` query param. An empty url (the plain `OPEN_AI` type, which
  `UpsertReceiptProcessingSettingsCommand.Validate` requires to be empty) falls back to go-openai's
  `https://api.openai.com/v1`. Ollama is the same — `internal/ai/ollama.go` posts to the url exactly
  as entered, with no suffix.
- **Azure must be configured with its OpenAI-compatible endpoint**, i.e.
  `https://<resource>.services.ai.azure.com/openai/v1` (with `Model` set to the *deployment* name), not
  a bare resource origin. Earlier versions sniffed the url for the substring `azure` and switched
  go-openai into Azure mode (`DefaultAzureConfig`), which rewrote the path to
  `/openai/deployments/<model>/chat/completions` and pinned `api-version=2023-05-15`. That heuristic
  was **removed**: it mangled the modern Foundry `/openai/v1` endpoint into a 404, and it could never
  match an Azure resource behind a custom domain. Pinned by
  `TestOpenAiGetChatCompletion_AzureUrlIsUsedVerbatim`.

### Configuration
- Configuration loaded from JSON files in `config/` directory
- Environment variables override config file settings
- Sample configuration in `config/config.sample.json`

## Filesystem Access & Path-Traversal Safety

Group and receipt files live under the app's **data directory** (`<cwd>/data/<groupId>-<groupName>/…`).
The group name is **user-controlled**, so any path built from it is a path-traversal vector (CWE-22,
GHSA-966v-m9rv-j5cx): `filepath.Join` collapses `..` *after* the name is joined, so an unsanitized name
like `../../../../tmp/x` escapes `data/`.

**Rule: for any file under the data directory, always use the sanitized helpers in
`internal/utils/data_files.go` — never raw `os.*` (and never the generic `utils.WriteFile` /
`ReadFile` / `MakeDirectory` / `DirectoryExists`) on a data path.**

- Path building: `GetDataDir()` (the single source of truth for the data root) and
  `BuildGroupPathString` / `FileRepository.BuildFilePath`, which **reject any name that escapes `data/`**.
- File operations (each re-asserts containment before touching disk):
  `MakeDataDirectory`, `EnsureDataDirectory`, `WriteDataFile`, `ReadDataFile`, `RemoveDataPath`,
  `RemoveAllInDataDir`, `RenameDataPath`.
- Input validation: `UpsertGroupCommand.Validate` rejects group names containing path separators or
  `.` / `..` up front (400), so a malicious name is never persisted.

These are the single, centralized defense — do not sanitize ad hoc at call sites, and do not reintroduce
a raw `os.Rename` / `os.RemoveAll` / `os.ReadFile` etc. on a data path.

**Boundary (the exception):** server-generated, non-attacker-influenced paths — `temp/`, `logs/`,
`sqlite/`, and OCR / HTML-to-PDF / email scratch files — legitimately use the generic `utils` helpers or
raw `os.*`. They are **not** data paths and must **not** be forced through the data-scoped helpers (which
would reject a non-`data/` path). The deliberate raw reads are the AI image readers in
`services/receipt_processing.go` (the OpenAI and Gemini readers; Ollama reads via imagick), whose path may
be a `temp/` file or a data file depending on the PDF branch — and the data-file case is now guaranteed
contained at construction because `BuildFilePath` asserts containment on the full path.

## Testing Patterns

Each package typically has:
- `main_test.go` - Test setup and teardown
- `*_test.go` - Unit tests for specific functionality
- Test utilities in `internal/utils/testing.go` and `internal/repositories/testing.go`

Tests use dependency injection patterns and mock implementations for external services.

## Testing Guidelines for Claude

When working with tests in this codebase, follow these critical requirements:

### Test Execution Requirements
- **ALWAYS run tests after writing them** - When asked to write tests, you MUST run them to verify they pass
- **Report coverage** - Always report the coverage of files impacted by the tests using `go test -coverprofile=coverage.out -covermode=atomic`
- **Verify all tests pass** - Never consider test writing complete until all tests are verified to pass

### Test Database Cleanup
- **Failed tests may leave behind `app.db` files** in test directories (e.g., `services/app.db`, `handlers/app.db`)
- **These MUST be removed** before rerunning tests to avoid conflicts
- **CRITICAL**: Only remove `app.db` files from test directories, NEVER delete anything from the `sqlite/` directory
- Example cleanup locations: `internal/services/app.db`, `internal/handlers/app.db`, etc.

### Test Workflow
1. Write tests following existing patterns in the codebase
2. Run tests to verify they pass: `go test -v ./...`
3. Generate and report coverage: `go test -coverprofile=coverage.out -covermode=atomic -v ./...`
4. If tests fail, check for and remove any `app.db` files in test directories
5. Re-run tests until all pass
6. Report final coverage results for impacted files

## OCR and Image Processing

- Tesseract OCR integration via `otiai10/gosseract`
- ImageMagick integration for image processing and format conversion
- Supports HEIC format conversion to standard image formats
- Python dependencies for additional image processing capabilities

## Email HTML to PDF Rendering

- HTML email bodies are rendered to PDF via `chromedp` running headless Chromium
- The Chromium binary path is read from the `CHROMIUM_BINARY_PATH` env var
  (defaults to `/usr/bin/chromium`); installed by `set-up-dependencies.sh`
- The Chromium process sandbox is **off by default** because the supported
  docker images run as root, where chromium's sandbox refuses to start.
  Operators running the API as a non-root user can opt back in by setting
  `CHROMIUM_SANDBOX=true`
- External network resource loads (remote images, CSS, fonts) are
  **blocked by default** to remove an SSRF / tracking-pixel surface.
  Inline `data:` URIs and the file:// page itself remain allowed. To
  permit remote loads (useful when receipts depend on remote logos or
  product imagery), set `CHROMIUM_ALLOW_EXTERNAL_RESOURCES=true`
- Implementation: `internal/services/html_to_pdf.go` (HtmlToPdfService.Render)
- The rendered PDF is saved on the receipt as a `FileData` and routed
  through the existing `repositories.ConvertPdfToJpg` pipeline so vision and
  OCR models receive an image, exactly like a PDF email attachment
- Gating: only runs when `EmailBodyProcessingEnabled` is true on at least one
  consuming group (per `shouldRenderEmailBodyPdf` in `wranglerasynq/email.go`)
- For an email with both an attachment and an HTML body, each per-attachment
  receipt is augmented with a copy of the body PDF and both images are sent
  to the LLM together; when the body is sent as an image, its text is dropped
  from the prompt to avoid duplication

## Testing Requirements

**All new code must have accompanying unit tests.**

Before considering any work complete:

1. Write unit tests for all new functions and endpoints
2. Follow existing test patterns in the codebase (see `main_test.go` files for setup)
3. Mock external dependencies (database, services, etc.)
4. Run the full test suite: `go test -v ./...`
5. Ensure all tests pass before submitting changes

Tests should cover:

- Happy path scenarios
- Error handling and edge cases
- Input validation
- Authentication/authorization logic

## API Documentation

- OpenAPI 3.1 specification in `swagger.yml`
- API serves on port 8081 by default
- All endpoints require JWT authentication except login/signup

## Roles & Permissions

A configurable role/permission system. Administrators can define roles from granular permission
strings at two scopes — **application** and **group** — and assign them to users / group members.
**Handlers now enforce these permissions** (see "Enforcement status" below). The legacy
`models.UserRole` (`ADMIN`/`USER`) and `models.GroupRole` (`OWNER`/`EDITOR`/`VIEWER`) enums have
been **removed from the backend** — the Go enum types, the `User.UserRole` model field, the
`Claims.userRole` JWT field, and the `DeriveLegacy*` shims are all gone. "Admin" is now
defined by the app permission `app.users.read` (the seeded **Legacy Admin** role grants it; **Legacy
User** omits it). The legacy remnants are the **physical** `user_role`/`group_role` DB columns,
retained on existing installs so the one-time data migration can still back-fill from
them (see "Legacy role assignment" below). `user_role` is purely physical (no Go field; GORM never
creates it on fresh installs). `group_role` is the one exception: `GroupMember.GroupRole` is
**temporarily re-declared** on the model as a plain nullable string (`json:"-"`, never read) — see
"Legacy `group_role` column on upgrade" below — so AutoMigrate manages it on all installs again. Both
will be dropped in a later release.

### Permission registry

- `internal/permissions/registry.go` is the **hardcoded source of truth** for every permission.
  Each entry is a `Descriptor{ Key, Label, Description, Category, Scope }`; `Scope` is `APP` or
  `GROUP`. Helpers: `All()`, `Get(key)`, `Exists(key)`.
- **String format:** `scope.domain[.subdomain].action` — e.g. `app.users.create`,
  `group.receipts.read`. Permissions are **CRUD-granular** (`create`/`read`/`update`/`delete` per
  domain); distinct non-CRUD actions stand alone (e.g. `app.system-settings.restart-task-server`,
  `group.activities.rerun`, `group.receipts.duplicate`).
- Exposed to clients via `GET /permission` (returns the descriptors for the role-editor UI) and
  mirrored in the `swagger.yml` `Permission` enum. `permissions/registry_test.go` enforces that the
  registry and the swagger enum stay in sync.
- **Adding a permission:** add the constant + `Descriptor` in `registry.go`, add the key to the
  `Permission` enum in `swagger.yml`, then regenerate clients (see "API Client Generation") —
  **including `mobile/api/`, in the same change**. A `ScopeGroup` key is picked up automatically by
  the seeded **Legacy Owner** (`LegacyGroupOwnerKeys()` = every group-scope key) on the next boot, so
  it lands in essentially every user's `AppData` the moment the server upgrades.
- **Catalog vs. data — do not `$ref` the `Permission` enum onto `AppData`.** The enum is the
  **catalog** of which permissions exist (`Role.permissions`, `UpsertRoleCommand.permissions`,
  `PermissionDescriptor.key`, the `permission` query param). A user's **effective** permissions
  (`AppData.appPermissions` / `.groupPermissions`) are server-resolved **data** and ride as plain
  `type: string`. Generated clients render an enum as a *closed* set — the Dart `EnumClass` throws on
  an unknown value and fails the entire `AppData` parse, which hard-fails login on every
  already-released mobile build. That shipped twice (2026-07-24 `group.members.create`; 2026-08-06
  `group.members.grants.update`). Granted strings may also be **wildcards** (see "Matcher"), which an
  enum cannot represent. Pinned by `permissions/registry_test.go` →
  `TestAppDataEffectivePermissionsAreUntypedStrings`.

### Matcher

- `internal/permissions/matcher.go` is a **pure** matcher over a granted `[]string`:
  - `HasAll(granted, required...)` — logical AND (the default; a single-permission check is just
    `HasAll(granted, "x")`).
  - `HasAny(granted, required...)` — logical OR.
- Wildcards in a *granted* string are honored: `*` matches anything, a trailing `group.*` matches
  any deeper key, and a mid-segment `*` matches exactly one segment. Both helpers deny when no
  required permission is supplied. The `:sub-scope` suffix (e.g. `read:any`) is matched literally —
  `:any` superset semantics are **not implemented yet**.

### Data model

- **App roles:** `AppRole` + `AppRolePermission` (permission strings in a child table).
- **Group roles:** `GroupRoleDefinition` + `GroupRolePermission`, plus `GroupRoleCategoryGrant` /
  `GroupRoleTagGrant` (composite-PK join rows) for per-role category/tag visibility. A group role
  with **no** grant rows is **unrestricted** (sees every category/tag); a non-empty grant set
  restricts members to exactly those ids — restriction is opt-in, so legacy/system roles (no grants)
  keep seeing everything and no data migration is needed.
  - **Per-member grants** (`GroupMemberCategoryGrant` / `GroupMemberTagGrant`, composite-PK
    `{UserID, GroupID, CategoryID|TagID}`) are a second, finer layer hanging off the **membership**,
    for the case a shared role cannot express ("Alice sees Child A and B, Bob sees Child C"). The two
    layers compose by **intersection** — the role is a ceiling, the member narrows within it. See
    "Category/tag grant resolution" below for the full rule, the fail-closed flags, the dedicated
    write endpoint and its own permission, and the grant-row lifecycle hazards. Categories/tags are **global** (no
  `GroupId`); a grant is a per-group-role slice of the global pool. CRUD persists/returns the grants
  and `PermissionService` resolves them (see "Category/tag grant resolution" below); wiring the
  resolved sets into AppData delivery and request enforcement is rolled out in later slices.
  - **Paid-by visibility** is a third, **row-level** grant type: `GroupRolePaidByUserGrant`
    (composite-PK `{GroupRoleID, UserID}`) plus an `IncludeOwnPaidReceipts` bool column on
    `GroupRoleDefinition`. It restricts **which receipts** a member sees by the receipt's
    `paid_by_user_id` (vs. category/tag grants, which strip fields off a still-visible receipt). Same
    opt-in rule: no grant rows **and** `IncludeOwnPaidReceipts == false` ⇒ unrestricted (see every
    payer). The bool is the relative "their own receipts" token — the member's own id is unioned in at
    **resolution** time (never stored/cached, since the cache is role-keyed and shared across members),
    so a role granting only specific users (bool false) is a "pure reviewer" that can't see its
    holder's own receipts. `UpsertRoleCommand`/`RoleView`/swagger carry `paidByUserGrants` +
    `includeOwnPaidReceipts`; the granted user ids are existence-validated like category/tag grants
    (`ErrInvalidGrant`). See "Paid-by visibility enforcement" below.
  - **Report-template access** is a fourth grant type: `GroupRoleReportTemplateGrant` (composite-PK
    `{GroupRoleID, ReportTemplateID, Permission}` — the action folded into the key) restricting which
    saved report templates a role's members may act on, per action. Same opt-in rule (empty =
    unrestricted) with a `ReportTemplateGrantsRestricted` fail-closed flag; carried on
    `UpsertRoleCommand`/`RoleView`/swagger as `reportTemplateGrants` (a per-template action list). See
    "Report-template access" in the Reports section for the full three-layer model + the `*All` bypass.
- **Assignment:** nullable FKs `User.AppRoleID` and `GroupMember.GroupRoleID` (one app role per
  user; one group role per group membership). Nullable because per-create assignment is best-effort
  (the FK is left `nil` rather than failing creation when no role can be resolved, e.g. an unseeded
  test DB) and because the one-time migration back-fills pre-existing rows that start `nil`.
- `IsSystem` marks protected, non-editable/non-deletable roles; `IsDefault` (on **both** `AppRole`
  and `GroupRoleDefinition`) marks the single default role for that scope — the role assigned to new
  accounts (app) or to group creators (group). See "Default roles" below.

### Role CRUD

- Data access in `repositories/roles.go`; business logic in `services/roles.go`. Guards: system
  roles are immutable/undeletable, a role's scope can't be changed (type-mismatch error), an
  assigned role can't be deleted, and the **default role for a scope can't be deleted**
  (`ErrRoleIsDefault`) — pick another default first.
- Endpoints (`routers/role.go`, gated by `app.roles.*`): `GET /role`, `POST /role`,
  `PUT /role/{roleId}`, `PUT /role/{roleId}/default?scope=APP|GROUP` (make this role the scope's
  default; allowed on system roles; gated by `app.roles.update`), `DELETE /role/{roleId}?scope=APP|GROUP`.
- `commands.UpsertRoleCommand` validates that every permission exists in the registry and matches
  the role's scope. It also carries `categoryGrants` / `tagGrants` (category/tag ids): grants are
  rejected on APP scope and dedup-checked in `Validate()`, and their existence is verified against
  the DB in `RoleService` (`ErrInvalidGrant` → 400). `repositories/roles.go` syncs grant rows with
  the same delete-then-insert pattern as permissions (`replaceGroupRoleGrants`, with the nested
  Category/Tag association `Omit`-ted so only join rows are written), preloads them in
  `GetGroupRoleById` / `GetAllRoles`, and cascades them on role delete. `structs.RoleView` is the
  read model (includes `assignedCount`, `isDefault`, and `categoryGrants` / `tagGrants` — empty
  slices for app roles).

### Seeded system roles (legacy-equivalent)

- `repositories.SeedSystemRoles` (`repositories/seed_roles.go`) seeds five immutable
  (`IsSystem = true`) roles on startup — wired into `InitDB`, runs in all deploy envs (it is
  structural, unlike the bootstrap admin user). Their permission sets reproduce the capabilities of
  the historical `ADMIN`/`USER` app roles and `OWNER`/`EDITOR`/`VIEWER` group roles **exactly**, so
  upgrading installs see **zero behavior change**: **Legacy Admin** (every app permission), **Legacy
  User** (the app actions a plain `USER` could do), **Legacy Viewer** / **Legacy Editor** / **Legacy
  Owner** (the group VIEWER / EDITOR / OWNER tiers; Owner = every group permission). The sets live in
  `permissions/legacy.go` (`Legacy*Keys()` helpers) and were derived from the actual handler-level
  gating, not the desktop UI presets. **Deliberate exceptions** (Legacy User omits these): `app.users.read`
  — it gates the admin user listing (the unpaged `GET /user/` **and** the paged `POST /user/getPagedUsers`
  that the desktop "Manage Users" page reads from); user dropdowns instead read from AppData via
  `app.account.read`, so granting it would only expose the admin "Manage Users" page to normal users; and
  `app.categories.read` / `app.tags.read` — omitted as part of the category/tag grant lock-down, since they
  gate the GLOBAL category/tag lists; normal users now get only the per-group filtered catalogs (the
  `app.categories.create` / `app.tags.create` permissions are retained for inline creation); and
  `app.reports.read` (the app-level gate on the desktop report builder route + nav and the saved-template
  read endpoints), `app.reports.generate` (the app-level gate on report generation, ANDed with the
  per-group check), and `app.reports.duplicate` (duplicating a saved template). Reporting is
  **admin-by-default**: Legacy Admin picks these up automatically (its set is every app permission), while
  non-admins get them only via a custom role that grants them. Per-group generation additionally stays
  gated by the group-scoped `group.reports.read`. See "Category/tag delivery on AppData" below.
- `SeedSystemRoles` creates the roles with `IsDefault = false`; the **default** per scope is set
  separately by `EnsureDefaultRoles` (see "Default roles" below), the one-time data migration assigns
  the roles to existing users/members, and enforcement is wired in `HandleRequest`.
- Idempotent and **self-reconciling**: keyed on role `Name` (a `uniqueIndex`), safe on every boot. A
  missing role is created; an existing one has any permissions it lacks **added — add-only, so a
  permission already on the role is never removed** (`missingPermissions` computes the add set). The
  five role names are shared constants (`repositories/system_role_names.go`, `Legacy*RoleName`) used
  by both the seeder and the migration.
- Because reconciliation is add-only, a permission **added** to the registry later flows into an
  already-seeded Legacy Admin / Legacy Owner on the next boot — both sets are dynamic over the
  registry (`LegacyAppAdminKeys` / `LegacyGroupOwnerKeys`) — while a capability an install already
  holds is never stripped.

### Default roles

- Exactly one app role and one group role are the **default** (`IsDefault = true`): the role assigned
  to a new account on signup/admin-create, and to a group's creator on group creation. This is a
  required invariant — there is always exactly one default per scope.
- `repositories.EnsureDefaultRoles` (`repositories/seed_roles.go`) enforces it on every boot,
  immediately after `SeedSystemRoles` in `InitDB`: if a scope has **no** default, it flags the
  legacy-equivalent role (**Legacy User** for app, **Legacy Owner** for group), so upgrades and fresh
  installs behave exactly as before. It only acts when no default exists, so it never overrides a
  default an admin chose, and it self-heals dev DBs created before the `app_roles.is_default` column.
- Admins change the default via `PUT /role/{roleId}/default?scope=…`
  (`RoleService.SetDefaultRole` → `SetDefaultAppRole`/`SetDefaultGroupRole`), which clears the prior
  default and sets the new one in one transaction. System roles are eligible (the legacy defaults are
  system roles). The current default cannot be deleted (`ErrRoleIsDefault`).
- **Per-create assignment** uses the defaults: `UserRepository.CreateUser` sets `User.AppRoleID`
  (Legacy Admin for the first user — `isAdmin := usrCnt == 0`, resolved by `resolveAppRoleId(tx,
  isAdmin)` — so the bootstrap admin is never locked out; the default app role otherwise), and
  `GroupRepository.CreateGroup` sets the creator's `GroupMember.GroupRoleID` to the default group
  role. Both are **best-effort**: if the role can't be resolved (e.g. an
  unseeded test DB), the FK is left `nil` rather than failing creation. Members added to *existing*
  groups (and explicit role choices in the admin user/group-member forms) are assigned via the
  modern-role authoring flow — see "Modern role assignment in authoring flows" under "Enforcement
  status".

### Skipping the personal group per app role

- Every new user normally gets **two** groups in `UserRepository.CreateUser`: the personal
  `"My Receipts"` group and the virtual `"All"` aggregate group (`IsAllGroup`, which backs the
  dashboard's all-receipts view). For accounts that are only ever meant to live in a few specific
  shared groups, the personal group is dead weight that accumulates as clutter in group management.
- **`AppRole.SkipDefaultGroupCreation`** (`models/app_role.go`, `not null;default:false`) opts a
  role's new users out of the personal group. The `"All"` group is **always** created, so the
  account still has a working dashboard landing page and sees receipts from groups an admin adds it
  to. AutoMigrate adds the column with default `false`, so existing roles/installs are unchanged —
  **no data migration**.
- **App-scope only**, mirroring the group-scope-only `seesAllMembers` in the opposite direction:
  `UpsertRoleCommand.Validate` rejects it on GROUP scope (`skipDefaultGroupCreation` error key), and
  `structs.RoleView` serializes `false` for every group role. Carried on `swagger.yml` for both
  `Role` and `UpsertRoleCommand`.
- **Creation-time only.** `CreateUser` resolves the flag from the user's just-assigned
  `AppRoleID` via `UserRepository.appRoleSkipsDefaultGroup` → `RoleRepository.AppRoleSkipsDefaultGroup`
  (a single-column read, mirroring `GetGroupRolePaidByConfig`) and skips only the `"My Receipts"`
  `CreateGroup` call. Assigning the role later — or unchecking the box — never adds or removes a
  group for an existing user. Resolution is **best-effort** in the same way as `resolveAppRoleId`: a
  nil or **missing** role (`gorm.ErrRecordNotFound`) falls back to creating the group rather than
  failing user creation, while any *other* lookup error propagates and rolls the creation back — a
  transient DB failure must not silently decide which groups an account is created with. (The
  missing-role branch is defensive: `User.AppRoleID` is an `OnDelete:RESTRICT` FK, so a non-nil id
  always references a live row.)
- Because the flag lives on the role, it applies at **every** creation path — admin create (explicit
  `appRoleId`) and public self-signup (which uses the configured **default** app role, so flagging
  the default makes signups skip the personal group too). The bootstrap admin resolves to Legacy
  Admin, a seeded system role whose column is the `false` zero value and which the UI opens
  read-only, so it is never affected.
- `UpdateAppRole` writes the column through the **map form** `Updates` — GORM's struct form skips
  zero-value bools, which would leave a toggled-off flag stuck on (the same reason
  `UpdateGroupRole` uses it).
- Tests: `commands/upsert_role_command_test.go` (APP accepted / GROUP rejected),
  `repositories/roles_test.go` (`TestAppRoleSkipDefaultGroupCreationRoundTrips` incl. the
  toggle-off case, `TestGetAllRolesReturnsSkipDefaultGroupCreation`), `repositories/users_test.go`
  (`TestCreateUserSkipsDefaultGroupForFlaggedAppRole` — only `"All"` remains — and
  `TestCreateUserCreatesDefaultGroupForUnflaggedAppRole`), and `services/roles_test.go`
  (service round-trip + group roles never carrying it).

### Legacy role assignment (one-time data migration)

- A startup data migration back-fills the new role assignments from the legacy role values so
  existing installs upgrade with **zero behavior change**: each user's `user_role` maps onto the
  matching `User.AppRoleID` (`ADMIN` → Legacy Admin, `USER` → Legacy User) and each member's
  `group_role` onto `GroupMember.GroupRoleID` (`OWNER`/`EDITOR`/`VIEWER` → Legacy
  Owner/Editor/Viewer). Lives in `repositories/data_migrations.go` (`assignLegacyEquivalentRoles`).
- **Reads the physical columns as plain strings, not enum Go fields.** The legacy enum types were
  removed, so the migration matches the `user_role`/`group_role` values as plain strings and **guards
  each back-fill loop with `tx.Migrator().HasColumn(...)`**. For `user_role` (no Go field) the guard
  is the real safety net — upgrading installs keep the physical column so the back-fill runs, while
  fresh installs never created it so the guard skips cleanly instead of erroring with "no such
  column". For `group_role` the guard is now effectively always-true, since `GroupMember.GroupRole`
  is re-declared on the model (so AutoMigrate always creates the column); the back-fill still no-ops
  on fresh rows because their value is `""`, which matches no legacy enum. There is deliberately **no
  drop-column migration**, to preserve this upgrade path.
- **Tracking:** one-time data migrations are recorded in a `data_migrations` ledger
  (`models.DataMigration`, keyed by unique `name`) — distinct from GORM schema AutoMigrate. The
  runner `RunDataMigrations` skips any migration already in the ledger and otherwise runs it **and**
  writes the ledger row in a single `db.Transaction`, so a failure rolls back and retries next boot.
  Append new one-time migrations to the `dataMigrations` registry slice.
- Wired into `InitDB` **after** the bootstrap-admin step (roles must be seeded first, and this order
  also assigns a fresh install's first admin). `InitDB` is only called from `main.go`, never the
  test harness, so the migration does not auto-run in tests.
- Updates are guarded with `... IS NULL`, so a role an admin has already set through the new UI is
  never overwritten — defense-in-depth on top of the ledger. The migration is one-time over existing
  rows; per-create assignment for *new* users / group creators is handled by the default-role wiring
  (see "Default roles" above).
- Tests: `repositories/data_migrations_test.go` (assignment, idempotency, ledger short-circuit,
  no-clobber, and the `HasColumn`-guard skip path when the legacy `user_role` column is absent). The
  tests seed `user_role` via raw `ALTER TABLE ... ADD COLUMN` DDL since AutoMigrate no longer creates
  it; `group_role` is left to AutoMigrate (it is back on the model) and is never added/dropped by the
  test helpers — dropping it would leave a field-without-column state that breaks later group_members
  inserts in the package's test run.

### Legacy `group_role` column on upgrade

- On databases upgraded from before the role rework, the obsolete `group_members.group_role` column
  survives **`NOT NULL` with no default** (AutoMigrate never drops columns). Because v7 stopped
  writing it, every new `group_members` INSERT violated the constraint (HTTP 500 on creating a group,
  a user — which auto-creates personal groups — or adding a member). Existing data, logins, reads,
  and receipt CRUD were unaffected.
- **Fix:** `GroupMember.GroupRole` is **temporarily re-declared** as a plain, nullable, `json:"-"`
  string (`internal/models/group_member.go`). Two effects: GORM writes the zero value (`""`) on every
  INSERT, satisfying any leftover `NOT NULL`; and because the model field is nullable, AutoMigrate
  relaxes the existing `NOT NULL` column to nullable on upgraded installs (Postgres `ALTER COLUMN …
  DROP NOT NULL`, MariaDB `MODIFY …`). `json:"-"` keeps it off the API contract, so **no swagger /
  client regeneration**. Nothing reads the field. Targets PostgreSQL/MariaDB; SQLite is not a goal
  (it is handled incidentally by GORM's portable migrator). `users.user_role` needs no such treatment
  — it is already nullable with a default.
- **Removal plan:** drop this field together with a one-time migration that drops the `group_role`
  column once all installs have upgraded past this release.

### Permission checks (`PermissionService`)

- `services/permission.go` exposes four **scope-separated** entry points:
  `HasAppPermissions` / `HasAnyAppPermission` and `HasGroupPermissions` / `HasAnyGroupPermission`
  (each App/Group pair = AND default + OR variant).
- Each call **resolves the user's current permissions from the database** (the user's app role for
  app checks; the group membership's group role for group checks) and matches them with the pure
  matcher. The **JWT is never trusted for authorization** — permissions are always re-read. A user
  with no assigned role, or a non-member of the group, resolves to no permissions (deny).
- Required keys are scope-guarded: passing an `app.*` key to a group check (or an unknown key)
  returns an error, catching call-site bugs.
- Backed by a small role-permission cache (`services/permission_cache.go`) keyed by `scope + roleId`
  and invalidated in `RoleService.UpdateRole` / `DeleteRole`. Only a role's permission *list* is
  cached; a user's role *assignment* is resolved fresh on every check, so re-assigning a user takes
  effect immediately.

### Category/tag grant resolution (`PermissionService`)

Category/tag visibility resolves from **two grant layers**, composed by **intersection**:

1. **Group role grants** — `GroupRoleCategoryGrant` / `GroupRoleTagGrant`, shared by everyone holding
   the role. This is the **ceiling**.
2. **Group MEMBER grants** — `GroupMemberCategoryGrant` / `GroupMemberTagGrant`, composite-PK
   `(UserID, GroupID, CategoryID|TagID)`, per individual. This **narrows within** the ceiling.

The member layer exists because a role is shared and so cannot express "Alice sees Child A and B, Bob
sees Child C" — the foster-care requirement. It is scoped to the **membership**, not the user, because
categories are global but visibility always resolves in a group context: a flat per-user list would
follow the member into unrelated groups and blank their catalog there.

**Why intersection, not union.** Union can only ever *add*, so any role carrying grants would floor
every member at the role's full set and an individual assignment could never *restrict* anyone —
which is the entire point of the member layer. Intersection also keeps a role an auditable ceiling
("no holder of this role can see outside its set") and makes role grants safe to widen, since widening
a role never widens an individually-assigned member.

- `services/grant.go` — `resolveEffectiveGrants(userId, groupId)` composes the layers and returns an
  `effectiveGrantSet`; `composeGrantLayer` applies the rule per resource:

  ```text
  role requires individual && member unconfigured -> see nothing (fail closed)
  effective = ALL
  if the role grants a non-empty set          -> effective ∩= role set
  if the membership opted into restriction    -> effective ∩= membership set
  if neither narrows                          -> unrestricted
  ```

  `GetGroupCategoryIdsForUser` / `GetGroupTagIdsForUser` keep their `(allowedSet, unrestricted, err)`
  signature, so **every existing call site inherits the member layer with no change** —
  `grant_filter.go`, AppData (`auth.go`), the AI candidate lists (`receipt_processing.go`),
  `receipts.go`, and MCP `tools.go`. **`unrestricted == true` means see-all** and the set is `nil`; a
  `false` flag with an **empty** set means **see nothing**. A non-member is unrestricted — grants only
  *narrow* within an already-permitted group; they never *grant* access (the handler permission gate is
  the access control). Categories and tags are independent. `GetVisibleCategoriesForUser` /
  `GetVisibleTagsForUser` filter a full slice to the visible subset (pass-through when unrestricted).
- **Fail-closed flags.** `GroupMember.CategoryGrantsRestricted` / `TagGrantsRestricted`
  (`json:"-"`, derived on save from the *submitted* sets) keep a configured membership restricted after
  its grant rows are emptied by a category deletion cascade, instead of silently widening back to the
  role's set. Same purpose and pattern as `GroupRoleDefinition.PaidByVisibilityRestricted`.
- **`RequiresIndividualCategoryGrants` / `RequiresIndividualTagGrants`** on `GroupRoleDefinition`
  (default **false**, so existing roles are unchanged) make per-member assignment mandatory: a member
  holding the role with no individual grants sees **nothing** rather than the role's set — so
  forgetting to assign a newly added member fails closed. Carried on `UpsertRoleCommand`/`RoleView`,
  group scope only (rejected on APP), persisted by
  `RoleRepository.SetGroupRoleIndividualGrantConfig` — a **separate** method from
  `CreateGroupRole`/`UpdateGroupRole` (whose positional signatures already end in two bools; appending
  two more would make four adjacent transposable bools), mirroring
  `ReplaceGroupRoleReportTemplateGrants`.
- **Caching.** The role layer is cached (`services/grant_cache.go`, keyed by **group-role id**, same
  generation-counter invalidation as the permission cache, evicted in `RoleService.UpdateRole` /
  `DeleteRole`); the `requiresIndividual*` flags ride that entry. The **member layer is deliberately
  NOT cached** — one indexed lookup, resolution is already batched once per group per pass by
  `newReceiptGrantFilter`, and caching would add a second invalidation surface across four write paths.
  A user's role *assignment* is resolved fresh each call. A category/tag deleted out from under a
  cached grant id is benign — a stale id never matches a real row when filtering.
- **Write surface: `PUT /group/{groupId}/member/{userId}/grants`** (`handlers.UpdateGroupMemberGrants`
  → `GroupService.UpdateMemberGrants`), body `{categoryGrants, tagGrants}`. A **dedicated endpoint,
  not a field on the group-member upsert**, for two reasons: it carries its own permission
  **`group.members.grants.update`** (separate from `group.members.update`, so a restricted member who
  can manage the roster cannot thereby widen their own visibility — the feature's main escalation
  hole), and it keeps grant writes off `UpdateGroup`'s wholesale roster replace, where the user form
  and group form posting partial rosters would clobber each other. Both ids come from the **URL only**.
  Submitted ids are checked for existence (`ErrInvalidGrant` → 400) and against the member's role
  ceiling (`GrantCeilingViolation`, naming the offending ids → 400), which is what makes the
  empty-intersection state unreachable through the API. `Legacy Owner` picks the permission up
  automatically (`LegacyGroupOwnerKeys()` = every group-scope key) via add-only seed reconciliation.
- **Lifecycle (the sharp edge).** Grant rows are keyed by `(user, group)` with **no FK to
  `group_members`**, and the teardown paths delete memberships with a raw `tx.Delete` (no Go-side
  cascade; SQLite does not enforce FKs unless `PRAGMA foreign_keys=ON` per connection). Orphaned rows
  would be **silently re-adopted when a removed user rejoins the same group**, restoring revoked
  visibility with nothing in the UI to show it. So `repositories/group_members.go` exposes
  `DeleteMemberGrants` / `DeleteMemberGrantsForUser` / `DeleteMemberGrantsForGroup` /
  `DeleteOrphanedMemberGrants`, called explicitly from `UserService.DeleteUser`,
  `GroupService.DeleteGroup`, and `GroupRepository.UpdateGroup` (after its association replace, phrased
  as "whatever no longer has a membership" so it is correct regardless of how GORM's replace behaves,
  and leaves retained members untouched). **Do not remove these calls.**
- **`UpdateGroup` must carry the restriction flags forward.** It persists `GroupMember` rows rebuilt
  from the request command (which carry both flags at their **zero value**) twice — once via the
  `FullSaveAssociations` `Updates`, once via the association `Replace`. So `UpdateGroup` reads
  `GetMemberGrantFlagsForGroup` **at the very top of the transaction, before either write**, and
  re-applies each retained member's flags to the rebuilt structs. Without it, a plain group edit —
  even just a rename — clears every member's restriction and silently widens them back to their
  role's full set. The flags are **carried forward, not recomputed** from the surviving grant rows: a
  membership configured and then emptied by a category deletion must stay restricted, which is the
  whole reason the flags exist apart from the rows. Pinned by
  `repositories/group_member_grants_test.go` → `TestUpdateGroupPreservesRetainedMemberGrantFlags`.
- **`GroupMember.CategoryGrants` / `TagGrants` are transient view fields (`gorm:"-"`)**, populated on
  read by `LoadMemberGrantsForGroup(s)` at the roster serialization boundary (AppData,
  `GetGroupsForUser`, `GetGroupById`) — deliberately **not** GORM associations, which would put the
  grant tables inside `UpdateGroup`'s wholesale association write.
- Tests: `services/group_member_grants_test.go` (intersection matrix incl. the disjoint "see nothing"
  case, the fail-closed toggle, restriction surviving a category delete, **no resurrection on rejoin**,
  orphan cleanup keeping retained members, and the endpoint's ceiling/existence/non-member rejections).
- `services/grant_filter.go` is the **single** shared enforcement mechanism, reused by every read and
  write surface: `FilterReceiptCategoriesTags` / `FilterReceiptCategoriesTagsForReceipt` strip a
  receipt's `Categories`/`Tags` in place to the visible subset (resolving each group's grants at most
  once per pass); `ValidateCategoryTagSelection` checks receipt create/update ids against the allowed
  set. Both short-circuit on unrestricted resources and on **admin bypass** — `userBypassesGrants`
  treats a holder of `app.categories.read` / `app.tags.read` as exempt (they can already see the
  whole pool), keeping their view consistent with the global lists.
- **Receipt enforcement wiring:** every receipt surface that returns or accepts categories/tags is
  gated. **Reads** strip via `FilterReceiptCategoriesTags` (receipt-, item-, and linked-item-level):
  `GetReceipt`, `GetPagedReceiptsForGroup`, `GetReceiptsForGroupIds`, and both CSV export handlers;
  `DuplicateReceipt` strips the source before copying. **Aggregation** reads substitute instead of
  stripping via `SubstituteRestrictedCategoriesTags` (receipt-level; a hidden category/tag becomes a
  `(Restricted)` bucket rather than collapsing into `(None)`): the pie-chart service and the reporting
  data source. **Writes**
  (`CreateReceipt` / `UpdateReceipt`) call `enforceReceiptGrantSelection` — existing ids must be in
  the caller's grants (else 403) and a new-by-name category/tag requires `app.categories.create` /
  `app.tags.create`. **List/pie/export filters** are narrowed by `IntersectReceiptFilterWithGrants`
  so a restricted user can't probe receipt existence via a hidden category/tag filter. Search returns
  no categories/tags (`SearchResult` omits them), so it needs no filtering.
- **Update preserves hidden associations:** `UpdateReceipt` does a full association *replace*, so a
  restricted user's submission (missing the categories/tags they can't see) would drop them.
  `MergeHiddenReceiptCategoriesTags` re-adds the receipt-level hidden categories/tags before the
  replace (and runs *after* the selection check, which would otherwise reject them); the response is
  then stripped so the user still doesn't see them. **Known limitation:** this merge is
  **receipt-level only** — receipt items have no stable id across an update (they are deleted and
  recreated), so hidden *item-level* categories/tags cannot be matched back and are dropped when a
  restricted user edits a receipt. Closing that needs item identity (a separate change).
- **AI prompt:** `ReceiptProcessingService` carries a `UserId` (the user who triggered processing; 0
  for system-initiated, e.g. email polling). When set together with a `Group`,
  `getCategoriesString` / `getTagsString` restrict the candidate categories/tags fed to the model to
  that user's grants (via `GetVisibleCategoriesForUser` / `GetVisibleTagsForUser`), so a quick scan
  can't surface or auto-assign a category/tag the user isn't allowed to see. `MagicFillFromImage`
  takes the triggering `userId` and sets it on the service; system/email processing passes 0 and
  stays unrestricted (the resulting receipts are covered by read-stripping when any user views them).

### Paid-by visibility enforcement (`PermissionService`, `services/paid_by_filter.go`)

Row-level visibility by the receipt's `paid_by_user_id`, layered on top of the existing
group-membership scoping. `GetGroupPaidByUserIdsForUser(userId, groupId)` returns `(allowedSet,
unrestricted, err)` and returns a **freshly allocated** set (the requesting user's own id is unioned
in when the role sets `IncludeOwnPaidReceipts`) so it never mutates the role-keyed grant cache. There
is **no app-level bypass** — every receipt read gates on `group.receipts.read` (the very permission a
restricted member holds), and an admin who isn't a group member has no group role there so resolves to
unrestricted.

**Fail closed (`PaidByVisibilityRestricted`).** "Unrestricted" is keyed off a persisted
`GroupRoleDefinition.PaidByVisibilityRestricted` flag — set on save to `includeOwn || len(paidByUserGrants) > 0`
— **not** the live grant count. This matters because the `user_id` FK is `ON DELETE CASCADE`: a role
restricted to only `[X]` (include-own false) would otherwise become *unrestricted* (see-all) once X is
deleted and its grant row cascades away — a silent privacy widening. With the flag, a configured role
stays restricted and resolves to an **empty** allowed set → the `IN (0)` sentinel → "see nothing".
(A user delete does not evict the role grant cache, but that is benign: the deleted user's receipts are
gone too, and the cached flag stays true.) The flag is internal/derived — not on the API contract.

Because paid-by hides the **whole** receipt (not just fields), enforcement differs by surface:

- **Paged list** (`GetPagedReceiptsByGroupId`): a `PaidByAllowedResolver` is passed in from the
  handler/service (the repo can't import the service). The WHERE is added **before** the count, so
  `totalCount` stays correct — single-group adds `paid_by_user_id IN (allowed)` (empty restricted set
  ⇒ `IN (0)` no-match sentinel); the all-group view builds a per-group **disjunction**
  `(group_id=G AND paid_by_user_id IN s_G) OR (group_id=G2) …` so each group applies its own role.
- **Single receipt + dependent reads** (`HandleRequest` chokepoint): the `ReceiptId`/`ReceiptIds`
  blocks also select `paid_by_user_id`; `enforcePaidByVisibility` denies **403** when any resolved
  receipt is outside the caller's allowed set. This one place covers `GetReceipt`, image
  get/download/remove, comments, duplicate-source, update/delete, and the multi-id export (deny if
  **any** id is hidden). The `HasAccess` probe (the desktop receipt-route guard's check) does its own
  in-handler check, so it **also** calls `ReceiptPaidByVisible` after its group-permission check —
  otherwise the guard would admit the member to a hidden receipt that then 403s on fetch.
- **Search** (`handlers/search.go`): applies the predicate in SQL **before** `Limit(100)` via the
  shared `ReceiptRepository.ApplyPaidByDisjunction(query, memberGroupIds, resolver)` (the same per-group
  disjunction the all-group paged read uses). A post-fetch filter would be wrong here: hidden receipts
  filling the first 100 by date would drop a restricted user's visible matches.
- **`GetReceiptsForGroupIds`**: `FilterReceiptsByPaidBy` post-filters the returned slice — fine because
  it has no `LIMIT` (it returns all receipts for the groups). **Pie chart** and **CSV export** pass the
  resolver through `GetPagedReceiptsByGroupId`.
- **No request-filter intersection needed** (unlike category/tag grants): the paged query already
  row-filters on the allowed set, so a caller filtering by a payer they can't see intersects to
  nothing and can't probe receipt existence.
- **Scope boundary (intentional):** group totals / splitting / settlement aggregates stay unfiltered
  (paid-by restricts *browsing*, not the group's accounting — the per-viewer pie chart *is* filtered);
  write-side (which payer a member may *set*) is unchanged — this is read-visibility only. The
  settlement endpoint `GetAmountOwedForUser` sets `ReceiptIds` (for the permission gate) but marks the
  handler `SkipPaidByVisibilityCheck: true`, so the amount-owed total is identical for every member
  regardless of their paid-by filter — `HandleRequest` honors that flag to skip `enforcePaidByVisibility`
  while still enforcing `group.receipts.read`. Any future accounting endpoint should set the same flag.
- Tests: `services/paid_by_filter_test.go`, `repositories/receipts_test.go` (single + all-group
  disjunction count correctness), `handlers/receipt_paid_by_enforcement_test.go` (single-GET 403),
  plus the round-trip/validation cases in `repositories/roles_grants_test.go`,
  `services/roles_test.go`, and `commands/upsert_role_command_test.go`.

### Group member management (self-escalation guard)

`PUT /api/group/{groupId}` (`UpdateGroup`) replaces the whole member roster, including each member's
`GroupRoleID`, from the request body. It is still gated by `group.update` (you must be able to edit
the group to reach it), but the roster changes are additionally authorized by
`GroupService.AuthorizeGroupMemberChanges` (`services/groups.go`), called from the handler **before**
the repository write and returning **403** (`ErrGroupMemberChangeForbidden`) on any violation. This
closes GHSA-89mm-9qfv-cjg3 (a `group.update` member rewriting their own `groupRoleId` to escalate to
owner, or evicting the owner). The guard diffs the submitted roster against the current one and applies
two checks to every added / role-changed / removed row (unchanged rows are skipped, so a plain
name/settings edit never trips it):

- **CRUD gate:** adding a member requires `group.members.create`, changing a member's role requires
  `group.members.update`, removing a member requires `group.members.delete`.
- **Privilege ceiling** ("you can neither grant nor strip a privilege you do not hold"): the caller
  may only assign, or remove/replace, a role whose permission set is a **subset** of the caller's own
  current group permissions (resolved via `GetGroupPermissionsForUser`; a `nil`/empty role is always
  wieldable). This is what actually prevents self-escalation, independent of the CRUD gate.

A fourth, **separate** permission `group.members.grants.update` gates per-member category/tag
assignment (`PUT /group/{groupId}/member/{userId}/grants`). It is deliberately **not** folded into
`group.members.update`: bundling them would let a restricted member who can manage the roster edit
their own grants and lift the very restriction the feature enforces. See "Category/tag grant
resolution" above.

The four `group.members.*` permissions are group-scoped registry entries; **Legacy Owner** holds them
(it is the full group scope), Legacy Editor/Viewer do not (member management was historically
owner-only). Upgraded installs — whose Legacy Owner was seeded before these keys existed — pick them
up on the next boot from `SeedSystemRoles`' add-only reconciliation (see "Seeded system roles"
above); no dedicated data migration is needed, and
`repositories/seed_roles_test.go` → `TestSeedSystemRolesBackfillsGroupMemberPermissions` pins that
upgrade path. Tests: `handlers/group_member_authorization_test.go` (the PoC + happy paths) and
`services/group_member_authorization_test.go` (the guard's CRUD/ceiling matrix).

### Member isolation (presence privacy)

A group can be flagged **`IsolateMembers`** (`models.Group`): within it, a member cannot discover that
other members exist — not through the user directory, the group roster, receipts, comments, activities,
settlement, or notifications. A group role flagged **`SeesAllMembers`** (`models.GroupRoleDefinition`)
is the **supervisor** exemption: its holders see every member of an isolated group **and** are visible
to every member. Both default `false`, so existing groups/roles/installs are unchanged (AutoMigrate adds
the columns; no data migration).

**Isolation is resolved PER GROUP — "isolated means isolated."** Visibility is a function of
**(viewer, group)**, never a union across the viewer's groups. Inside an isolated group you see only
yourself + that group's supervisors, regardless of any other group you belong to. Co-members become
visible **only** through a shared **non-isolated** group (open means open) — and even then only on that
open group's own surfaces; the isolated group never leaks presence or settlement. This is deliberately
allowed to leave a cross-group aggregate incomplete (a settlement/report total may omit a hidden group's
dollars) — the truthful isolation guarantee wins.

**Two resolvers** (`services/member_visibility.go`):
- **`GetVisibleUserIdsForUserInGroup(viewerId, groupId)`** → `(set, unrestricted, err)` — the per-group
  resolver used by **every group-scoped surface** and by settlement. `app.users.read` ⇒ unrestricted; a
  **non-isolated** group ⇒ unrestricted; an isolated group where the viewer holds a **`SeesAllMembers`**
  role ⇒ unrestricted; an isolated group where the viewer is a **plain member** ⇒ `{self} ∪` that group's
  supervisors; a **non-member of an isolated group ⇒ `{self}` restricted** (an isolated roster must not
  leak to a non-member reader — e.g. an `app.groups.read` holder hitting `GetGroupById` via
  `OrAppPermissions` — who lacks the `app.users.read` directory exemption); a **non-member of a
  non-isolated group ⇒ unrestricted** (open group, preserving the paid-by / reporting contract for
  non-members, whose surfaces gate membership at the handler). `GetViewerGroupRow` LEFT-JOINs `groups` so
  the isolate flag + membership are known even for a non-member. Not cached; a batch spanning groups
  memoizes per group via `groupVisibilityResolver` / `visibleUserIdsByGroup` (the latter resolves the
  AppData roster path in two queries total).
- **`GetVisibleUserIdsForUser(viewerId)`** (the original UNION resolver, unchanged) is now used by
  **exactly one** surface: the flat `appData.users` directory (`FilterVisibleUserViews`). The union is
  mathematically the union of the per-group sets, so it already honors isolation (it never lists a peer
  you share only an isolated group with) — it is the correct name-resolution table, since the roster
  (`GroupMember`) carries only a `userId` and clients resolve the display name/avatar from this flat list.

**Read enforcement — every surface a user identity can reach a client (all PER GROUP):**
- **Directory** — `appData.users` uses the union resolver (above). **Rosters** — each group's
  `GroupMembers` is filtered with **its own** per-group set at the serialization boundary
  (`services/auth.go` AppData via `FilterGroupMembersForGroups`, the `GET /group/{id}` handler via
  `FilterGroupMembersForGroup`, and MCP `list_groups`). **Not** inside `GroupService.GetGroupsForUser` /
  `UserRepository.GetAllUserViews` — those also feed internal accounting and receipt processing.
- **Receipts (row visibility)** — the per-group visible set is intersected into the paid-by allowed set
  inside `paidByAllowedForGroup` (`services/paid_by_filter.go`), so a receipt whose `paid_by_user_id` is
  non-visible **in its own group** disappears from every read surface at once (paged list, single GET,
  search, CSV export, duplicate, `GetReceiptsForGroupIds`, pie chart, report data).
- **Field masking** — `services/member_masking.go` resolves per `receipt.GroupId` (memoized across a
  batch) and **nulls** user-reference fields whose id ∉ that receipt's group's visible set:
  `BaseModel.CreatedBy`/`CreatedByString` on the receipt and every nested entity, and item/linked-item
  `ChargedToUserId`. Fields are nulled, **not** replaced with `(Restricted)` — presence must be hidden,
  not announced. `paid_by_user_id` is never masked (row visibility already guarantees a visible payer).
- **Comments / activities (row drop)** — comments authored by a non-visible user are dropped;
  `GetActivitiesForGroups` drops rows whose `RanByUserId` is non-visible **in that activity's group**.
  Activity visibility is filtered **in the SQL query, before `COUNT`/`LIMIT`**, via
  `applyActivityVisibilityDisjunction` (`repositories/system_task.go`) — a per-group disjunction mirroring
  `ReceiptRepository.ApplyPaidByDisjunction`, fed by `PermissionService.ActivityVisibilityResolver`
  (the activity analogue of `PaidByListResolver`). So `TotalCount` and the returned page both reflect only
  visible rows with DB-side `LIMIT/OFFSET` preserved — a restricted member cannot infer hidden-peer
  activity from the total, and there is no unpaged in-memory fetch. The comment-notification fan-out has
  a **non-isolated fast path** (`recipientsWhoCannotSeeAuthor` reads the group's `isolate_members` once
  and skips the roster fetch + per-member resolver loop when the group isn't isolated).
- **Settlement** — `GetAmountOwedForUser` resolves visibility **per the receipt's group as it folds each
  item** into the counterparty balance map: a contribution counts only if the counterparty is visible in
  that receipt's group. So an isolated group contributes nothing about a hidden member even in the
  all-groups aggregate, while a shared open group contributes normally. This overrides the endpoint's
  paid-by exemption for isolated viewers.
- **Notifications** — the comment add/reply fan-out (`repositories/comments.go`, injected
  `AuthorVisibilityResolver func(authorId, recipientId, groupId)`) suppresses delivery to a recipient who
  can't see the comment author in the receipt's group.
- **Reporting + pie** — no dedicated code: reports read only through `ReportDataService.Rows` →
  `PaidByListResolver` (per group), called once per group by `ReportService.loadRows`, so an isolated
  group contributes only visible-payer receipts and an open group contributes everyone; the only
  user-identity a report row exposes is `paid_by` (row-hidden, never masked).

**Write-side guards:**
- **Receipt upsert** — `enforceReceiptMemberVisibilitySelection` rejects a `paidByUserId` or item
  `chargedToUserId` outside the creator's visible set **for the receipt's group** (403).
- **Update preservation** — `enforceReceiptChargedToPreservation` blocks a restricted editor from saving
  a receipt whose **stored** items are charged to a member they can't see in that group: the wholesale
  item replace + no stable item identity would silently drop the hidden charge, so the edit is rejected
  rather than lose data. **Known limitation** (same as item-level category/tag grants): preserving the
  hidden charge needs stable item identity — a separate change.
- **No member-invitation visibility guard.** Under per-group isolation, adding a user to a group never
  widens visibility in any *other* group (and adding to an isolated group where the caller is a plain
  member doesn't make the added user visible to the caller), so the old union-contamination attack is
  gone. The dedicated member-visibility invite checks were **removed**; the GHSA-89mm CRUD-gate +
  privilege-ceiling guard in `AuthorizeGroupMemberChanges` is untouched.

Co-members are visible only through a shared **non-isolated** group; an isolated group never leaks
presence or settlement — no cross-group operational discipline is required (per-group makes a shared open
group safe by design).

**Persistence:** `isolateMembers` rides `UpsertGroupCommand` → `CreateGroup`/`UpdateGroup` (the update
uses an explicit `Update("isolate_members", …)` so a toggle-off past the association `Omit` sticks);
`seesAllMembers` rides `UpsertRoleCommand` → `CreateGroupRole`/`UpdateGroupRole` and is surfaced on
`RoleView`, group scope only (mirroring `includeOwnPaidReceipts`). Both are on `swagger.yml`.

**Tests:** `services/member_visibility_test.go` (union + per-group resolver matrix, incl. the headline
"open-group peer hidden inside the isolated group"), `services/member_isolation_receipt_test.go`,
`services/report_data_test.go` (reporting isolation), `handlers/receipt_member_isolation_test.go`,
`handlers/receipt_charge_preservation_test.go`, `handlers/system_task_test.go` (per-group activity drop),
`handlers/users_test.go` (settlement incl. the cross-group isolated/open case),
`services/member_visibility_notifications_test.go`, plus persistence round-trips in
`repositories/groups_test.go` / `repositories/roles_grants_test.go` / `services/roles_test.go` /
`commands/upsert_role_command_test.go`.

### Enforcement status

Authorization is enforced centrally in `HandleRequest` (`handlers/generic_handler.go`) via the
`PermissionService`. Each handler declares its requirement on the `structs.Handler` it builds:

- `AppPermissions []string` — app-scoped permissions the caller must hold (logical AND).
- `GroupPermissions []string` — group-scoped permissions the caller must hold (AND) in **each**
  group resolved from `GroupId` / `GroupIds`, or from `ReceiptId` / `ReceiptIds` (the receipt's
  group is looked up automatically).
- `OrAppPermissions []string` — an app-scoped fallback; holding **any** of them bypasses the group
  check (e.g. an administrator viewing a group they aren't a member of). Replaces the old
  `OrUserRole`.

**`app.groups.delete` — deleting any group (Group Management):** the app-scoped counterpart to
`app.groups.read`. Reading all groups without being able to remove the abandoned/garbage ones left
admins unable to clean up, so `DeleteGroup` declares `OrAppPermissions: [app.groups.delete]` beside
its group-scoped `GroupPermissions: [group.delete]` — a holder deletes a group they are not a member
of, while an ordinary member still deletes their own via `group.delete`. Legacy Admin auto-includes
it (its set is every app permission); **Legacy User deliberately does not**. It is inert on its own:
the delete control lives on the Manage Groups page, which is reached behind `app.groups.read`. The
`IsAllGroup` → 400 guard is unaffected, so the virtual "All" group stays undeletable for everyone.

**`app.custom-fields.update` — editing a custom field (Custom Fields):** custom fields could be
created and deleted but never corrected, and deleting one also deletes every `CustomFieldValue`
attached to it across every receipt — so a typo in a name or an option was unfixable. `PUT
/customField/{id}` (`UpdateCustomField`, gated by `AppPermissions: [app.custom-fields.update]`) edits
the **name, description, and select options**. Legacy Admin auto-includes it (its set is every app
permission); **Legacy User deliberately does not** — it holds custom-fields create + read, matching
categories and tags, where create is granted and update/delete are not. No seed change and no data
migration: `SeedSystemRoles`' add-only reconciliation delivers it to an upgraded install on the next
boot.

Two invariants make the edit safe, enforced by `UpsertCustomFieldCommand.ValidateUpdate` (a **pure**
companion to `Validate` that takes the currently-stored `models.CustomField`, so the update rules sit
beside the create rules rather than in the handler) and re-asserted by the repository:

- **The type is immutable.** A `CustomFieldValue` stores its data in one of five type-specific columns
  (`StringValue` / `DateValue` / `SelectValue` / `CurrencyValue` / `BooleanValue`), so re-typing a
  field would silently mis-column every value already recorded against it. A differing `type` is a
  **400** (`{"type": "Type cannot be changed"}`), and `UpdateCustomField` never puts the column in its
  `Updates` map, so a caller that skips validation still cannot change it.
- **Options are renamed or appended, never removed.** `CustomFieldValue.SelectValue` holds an option
  **id**, so options are matched by a new `UpsertCustomFieldOptionCommand.Id` (zero = append) and a
  rename keeps the id — every receipt that picked it still resolves, now showing the corrected text.
  An option the command **omits is left untouched, not deleted**: that is the no-orphan guarantee, and
  it also makes a stale form (another admin appending an option meanwhile) a no-op rather than a 400.
  An option id the field does not own is a **400**, and the repository's option update is additionally
  scoped `WHERE id = ? AND custom_field_id = ?` so a foreign option is unwritable even if that check
  raced.
- **An option value is required and blank-checked after trimming.** Because a rename keeps the option
  id, an empty value would preserve the id and silently blank the label on every receipt whose
  `SelectValue` points at it. The rule lives in `Validate()` — **not** `ValidateUpdate()` — so create
  and update share one check (`UpdateCustomField` runs `Validate` first); it therefore also closes the
  older hole where a SELECT field could be *created* with blank options. `strings.TrimSpace` is used so
  the server agrees with the desktop's `trimmedRequiredValidator()`; the stored value is not trimmed.

`UpdateCustomField` uses the **map form** of `Updates` so a cleared description actually clears (the
struct form skips zero values — the same reason `UpdateAppRole` does), and maps
`gorm.ErrRecordNotFound` to **404**. (The older `GetCustomFieldById` handler still returns 500 for a
missing id; its test pins that and it was left alone.) `UpsertCustomFieldOptionCommand.CustomFieldId`
also had its json tag corrected from `custom_field_id` to `customFieldId` to match swagger — it is
never trusted either way, since create derives the FK from the parent association and update from the
request URL.

**`middleware.CanDeleteGroup` is permission-aware.** That middleware (`middleware/group.go`, wrapping
only the DELETE route) rejects when the **caller** belongs to ≤1 group — self-protection so a user
can't delete themselves out of every group. It ignores `{groupId}` and runs *before* `HandleRequest`,
so it would have blocked an administrator cleaning up a group they aren't in. It now short-circuits
for holders of `app.groups.delete` (resolved from the DB via `PermissionService`; a lookup error
fails closed with 500). It is **not** the authorization gate — that is still `HandleRequest`.

`HandleRequest` resolves the caller's effective permissions from the database (never the JWT) and
denies with `403` on any failure. The legacy `UserRole` / `GroupRole` / `OrUserRole` handler fields
and their checks have been **removed**. Every authenticated endpoint that previously had a legacy
role gate now has an equivalent permission gate; endpoints that touch only the caller's own data
(notifications, user preferences, own profile/claims/app-data, API keys, group lists) are gated by
dedicated self-service permissions (`app.notifications.*`, `app.user-preferences.*`,
`app.account.*`, `app.receipts.search`) included in the Legacy User set so existing users are
unaffected. Two endpoints are intentionally **not** permission-gated: the username-availability
lookup (used pre-auth during signup) and `ConvertToJpg` (a stateless image utility with no stored
resource to scope against). The role/permission management endpoints (`/role`, `/permission`) are
gated by `app.roles.*`.

**Effective permissions on AppData (desktop UI gating):** `GetAppData` (`services/auth.go`) includes the
caller's resolved permissions so the desktop can gate UI with them — `AppPermissions []string` and
`GroupPermissions map[uint][]string` (keyed by group id) — built via
`PermissionService.GetAppPermissionsForUser` / `GetGroupPermissionsForUser` (thin exported wrappers over
the cached `resolveAppPermissions` / `resolveGroupPermissions`). The JWT no longer carries any role
field at all (it holds only identity claims); the server always re-checks real permissions from the
DB on every request.

**Category/tag delivery on AppData (grant lock-down):** `GetAppData` also returns `GroupCategories
map[uint][]Category` and `GroupTags map[uint][]Tag` (keyed by group id) — each group's catalog
filtered to the caller's grants (the full pool when unrestricted), built via
`GetVisibleCategoriesForUser` / `GetVisibleTagsForUser`. The flat `Categories` / `Tags` arrays (and
the global `GET /category` / `GET /tag` endpoints) are **admin-only**: they're populated solely for
callers holding `app.categories.read` / `app.tags.read`, and empty otherwise. Because
`LegacyAppUserKeys` no longer grants those reads (only the create permissions remain), normal users
receive categories/tags **only** through the per-group `GroupCategories` / `GroupTags` maps — the
desktop receipt form sources its pickers from there.

**`app.api-keys.read-any` (Security):** listing *all* users' API keys (`GetPagedApiKeys` with
`associatedApiKeys=ALL`) requires `app.api-keys.read-any`, checked in the handler body via the
`PermissionService` — the legacy `token.UserRole == ADMIN` check in
`commands/paged_api_key_request_command.go` was removed. Legacy Admin auto-includes the new permission
(its set is every app permission); Legacy User does not.

**Per-create role assignment (done):** a new account (signup or admin-create) is assigned the
default app role, and a group's creator is assigned the default group role, via the default-role
wiring (see "Default roles" above), so accounts created after the one-time migration are no longer
locked out.

**Modern role assignment in authoring flows (done):** admin user-create/update and group-member
create/update assign **modern roles directly**. `SignUpCommand` carries `AppRoleID` and
`UpsertGroupMemberCommand` carries `GroupRoleID`; `UserRepository.CreateUser`/`UpdateUser` and
`GroupRepository.CreateGroup`/`UpdateGroup` honor them when present. The legacy-enum bridge has been
**removed**: `DeriveLegacyUserRole`/`DeriveLegacyGroupRole` are deleted and no `UserRole`/`GroupRole`
is derived or written anymore. The admin create endpoint's role-required validation
(`middleware.ValidateUserData`) now accepts **only** `appRoleId`. Public `SignUp` strips a
caller-supplied `AppRoleID` so a sign-up can never self-assign a role.

### Tests

`permissions/matcher_test.go`, `permissions/registry_test.go`, `permissions/legacy_test.go`,
`services/permission_test.go`, `services/roles_test.go`, `repositories/roles_test.go`,
`repositories/seed_roles_test.go` (default-role seeding), the per-create assignment tests in
`repositories/users_test.go` / `repositories/groups_test.go`, and the handler authorization tests in
`handlers/generic_handler_test.go` (with shared helpers in `handlers/auth_test_helpers_test.go`).
`handlers/group_delete_authorization_test.go` covers the `DeleteGroup` matrix (member with
`group.delete`, non-member with `app.groups.delete`, both denials, and the All-group 400).

## MCP Server & OAuth 2.1

Receipt Wrangler can expose a remote **Model Context Protocol** server so clients such as
Claude can read a user's data. It is **off by default** and Go-native (no separate service).

- **Configuration (System Settings, not env)**: the enable toggle (`mcpEnabled`) and the public
  URL (`mcpPublicUrl`) live on `models.SystemSettings` and are edited via the System Settings UI.
  `mcpPublicUrl` is the externally reachable origin (e.g. `https://receipts.example.com`) used to
  build the OAuth issuer/metadata/redirect URLs and the MCP token audience; it defaults to
  `http://localhost:8081` in dev. Both are read **live** (`services.IsMcpEnabled`,
  `services.GetMcpPublicUrl`/`GetMcpResourceUrl`), so toggling the server on/off or changing the
  URL takes effect without a restart. (There is no `MCP_ENABLED` / `MCP_PUBLIC_URL` env var.)
- **Live start/stop**: HTTP routes can only be mounted once at startup, so unlike the background
  workers (email polling / task server) the MCP routes are **always mounted** in
  `routers.BuildRootRouter` → `mountMcpRoutes` and gated at request time by `mcpEnabledMiddleware`,
  which 404s every MCP/OAuth path while `mcpEnabled` is off.
- **Endpoints** (mounted at the server root):
  - `/.well-known/oauth-protected-resource` + `/.well-known/oauth-authorization-server` —
    OAuth discovery (RFC 9728 / RFC 8414)
  - `/oauth/register` (Dynamic Client Registration, RFC 7591), `/oauth/authorize`
    (login form backed by `services.LoginUser`), `/oauth/token` (authorization_code +
    refresh_token grants, PKCE S256)
  - `/mcp` — Streamable HTTP MCP endpoint, guarded by bearer-token auth (401 +
    `WWW-Authenticate` advertising the protected-resource metadata)
- **Auth model**: the OAuth tokens are Receipt Wrangler HS512 JWTs, but **MCP-audience bound**.
  `services.GenerateMcpJWT` mints the access **and** refresh token with the audience set to the MCP
  resource URL (`GetMcpResourceUrl`, i.e. `mcpPublicUrl` + `/mcp`) instead of the normal
  `https://receiptWrangler.io` audience — the audience is *replaced, not appended*. The MCP
  endpoints verify that exact audience (`services.InitMcpTokenValidator`); the REST API keeps
  verifying the normal audience (`services.InitTokenValidator`). So an MCP token is rejected
  everywhere except `/mcp`, and an MCP refresh token can't be traded for a full-access token at
  `/api/token` (the refresh/rotation path also verifies the MCP audience). Because the audience is
  derived from a live setting, changing `mcpPublicUrl` intentionally invalidates existing connector
  tokens.
- **`mcp:read` scope is currently dead**: the bearer middleware requires it but every token carries
  it, so it gates nothing. Read-only is guaranteed *structurally* by only registering read tools
  (see notes on `readScope` in `internal/oauth/oauth.go` and `mcpReadScope` in
  `internal/mcp/server.go`). Adding any write/delete tool removes that guarantee and requires real
  per-tool scope enforcement.
- **Packages**: `internal/oauth/` (authorization server) and `internal/mcp/`
  (server + read-only tools). Tools call the service/repository layer in-process with the
  authenticated user's claims and enforce the same authorization as the REST handlers — not just
  group-scope but also the category/tag grants and paid-by visibility. The tools don't pass through
  `HandleRequest`, so for the two operations that have a REST twin the enforcement is **shared via
  `ReceiptService`** (the single source of truth, so the two ingress points can't drift):
  `get_receipt` and the REST `GetReceipt` handler both call
  `ReceiptService.GetReceiptForUser(userId, id)` (fetch → `group.receipts.read` →
  `ReceiptPaidByVisible` → `FilterReceiptCategoriesTagsForReceipt`, returning `ErrReceiptAccessDenied`
  on any miss/deny — mapped to a non-leaking MCP "receipt not found" / REST 403); `search_receipts`
  and the REST `Search` handler both call `ReceiptService.SearchReceiptsForUser(userId, query, limit)`
  (`app.receipts.search` → group scope → paid-by disjunction in SQL before the limit → `SearchResult`
  mapping; `ErrSearchForbidden` → MCP "unauthorized" / REST 403; blank query → empty). These two REST
  handlers therefore intentionally omit the declarative `HandleRequest` permission/`ReceiptId` gates —
  enforcement lives once, in the service. v1 tools are read-only: `search_receipts`, `get_receipt`,
  `list_groups`, `list_categories`, `list_tags`, `list_dashboards`. `list_categories`/`list_tags` have
  no REST twin and stay MCP-local in `tools.go`: they return the caller's grant-visible catalog (the
  full pool only for `app.categories.read`/`app.tags.read` holders, else the union of their group
  roles' grants — `visibleByGrants`).
- **Storage**: `models.OAuthClient` + `models.OAuthAuthorizationCode` (registered in
  `MakeMigrations`). Refresh tokens reuse the existing `models.RefreshToken` flow.
- **Production**: `docker/default.conf` proxies the new root paths to the backend; the `/mcp`
  location disables buffering and raises the read timeout for SSE streams.

## Session lifetime (configurable refresh-token expiry)

How long a signed-in user stays signed in is an admin setting rather than a compile-time constant.
Two System Settings drive it (`models.SystemSettings`, edited via the System Settings UI, validated
in `commands.UpsertSystemSettingsCommand.Validate`):

- `refreshTokenValidForHours` (int, default 24) — the REST/app refresh-token lifetime.
- `mcpRefreshTokenValidForHours` (int, default 24) — the same for MCP/OAuth connector tokens, kept
  **separate on purpose** so a long window chosen for human convenience does not silently extend
  tokens held by third-party clients.

Both are **whole hours**, bounded to **1-720 (30 days)**. There are two distinct "no value" cases,
and the difference matters:

- **Key omitted from the PUT body** (`nil` on the command) — the stored value is left alone. The
  command fields are `*int`, because `UpdateSystemSettings` writes every column with `Select("*")`:
  a plain `int` would persist as `0` and silently reset a configured lifetime whenever any client
  PUT a partial body. Same hazard and same fix as the pointer fields on
  `UpdateGroupReceiptSettingsCommand` (see the root `CLAUDE.md` → "Group Default Custom Fields").

  **An omitted lifetime is dropped from the UPDATE entirely**, via
  `OmittedLifetimeColumns()` feeding the existing `Omit(...)` on that statement — *not* by copying
  the stored value onto the row. The distinction matters under concurrency: `existingSettings` is
  read before the transaction, so two admins each setting one lifetime and omitting the other would
  both write a full row and clobber each other's explicit value. A column that is never written
  cannot be clobbered. This is deliberately **not** a `SELECT ... FOR UPDATE` row lock — SQLite is a
  supported engine and rejects that syntax, there is no locking precedent anywhere in the codebase,
  and a lock would only protect these two fields while every other column on this full-object upsert
  stays last-write-wins. `ApplyOmittedLifetimes` still runs, but only so the response body echoes the
  stored value instead of a misleading `0`.

  Guarded by `TestUpdateSystemSettingsPreservesOmittedRefreshTokenLifetimes` (a **raw JSON body**
  through the handler — a typed command literal cannot express an absent key),
  `TestUpdateSystemSettingsOmitsUnsentLifetimeColumns` (the generated SQL), and
  `TestUpdateSystemSettingsConcurrentLifetimeUpdatesBothSurvive` (two concurrent updates; fails on
  the first attempt without the column skip).
- **Explicit `0`** — means "unset, use the built-in default" (the `pdfDpi` convention). Valid, and
  resolved by the clamp on read.

**It is an inactivity timeout, not an absolute session cap.** Refresh tokens rotate on every use and
both clients proactively refresh on a 15-minute timer, so an actively-used session re-mints itself
indefinitely; the window only bites once a client has been away longer than it. There is no
absolute cap — adding one would need a session-start timestamp threaded through every rotation.
A useful corollary: **lowering** the setting takes effect for active sessions at their next refresh
(~15 min), so only genuinely idle sessions coast on their original expiry. Nothing is bulk-revoked.

**The access token is deliberately NOT configurable** and stays at 20 minutes
(`utils.GetAccessTokenExpiryDate`). It is stateless and unrevocable, and both clients size their
15-minute refresh timer against that fixed window — making it configurable would break them and
widen the blast radius of a leaked token. This is also why `oauth.accessTokenExpiresIn` (the OAuth
`expires_in`, RFC 6749 — which describes the access token only) still mirrors 20 minutes correctly.

**Where the value is resolved.** `internal/utils` imports nothing from `internal/`, so it cannot
read System Settings without an import cycle; `utils.GetRefreshTokenExpiryDate(lifetime)` therefore
takes the duration as a parameter. `services.GetRefreshTokenLifetime()` /
`GetMcpRefreshTokenLifetime()` do the live read and hand it to `generateTokenPair`, falling back to
24h on error/unset/out-of-range via `clampRefreshTokenLifetime` — the same clamp-with-fallback shape
as `repositories.pdfRasterizationDpi`. **The clamp, not the validator, is the real safety net**: it
stops a bad stored value from ever producing a token regardless of how it got there. The bounds
themselves live in `commands` (`Min`/`MaxRefreshTokenValidForHours`) and are imported by `services`,
so validation and the clamp cannot drift. `BuildTokenCookies` resolves the app lifetime itself for
the refresh cookie's `Expires` — both its call sites (login, token refresh) are REST-only, never
MCP.

The `@every 24h` refresh-token cleanup cron is unaffected: it deletes on `is_used = true` as well as
expiry, and rotation marks the previous token used immediately, so the table stays small at any TTL.

**Refresh-token consumption is atomic.** `middleware.RevokeRefreshToken` marks the token used with a
single `UPDATE ... WHERE token = ? AND is_used = false` and checks `RowsAffected`, mirroring
`oauth.rotateRefreshToken`. It used to read then update separately, which let two concurrent
refreshes both observe `is_used = false` and rotate — defeating replay detection, and logging the
loser out of every tab once the rotation it lost landed. Unknown and already-used tokens are
deliberately indistinguishable (both already returned the same response). Guarded by
`TestRevokeRefreshToken_ConcurrentRefreshesOnlyOneSucceeds`.

Tests: `services/refresh_token_lifetime_test.go` (clamp table, per-field isolation, expiry stamped
on both the JWT claim and the persisted row, MCP using its own setting, access token unchanged),
`commands/upsert_system_settings_command_test.go` (bounds + the `0` escape hatch),
`middleware/refresh_token_test.go` (the concurrency guard), `utils/auth_test.go` (the helper honors
whatever lifetime it is handed), `handlers/system_settings_handler_test.go` (the omitted-key and
explicit-value endpoint round trips), `repositories/system_settings_test.go` (the omitted-column SQL
shape and the concurrent-update guard), and `desktop/e2e/session-lifetime.spec.ts` (the only
end-to-end proof that the setting reaches the `Set-Cookie` header).

## Login QR & mobile deep link

The desktop login page can show a self-contained QR that sets up the mobile app. Two System Settings
drive it (`models.SystemSettings`, edited via the System Settings UI, validated in
`commands.UpsertSystemSettingsCommand.Validate`):

- `showLoginQr` (bool, default false) — opt-in toggle.
- `mobileServerUrl` (string) — the server/API URL mobile clients connect to; required (absolute
  http/https, validated by `isValidAbsoluteUrl` — the former `isValidMcpPublicUrl`, generalized and
  now shared by both settings) when `showLoginQr` is on. `isValidAbsoluteUrl` also rejects **embedded
  credentials** (`https://user:token@host`): both settings it guards are published to unauthenticated
  clients — this one inside the login QR on the public `/featureConfig`, `mcpPublicUrl` in the OAuth
  discovery metadata — so userinfo would be handed out verbatim. **http stays valid** on purpose:
  LAN / bare-IP self-hosting is a supported deployment (the mobile Connect screen accepts http too).

`SystemSettingsService.GetFeatureConfig()` derives a single **`FeatureConfig.LoginQrUrl`** from them
via `BuildLoginQrUrl` (`services/system_settings.go`): the App Link / Universal Link
`https://receiptwrangler.io/app/setup#url=<percent-encoded mobileServerUrl>` when the toggle is on and
a URL is set, else `""`. The server URL rides in the **fragment** so it never reaches
receiptwrangler.io's logs in the app-not-installed web fallback. `loginQrUrl` is the **only** login-QR
value exposed on the public, unauthenticated `GET /featureConfig` payload — the raw setting and toggle
stay behind auth. The desktop renders the QR locally from `loginQrUrl` (no new endpoint/handler).

The `receiptwrangler.io/app/setup` host/path is a fixed, project-owned constant that MUST stay in sync
with the mobile App Link config and the `.well-known/assetlinks.json` / `apple-app-site-association`
files hosted on **receiptwrangler.io** (served from that domain's own nginx/docs infra — not this
repo; the same layer also serves a `/app/setup` platform redirect for the app-not-installed fallback).
See `mobile/CLAUDE.md` → "App Links / Universal Links — server-URL pre-fill (login)". Tests:
`commands/upsert_system_settings_command_test.go` (validation), `services/system_settings_test.go`
(`BuildLoginQrUrl` compose/encoding + `GetFeatureConfig` mapping).

## Receipt statuses

`models.ReceiptStatus` (`internal/models/receipt_status.go`) is a plain Go `string` type — there is
no DB enum or CHECK constraint anywhere, so **adding a value needs no migration**, only these four
edits, which nothing forces to agree with each other:

1. the `const` block,
2. the `Value()` guard chain (a missing value makes every write fail as a generic 500),
3. `ReceiptStatuses()` — the single membership list, which drives the only two server-side
   validations of a caller-supplied status: `handlers/receipts.go` `BulkReceiptStatusUpdate` and
   `commands/update_group_receipt_settings_command.go` `isValidReceiptStatus`,
4. the `ReceiptStatus` enum in `swagger.yml` — then regenerate **both** clients, `mobile/api/`
   included, in the same change (see "API Client Generation").

`models/receipt_status_test.go` → `TestReceiptStatuses` asserts the exact list **and order**, so it
fails by design on any addition. Update it rather than weakening it.

**`AfterReceiptUpdated` (`repositories/receipts.go`) is the only transition logic** — there is no
state machine, any status may follow any other, and all three write paths reach it (`UpdateReceipt`,
`CreateReceipt`, and the bulk handler). Two statuses are **terminal** and cascade the receipt's items
to `ITEM_RESOLVED`:

- **`RESOLVED`** also stamps `resolved_date`.
- **`DECLINED`** deliberately does **not** — a decline is not a resolution, and `resolved_date` is a
  reporting dimension (`receiptsource.KeyResolvedDate` plus its derived day/month/year fields) and a
  sortable/filterable column. It falls into the existing `Status != RESOLVED` branch, which clears
  `resolved_date`.

The cascade is what removes a terminal receipt from settlement: `handlers/users.go`
`GetAmountOwedForUser` filters on **`ITEM_OPEN`**, i.e. the *item* status, never the receipt's. A new
status that should not count as owed therefore has to be added to that cascade, not just the enum.
`OPEN` and `NEEDS_ATTENTION` have no cascade and leave items alone.

**Two validation gaps to know about** (both pre-existing): `UpsertReceiptCommand.Validate` and
`UpdateGroupSettingsCommand.Validate` check a status is **non-empty**, never that it is a member — an
arbitrary string reaches the DB layer and is caught only by `Value()`, surfacing as a 500. And
`ReceiptStatus.Scan` never returns an error, so `QuickScanCommand.LoadDataFromRequest` cannot reject
a bogus status either.

## Quick Scan Field Configuration

Group admins configure the quick-scan workflow per group on `GroupReceiptSettings` (gated by the
group `GroupUpdate` permission, edited via `PUT /group/{groupId}/groupReceiptSettings`). For each of
**paid-by, status, categories, tags, comment** there is an `*Enabled` (show) and `*Required` toggle. When
paid-by/status is **not** both shown and required, a default backfills it, so **a receipt always has
a real paid-by and status — neither field is ever null/empty**. This is why the `Receipt` model,
`UpsertReceiptCommand.Validate`, and the email workflow are all unchanged by this feature.

- **Model** (`models/group_receipt_settings.go`): the `QuickScan*` columns. Paid-by default is a
  `QuickScanDefaultPaidByType` (`UPLOADER` | `USER`, `models/quick_scan_default_paid_by_type.go`) plus
  a nullable `QuickScanDefaultPaidById`; status default is `QuickScanDefaultStatus`. Backwards-compatible
  GORM defaults: paid-by/status ship **enabled + required** (no default needed), categories/tags/comment
  **hidden** — so existing groups are unchanged until an admin opts in. AutoMigrate adds the columns; no
  data migration.
  - **`UpdateGroupReceiptSettings` assigns every field one at a time** — a new setting that isn't added
    to that block silently never persists (no compile error, no test failure). `repositories/group_receipt_settings_test.go`
    round-trips the flags specifically to catch that.
- **Comment (`QuickScanCommentEnabled` / `QuickScanCommentRequired`)** captures a receipt comment at
  scan time. Two things gate it beyond its own toggle, both resolved in **one** place per runtime so
  they can't drift — `GroupReceiptSettings.IsQuickScanCommentShown()` / `IsQuickScanCommentRequired()`
  on the server, `resolveQuickScanFieldConfig` on mobile, `showComment(i)` on desktop:
  - **`HideComments`** (the group-wide "hide comments" setting) overrides the toggle. It is derived,
    never written: the stored toggles are untouched, so turning `HideComments` off restores them. The
    desktop config UI greys the two checkboxes out for the same reason — which is why its `submit()`
    must use `getRawValue()` (a disabled control is omitted from `form.value` and would unmarshal as
    `false`, wiping the admin's configuration).
  - **`group.comments.create`** acts as an extra AND on "enabled". A caller without it never sees the
    field, is **not** subject to the required check (or they could never quick scan at all), and any
    comment they submit anyway is **silently dropped — deliberately not a 403**, unlike
    `enforceQuickScanGrantSelection`'s treatment of out-of-grant categories/tags: a comment is
    incidental to the receipt, so refusing the whole fire-and-forget scan over it is the worse
    outcome. The permission is resolved only when the field is enabled (cached per group id), so the
    default-off config costs no extra queries.
  - **Required is enforced strictly**, like categories/tags: a client that omits `comments` while the
    group requires one gets a 400. Backwards compatibility comes from the columns defaulting to
    `false/false`, so an already-released mobile build is unaffected until an admin opts in — at which
    point those builds must update. The desktop config UI says so inline.
- **Config validation** (`commands/update_group_receipt_settings_command.go` `Validate`, now wired into
  the handler): a default is **required** for paid-by/status unless `(enabled && required)`; categories/
  tags/comment never need a default (empty is fine).
- **Ingress** (`handlers/receipts.go` `QuickScan` → `ReceiptService.ResolveQuickScanFields`,
  `services/quick_scan_fields.go`): loads each file's group
  settings, **enforces required fields synchronously (400 before enqueue)** since quick scan is
  fire-and-forget, and resolves paid-by/status defaults (`UPLOADER` ⇒ the caller's user id). Categories/
  tags ride the multipart command as **per-file comma-joined id strings** (`QuickScanCommand.CategoryIds`/
  `TagIds`; an empty paid-by string parses to `0` = unset). The **comment** rides as a parallel
  `comments` string array, one entry per file — free text, so it is **not** comma-split (a comment may
  contain commas and newlines); entries are trimmed on parse so a whitespace-only comment counts as
  empty. It is length-checked against `models.MaxCommentLength` (500, matching the `Comment` column)
  here rather than at the database: the receipt is created in a background task, so an over-length
  comment would otherwise fail the insert where the user only ever sees a "queued" toast. The check
  counts **runes, not bytes** (`utf8.RuneCountInString`) — the column is `varchar(500)`, which MySQL
  and Postgres both measure in *characters*, so a `len()` check would be stricter than the column it
  mirrors and would reject an accented or non-Latin comment well inside its real capacity. The limit
  is published on the contract as `maxLength: 500` on `QuickScanCommand.comments.items` (OpenAPI
  measures `maxLength` in characters too), and both clients cap at 500 client-side — desktop via
  `Validators.maxLength`, mobile via `FormBuilderTextField.maxLength`.
  `ReceiptService.QuickScan` **appends** it to `receiptCommand.Comments` (with `UserId` set — a nil one
  fails `UpsertCommentCommand.Validate` and would take the whole receipt down) rather than replacing:
  the default prompt doesn't emit comments, but the AI response is unmarshalled straight into an
  `UpsertReceiptCommand`, so a group running a custom prompt can produce them.
- **`ReceiptService.QuickScan` takes a `QuickScanParams` struct**, not positional arguments — the
  positional form ended in several interchangeable strings (temp path / file name / task id / comment),
  which is silently transposable. The async `ReceiptService.QuickScan`
  **resolves** the union of the user's category/tag picks and the AI-assigned ones
  (`resolveQuickScanCategories`/`resolveQuickScanTags`): the ids (AI first, then user, deduped) are looked
  up via `CategoryRepository.GetByIds`/`TagsRepository.GetByIds` so each carries its real **name**. This is
  load-bearing — the default prompt tells the model to return categories/tags **by id only** (`{ "id": N }`,
  no name), which `UpsertReceiptCommand.Validate` (name required) would otherwise reject, silently dropping
  the whole receipt. Ids that don't resolve (hallucinated / deleted) are **dropped**, as are ids the
  triggering user isn't allowed to see (`resolveAllowedCategoryIds`/`resolveAllowedTagIds` reuse
  `userBypassesGrants` + `GetGroupCategoryIdsForUser`/`GetGroupTagIdsForUser` — the same bypass+set logic as
  the write-side check, dropping instead of rejecting). Validation deliberately stays strict (name required):
  an id-only category reaching `UpdateReceipt` — which persists associations with `FullSaveAssociations: true`
  — would upsert the category row and **blank its `not null; uniqueIndex` name**, so resolving names (not
  relaxing validation) is the fix.
- **Failed-task recording on pre-transaction errors:** `CreateReceiptUploadedSystemTask` only runs
  **inside** the create-receipt transaction, so an error that aborts `QuickScan` *before* it (category/tag
  resolution or receipt validation) used to leave the AI-processing tasks marked SUCCEEDED and **no record**
  of why no receipt appeared — the failure only hit the log via `HandleError`. `QuickScan` now routes those
  early returns through a `recordEarlyQuickScanFailure` closure that writes a FAILED `RECEIPT_UPLOADED` task
  chained to the QUICK_SCAN parent; because `SystemTaskRepository.CreateSystemTask` flips a SUCCEEDED parent
  to FAILED when a FAILED child references it, the QUICK_SCAN activity now surfaces the failure. `RanByUserId`
  is left nil so the child stays hidden and only the flipped parent shows (mirroring the successful child).
- **Grant enforcement (write-side):** because quick scan creates receipts through the service layer
  (bypassing the receipt-upsert path's `enforceReceiptGrantSelection`), the handler **synchronously
  grant-validates the user's per-file category/tag picks** before enqueue via
  `enforceQuickScanGrantSelection` (`handlers/receipt_grant_enforcement.go`), which reuses
  `PermissionService.ValidateCategoryTagSelection` — an out-of-grant pick returns **403** (unrestricted
  and admin-bypass callers are a no-op). This closes a write bypass: the UI picker already limits picks
  to the granted catalog, but a crafted request could otherwise attach a category/tag the caller can't
  see.
- **End-to-end ingest tests** (`services/quick_scan_ingest_test.go`): the AI is mocked at the HTTP
  transport layer (a new mutable-body httptest server `newMutableOllamaServer` + `ollamaReceiptResponse`,
  reusing `seedReceiptImagePipeline` from `receipt_image_test.go`), a controlled
  receipt JSON is driven through the real `ReceiptService.QuickScan`, and the persisted receipt is read
  back via `GetFullyLoadedReceiptById` to prove **every** field survives — name/amount/date/paid-by/
  status, receipt- and item-level categories/tags, items (amount / status / `chargedToUserId` /
  `isTaxed` / linked items), comments, and all five custom-field value types. This pins the
  AI→`CreateReceipt` contract (which passes items/comments/customFields through untouched, so a future
  prompt emitting them just works), plus the paid-by/status backfill, the category/tag resolution, refunds
  (negative amounts), and the all-or-nothing failure path (bad item status / invalid JSON persist
  nothing). Note: `UpsertReceiptCommand.Validate` deliberately does **not** validate `CustomFields`
  (unlike categories/tags/items/comments), so a custom-field value is persisted unchecked — the test
  is the only guard against a regression that silently drops or mis-columns it. The resolve/failure
  behavior above is pinned by `TestQuickScan_ResolvesIdOnlyAiCategory` (id-only AI category gets its name
  filled), `_DropsUnresolvableAiCategory` (hallucinated id dropped), `_DropsOutOfGrantAiCategory` (a
  grant-restricted member's AI category dropped — seeds a restricted group role via the `postSeed` hook on
  `runQuickScanWithSetup`), and `_ValidationFailureRecordsFailedSystemTask` + the extended
  `_MissingItemStatusFailsAndPersistsNothing` (a pre-transaction failure flips the QUICK_SCAN task to
  FAILED and records a FAILED RECEIPT_UPLOADED task, asserted via `quickScanSystemTaskByType`), plus the
  unit-level `TestResolveQuickScan{Categories,Tags}`. **Known gap (follow-up):** resolution is
  **receipt-level only** — item-level categories/tags (`receiptCommand.Items[].Categories/Tags`) aren't
  produced by the default prompt and aren't resolved here, so a future prompt emitting id-only item-level
  categories would hit the same validation failure.

## Group Default Custom Fields

Group admins declare, in **Group Receipt Settings**, the set of custom fields that should always be
present on that group's receipts (a tax field, a PO number, a cost centre). The receipt form pre-adds
them on create and re-applies the target group's set when the user switches groups — each group is
effectively its own receipt template. A second, opt-in toggle
(`applyDefaultCustomFieldsOnIngest`) extends the same set to receipts the **server** creates (quick
scan and email integration).

A default is a **starting point, never a requirement**: the user may remove any of them and nothing
validates their presence. There is deliberately **no notion of a required custom field** — custom
field validation does not exist in this codebase and would belong on the custom field itself, not on
this screen. `UpdateGroupReceiptSettingsCommand.Validate` therefore checks nothing here.

- **Join model** (`models/group_receipt_settings_custom_field.go`): `GroupReceiptSettingsCustomField
  {GroupId, CustomFieldId}`, composite primary key, `CustomFieldId` indexed. Two deliberate choices,
  both spelled out in the model's doc comment:
  - **Keyed on `GroupId`, NOT the `GroupReceiptSettings` row id.** `GroupRepository.GetGroupById`
    lazily creates a missing settings row and **discards** the created record, so
    `group.GroupReceiptSettings.ID` is still `0` on the very call that created it — keying on it
    would silently write rows against id `0`. `GroupId` is `not null;unique` on the settings model
    and is what every caller already holds (quick scan, the email handler, group delete, the PUT).
  - **An explicit join model, not a GORM `many2many []CustomField`.** `UpdateGroupReceiptSettings`
    does `Preload(clause.Associations)` then `db.Select("*").Model(...).Updates(...)`, which
    full-save-associates every loaded association — that would upsert the joined `CustomField` rows
    and can blank a `not null` `CustomField.Name` (the same failure documented above for AI-assigned
    categories). Inserts therefore use **`Omit("CustomField").Create(&rows)`**, matching
    `replaceReportTemplateGroups`. Pinned by
    `TestUpdateGroupReceiptSettingsDoesNotBlankCustomFieldName`.
- **Model** (`models/group_receipt_settings.go`): `ApplyDefaultCustomFieldsOnIngest bool`
  (`not null;default:false`, so upgrading installs are unchanged until an admin opts in) and the
  transient `DefaultCustomFieldIds []uint` (`gorm:"-"`).
- **Hydration is an explicit batched loader, NOT a GORM `AfterFind` hook.**
  `GroupReceiptSettingsRepository.LoadDefaultCustomFieldIds([]*GroupReceiptSettings)` runs **one**
  `WHERE group_id IN (?)` query (`ORDER BY custom_field_id` for a deterministic order) and mutates
  the rows in place; `LoadDefaultCustomFieldIdsForGroups([]Group)` is the convenience wrapper. This
  mirrors `GroupMemberRepository.LoadMemberGrants`, which fills `GroupMember.CategoryGrants`/
  `TagGrants` the same way. A hook was considered and rejected: it would be the only `AfterFind` in
  the codebase, adds an N+1 inside `Preload(clause.Associations)`, needs a `NewDB` session to avoid
  inheriting the preload's statement, **and would not even be correct** —
  `UpdateGroupReceiptSettings` returns the in-memory struct it mutated rather than re-reading, so
  the PUT response would carry the *old* ids, which the desktop writes straight into its group state.
  Called at every boundary that serializes a group's receipt settings, mirroring the
  `LoadMemberGrantsFor*` call sites: `services/auth.go` (AppData, beside
  `LoadMemberGrantsForGroups`), `handlers/groups.go` `GetGroupsForUser`, `GetPagedGroups` and
  `CreateGroup`, `repositories/groups.go` `GetGroupById` (which `UpdateGroup` also returns through),
  `GetGroupReceiptSettingsByGroupId`, and the `UpdateGroupReceiptSettings` return value.
  - **The two easy ones to miss are `GetPagedGroups` and `CreateGroup`**, and both shipped without
    the call at first. `GetPagedGroups` looks covered because it uses `Preload(clause.Associations)`
    — which loads the settings *row* but cannot touch a `gorm:"-"` field — and its load must run
    **before** the `anyData[i] = groups[i]` copy, which takes each group by value. `CreateGroup`
    returns a `Find` that preloads only `GroupMembers`, so its settings object is zero-valued
    **including `GroupId`**; set that first or the loader keys its lookup on group 0 and the empty
    result is right by accident. Both are pinned by response-shape tests (below).
- **An empty set must serialize as `[]`, never `null`.** `defaultCustomFieldIdsOrEmpty` normalizes it
  inside the loader (mirroring `grantIdsOrEmpty`) so every read path is covered at once. The
  generated Dart deserializer has **no null guard**, so a `null` would fail the **whole** AppData
  payload on already-released Android builds — the exact class of the two documented login outages.
  Guarded by a `json.Marshal` test in `repositories/group_receipt_settings_test.go` and a
  response-shape test in `handlers/group_default_custom_fields_test.go`.
- **Command semantics: BOTH new fields are POINTERS, and `nil` means LEAVE UNCHANGED.**
  `UpdateGroupReceiptSettingsCommand.DefaultCustomFieldIds *[]uint` and
  `ApplyDefaultCustomFieldsOnIngest *bool`. The desktop hides this whole section from an admin
  without `app.custom-fields.read`, so its `getRawValue()` omits both keys; a non-pointer `bool`
  would unmarshal as `false` and the repository's unconditional field-by-field assignment would
  **wipe the stored toggle** — byte for byte the `hideComments` bug documented in
  `desktop/CLAUDE.md`. An explicit `[]` clears the set.
  - **Same field-by-field assignment hazard as the quick-scan config:**
    `UpdateGroupReceiptSettings` assigns every setting one at a time, so a new setting that isn't
    added to that block silently never persists (no compile error, no test failure). The
    round-trip + `nil`-leaves-unchanged tests exist specifically to catch that.
- **Transaction**: `UpdateGroupReceiptSettings` wraps the scalar update **and** the join replace in
  one `db.Transaction(...)` (precedent: `CreateReportTemplate`), so a rejected save changes nothing.
  Callers pass a `nil` TX today, and a nested GORM transaction degrades to a savepoint, so this is
  safe either way. `replaceGroupDefaultCustomFields` is delete-by-`group_id`-then-insert, deduping
  the submitted ids so a repeated id can't violate the composite primary key.
- **Handler guards** (`handlers/groups.go` `UpdateGroupReceiptSettings`) — the endpoint is gated on
  the group's `GroupUpdate` only, so two extra checks run when `DefaultCustomFieldIds != nil`
  (a `nil` pointer means the client didn't touch the selection and stays allowed):
  - **403** without the app-level `app.custom-fields.read`. Without this a group admin could attach
    fields whose catalog they cannot read, mirroring `enforceReceiptCustomFieldSelection`.
  - **400** with error key `defaultCustomFieldIds` when any submitted id doesn't exist (via the new
    `CustomFieldRepository.GetCustomFieldsByIds`). An unknown id would otherwise persist as a row
    nothing can ever resolve, and the receipt form would silently skip it forever.
- **Delete cascades**, both explicit rather than FK-driven, matching every other cascade nearby:
  - `CustomFieldRepository.DeleteCustomField` gains a fourth delete inside its existing transaction,
    before the `CustomField` row: the field is removed from **every** group's default set.
  - `GroupService.DeleteGroup` deletes the join rows beside the member-grant cleanup. The receipt
    settings delete uses `Select(clause.Associations)`, which cannot reach them — `gorm:"-"` means
    the join is not an association of the settings model at all.
- **Server-side ingest** (`services/default_custom_fields.go`) — one helper shared by both ingest
  paths so they cannot drift. `ApplyDefaultCustomFields(settings, *command)` is pure: it no-ops
  unless `ApplyDefaultCustomFieldsOnIngest`, then appends an **empty**
  `UpsertCustomFieldValueCommand{CustomFieldId: id}` for each default id the command doesn't already
  carry. **De-duplication matters** — the AI response is unmarshalled straight into an
  `UpsertReceiptCommand`, so a group running a custom prompt can already have produced a value for a
  defaulted field. `ApplyGroupDefaultCustomFields(tx, groupId, *command)` is the thin loader wrapper;
  it treats a **missing settings row as "not opted in" rather than an error**, because the row is
  created lazily and a group never opened in the settings UI legitimately has none. Call sites, both
  immediately before `Validate`:
  - `ReceiptService.QuickScan` — after the `resolveQuickScanCategories`/`resolveQuickScanTags`/
    comment-append block.
  - `wranglerasynq/email_process_handler.go` — after `command.CreatedByString = "Email Integration"`.
  `services/import.go` deliberately does **not** apply defaults — it replays existing receipts.
  `rerun_task_payload.go` routes through `ReceiptService.QuickScan`, so it inherits the behaviour.
- **Users without `app.custom-fields.read` are gated out client-side**: the desktop and mobile
  receipt forms skip ids absent from their loaded catalog, which is empty without the permission, so
  their save would never have passed `enforceReceiptCustomFieldSelection` anyway.
- **Tests**: `repositories/group_receipt_settings_test.go` (round-trip, replace, `[]` clears, `nil`
  leaves both unchanged, `json.Marshal` gives `[]`, the not-blanked custom field name, the batched
  loader across two groups, `GetGroupById` hydration), `repositories/custom_fields_test.go` (a
  deleted field leaves every group's set, two groups seeded), `services/default_custom_fields_test.go`
  (toggle-off no-op, appends only missing ids, no duplicate, missing settings row, group delete
  leaves zero join rows), `services/quick_scan_ingest_test.go`
  (`TestQuickScan_{Applies,Skips}GroupDefaultCustomFields*`), and
  `handlers/group_default_custom_fields_test.go` (400 unknown id writes nothing, 403 without the
  permission, a save omitting both keys leaves the stored config untouched, `[]` on the wire).

## Reporting Engine (`internal/reporting`)

A **pure** report engine: `(ReportSpec + FieldCatalog + []Row + MetaInput) → ReportModel`. It
fetches nothing, renders nothing, reads no clock, and consults no global. Renderers (CSV/XLSX/PDF), a
dashboard widget, template persistence and HTTP delivery all *call* it; none of them are part of it.

**`internal/reporting/README.md`** is the narrative guide for engineers: the pipeline, the type
vocabulary, and a worked example carried from input rows through the report tree to the rendered
table. The section below is the rules digest; the README is where the diagrams are.

**The purity rule is enforceable and must stay true:**

```bash
go list -deps receipt-wrangler/api/internal/reporting | grep -E 'gorm|repositories|internal/models'
# must print nothing
```

`internal/reporting/receiptsource/` is the **only** package that imports `models`. It maps
`[]models.Receipt` + `[]models.CustomField` into `[]reporting.Row`. Item-grain reporting or a
non-receipt widget adds a *sibling* of it; the engine core never changes.

```go
source, err := receiptsource.New(customFields)          // needs CustomField.Options loaded
model, err := reporting.Run(spec, source.Catalog(), source.Rows(receipts), meta)
```

**Caller must preload** `PaidByUser`, `Group`, `Categories`, `Tags`, `CustomFields`. An unloaded
association resolves to no value, which surfaces as a `(None)` bucket — not an error.

**Date period grouping — derived string fields.** `date`, `resolved_date` and `created_at` are
`TypeDate`, so grouping by one buckets on the exact instant (one group per receipt). To group by
calendar period, `receiptsource` also offers derived **string** fields `<base>_day` / `_month` /
`_year` (e.g. `date_month`, `created_at_year`, `resolved_date_day`) — zero-padded ISO in **UTC**, so
they sort chronologically as plain text. A report groups by one of these instead of the raw date field;
a nil `resolved_date` emits none of its period fields (→ `(None)`). A **date custom field** gets the
same treatment: `CustomFieldPeriodKeys(id)` yields `custom_<id>_day` / `_month` / `_year`, labelled off
the field's own name (`"Due Date (Month)"`) through the shared `dateFieldRefs` / `setDateParts` helpers.
They cannot collide with another field's key — `CustomFieldKey` is always `custom_` plus digits alone —
and only a `models.DATE` field derives them. When duplicate values force the lowest-id-wins tie-break,
the period parts are rewritten with the winner, so they always describe the value that survived.

**Every field can cut; only numeric fields can be measured.** `Role` (`field.go`) bounds *measuring*
only — `compileGroupBy` / `compileDetail` accept any field in the catalog, so a **currency custom field
is both groupable and summable** (and a plain label column always accepted any field). Grouping by a
number is well defined: bucket keys are canonical for decimals and `compareValues` orders them, so
`15.60` and `15.6` share one bucket. `ErrGroupByNotDimension` / `ErrDetailByNotDimension` are **gone** —
a spec that used to be rejected for grouping by a measure now compiles. Tests that needed an
*invalid* spec use an unknown field key (`ErrUnknownField`) instead; don't reintroduce "group by
amount" as the invalid-spec fixture.

**Typed dimension and label rendering.** The engine emits raw typed values, so a renderer needs the
field's type to present them: without it a bool prints `true`, a date an RFC 3339 instant, and money a
bare decimal. `render.Dimension` therefore carries a `DataType` (populated by `buildDimensions` from the
catalog), and `render/value.go`'s shared `formatLabelValue` is used by **every** label-ish cell —
`formatDimension` (the CSV walk *and* `walk.go`, so HTML/PDF and XLSX too) and `joinLabel` (XLSX) —
mapping bool → `Yes`/`No`, date → `2006-01-02` **UTC** (matching the derived `_day` field), currency →
`Meta.Currency`'s format (bare 2dp when unset). A declared type that disagrees with a value's payload
falls through to the value's own rendering rather than substituting a zero. A label column is still a
**string** cell in XLSX even when it holds money — it names a bucket rather than measuring one, and a
multi-value label has no numeric reading; aggregate columns are what carry native numbers. Note this
changed existing reports that group by the built-in `date` / `resolved_date` / `created_at`: their
bucket text went from RFC 3339 to a plain calendar day.

**Grouping-level column headings are overridable.** Because a grouping level renders as a *leading
column*, the request may name that column itself: `ReportRequestCommand.GroupByLabels`
(`map[string]string`, `omitempty`) overrides `render.Dimension.Label` per grouping field key, resolved
in `buildDimensions` (`services/report_service.go`). It is **keyed by field key rather than
index-aligned** with `GroupBy` — grouping keys are unique (`reporting.ErrDuplicateGroupBy`), so
reordering the levels can't desynchronize it. A key that is absent, blank after trimming, or names no
grouping level falls back to the catalog label; a stale entry is ignored rather than failing an
otherwise valid report, so there is **no validation** for it (matching the unvalidated
`ReportColumn.Label`). Only the heading changes — `DataType` is untouched, so buckets still format by
type. It is purely additive and optional, so stored templates deserialize unchanged and
`CurrentReportConfigurationVersion` stays `1`. The user-supplied text is already covered by the two
escaping paths every column label goes through: `html/template` for HTML/PDF and `SanitizeCSVField` in
`dimensionHeading` for CSV. Nothing in `internal/reporting/` is involved — this is a renderer concern.
The desktop drives it from the pencil on each Grouping row (see `desktop/CLAUDE.md` → "Reports").

**`services.ReportDataService`** (`internal/services/report_data.go`) is the first DB-backed caller of
the engine — it follows the `pie_chart.go` pattern. `Rows(userId, groupId, filter)` fetches the group's
receipts unpaged and applies the reporting access controls **in order**: narrow the request filter to
the caller's grants (`IntersectReceiptFilterWithGrants`), hide whole receipts in the query by paid-by
(`PaidByListResolver`), then substitute the categories/tags the caller can't see. It returns the
engine's `(FieldCatalog, []Row)` — it does **not** build a `ReportSpec` or call `Run`; a caller does.

**Renderers** (`internal/reporting/render`) are pure downstream consumers of a `ReportModel` — they
fetch and compute nothing, and never reach back into the engine, so a new format (or a future layout)
touches no upstream code. `render.CSV(model, groupBy)` is the first: a flat, **data-only** CSV — the
group-by dimensions become leading columns (`groupBy []render.Dimension` supplies their order and header
labels, since the model carries dimension keys but not labels), each detail leaf is one row, and a
leading `Row Type` column marks each row `Detail` / `Subtotal` / `Grand Total`. It renders only what the
model carries (subtotal/grand-total rows appear only when the spec toggled them on), draws no document
chrome, and is unit-tested in isolation against models built via `reporting.Run` (there is no orchestrator
or handler yet). Per `docs/engine-design.md` §5, CSV is deliberately the minimal renderer; the
grouped/visual "looks like the on-screen report" layout belongs to the XLSX/PDF renderers, each a separate
consumer of the same tree. Currency renders per the app's custom currency configuration — symbol, symbol
position, thousands/decimal separators, and hide-decimal-places — when the caller supplies `Meta.Currency`
(a bare 2dp otherwise); other numbers at full precision; `(None)` buckets use `Meta.NoneLabel`.

`render.XLSX(model, groupBy)` (via `github.com/xuri/excelize/v2`) is the **faithful, grouped** counterpart:
the group-by dimensions are leading columns with each value shown **once per group** (blanked on repeats),
and a subtotal/grand-total row carries a `Total`/`Grand Total` marker in the column at the group's depth
(the "staircase"). Numbers are written as **native, typed cells** with a number format — for currency an
Excel format code built from `Meta.Currency` (symbol, position, and decimal-places; the group/decimal
glyphs follow the opener's locale, an Excel constraint), overridable per column via
`ColumnDescriptor.Format`, defaulting to `#,##0.00` when neither is set — so the workbook stays analyzable,
and header/subtotal/grand-total rows are bold; sheet name `Report`. It writes the engine-computed values
**statically** — live `=SUM`/expression formulas (the reason `ColumnDescriptor.Expr` is an exported AST)
are a later slice, and document chrome/slots (logo) await the template work. It shares the `Dimension`
type and the `groupBy`-depth guard (`validateGroupByDepth`) with the CSV renderer, and the faithful
**walk** with the HTML renderer (`faithfulWalk` in `render/walk.go`, driving a per-format `faithfulSink`);
CSV keeps its own flat walk. Tested by round-tripping the bytes back through `excelize.OpenReader`.

`render.HTML(model, groupBy, chrome)` is the **PDF format's HTML stage** (via `html/template`): a
self-contained document — a `Meta.Title` heading, an optional authored intro, a preamble of the resolved
`Meta.Params`, the same faithful table as XLSX (through the shared `faithfulWalk`), and a footer, each
omitted when there is nothing for it. The `render.DocumentChrome{Intro, Footer}` argument is authored
presentation copy layered on **at render time** — kept out of the pure model (which stays
presentation-free); a zero value is byte-identical to the data-only rendering, and an authored footer
replaces the automatic `Meta.GeneratedAt` note. All CSS is inline and it references no external resources
or scripts, so it renders through the headless-Chromium HTML-to-PDF pipeline (`services/html_to_pdf.go`,
which blocks network loads and disables JS by default). It returns **HTML bytes** — the reporting package
is pure, so the chromedp conversion to PDF stays in the services layer (`ReportService`, below). Unit-tested
against the same faithful golden grids as XLSX (parsed with `golang.org/x/net/html`) plus document-chrome
and HTML-escaping cases.

**`services.ReportService` + `POST /api/report/generate`** are the orchestrator and endpoint that join
the pieces. `ReportService.Generate(userId, command)` resolves the request's period into a date filter,
loads rows across every group in the request under the one (global) catalog, builds a `ReportSpec`, runs
the pure engine, resolves the document's `{{period}}`/`{{group.name}}`/`{{generatedAt}}`/`{{currentUser.name}}`
variables (the rendered heading is the authored `document.title`, falling back to the report `name` when
blank so the report — and its live preview — is never headingless), renders each requested format
(bridging PDF through `HtmlToPdfService.Render`), and returns a
single file or a **zip** of several. It reads the clock exactly once. The handler (`handlers/report.go`)
parses+validates the `ReportRequestCommand` **before** building the `structs.Handler` — because the
`groupIds` it carries drive the gate: it declares `GroupIds` + `GroupPermissions: [group.reports.read]`,
so `HandleRequest` re-checks that permission (and membership) in **every** covered group before generating.
A malformed spec surfaces from `reporting.Run` as a `ReportService`-typed `*ReportSpecError`, which the
handler maps to a 400. The request contract mirrors the engine (columns carry a machine `name` that
formulas reference by name with ASCII operators; group-by/detail carry engine field keys), so the Angular
client maps its builder UI onto engine-shaped values before submitting. Report generation is **synchronous**
(streamed download); an async job + live progress + stored-results download is a possible later slice.

**`POST /api/report/preview`** drives the desktop builder's live preview. It shares GenerateReport's
front-loaded parse/validate and the same per-group `group.reports.read` gate (the shared
`loadReportCommand` handler helper), and `ReportService.Preview` runs the **same** pipeline as `Generate`
up through the engine (both call the shared `buildModel`) but renders only `render.HTML` — no PDF bridge,
no zip — and **row-caps** the sample (`reportPreviewRowCap`, currently 1000; `ReceiptCount` still reports
the true total). It returns a JSON `ReportPreviewResponse { html, receiptCount }`, so the preview is the
engine's own output rather than a client-side re-implementation. A separate **app-level `app.reports.read`**
permission gates the desktop report-builder route/nav and the saved-template read endpoints (Legacy Admin
picks it up via add-only role reconciliation; reporting is admin-by-default). Generate additionally
requires the app-level **`app.reports.generate`** (ANDed with the per-group `group.reports.read`), so a
non-admin needs both a custom role granting `app.reports.generate` and per-group generation access.
**Preview enforces that same app-level read gate** — `app.reports.read` **or** `app.reports.readAll`,
declared on the handler as an `AnyAppPermissions` any-of and ANDed with the per-group
`group.reports.read` — so the builder's live feedback loop is reachable by exactly the users who can
open the builder, and preview is never merely group-scoped.

**`POST /api/report/template/{id}/render`** backs the **dashboard report widget** (a view-only widget
that pins a saved template — see `desktop/CLAUDE.md` → "Reports"). It mirrors
`GenerateReportFromTemplate` (loads the stored config server-side) but emits the same JSON
`ReportPreviewResponse { html, receiptCount, allowedActions }` as `/preview`, with two deliberate
differences from the builder preview: it renders the **full dataset** (`ReportService.RenderTemplateForUser`
calls the shared `renderHTML` helper with `rowLimit = 0`, like `Generate` — **not** the capped
`reportPreviewRowCap` sample the builder's `Preview` uses), and it **returns restricted-notice HTML at a
normal 200** (with empty `allowedActions`) when the caller may not view the template or it was deleted,
rather than a 403/404. That's because the widget always drops whatever HTML it gets into its sandboxed
iframe, so there is no client-side "restricted" branch. Authorization is resolved **in the service** via
`PermissionService.AllowedActionsForTemplate` (the single source: the base/`*All` app perms + the
per-group ceiling + the per-template matrix); a result lacking `"read"` yields the restricted notice, and
the returned `allowedActions` (which include `"generate"` iff the caller may generate) gate the widget's
download button off the server result — the download itself reuses the enforcing
`/report/template/{id}/generate`. The `restrictedReportHTML` notice and the extracted `renderHTML` helper
(shared by `Preview` and `RenderTemplateForUser`) live in `services/report_service.go`.

**Custom currency formatting.** `buildModel` loads System Settings (`SystemSettingsRepository.GetSystemSettings`,
a get-or-create singleton) and passes the app's currency configuration — symbol, symbol position (START/END),
thousands/decimal separators, and hide-decimal-places — through `MetaInput.Currency` (mapped by
`currencyFormat`). Because Generate and Preview share `buildModel`, **every** rendered output (the live
preview, PDF/HTML, CSV, and XLSX) presents money exactly as the rest of the UI does — matching the desktop
`customCurrency` pipe that the report's receipts drill-in dialog already uses. The engine stays pure (it
carries `Currency` through untouched); the settings load lives in the service and the formatting in the
`render` package (`render/currency.go`).

**Dynamic report-generator paid-by.** `buildModel` also resolves a **report-generator paid-by sentinel**
before it fans out to `loadRows`: `resolveReportGeneratorPaidBy(&filter, userId)` replaces the value `-1`
in `filter.PaidBy.Value` with the generating user's id (the desktop report builder stores `-1` for its
"Whoever generates the report" paid-by option; negative so it can't collide with a real id). So a saved
template stays **dynamic** — whoever runs it filters to their own receipts, regardless of who authored it
(User A running User B's saved report sees User A's receipts). It runs once, upstream of
`IntersectReceiptFilterWithGrants` + `BuildGormFilterQuery` (which then see a normal numeric
`paid_by_user_id IN (...)`), on a by-value copy of `command.Filter` so the request is never mutated; and
because Generate and Preview share `buildModel`, the substitution covers download and the live preview
alike. Values arrive as `float64` (JSON), matched/substituted as such — the same shape a static user-id
filter already flows through.

**Report templates.** `POST /api/report/template` saves a report configuration for reuse. It reuses the
shared `loadReportCommand` parse+validate and stores the whole `ReportRequestCommand` verbatim as a
`json.RawMessage` blob on `models.ReportTemplate` (name + owner taken from the request / JWT), so a
template round-trips back into the builder unchanged. A stored config may include a dimension column
that is currently **disabled** in the builder (aggregate mode, a label whose field is neither the
aggregate dimension nor a grouping level) — persisted deliberately so it round-trips and self-heals
rather than being silently dropped on save. `buildReportSpec` (`services/report_service.go`,
`isDisabledDimensionColumn`) **projects such a column out** before running the engine, mirroring the
desktop `isDimensionColumnDisabled` and the engine's own `ErrLabelColumnUnresolvable` rejection — so
verbatim generation of a stored template (including `POST /report/template/{id}/generate`) succeeds
with the column omitted instead of returning a 400. Because this is the first feature to **re-serialize
a `ReceiptPagedRequestFilter` back to a client**, the filter's json tags must be correct:
`PagedRequestField.Value` carries `json:"value"` and the filter's `Tags` field `json:"tags"` (both
lowercase, matching swagger) — a capitalized key would deserialize fine (Go is case-insensitive) but
serialize a `Value`/`Tags` the desktop can't read, so the operation would hydrate while the value
silently dropped (see `paged_request_command_test.go`). Unlike generate/preview it is **app-scoped** behind
a new `app.reports.create` permission — `handlers.CreateReportTemplate` gates on `AppPermissions` and calls
`repositories.ReportTemplateRepository` directly (handler→repo, like prompts, no engine involvement),
because it persists a configuration and touches no group's receipts. Legacy Admin picks the permission up
via add-only role reconciliation (reporting is admin-by-default, same as `app.reports.read`); per-group
generation stays gated by `group.reports.read`. `DELETE /api/report/template/{id}` removes a template
(`handlers.DeleteReportTemplate` → `ReportTemplateRepository.DeleteReportTemplateById`, mirroring
`DeletePromptById`; deleting a non-existent id is a 200 no-op), gated by a separate CRUD-granular
`app.reports.delete` (Legacy Admin auto-gains it; no ownership scoping yet — any holder may delete any
template). `POST /api/report/template/list` (`getReportTemplates`) returns a paged, sorted list and
`GET /api/report/template/{id}` (`getReportTemplate`) one template — both gated by `app.reports.read`
(the same read gate as the builder route); the list mirrors the prompt paged-read pattern
(`GetPagedReportTemplates` with a `name`/`created_at`/`updated_at` order-by allow-list, since the column
is interpolated raw), and get-by-id maps `gorm.ErrRecordNotFound` to a 404.
`POST /api/report/template/{id}/duplicate` (`duplicateReportTemplate`) copies a template into a new row
owned by the caller (name suffixed `" duplicate"`, configuration/version carried verbatim, a fresh id),
gated by a separate CRUD-granular **`app.reports.duplicate`**. `PUT /api/report/template/{id}`
(`updateReportTemplate`) overwrites a template in place — it mirrors `CreateReportTemplate` (shared
`loadReportCommand` parse+validate + the same non-empty-name guard) but loads the existing row first
(`UpdateReportTemplate` repo method → `GetReportTemplateById`, so a missing id is a 404, not a silent
insert), preserving its id and owner while replacing name/config/version and refreshing `UpdatedAt`.
Gated by a separate CRUD-granular **`app.reports.update`** (Legacy Admin auto-gains it; no ownership
scoping — any holder may update any template, matching delete/duplicate). Each template carries a
`configurationVersion` (currently `1`, DB default `1`, stamped from
`commands.CurrentReportConfigurationVersion`) marking the schema its stored config was written under, so
a future breaking change to the `ReportRequestCommand` shape can upcast — or fail loud on — old blobs
instead of silently misdeserializing them; upcasters + a migration are deferred until that first break.
The desktop **template-management UI** is delivered: `/reports` lists saved templates and each row can
generate, open-in-builder, duplicate, or delete. Opening a template rehydrates the builder from its
stored config; **the builder's Save updates in place on the edit route** (`app.reports.update`) and
**creates on the new route** (`app.reports.create`) — save-as-new is retired (Duplicate copies).

**Report-template access (group-scoping + per-template matrix + `*All` bypass).** The flat "any holder may
act on any template" model above is **superseded**: the six template handlers now resolve access through
`services.PermissionService` (`internal/services/report_authz.go`) — `CanActOnTemplate`,
`VisibleTemplateIds`, `AllowedActionsForTemplate`, `CanReportOverGroups` — which AND three layers:
1. **App action permission** — the base `app.reports.<action>`, moved *into* the authz service so it can
   be OR-ed with the bypass (the declarative `AppPermissions` gate is AND-only, so the six template
   handlers dropped it and enforce in-body).
2. **Group-access ceiling** — the caller must hold `group.reports.read` in *every* group the template
   covers (most-restrictive-wins), read from the denormalized `report_template_groups` index (synced on
   create/update/duplicate, cascades on delete) since the group ids otherwise live only in the config blob.
3. **Per-template matrix** — `GroupRoleReportTemplateGrant {group_role_id, report_template_id, permission}`,
   a group-role grant alongside category/tag/paid-by. Empty for a role = unrestricted; non-empty = only the
   listed (template, action) pairs. A `ReportTemplateGrantsRestricted` flag fails closed once the last
   granted template is deleted (paid-by style). Resolution rides the per-role grant cache; the delete
   handler flushes it (`services.EvictAllGroupRoleGrants`).

For each action there is an app-scoped **`app.reports.<action>All`** bypass (readAll/createAll/updateAll/
deleteAll/duplicateAll/generateAll) short-circuiting both the ceiling and the matrix — auto-granted to
Legacy Admin (ScopeApp), so admins keep full reach with no migration. The list returns per-row
`allowedActions` so the desktop gates each row's buttons off the server result (never re-AND-ed with a
client permission check). New endpoints: **`POST /report/template/{id}/generate`** enforces the per-template
generate grant (the ad-hoc `/generate` stays app + per-group only — it carries no template id, so "view but
not generate" is only a real boundary via this path); **`GET /report/template/options`** (gated on
`app.roles.read`, the role editor's own gate) feeds the role-form access matrix. Create/update additionally
require `CanReportOverGroups` on the attached groups (createAll/updateAll bypass).

The per-template matrix is **UX scoping** (which saved templates a member sees/acts on) layered on the hard
data-access controls — the group-access ceiling plus the category/tag/paid-by grants — which alone bound the
receipt data any report can reach; so the ad-hoc `POST /report/generate` (no template id, so the matrix never
applies) and a `duplicate`'s resulting copy (which starts matrix-unrestricted) are intentionally not bound by
the per-template matrix and never widen data access beyond those hard controls, which every generation re-resolves.

**`(Restricted)` vs `(None)`.** Aggregation uses `PermissionService.SubstituteRestrictedCategoriesTags`
(not the strip variant): a category/tag the caller may not see is replaced with a single `(Restricted)`
marker, so the receipt still counts toward the totals in its own bucket instead of vanishing. `(None)`
stays reserved for a receipt that genuinely carries no category/tag. The **pie chart** (`pie_chart.go`)
substitutes the same way, so a restricted viewer sees a `(Restricted)` slice rather than hidden spend
folding into `Uncategorized`/`Untagged`. (The CSV export handlers still strip — export is a per-receipt
listing, not an aggregation.)

### Semantics that are easy to get wrong

| Concern | Rule |
|---|---|
| Aggregate rollup | A parent **merges its children's accumulators**, never their finalized values. This is what makes `AVG` at a subtotal `sum(all descendants)/count(all descendants)` rather than the average of its children's averages. |
| Arithmetic rollup | Recomputed from the **same row's** other columns at every level (detail, subtotal, grand total). Never summed. Additive formulas agree either way; ratios and averages do not. |
| `SUM`/`COUNT` of nothing | `0` — a category with no tax shows `0.00`, not an empty cell. |
| `AVG`/`MIN`/`MAX` of nothing | `Null` (empty cell). Zero would be a lie. |
| `COUNT()` | Counts **records**, not values. A row with a null measure still counts. It takes no field. |
| Nulls | Skipped by `SUM`/`MIN`/`MAX` and excluded from `AVG`'s divisor. Any null operand makes an arithmetic result null. |
| Division by zero | `Null` cell, **never a panic** — `shopspring` panics on a zero divisor, so `evalBinary` guards `IsZero()` first. |
| Division precision | `DivRound(x, spec.Config.DivisionScale)`. **Never read or write `decimal.DivisionPrecision`** (a mutable process-wide global). |
| Multi-value fan-out | A receipt with two tags is attributed to **both** buckets in full, so it double-counts, and that double count propagates to the grand total. Intended — matches `services/pie_chart.go`. Two multi-value levels produce a cross product. But only **distinct** values fan out: a receipt tagged `"Alex"` twice is attributed once. |
| Multi-valued measures | Refused (`ErrMeasureIsMultiValued`). Summing a field that resolves to several values would silently read the first and drop the rest. A `Multi` field is still a fine dimension and a fine display label. Note `Multi` is a **producer's declaration**, never checked against the rows: a catalog that lies loses values in `Row.Measure`. |
| Bucket keys | Two values share a bucket key **exactly when `compareValues` finds them equal**. Numbers key on `decimal.String()` (canonical, lossless); dates key on `Unix()` seconds + `Nanosecond()` and **never `UnixNano()`**, which is undefined outside 1678–2262 — a zero `time.Time` and a date in 585 share one. A coarser key merges buckets *and* makes the surviving bucket's value depend on input order. |
| Bucket **values** | A bucket keeps `Value.canonical()`, not whichever member arrived first. Agreeing with `compareValues` is only half the job: dates compare by instant, so one instant in two zones merges, and `Value.Time()` would otherwise hand a renderer an arbitrary one of them (`2026-05-01T00:00:00Z` vs `2026-04-30T19:00:00-05:00` format as different **calendar days**). **Date buckets are emitted in UTC.** A producer wanting calendar-day grouping in a local zone must truncate before handing rows over. |
| Unknown enums | `Validate` rejects an `AggFunc` or `DetailMode` outside the known set (`ErrUnknownAggFunc`, `ErrUnknownDetailMode`). Both used to fail silently — an unknown reduction fell through `finalize` and blanked the column; an unknown detail mode skipped the label-column check. Ask "is this aggregate mode?" only through `DetailSpec.isAggregate()`; two spellings of it once disagreed. |
| Enum switch drift | `AggFunc` is switched on in **four** places (`String`, `valid`, `aggFuncFromName`, `accumulator.finalize`) and Go forces none of them to agree. `enums_test.go` pins them to each other exhaustively over all 256 values. **Adding a reduction means updating all four** — `valid()` alone would let `Validate` accept a column `finalize` blanks. `finalize`'s fallthrough stays `Null()`, not a `panic`: nothing in this package panics, the line is unreachable through `Validate`, and drift is a compile-time mistake caught at `go test` time. |
| Formula size | `ParseArithmetic`/`ParseAggregate` refuse source over `maxFormulaLength` (1 KB) **before parsing**. expr's 10k-node cap does *not* bound this: a parenthesis builds no node, so nesting is bounded only by the goroutine stack (~640 B/paren; ~1.6M overflows the default 1 GB), and a Go stack overflow is a fatal error `recover` cannot catch. |
| Field keys | `NewFieldCatalog` rejects a key that is not a plain identifier (`ErrInvalidFieldKey`). `unit-price` would tokenize as a subtraction. |
| Empty dimension | Explicit `(None)` bucket (`IsNone: true`, `Value` null). Never dropped. |
| Ordering | Buckets sort by typed value with `(None)` last; records preserve input order. **Never range a Go map to produce output** (`pie_chart.go` has exactly this bug). |
| Formatting | The engine emits raw typed values. Currency symbols and decimal places are a renderer's job. |
| `GeneratedAt` | An **input** (`MetaInput`), never `time.Now()`. Determinism depends on it. |
| `IsNone` | Equals `Value.IsNull()` for group nodes and **aggregate-mode** detail rows only. The synthetic `Root` and every **records-mode** detail row carry a null `Value` with `IsNone: false`. It is **not** a global biconditional — do not "fix" the engine to make it one. |

### Columns and formulas

Three column kinds, declared not inferred: `ColumnLabel` (displays a field, blank on subtotal rows),
`ColumnAggregate` (`SUM(amount)`, `COUNT()` — backed by a mergeable accumulator), and
`ColumnArithmetic` (`Subtotal + Hst`, `ROUND(Total / Count, 2)`).

`Column.Name` is what formulas reference and must be a plain identifier that is not a reserved word;
`Column.Label` is the free-text heading. They are separate so renaming a heading (`Avg/Receipt`)
cannot break a formula (`AvgPerReceipt`).

Formulas are parsed by **`expr-lang/expr` used as a parser front-end only** (`parser.Parse` →
`ast.Node`); its VM is never run, because it evaluates over `float64` and money must not touch a
float. The tree is then checked against an allow-list implemented as a **closed type switch whose
`default` rejects**. Two traps: `??`, `in`, `**` and `%` all parse as `*ast.BinaryNode` (so the
*operator* is whitelisted too), and expr ships lower-case builtins named `sum`, `min`, `max`,
`count`, `round`, `len` — so `round(a,2)` arrives as an `*ast.BuiltinNode`, not a `CallNode`.

Arithmetic columns form a **DAG**, topologically sorted with cycle detection, so a column may be
declared before the aggregate it reads. `ColumnDescriptor.Expr` keeps the parsed tree so the future
XLSX renderer can translate a column expression into a live `=SUM(...)` cell formula.

### Tests

Both packages are pure — **no `main_test.go`, no DB, no `app.db` cleanup.** The engine suite builds
synthetic `Row` literals, which is what makes the scenario matrix cheap; `receiptsource` uses real
`models.Receipt` fixtures.

**Compare money with `decimal.Equal`/`StringFixed`, never `reflect.DeepEqual`** — `NewFromInt(200)`
and `200.00` are equal but carry different internal exponents. Determinism is compared through
`serializeModel` for the same reason.

Four layers, each catching what the one before it cannot:

| Layer | File | What it does |
|---|---|---|
| Invariants | `invariants_test.go` | `assertModelInvariants` re-derives the model's structure from the spec. `mustRun` calls it, so **every** engine test checks it. |
| Golden | `golden_test.go` | Renders a whole report as a text table. One assertion covers grouping, ordering, values, subtotal placement and blank cells. |
| Properties | `property_test.go` | 400 randomized specs + rows from a fixed seed (`REPORTING_SEED` overrides). Conservation, rollup, AVG exactness, arithmetic recompute, bucket cardinality, determinism. |
| Fuzz | `fuzz_test.go`, `bucketkey_test.go` | Seed corpora run under plain `go test`. `-fuzz` is opt-in. |

**A law must not be derived from the code it judges.** This rule has now been broken twice, in two
different files, and both times the suite went green over a real defect:

- `property_test.go` computes bucket identity with `compareValues` and reads rows with `Row.Get`,
  deliberately *not* `bucketKey`/`dimensionValues` — an earlier draft used the engine's own helpers
  and consequently agreed with the bug it was meant to catch.
- `invariants_test.go`'s `serializeValue` rendered a date as `bucketKey(value)`, so the determinism
  and permutation laws agreed that two zones naming one instant were interchangeable — which is the
  one thing they exist to deny. It now serializes the wall clock and zone offset a **renderer**
  reads. `Value.String()` normalizes to UTC, so the golden test could not have seen it either.

Likewise the random date alphabet contains both the two instants that share a `UnixNano` **and** one
instant expressed in two zones that straddle midnight. A generator that only produces plausible data
only tests the plausible paths — and note the two `12:00` entries look like such a pair and are not
(one is `12:00Z`, the other `11:00Z`), which is exactly how the gap survived.

### Mutation checking

```bash
./internal/reporting/mutation-check.sh          # all 43 mutations, engine + receiptsource
./internal/reporting/mutation-check.sh avg      # only those matching "avg"
```

Breaks the engine one way at a time and asserts a test objects. Uses `go test -overlay`, so it
**never writes to the working tree**. Run it before merging any change to the engine; a survivor
means the engine can be broken that way without a single test noticing.

Its three self-imposed rules, each learned the hard way: a mutation whose search text no longer
matches is a **failure**, not a pass; a mutant that **does not compile** was never tested, so only
`--- FAIL` counts as caught; and `-count=1`, or the build cache serves the last verdict.

- `internal/reporting/{value,bucketkey,row,field,formula,eval,accumulator,model,validate,engine,invariants,golden,property,fuzz,passthrough,adversarial}_test.go`
- `internal/reporting/receiptsource/receiptsource_test.go`
