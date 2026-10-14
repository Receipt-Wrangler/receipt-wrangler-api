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

**Exception to the exception: a `temp/` path that came back out of an asynq payload.** The exemption
above is for paths the *server builds*. A path read out of a task payload is JSON in Redis and is
attacker-adjacent, so it must go through **`FileRepository.AssertWithinTempDirectory`** before anything
opens or serves it (the activity source-file endpoints are the current callers). The data-scoped
helpers cannot be used — they resolve against `data/` and would reject every temp path — so both
checks now delegate to the shared **`utils.AssertWithinDir(baseDir, path)`**, with the base as a
parameter because the roots genuinely differ: `data/` hangs off `os.Getwd()`, `temp/` off
`config.GetBasePath()`, which `BASE_PATH` can override. Prefer a scoped wrapper over naming a base
yourself, so a new trust boundary is always deliberate. The check is purely lexical (`filepath.Rel`,
no `EvalSymlinks`), matching how every path here is constructed.

**Reading is per-boundary; encoding for a browser is not.** The split above is
about *getting* bytes safely, and it is genuinely different per root. Turning those
bytes into something an `<img>` can render is identical everywhere, so it has one
home: `FileRepository.BuildDisplayImageString`. Any handler returning an image to a
client calls it — never `BuildEncodedImageString` directly, which encodes whatever
it is handed and (via `GetFileType`) will happily label an unconverted PDF
`image/jpeg`. See "Activity source files" below for the bug that cost.

**`utils.ReadFile` returns `(nil, nil)` on ANY read error** — it swallows it. Never use it where a
missing file must be an error: the caller sees empty bytes and a nil error and carries on. That cost
a real bug in `wranglerasynq/email_process_handler.go`, where a deleted attachment sailed past the
read (`hasAttachmentImage` is derived from the payload *string*, not the file) and, if processing then
succeeded, persisted a **zero-byte receipt image carrying the real name and size**. That call site and
the source-file endpoints use `os.ReadFile`; `utils.ReadDataFile` is the propagating equivalent for
data paths.

## Testing Patterns

Each package typically has:
- `main_test.go` - Test setup and teardown
- `*_test.go` - Unit tests for specific functionality
- Test utilities in `internal/utils/testing.go` and `internal/repositories/testing.go`
- The test database enforces foreign keys, so a cleanup test passes on `ON DELETE CASCADE` alone.
  To prove an explicit delete, run it under `repositories.UseTestDbWithoutForeignKeysForTests()`
  (`repositories/db.go`) and `defer` the restore it returns

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
  **blocked by default** to remove an SSRF / tracking-pixel surface. To
  permit remote loads (useful when receipts depend on remote logos or
  product imagery), set `CHROMIUM_ALLOW_EXTERNAL_RESOURCES=true`. Because an
  `about:blank` document (see below) never fetches network sub-resources at all,
  the external mode instead serves the HTML from an **ephemeral loopback HTTP
  server** (`127.0.0.1:0`, one per render, `defer`-closed right after) and
  navigates to it, giving the document a real `http` origin so remote images
  actually load; `chromedp.Navigate` waits for the load event, so slow remote
  images finish before printing.
- **External mode is SSRF-guarded: only PUBLIC hosts may load.** The email body is
  attacker-controlled and the render now has a real `http` origin, so a naive
  external mode would let `<iframe>`/`<img>` reach internal targets (verified: an
  internal response was baked into the PDF). Sub-resource requests are therefore
  intercepted via the CDP **Fetch** domain (`isRequestAllowed` in `html_to_pdf.go`)
  and each is allowed or failed by its resolved address: the render's own loopback
  page and inline `data:` URIs are allowed; loopback/private/link-local/multicast/
  unspecified targets — `127.0.0.1:*`, `169.254.169.254` (cloud metadata), RFC1918,
  `::1`, `fe80::/10`, `0.0.0.0` — plus the non-public ranges the stdlib `net.IP`
  predicates miss (`ipIsInternal`'s `extraInternalNets`: `0.0.0.0/8`, CGNAT
  `100.64.0.0/10` incl. Alibaba metadata `100.100.100.200`, benchmarking
  `198.18.0.0/15`, reserved+broadcast `240.0.0.0/4`, NAT64 `64:ff9b::/96`) and the
  `localhost` / `*.localhost` names (which chromium resolves to loopback itself,
  never asking DNS) — and every non-`http(s)` scheme (incl. `file://`) are
  refused. A URL blocklist can't express this (it can neither allow our own
  loopback page while denying other internal hosts, nor catch a hostname that
  resolves to an internal IP). Residual: **DNS rebinding** — the host is resolved
  at check time and re-resolved by chromium at fetch time, so operators enabling
  external mode in a cloud should also restrict egress at the network layer / use
  IMDSv2. Regressions: `TestHtmlToPdfService_Render_ExternalMode_BlocksInternalSSRF`,
  `TestHtmlToPdfService_Render_ExternalMode_EmbedsDataImage`, and the hermetic
  policy unit test `TestIsRequestAllowed`.
- **`file://` loads are ALWAYS blocked, in BOTH modes.** In the default mode the
  HTML is injected into an `about:blank` page via `page.SetDocumentContent` rather
  than being written to a temp file and navigated to as `file://<path>`. The email
  body is attacker-controlled, so a `file://` document origin let it read local
  files (other groups' receipt images under `data/`, logs, `/etc/passwd`) into the
  PDF via `<iframe src="file:///…">`/`<img>` — even with JavaScript disabled.
  `about:blank` (and the loopback `http` origin used in external mode) has no
  local-file origin, so those sub-resources are cross-origin-blocked; `file://*`
  is also on the default mode's network blocklist and denied by the external
  mode's Fetch policy, as defense in depth (blocked in BOTH modes). Inline `data:`
  URIs still load
  (`SetDocumentContent` also sidesteps the data-URL length cap that motivated the
  old temp-file approach) and decode synchronously, so the default path needs no
  explicit load wait. Regressions:
  `TestHtmlToPdfService_Render_BlocksLocalFileRead` and
  `TestHtmlToPdfService_Render_ExternalMode_BlocksLocalFileRead`.
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
- **Moving a receipt between groups authorizes the DESTINATION.** `UpdateReceipt` runs through
  `HandleRequest` with `ReceiptId` set, which only verifies `group.receipts.update` on the receipt's
  **current** group. Both clients let a user change a receipt's group on edit, so when
  `command.GroupId != currentReceipt.GroupId` the handler additionally requires `group.receipts.create`
  on the destination group (403 otherwise) and runs the grant / member-visibility **selection** checks
  against the destination (the group the receipt will live in). Without this a member with update rights
  in one group could relocate receipts into any group. The **synthetic All group is rejected as a
  destination** (400) *before* that create-check: a caller's All-group membership carries the default
  unrestricted role, so its `group.receipts.create` would otherwise satisfy the check and persist the
  receipt into the cross-group view under a role that sidesteps every real group's grant / visibility
  controls. Tests: `handlers/receipt_group_move_authz_test.go` (incl.
  `TestUpdateReceipt_MoveToAllGroupRejected`).
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
- **All-group expansion is gated per group, not by All-group membership.** The synthetic "All" group
  is a real membership where the caller holds an unrestricted role, so expanding it to every membership
  without re-checking each group let a member read receipts (and category/tag names) in groups whose
  role denies it. `GetPagedReceiptsByGroupId` now takes a `GroupReadableResolver`
  (`PermissionService.GroupPermissionResolver(userId, perm)`, cache-backed) and, for the All view,
  drops any group where the caller lacks the permission its **direct** single-group path requires
  (`group.receipts.read` for list/export/summary, `group.reports.read` for reports, `group.widgets.read`
  for the pie chart) — before the `group_id IN` scope, the paid-by disjunction, and the count. The
  `GroupReadableResolver` and `CategoryTagVisibilityResolver` must be passed together for the All view
  (`GetPagedReceiptsByGroupId` fails closed if only one is supplied) — the read gate without the
  category/tag resolver would let a category/tag filter fall through to the flat, group-unscoped subquery.
  The same per-group read
  filter is applied in `SearchReceiptsForUser` (search is gated on `app.receipts.search`, which does
  not grant per-group read) and in `GetAmountOwedForUser`'s All aggregation. It also carries a
  `CategoryTagVisibilityResolver`: for the All view the category/tag **filter** is applied as a
  per-group disjunction (`applyAllGroupCategoryTagFilter`) so a filter on a category id only matches
  receipts in groups where that id is visible to the caller — closing a probe where the flat,
  group-unscoped filter subquery matched a restricted category across groups (`IntersectReceiptFilterWithGrants`
  is a no-op for the All view because the All membership resolves unrestricted). AppData's
  `groupCategories`/`groupTags` for the All group is likewise the **union** of the caller's per-real-group
  visible sets, not the whole pool. Tests: `repositories/receipt_all_group_gate_test.go`,
  `services/auth_test.go` (`TestGetAppData_AllGroupCatalogIsUnionOfRealGroups`), `handlers/users_test.go`
  (`TestGetAmountOwedForUserAllGroupExcludesUnreadableGroup`).
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

**`UserVisibleInGroup(viewerId, targetId, groupId)`** is the single-row form of the
per-group resolver below, for a gate on one record rather than a batch. It exists
because that closure was being re-inlined per call site (the comment-notification
fan-out, and again in its test) while the in-package equivalent,
`groupVisibilityResolver.isVisible`, is unexported and memoized for batch use. Use
it for any per-request check; use the resolvers below when filtering a set.

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

### API key scopes

An API key carries a scope — `r`, `w` or `rw` — and it is **enforced in
`middleware.UnifiedAuthMiddleware`**, the one place a key becomes `structs.Claims`. The scopes are
**literal**: `r` denies writes, `w` denies reads, `rw` allows both.

It was not enforced anywhere until this change. The value was stored on the key, copied into the
claims, and validated for well-formedness — and then read by nothing. Both `IsValid()` calls check
that the value is one of the three, never that the request is allowed, and `enforcePermissions`
resolves everything from the caller's user id without consulting it. So a key created as "Read" could
create, update and delete anything its owner could. That was not a dormant field: the desktop form
presents a required Read / Write / Read-Write picker and **defaults to Read**, so the most restrictive
label was the default and it meant nothing.

**The match on the scope string is EXACT, and that is load-bearing rather than tidy.** It was
`strings.Contains(scope, "r")` / `"w"`, which read the `r` in `rx` and both letters in `wr`, `rww`,
`read`, `write` and `r,w` — the last of which granted every request class. `apiKeyScopePermits`
switches on `models.ApiKeyScope`'s three constants and denies anything else, because this gate is
the **only** check that runs at request time: `Claims.Validate` is the
`validator.WithCustomClaims` hook and never fires on the API-key path (`GetClaimsFromApiKey`
hand-builds the claims), and `models.ApiKey.Scope` is a plain `string` column — not
`models.ApiKeyScope` — with no `driver.Valuer`, so GORM never checks it either. That leaves
`UpsertApiKeyCommand.Validate` at create/update time as the only other guard, and anything writing
that column by another route (a migration, a seed, manual SQL) arrives here unvalidated. Do not
"simplify" it back to a substring test.

**The HTTP method decides, with an allowlist.** GET/HEAD/OPTIONS are reads; everything else is a
write unless the path is on `readOnlyPostPaths` / `readOnlyPostPatterns`
(`middleware/api_key_scope.go`). This fails closed — a write endpoint added later is denied to an `r`
key with nobody having to remember — and the opposite mistake, a new read-over-POST wrongly denied,
surfaces as a broken call rather than a silent hole.

**The allowlist is 23 entries, not a handful.** This API serves that many reads over POST because
`PagedRequestCommand` and friends read their filter from a JSON body: every paged list, five report
endpoints, both exports, the pie chart, the receipt summary and `convertToJpg`.

**"Read" means DISCLOSURE to the caller, never "touched a row".** The gate operates on HTTP
requests, not on queries. `UpdateReceipt` internally reads the receipt, re-resolves permissions and
merges hidden associations — none of that is a read in this sense, because none of it reaches the
caller. A `w` key may PUT all day and the server reads whatever it needs to serve it.

**The corollary is the one easy to miss: a write that RETURNS data is a read too.** The method-based
model takes for granted that a write does not disclose, and that is an assumption rather than a
guarantee. `POST /api/receipt/{id}/duplicate` and `POST /api/report/template/{id}/duplicate` break
it — each takes nothing but an id, supplies no content of its own, and answers with the **source's**
contents. Classified as plain writes, a `w`-only key could read any receipt or template it may
duplicate simply by duplicating it and reading the answer, which defeats the one thing
write-without-read exists for. Both are named in `readWritePostPatterns` and classify as
`classReadWrite`, which only `rw` satisfies. An endpoint that both discloses and changes belongs on
that list rather than forced into a binary that does not fit it.

**Do not derive read-vs-write from the declared permissions.** It is the obvious shortcut and it
fails OPEN. Both `checkConnectivity` endpoints persist a `SystemTask` while carrying a `.read`
permission (`services/ai.go`, `services/system_email.go`), and `app.reports.generate` sits on a pure
read. `createDefaultPrompt`, `pollGroupEmail` and `rerunActivity` are likewise writes with read-ish
names.

**`UpdateApiKeyLastUsedDate` must stay outside the gate.** It fires on every API-key request
including plain GETs, from a goroutine in the same function — it is authentication bookkeeping, not
the caller's operation, so a read-only key must still stamp its own last use. Phrasing the predicate
as "does this request write anything" gets that backwards.

**Scope is not the only thing a credential decides.** Some operations are ordinary for a signed-in
human and an escalation for a key, whatever its scope, so `UnifiedAuthMiddleware` also stamps
**`structs.Claims.AuthSource`** (`AuthSourceJwt` / `AuthSourceApiKey`) — it is the only thing that
knows which branch ran, and the only writer. Its consumers are the OIDC link pair — see "OIDC
(OpenID Connect) login" → "Linking an existing account" for the escalation it closes — plus
self-delete (`handlers.DeleteAccount`) and its provider confirmation (`oidc.DeleteConfirmationStart`),
so a key cannot destroy the account that issued it. Three rules
travel with it: the field is **`json:"-"`**, so it cannot ride on a token, be set by whoever minted
one, or be inherited by a legacy one; a check is written as **"require `AuthSourceJwt`"**, never
"reject `AuthSourceApiKey`", so the zero value denies rather than slips through; and it is **not**
inferred from `ApiKeyScope` being non-empty, which happens to hold, is serialized, and fails open.

**Out of scope by construction:** login, logout, token-refresh and signup never mount this
middleware, so an API key cannot reach them; MCP authenticates an audience-bound bearer token rather
than an API key.

This is an **enforcement change with teeth** — an existing `r` key being used for writes starts
getting 403s. That is the promise finally being kept, but it will break integrations on upgrade.

Tests: `middleware/api_key_scope_test.go` — a table over the whole route inventory × all three
scopes (expectations stated as an explicit matrix, not derived from the function under test), the
classifier's near-misses for both pattern lists, the path matcher's near-misses, the literal rule
including fail-closed on an unknown scope — the invalid set carries the substring near-misses
(`rx`, `wr`, `rww`) and the plausible values (`read`, `write`, `r,w`, borrowed from
`models/api_key_scope_test.go`'s own invalid table) rather than only values that never contained an
`r` or a `w`, which is why the original four passed trivially — the last-used stamp surviving on a
read, and a JWT caller being unaffected.
`TestUnifiedAuthMiddleware_ApiKeyScopes` used to assert a `w` key succeeding on a **GET** — it
documented the gap, and its name is a large part of why the gap survived review.
`middleware/auth_source_test.go` covers the marker: both JWT shapes (header and cookie) and every
key scope, plus `TestAuthSourceNeverRidesOnAToken`, which signs a JWT that *claims* to be an API key
and asserts the request is still read as the session it is.

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
    - **`services.LoginUser` itself rejects dummy users and empty passwords**, so the OAuth login form
      cannot be used to authenticate a passwordless placeholder account. The REST login enforced this
      via middleware + a post-check, but the OAuth authorize form called `LoginUser` directly and did
      neither — a dummy user (stored as `bcrypt("")`) could obtain an MCP token with an empty password.
      Centralizing the guard in `LoginUser` covers every caller. Tests: `services/auth_test.go`
      (`TestLoginUserRejectsDummyUser`, `TestLoginUserRejectsEmptyPassword`), `oauth/oauth_test.go`
      (`TestAuthorizeRejectsDummyUserAndEmptyPassword`). It also refuses an **OIDC-only account** with
      no password of its own — see "OIDC (OpenID Connect) login" → "A provisioned account's password".
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

## OIDC (OpenID Connect) login

Users can sign in with an external identity provider. The API is the **relying party** — a
confidential client that owns the whole exchange (discovery, PKCE, state, nonce, code exchange, ID
token verification), so no desktop or mobile client ever handles an IdP token or the client secret.
This is the mirror image of `internal/oauth`, which is an OAuth 2.1 **authorization server** for MCP;
they share no flow, only two primitives in `internal/utils`.

- **Package**: `internal/oidc/` (`flow.go` login/callback, `provision.go` identity resolution,
  `exchange.go` the mobile handoff, `store.go` session/code storage, `provider_cache.go`,
  `claims.go`). Provider CRUD follows the usual chain:
  `routers/oidc.go` (both the flow and the CRUD routers) → `handlers/oidc_providers.go` →
  `services/oidc_provider.go` →
  `repositories/oidc_providers.go`, gated by `app.oidc-providers.{create,read,update,delete}`.
- **Library**: `github.com/coreos/go-oidc/v3` + `golang.org/x/oauth2`.

### The identity anchor: (provider, subject)

`models.OidcIdentity` is unique on **both** `(oidc_provider_id, subject)` and
`(oidc_provider_id, user_id)`. `sub` is the only claim OIDC guarantees stable and unique within an
issuer, so every login after the first resolves through that one lookup and consults no other claim.
A user renaming themselves at the IdP — or an IdP recycling a released username — therefore cannot
re-point an existing account. `Subject` is capped at 255 so the composite index stays under MySQL's
3072-byte limit; that is also why `PreferredUsername` is deliberately outside it.

**First login only**, when no link row exists yet, `resolveUser` may:

1. attach to an existing local account whose `username` equals the `preferred_username` claim, if the
   provider's **`LinkByUsername`** is on — **default false, deliberately**. OIDC does not promise that
   claim is stable or unique, and public providers (Twitch notably) let users change their username
   and recycle released names, so an always-on match is an account-takeover path against a local
   `admin`. The safe way to attach an existing account is the authenticated link flow below.

   **The match is re-asserted byte for byte in Go** (`usernameMatchesClaim`), not left to the SQL
   lookup that found the row. `users.username` pins no collation, so the same claim resolves
   differently per engine: MySQL and MariaDB default to a case-**insensitive** collation, where an
   identity provider account named `ADMIN` resolves the local `admin`, while SQLite and Postgres
   compare case-sensitively and would not. A boundary that holds on one supported engine and not
   another is not a boundary. An inexact match is treated as a **miss**, so it falls through to
   provisioning — where `deriveUsername` lower-cases the candidate, collides, and refuses with
   `ErrAccountExists` rather than quietly creating a lookalike account. It is a named function
   precisely because the test database is SQLite: inlined, the rule would be unreachable there and
   would ship with no test that can fail on it.
2. **provision** a new account, if `AllowProvisioning` is on.
3. otherwise refuse (`ErrNoAccount` → a `no_account` error redirect).

**A username collision refuses rather than suffixing** (`ErrAccountExists`). Silently creating
`bob-2` is indistinguishable from data loss from the user's side, and on a self-hosted install with
one IdP the collision is almost always the same person who has not linked yet. Uniqueness is normally
enforced by `middleware.ValidateUserData`, which this path does not go through.

**The account and its link row are written in ONE transaction**, and the failure mode is why.
`CreateUser` and the identity insert used to be separate statements, so losing the second left an
account with nothing pointing at it — and `deriveUsername` is deterministic, so the next login for
that subject derives the same name, `GetUserByUsername` finds the orphan, and `provisionUser`
answers `ErrAccountExists`. Permanently: it cannot tell an orphan from a real local account of that
name, and must not try. The identity insert is far from unfailable — two first logins for one
subject (two tabs, a double-tapped button, a retried callback) race on the
`(oidc_provider_id, subject)` unique index, and the loser lands exactly there. `CreateUser` opens
its own `db.Transaction` off `repository.GetDB()`, which is the injected tx, so GORM nests it as a
**savepoint**. Guarded by `TestProvisionLeavesNoOrphanWhenTheIdentityInsertFails`, whose second
assertion reproduces the lockout itself.

**A provisioned account's password is real randomness, generated and discarded.**
`UserRepository.CreateUser` bcrypts whatever it is handed, so a sentinel such as `"!oidc"` would
become a *working* password for anyone who knew it — through the normal login form and through the
MCP OAuth login form, which shares `services.LoginUser`. Guarded by
`TestProvisionedPasswordIsUnusable`.

**And `LoginUser` refuses such an account outright** (`ErrPasswordLoginUnavailable`), whatever
password is presented. The randomness makes the password unguessable; the refusal makes the rule hold
even if that ever changed. "Such an account" is the passwordless fact below,
`OidcIdentityRepository.IsAccountPasswordless`, the same question the unlink guard asks. Two details:

- **The check runs AFTER the password verifies**, and both callers replace the error with their
  generic wrong-password text. Run first, it would answer faster than a bcrypt compare and tell
  anyone which usernames belong to identity provider accounts.
- **An administrator's password reset is the way back in**: `ResetPassword` clears the flag, so the
  account has a password of its own and signs in either way.

Tests: `TestLoginUserRejectsPasswordlessOidcAccount` and
`TestLoginUserAllowsLinkedAccountWithItsOwnPassword` (`services/auth_test.go`), and
`TestLogin_PasswordlessOidcAccountLooksLikeWrongPassword` (`handlers/login_test.go`), which asserts
the refusal's status and body match a wrong password's.

**Known gap: deleting a provider deletes the passwordless fact with it.** The flag lives on identity
rows, and `DeleteOidcProvider` removes that provider's rows. An OIDC-only account that only had that
provider then reads as a local account: password login is no longer refused, though nobody knows the
password, and only an administrator's reset gets the user back in. The fix is to hold the fact on
`User` instead, which needs a schema change and a migration. Deferred.

**Deleting a user removes its OIDC rows explicitly** (`services.DeleteUser`): its identities, any
pending signed-in session (`link_user_id`, a link or a delete confirmation), link launches, exchange
codes and delete confirmations. Do not leave these to
`ON DELETE CASCADE`. Two of those tables have no foreign key at all. And an orphaned identity is an
account-takeover path: it would sign its subject into whichever account later reuses the id.
`TestDeleteUser_RemovesOidcRecords` runs the delete with foreign keys **off**, so it fails if the
cascade is the only thing doing the work. It does that through
`repositories.UseTestDbWithoutForeignKeysForTests`, a second pool on the test database opened with
the pragma off. The obvious alternative, pinning the existing pool to one connection and running
`PRAGMA foreign_keys = OFF`, **deadlocks**: `DeleteUser` opens a second connection mid-transaction
(`GroupService.GetGroupIdsForUser` reads through a repository built without the transaction).

### The callback's security checks, in order

`internal/oidc/flow.go` `Callback`, failing closed at every step:

1. An IdP-reported `error` → error redirect. Never echo the upstream text.
2. **Consume the auth session atomically** (`UPDATE ... WHERE state_hash = ? AND used = false AND
   expires_at > ?`, checking `RowsAffected`) — **before any network call**, so a replayed state cannot
   even reach the identity provider. Unknown, expired and used states are indistinguishable.
3. The session's provider must match the callback's provider.
4. **Browser binding.** A short-lived HttpOnly cookie set at login start, compared by hash in
   constant time. Without it, state alone does not stop an attacker starting a flow, harvesting the
   state and code, and planting them in a victim's browser — silently signing the victim into the
   *attacker's* account. **The cookie is `SameSite=Lax`, deliberately unlike `BuildTokenCookies`'
   `Strict`**: the callback is a cross-site top-level GET, and Strict would drop the cookie on exactly
   the request that needs it — a failure that appears only in production.
5. `Exchange` with our PKCE verifier.
6. **A missing `id_token` is a hard failure** — an OAuth-only response carries no verified identity,
   and userinfo is not signed.
7. `Verify` (signature via JWKS, `iss`, `aud`, `exp`).
8. **Compare the nonce ourselves.** go-oidc documents that `Verify` does *not* validate it. An empty
   nonce is rejected before the compare, because it hashes to a fixed value and must never match.

**A failure returns to where the flow started, which is not the same as "to the right client".**
After the session is consumed, `failCallback` routes by `session.IsLink()` — a link's caller is
already signed in, so the login screen reads as "you have been signed out" rather than "that
provider could not be connected" (the rule `redirectLinkError`'s own comment states). Only
`finishLink` honored it at first, leaving every check between the provider match and the nonce
compare bouncing a signed-in user to a login form; a desktop LINK and a desktop LOGIN are both
"desktop", so the client type alone cannot tell them apart. **Before** the consume there is no
session to read, and the identity provider appends only `code`/`state`/`error` to the fixed
redirect URI, so `pendingClientType(state)` recovers the client type from the pending row —
read-only, scoped to a LIVE session, one column, and anything unknown degrades to desktop. It
matters for the branch that is not an attack at all: a user who declines consent. Answered as
desktop, a mobile flow strands the app waiting on its private-use scheme while the external browser
sits on the web login page.

**Redirect targets are fixed server-side constants.** No `redirect_uri` or `returnUrl` is ever taken
from the request — that is the open-redirect hole this design does not have.

### Storage

State, nonce and the mobile exchange code are stored **hashed** (`utils.Sha256Hash`), like
`RefreshToken.Token`: they are bearer secrets we only compare, so a database dump cannot complete a
pending login. The PKCE **verifier cannot be hashed** — it must be replayed to the IdP verbatim — so
it is AES-GCM encrypted, as is the provider's client secret (the `SystemEmail.Password` precedent).
TTLs: auth session **10 min** (must cover an MFA prompt), exchange code **2 min** (the app already
holds the browser result), link launch **2 min**, delete confirmation code **5 min**. All are swept by
an `@every 24h` cron, which is hygiene only — `expires_at` is inside every consume statement's WHERE
clause.

### Mobile handoff

`GET /oidc/{name}/login?client=mobile&codeChallenge=…` → the callback redirects to
`io.receiptwrangler://oidc?code=…` carrying a single-use, PKCE-bound code, **never tokens**: a
private-use scheme is unverifiable on Android, so any installed app can read the redirect; the
verifier the app kept is what makes an interception useless. `POST /oidc/exchange` then returns the
full `AppData` with `jwt` + `refreshToken` — byte-identical in shape to `POST /login/?tokensInBody=true`,
so mobile reuses `storeAppData` unchanged, and **no cookie is ever set** on that route.

The code is **loaded, verified, then burned** (never burned first): otherwise anyone who intercepted
the redirect could destroy a valid code just by presenting a wrong verifier. Pinned by
`TestExchangeDoesNotConsumeTheCodeOnAWrongVerifier`.

### Linking an existing account

`GET /oidc/link/{name}` (authenticated) runs the same flow with `LinkUserId` set, so the callback
**skips the whole match/provision tree** and writes the link directly — the session proves identity,
nothing is inferred from a claim. This is what makes `linkByUsername = false` a comfortable default.
`middleware.UnifiedAuthMiddleware` reads the `jwt` **cookie** before the header, so a plain top-level
link navigation authenticates. Unlinking refuses when it is the caller's last identity *and* the
**account** is passwordless — that account has only the discarded random password.

**`ProvisionedUser` is an account fact held on every identity, not a per-row one**, and that is load
bearing. It used to mean "this identity created the account", so exactly one row carried it — and
unlinking *that* row first destroyed the only record that the account had no password, letting the
next unlink strand it. Provisioned by Google, link GitHub, drop Google, drop GitHub, locked out. Three
places now maintain the invariant: `provisionUser` sets it, `resolveLink` **inherits** it when
attaching a second provider, and `ClearProvisionedFlagForUser` clears it across all of them when a
password is actually set. `UnlinkIdentity` also re-asserts it across what remains before deleting a
flagged row, so an account whose rows predate the invariant heals the first time it is touched rather
than needing a data migration. The wire name stays `provisionedUser` — it is `required` and
non-nullable in the generated Dart client, so renaming it would break released builds for nothing.

Clearing on password-set is what makes the refusal's own advice work. The message says "ask an
administrator to set one before disconnecting"; before this, the administrator could comply and the
unlink stayed refused, because nothing observed the reset (`handlers.ResetPassword`).

**It is gated on `app.account.update`**, the same permission as `DeleteOidcConnection`. Connecting a
provider ADDS a way to sign into an account, so it is at least as privileged as disconnecting one;
without the check an administrator could build a role that cannot remove a sign-in method but can
still add one, which is not a coherent thing to be able to configure. The check sits in the handler
body rather than on a `structs.Handler` because this route answers with a redirect — `HandleRequest`'s
403 would drop a raw error object into a browser the user navigated here. Same shape as
`GetPagedApiKeys`' `app.api-keys.read-any` check.

**Neither linking nor unlinking may be done with an API KEY** — both require an interactive
session (`structs.Claims.AuthSource == AuthSourceJwt`, stamped per request by
`UnifiedAuthMiddleware`). This is a privilege escalation, not a tidiness rule. `GET
/oidc/link/{name}` is a GET, so `middleware.classifyRequest` reads it as a read and the weakest
possible key — scope `r` — clears the scope gate; Legacy User holds `app.account.update`, so the
permission check clears too. From there the holder opens the launch URL (or follows the 302) in
their **own** browser, authenticates at the identity provider as themselves, and the callback
writes that identity against the key owner's account. "Log in with X" afterwards is a full
interactive session — one that survives revoking the key, and that surfaces only as an extra
Connected Accounts row. There is no pre-existing equivalent to shrug it off against: a non-admin's
key cannot reach `ResetPassword` (`app.users.update`, which Legacy User lacks) and
`UpdateProfileCommand` carries no password, so this would be the **first** route from a scoped,
revocable credential to an unscoped, unrevocable one. `DELETE /oidc/connections/{name}` carries the
same rule for the reason it already carries the same permission.

**The check is phrased as a REQUIREMENT, never as a rejection**, and `Claims.AuthSource` is
`json:"-"`. Both matter. `AuthSource`'s zero value is neither source, so "require `AuthSourceJwt`"
denies anything that reached a handler without passing through `UnifiedAuthMiddleware`, while
"reject `AuthSourceApiKey`" would wave it through. And a serialized field would be settable by
whoever minted the token and inheritable by any legacy one — "how did THIS request arrive" is not a
property of a token at all. Do **not** infer the same thing from `ApiKeyScope` being non-empty:
it happens to hold today, it *is* serialized, and the inference fails open. `swagger.yml` drops
`apiKeyAuth` from both operations; `GET /oidc/connections` deliberately keeps it, since listing the
caller's own providers is what a read key is for.

**A mobile link start answers with JSON, not a 302** — `GET /oidc/link/{name}?client=mobile` returns
`structs.OidcLinkStartView{launchUrl}`. This is not a stylistic split: the app authenticates with a
**bearer token**, and the external user agent RFC 8252 requires cannot carry one, so a redirect into
that browser arrives unauthenticated and is refused. The whole flow-start error vocabulary follows it
(`structs.OidcFlowError{errorCode}`, mapped to 404/400/403/502 by `writeOidcFlowError`), so a failure
reaches the app in the same shape as a success instead of as an unparseable redirect.
`resolveClientType` therefore runs **before** the provider lookup — otherwise an unknown-provider
failure took the desktop branch and the app got HTML.

**That start creates NO session — it hands out a launch handle, and the browser starts the session.**
This is the one genuinely subtle part of the feature, and it exists because of a real hole. It first
shipped returning the IdP's **authorization URL** directly, reasoned as "safe, because the caller
already proved who they are with a bearer token on the start request". That is true of the *start* and
false of the *callback*, which is a separate unauthenticated browser request — and the callback is
where the link is written. Identity providers leak the authorization URL by Referer from their own
login pages, so anyone who obtained it could authenticate as **themselves**, drive the callback, and
have their identity grafted onto the caller's account.

PKCE cannot fix that, and it is worth knowing why, because it looks like it should. Binding the link
to the app's own challenge is circular: the challenge would come from the victim's session, so the
victim's app holds a matching verifier by construction, and on Android the attacker can deep-link
their result into the victim's waiting `flutter_web_auth_2` (the manifest's `CallbackActivity` is
`exported` + `BROWSABLE`). **A post-callback artifact cannot retroactively bind an unbound callback.**

So: `LinkStart` creates a single-use `models.OidcLinkLaunch` and returns
`{serverPublicUrl}/api/oidc/link/{name}/launch?h=…`. The app opens *that*, the external browser
follows it to `oidc.LinkLaunch` (unauthenticated — the handle is the authorization), and **that**
request creates the auth session and sets the binding cookie. The handle never reaches the IdP: a 302
does not make the redirecting URL the next request's Referer.

**Every session is bound, mobile included** (`store.go`, `flow.go`). This too was wrong at first:
`BindingHash` was skipped for all mobile sessions because "an external user agent may not carry the
cookie", and the callback's check read `len(session.BindingHash) > 0 && …`, which turned it into a
no-op for exactly those sessions. The premise was false for **mobile login** — the app hands the login
URL itself to `flutter_web_auth_2`, so the browser fetches the start and round-trips the cookie fine,
and the gap was a live session-fixation bug there too. The check is now unconditional and fails closed
on an empty hash.

**KNOWN RESIDUAL — the launch endpoint can be pointed at someone else's browser.** `LinkLaunch` is
unauthenticated by necessity, and the handle it consumes names the account to link. So an attacker
holding a handle *for their own account* can send that URL to a victim: the victim's browser gets the
binding, the victim authenticates at the identity provider, and their identity is attached to the
**attacker's** account. The victim's later "Log in with X" then lands in the attacker's account.

It is narrower than what it replaced — it needs active phishing plus an identity-provider login,
where the previous shape needed only a leaked URL — and it fails outright if either party has already
linked that provider (`ErrAlreadyLinked` / `ErrIdentityLinkedElsewhere`). The 2-minute single-use TTL
raises the bar further. It is **not** closed.

Closing it needs the app to prove it ran the flow, and note *which* proof works. A handle-only
confirmation does not: the attacker holds the handle too. What cuts the attacker out is a secret the
**callback** mints and hands to the browser that completed the flow — they never receive it. So:
the callback stops linking inline for mobile, mints a completion code against the launch row, and
redirects with it; the app redeems it on an authenticated call that checks the caller owns the row
(and a PKCE challenge carried from the start, so a rogue app on the device that grabs the private-use
redirect cannot redeem it either). Note this only works **on top of** the launch hop: without the
binding, the same completion code is redeemable by the victim's own app on the attacker's behalf,
because the challenge would be the victim's own.

**A mobile link needs no `codeChallenge`**, unlike a mobile login — but for a different reason than
first documented. The challenge binds the login's one-time exchange code to the app that started the
flow; a link's callback hands back only `?linked={name}`, which is not a credential, so there is
nothing to bind. Binding and PKCE defend different things: binding stops an attacker **driving** your
flow, PKCE stops a rogue app **stealing its result**. `startFlow` requires the challenge only when
`linkUserId == nil`. The callback returns to `io.receiptwrangler://oidc?linked={name}` (or
`?error={code}`) — **never** the desktop profile path, which would strand the external browser on a
page the app never sees.

The static route segments (`link`, `exchange`, `connections`) come first so chi's resolution stays
unambiguous, and `commands.ReservedOidcProviderNames` rejects them as slugs from the other direction.
The provider **name is immutable on update**: it is baked into the redirect URI already registered at
the IdP.

### Confirming account deletion with the provider

Self-delete (`POST /user/deleteAccount`) needs a confirmation, and an OIDC-only account has no
password to give: its own was random and discarded, and password login refuses it outright. Apple
5.1.1(v) and Google Play both require in-app account deletion for any app that lets users create
accounts, and mobile's "Log in with X" does. So such an account confirms by **signing in at one of its
linked providers again**, and gets a single-use code to spend in place of the password.

**The flow is a third purpose on the link machinery, not a new flow.** `OidcAuthSession.Purpose` and
`OidcLinkLaunch.Purpose` (`""` login, `link`, `delete_account`) say what a session is for.
`LinkUserId` stays the column for "the signed-in account that started this", for both signed-in
purposes, which is why `DeleteUser`'s `link_user_id` cleanup and the failure routing back to the
profile page cover the new flow for free. `FlowPurpose()` reads it. A row from before the column
existed has `Purpose ""` plus a `LinkUserId`, and still reads as a link, so a session in flight across
the deploy keeps its meaning. `IsLink()` is `FlowPurpose() == link`; **never** test
`LinkUserId != nil` for "is a link" again.

- **Start: `GET /oidc/connections/{name}/delete-confirmation`** (`oidc.DeleteConfirmationStart`).
  It sits under the reserved `connections` segment, so it needs no new reserved provider slug. It
  shares `authorizeSignedInStart` with `LinkStart`: an interactive session (an API key may not begin
  destroying the account that issued it), then **`app.account.delete`**, the delete's own permission.
  Then two checks of its own, both refusals rather than fallbacks:
  - **The account must be passwordless** (`IsAccountPasswordless`), else `forbidden`. A password
    account confirms with its password, so a provider is never a second way around it.
  - **It must hold an identity at this enabled provider**, else `not_linked`. The callback only
    accepts that identity.
  Desktop gets a 302. Mobile gets the same `OidcLinkStartView{launchUrl}` as a link start, and the
  launch row carries the purpose through the unauthenticated `/link/{name}/launch` hop.
- **Fresh sign-in: `max_age=0` only.** It is the standard OIDC parameter, and the only one whose
  result can be checked: a provider that honours it must return `auth_time`. **Do not add
  `prompt=login`**: Google documents only `none`/`consent`/`select_account`, so it could fail the flow
  outright there. `TestLoginAndLinkDoNotAskForAFreshLogin` pins that nothing else sends it.
- **Callback: `finishDeleteConfirmation`.** Two checks, then the code:
  - **The identity must be the account's own.** The `(provider, sub)` lookup must land on the
    session's user, else `identity_mismatch`. That covers another user's identity and an unlinked one.
  - **When `auth_time` is present it must not predate the session's `CreatedAt`** (60s skew), else
    `stale_login`: the provider ignored `max_age` and reused an old session. When it is absent the
    round trip is accepted, deliberately. Refusing would leave every user of a provider that never
    sends it unable to delete their account, which is the requirement this exists to meet.
    `idTokenClaims.AuthTime` is a `tolerantTime` that never fails the decode (number, fraction or
    numeric string; anything else reads as absent), because go-oidc exposes no field for it and a
    stricter type would break every login on a provider that formats it oddly.
  - **Nothing is deleted in the callback.** It mints an `OidcDeleteConfirmation` (hashed, 5 min,
    single-use) and returns it: `/settings/user-profile/view?tab=user-profile&deleteConfirmation=…`,
    or `io.receiptwrangler://oidc?deleteConfirmation=…`. The client spends it on the normal delete
    endpoint, so deletion keeps one endpoint, one permission check and one session check.
  - Failures go to the profile page as `deleteError=…` (a link's are `oidcError=…`, so the page can
    say which attempt failed) and to the app as `?error=…`.
- **Its own table, not a kind of exchange code.** An exchange code buys a session. "Delete my
  account" and "give me a session" must never share a redemption path.
- **The delete: `DeleteAccountForUser(userId, command)`.** The ACCOUNT decides which confirmation
  applies, never which field the request filled.
  - A passwordless account must send `confirmationCode`. It is consumed atomically with the account
    in the WHERE clause (`ConsumeDeleteConfirmation`), so another account's code burns nothing.
  - A password account must send its password; a code is ignored.
  - A code refusal is a **400**, not the password path's 401, because the desktop reads 401 as "ask
    for the password again". The code is spent before `DeleteUser` runs, so a delete refused after
    that (`ErrLastAdmin`) needs a fresh confirmation.
- **Pre-consume failures route by purpose too.** `pendingSession` reads client type, `link_user_id`
  and purpose from the live row (`GetPendingAuthSessionRouting`). So a user who declines at the
  provider lands back on the profile page, not `/auth/login`. This also fixed the same gap for links.
- **Error codes are plain strings** on the contract (`OidcFlowErrorView.errorCode` is `type: string`),
  so `not_linked`, `identity_mismatch` and `stale_login` cannot break a released mobile build.

Tests: `oidc/delete_confirmation_test.go`:
- the start's four refusals;
- `max_age=0` and the session purpose;
- the desktop and mobile round trips;
- `identity_mismatch` for another user's identity and for an unlinked one;
- the `auth_time` matrix;
- declining at the provider returning to the profile, for delete and for link.

`oidc/claims_test.go` covers `TestAuthTimeDecoding`. `services/users_test.go` covers the code path
(valid, missing, unknown, spent, expired, foreign, single-use across `ErrLastAdmin`), and a password
account ignoring a code. `handlers/users_test.go` (`TestDeleteAccountWithAConfirmationCode`) covers
the API key still getting a 403. Every one of the new checks was mutation-tested: each, broken alone,
fails a test.

### Configuration

Providers are rows in `oidc_providers`, administered from System Settings. `serverPublicUrl` (a new
System Setting, `*string` on the command with the same omitted-key semantics as the refresh-token
lifetimes) builds each provider's redirect URI, `{serverPublicUrl}/api/oidc/{name}/callback`, shown
read-only on the provider form. It is kept **separate from `mcpPublicUrl`** so an OIDC deployment
tweak cannot invalidate live MCP connector tokens.

**That dependency is enforced from BOTH directions, and it has to be.**
`OidcProviderService.validateCrossSettingRules` refuses to *enable* a provider with no
`serverPublicUrl`; `handlers.serverPublicUrlStillNeeded` refuses to *clear* it while any enabled
provider exists (a 400 under `serverPublicUrl`, reusing `services.ErrServerPublicUrlRequired`). Only
the first existed at first, so the same broken configuration was reachable by doing the two steps in
the other order — and the result is invisible from here: `GetServerPublicUrl` falls back to
`http://localhost:8081`, every provider starts sending a `redirect_uri` nobody registered, and the
rejection happens at the identity provider. The check lives in the **handler**: there is no system
settings service, and the repository cannot import `services`. It triggers on a blank value only —
the same `len(TrimSpace(...)) > 0` test `IsServerPublicUrlConfigured` applies, so the two halves
cannot disagree about what "configured" means — leaving a *change* allowed (a deployment moves) and
an omitted key untouched.

`UpdateOidcProvider` uses the **map form** of `Updates`: GORM's struct form skips zero values, so
turning `allowProvisioning` / `linkByUsername` / `enabled` **off** would be silently ignored — and
each is a security toggle. For the same reason `OidcProvider.Enabled` carries **no `default:true`
tag**: GORM skips zero-value fields on Create when the column has a default, so a provider registered
with `enabled: false` came back enabled. Both are pinned by tests.

`GET /featureConfig` (public, unauthenticated) publishes `oidcProviders` as `{name, displayName}`
only — never the issuer, client id or anything else — and **always as `[]`, never `null`**, because
the generated Dart deserializer has no null guard and that payload is fetched before login.

### Provider cache

`oidc.NewProvider` makes a network round trip on every call, and the `*oidc.Provider` is what owns the
cached JWKS key set, so it is cached per provider id. A hit is valid only when the issuer URL **and
the row's `UpdatedAt`** match what was just read from the database — GORM bumps `UpdatedAt` on every
write and the row is read fresh per request, so an admin's edit invalidates the entry **by
construction, including on other replicas**. Do not add cross-process invalidation. Discovery uses a
fresh `context.Background()` with a timeout, never `r.Context()`: `NewProvider` stores the context's
`http.Client` inside the Provider for later JWKS fetches, so a cancelled request-scoped context would
poison the entry. `ClearProviderCacheForTests()` exists because the test DB reuses row ids.

### Off-spec providers

Twitch will not put `email` in the ID token from a scope alone — it needs a `claims` request
parameter. Since identity anchors on `sub` and provisioning falls back through `preferred_username`,
**Twitch works without email**; email is stored only for display on the Connected Accounts row. If it
is ever needed, the fix is one nullable `extraAuthParams` column fed through `oauth2.SetAuthURLParam`.

The claims decoder carries its own tolerant `email_verified` type (accepts `true` and `"true"`),
because go-oidc solves that internally with an unexported `stringAsBool`.

### Tests

`internal/oidc/` — `flow_test.go` is the security suite, run against a **real fake identity provider**
(`fake_idp_test.go`: an `httptest` server serving discovery, a JWKS, and genuinely RS256-signed ID
tokens via `golang-jwt`, already a dependency), so go-oidc is exercised rather than stubbed. Cases:
replayed / expired / cross-provider state, missing and foreign binding cookie, nonce mismatch, absent
nonce, missing `id_token`, a signature from a key not in the JWKS, a foreign audience, both happy
paths, and the IdP-error branch returning to the client that started the flow (with an unknown state
degrading to desktop). Plus `provision_test.go` (the resolution matrix, `TestSecondLoginUsesSubjectNotUsername`,
`TestProvisionedPasswordIsUnusable`, `TestUsernameMatchesClaim` and
`TestLinkByUsernameRequiresAnExactUsernameMatch`), `link_test.go` (the `app.account.update` gate, the API-key
refusal across all three scopes plus the unmarked-claims case, every callback failure landing a link
on the profile rather than the login page — with a login as the contrast — the
mobile JSON start and its denial/unknown-provider shapes, the `codeChallenge` split between login and
link, and the callback returning to the app's scheme), `exchange_test.go`, `store_test.go` (atomic
consume under 20 concurrent callers), `provider_cache_test.go`, `claims_test.go`. Elsewhere:
`handlers/oidc_providers_test.go` (permission gating, the secret never in a response body, the
boolean-flag regression), `repositories/oidc_providers_test.go` (the unique indexes, cascades, the
`enabled:false` regression), `services/oidc_feature_config_test.go` (the `[]`-not-`null` guard),
`commands/upsert_oidc_provider_command_test.go`, `permissions/legacy_test.go`.

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
  `OmittedColumns()` feeding the existing `Omit(...)` on that statement — *not* by copying
  the stored value onto the row. The distinction matters under concurrency: `existingSettings` is
  read before the transaction, so two admins each setting one lifetime and omitting the other would
  both write a full row and clobber each other's explicit value. A column that is never written
  cannot be clobbered. This is deliberately **not** a `SELECT ... FOR UPDATE` row lock — SQLite is a
  supported engine and rejects that syntax, there is no locking precedent anywhere in the codebase,
  and a lock would only protect these two fields while every other column on this full-object upsert
  stays last-write-wins. `ApplyOmittedValues` still runs, but only so the response body echoes the
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

**Access and refresh tokens are NOT interchangeable.** They are otherwise minted with identical
claims (same key/issuer/audience), so a refresh token would validate — and was previously usable — as
an API access credential. Each now carries a `structs.Claims.TokenType` (`"access"` / `"refresh"`).
The access paths reject refresh tokens (`middleware.UnifiedAuthMiddleware` JWT branch via
`isRefreshToken`, and `mcp.verifyToken`); the refresh paths require a refresh token
(`middleware.ValidateRefreshToken`, and the OAuth refresh grant guarded by the `refresh_tokens`
table). `Claims.IsRefreshToken()` also recognizes **legacy** tokens minted before this claim existed:
they have no `TokenType`, but only refresh tokens were ever given a `jti` (`RegisteredClaims.ID`), so
a typeless token WITH a jti is treated as refresh. That keeps already-issued legacy access tokens
working (no forced re-login) while still rejecting legacy refresh tokens immediately. Regression:
`middleware.TestUnifiedAuthMiddleware_RejectsRefreshTokenAsAccess`.

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

### Cookie `Secure` is a System Setting, off by default (`secureCookies`)

Whether the cookies the API sets carry `Secure` is the admin's call, per install:
`SystemSettings.SecureCookies` (`gorm:"default:false"`), edited on the desktop's System Settings
**Cookies** section. It is **off by default**, and that is deliberate:

- **Plain-http self-hosting is supported**, on a LAN or a bare IP (the same reason
  `isValidAbsoluteUrl` accepts `http`). A browser silently drops a `Secure` cookie over http, so
  setting it unconditionally **breaks login on those installs** — that has happened before, which
  is why this is opt-in rather than on for everyone. It is recommended for any install served over
  HTTPS.
- **`dev` still forces it on.** `BuildTokenCookies` uses `SameSite=None` there, and browsers reject a
  `None` cookie without `Secure`. Everywhere else SameSite stays `Strict` (session cookies) or `Lax`
  (the OIDC binding cookie) whatever the setting says.

**One resolver, every cookie.** `services.UseSecureCookies()` is the only thing any cookie builder
asks: `BuildTokenCookies` (jwt + refresh_token), `GetEmptyAccessTokenCookie` /
`GetEmptyRefreshTokenCookie` (the expiry cookies set on logout, refresh failure and account delete),
and `oidc.buildBindingCookie` / `clearBindingCookie`. Those six are every `http.SetCookie` in the API;
a new cookie must ask the same function. It reads the setting **live, with no cache** (the
`IsMcpEnabled` pattern), so a save applies to the next response, and a failed read falls back to the
default, off. The expiry cookies follow it too — a browser matches a deletion on name/domain/path,
so it is not needed for them to work over https, but it keeps every `Set-Cookie` consistent.

**The command field is a `*bool`**, on the same omitted-key machinery as the lifetimes (see "Session
lifetime" → `OmittedColumns` / `ApplyOmittedValues`). For a security toggle a plain `bool` is the
worst version of that hazard: any client that does not send the key — `services/import.go` builds a
four-field command — would silently turn Secure cookies back off.

**CodeQL still flags the default** as `go/insecure-cookie` (alerts #119-#122, medium). The alerts are
**known and accepted**: the flag is now the operator's to set, and on by default would break the
plain-http installs above. An **auto-detect** default (`r.TLS != nil || X-Forwarded-Proto == "https"`)
was considered and not adopted — behind a second TLS-terminating proxy, the container's nginx sets
`X-Forwarded-Proto $scheme` to `http`, so it reads wrong in a common deployment.

**Turning it on over plain http locks the web app out**, including the admin who would turn it back
off: their next token refresh sets cookies the browser drops. The desktop warns before saving when
its own page was loaded over `http:` — conservatively, since Chromium accepts `Secure` cookies from
`http://localhost`. Mobile is unaffected — it carries its tokens in the body, not in cookies.
**Recovery** goes through the API rather than the browser, with an admin's credential:

```bash
HOST=http://receipts.lan    # however the API is reached
# An API key with scope rw, owned by a user holding app.system-settings.read + .update.
# No key? Sign in for a token instead and send "Authorization: Bearer <jwt>" below:
#   curl -s -X POST "$HOST/api/login/?tokensInBody=true" \
#     -H 'Content-Type: application/json' -d '{"username":"admin","password":"..."}' | jq -r .jwt
AUTH='key.1.<id>.<secret>'
curl -s -H "Authorization: $AUTH" "$HOST/api/systemSettings/" \
  | jq '.secureCookies = false' \
  | curl -s -X PUT -H "Authorization: $AUTH" -H 'Content-Type: application/json' -d @- "$HOST/api/systemSettings/"
```

The PUT validates a whole settings body, which is why it round-trips the GET rather than sending the
one key; that also makes it read **and** write, so a key needs `rw` (see "API key scopes").

Tests: `services/auth_test.go` (`TestUseSecureCookiesFollowsTheSystemSetting`,
`TestCookiesAreSecureWhenTheSettingIsOn`, and `Secure=false` on the default path),
`oidc/flow_test.go` (`TestBindingCookieIsSecureOnlyWhenTheSettingIsOn` — the binding cookie, the
callback's session pair and the binding's expiry), `handlers/login_test.go`
(`TestLogin_SetCookieHeadersFollowTheSecureCookiesSetting`, the raw `Set-Cookie` headers — the wire
proof), the omitted/explicit round trips in `handlers/system_settings_handler_test.go`, and
`commands/upsert_system_settings_command_test.go`. `repositories.SetSecureCookiesForTests` flips it.
There is no e2e: turning the setting on globally would change every concurrently running spec's
session.

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

## App time zone

One admin-editable System Setting decides which calendar day an instant falls on, what "today" and
"this month" are, and therefore what every period means. Before it, the answer was split between the
server process's zone (report presets, `WITHIN_CURRENT_MONTH`, system-task day filters,
`{{generatedAt}}`), hard-coded UTC (report buckets and labels, the PDF footer) and the browser — so a
"Last Month" report run on a UTC container left out a receipt added at 9PM on Sep 30 Eastern, which
is 01:00 UTC on Oct 1.

**The rules (root `CLAUDE.md` → "App Time Zone" holds the same table):**

| Value | Kind | Zone its day is read in |
|---|---|---|
| `receipts.date`, custom field `date_value` | calendar day | UTC — stored as midnight UTC |
| `created_at`, `resolved_date`, task `started_at` / `ended_at` | instant | app zone |
| "today", "this month", report presets, `@currentYear`, `{{generatedAt}}`, the PDF footer | clock | app zone |

Unchanged: log timestamps, token expiry, `@every` schedules and sort order.

- **The setting.** `SystemSettings.TimeZone` (IANA name, `gorm:"default:UTC"`, so an existing install
  behaves exactly as before). `UpsertSystemSettingsCommand.TimeZone` is a `*string` and joins the
  pointer-field machinery of the lifetimes (see "Session lifetime"): `nil` ⇒ the column is in
  `OmittedColumns()` and dropped from the UPDATE, and `ApplyOmittedValues` echoes the
  stored value. A sent value must load with `time.LoadLocation`; `""` and `"Local"` are refused too
  (one reads as UTC, the other as the process zone). Failures are a 400 under `timeZone`.
  `main.go` imports `_ "time/tzdata"`, so a name that validates resolves identically on an image
  without `/usr/share/zoneinfo` instead of silently becoming UTC.
- **The resolver.** `repositories.GetAppLocation()` reads the setting **live, with no cache** (the
  `IsMcpEnabled` / `GetRefreshTokenLifetime` pattern), so a save applies on the next request.
  `BaseRepository.AppLocation()` is the same read through the repository's transaction.
  `ResolveAppLocation(name)` maps `""`, `"Local"` or an unloadable name to UTC and logs a bad name
  once; it caches parsed zones by **name**, never the setting. It lives in `repositories` so filter
  builders can call it without a parameter threaded through every caller; services use it too.
  **`AppData.timeZone`** carries the *resolved* name (`"UTC"`, never `""`).
- **Day filters: `BaseRepository.BuildDayFilterQuery(query, value, op, column, columnLoc, appLoc)`.**
  The one builder for every date filter — receipts (`date` with `time.UTC`, `resolved_date` /
  `created_at` with the app zone) and system tasks (`started_at` / `ended_at`, app zone). It replaced
  `applyTimestampDayFilter` / `startOfDayValue` and the receipt builder's raw `date = ?` comparisons.
  - Ranges: `EQUALS d` ⇒ `[start(d), start(d+1))`, `GREATER_THAN d` ⇒ `>= start(d+1)`, `LESS_THAN d`
    ⇒ `< start(d)`, `BETWEEN [a,b]` ⇒ `[start(a), start(b+1))`, where `start` is midnight in
    `columnLoc`. Adding a calendar day (not 24h) keeps DST days 23 or 25 hours long.
    `WITHIN_CURRENT_MONTH` ⇒ month start through the end of today, **today being the app zone's**
    even for the UTC `date` column (`currentMonthBounds`).
  - Accepted values: a bare `yyyy-MM-dd` (what clients send), an RFC 3339 string read as the day
    **as written** (its own offset, never converted — this is what lands an older mobile build's
    fake-`Z` local wall time on the day the user picked), or a `time.Time` read in its own location.
    `CalendarDayOf` / `StartOfCalendarDay` are the exported pieces.
  - Every bound is converted to **UTC** before it reaches the driver, so it compares correctly
    against UTC-stored timestamps on every engine (SQLite compares them as text).
  - Every unwrap is comma-ok: a wrong-typed or unparseable value adds no predicate. The receipt
    builder's bare `.(string)` / `.([]interface{})` assertions on the three date fields, which
    panicked on a malformed body, are gone.
  - `BuildFilterQuery`'s own `WITHIN_CURRENT_MONTH` branch (any other caller) uses the app zone too.
- **Calendar dates on write.** `NormalizeReceiptCalendarDates` stores `Receipt.Date` and every custom
  field `DateValue` as midnight UTC of the as-written day (`utils.CalendarDayUTC`). It runs in
  `ReceiptRepository.CreateReceipt` and `UpdateReceipt`, right after `ToReceipt`, which every writer
  reaches: the form, single-call create, quick scan, email and AI. Duplicate copies stored values
  verbatim. **Existing rows are not migrated** — they are read by their UTC day.
- **Reports.** `buildModel` resolves the zone once, from the settings row it already reads for the
  currency format. `applyPeriod` works on `now.In(loc)` and writes the BETWEEN as **bare calendar
  days**, so the day builder decides per column what instants a day covers — the report, its
  drill-in and the receipts table cannot disagree. `{{generatedAt}}` and the period label are in the
  app zone; `MetaInput.Location` carries it to the HTML footer (`2006-01-02 15:04:05 MST`).
  `ReportDataService.Location` passes it to `receiptsource.New(customFields, loc)`, whose catalog
  declares each date field's `Location` (UTC for `date` and DATE custom fields, the app zone for
  `created_at` / `resolved_date`); `buildDimensions` copies it onto `render.Dimension`, and the
  engine copies a label column's onto `ColumnDescriptor.Location`. Bucket identity stays the
  canonical UTC instant — the zone only changes the derived period strings and how a day is named.
- **Elsewhere.** `receipt_csv.go` prints Added At / Resolved Date in the app zone and the receipt
  Date in UTC; `@currentYear` in AI prompts is the app zone's year.
- **Contract.** `timeZone` is a plain `type: string` (not an enum: tzdata changes, and a closed Dart
  enum would fail the whole AppData payload) on `SystemSettings`, `UpsertSystemSettingsCommand` and
  `AppData`, and in no `required` list. `commands/time_zone_contract_test.go` pins it.
- **Clients.** The desktop reads `AppData.timeZone`, sends filter dates as `yyyy-MM-dd`, sends a
  receipt Date as midnight UTC of the picked day, and displays instants in the app zone. Mobile only
  carries the regenerated client: its fake-`Z` filters and dates already land on the right day under
  the as-written rule; displaying in the app zone is a follow-up.

**Tests:** `repositories/app_time_zone_test.go` (resolver default / configured / invalid / live after
a real settings save; `BuildDayFilterQuery` bounds per operation and zone via `DryRun`, a DST day,
every wire form equal to the bare day, wrong-typed values, `currentMonthBounds`; the receipt filter
reading each column in its own zone, and the reported bug as a row count; write normalization on
create and update), `services/report_time_zone_test.go` (the reported bug end to end — UTC excludes,
New York includes on the next `buildModel`, UTC excludes again — plus calendar Dates unaffected,
presets / label / `{{generatedAt}}` per zone, buckets and footer, AppData), the CSV case in
`services/receipt_csv_service_test.go`, `handlers/system_settings_time_zone_test.go` (persist,
omitted key, 400), `commands/upsert_system_settings_command_test.go` (validation, omitted column),
and the pure `receiptsource` / `render` cases, each with a `mutation-check.sh` entry.

## Temporary file retention & cleanup

`temp/` holds the files the ingest pipelines work from. Everything that writes
there: the **quick-scan upload** (`handlers/receipts.go`, one per file in a
multi-file scan), the **email attachment and its `image-` OCR copy**
(`wranglerasynq/email.go`), the **email body image**, short-lived scratch from
`MagicFillFromImage` / `ReadImageWithEasyOcr` / `ConvertPdfToJpg` (each with a
`defer os.Remove`), and the **`DebugOcr` dumps** from `services/ocr.go`
`writeDebuggingFiles`, which are never removed by the code that writes them.

**`HandleTempFileCleanUpTask` (`wranglerasynq/temp_file_cleanup_handler.go`) is
the only thing that reclaims them.** It runs `@every 1h` on
`models.SystemCleanUpQueue`, registered from **`StartSystemCleanUpTasks`**, and
that placement is load-bearing: it used to live in `StartEmailPolling`, which
`main.go` calls only when `EmailPollingInterval > 0 && ReceiptProcessingSettingsId != nil`,
so an install that never configured email polling ran **no temp cleanup at all** —
including for quick-scan uploads, which have nothing to do with email.
`StartEmailPolling` also re-runs on every polling-interval change while
`scheduler.Register` adds an entry without removing the previous one, so the sweep
used to accumulate duplicate crons.

### The classification, and its precedence

Ordered — the order is the behaviour, not a formatting choice:

| # | Case | Action |
|---|---|---|
| 1 | any referencing task is still actionable (pending/active/scheduled/retry/aggregating) | keep |
| 2 | else any referencing task is `Archived` | keep until older than retention |
| 3 | else every referencing task is `Completed` | remove now |
| 4 | unreferenced | keep until aged, **and only when the reference map is complete** |

**Rule 2 must be tested before rule 3.** One email attachment fans out to a
sibling task per `groupSettingsId`, so a file can be referenced by both a
succeeded task and a permanently failed one; releasing it because something
succeeded is precisely the bug this sweeper replaces.

**`Archived` can never mean deletable** — it is the state that makes an activity
rerunnable (`SetActivityFlags`), and the previous cleanup released on it.

Age comes from **file mtime**, not task timestamps: uniform across queues, still
correct once the task ages out of Redis, and it doubles as the grace window for
the race where a file is written just before its task becomes visible.

### Two invariants that keep the orphan branch safe

The failure mode inverted with the rewrite. Previously a task the scan missed
meant "we fail to delete" (harmless); now it means the file looks unreferenced
and is deleted once aged. So:

1. **Every inspector listing pages.** Asynq defaults to **30 per page**
   (`defaultPageSize` in its `inspector.go`), and the old `getTaskInfo` passed no
   `ListOption` at all. `forEachTaskPage` loops with `asynq.PageSize`/`asynq.Page`
   until a short page, handing each page to a callback and then dropping it —
   the sweep wants each payload's paths and state, while the payloads are the
   large part (`EmailProcessTaskPayload` carries the body twice, copied per
   `groupSettingsId`). Accumulating a whole listing first held the archived email
   set — 10,000 tasks, kept 90 days — in memory at once inside an hourly job.
   Building the map incrementally makes the abort path load-bearing: a listing
   error must discard what it already has, or the unreached files look
   unreferenced and the orphan branch deletes them once aged.
2. **Any gap stands the orphan branch down for that run.** A listing error other
   than `ErrQueueNotFound` aborts the whole sweep with zero deletions; a payload
   that will not unmarshal, **or a listing that runs out the `tempFileMaxListPages`
   cap**, drops a `referencesComplete` flag that suppresses **only** rule 4. Rules
   1-3 rest on a state actually observed, so they stay live.
   (The old `buildAttachmentMap` returned the error, which would have aborted every
   future sweep after one malformed payload.)

   The cap is the easy one to overlook, because it reads like a loop guard rather
   than a correctness rule: running out of pages is the one case that silently skips
   **real** tasks, so `forEachTaskPage` reports it (`(scannedAll bool, err error)`)
   and logs it. Reaching it takes ~100,000 tasks in one (queue, state) and is close
   to unreachable in practice — asynq caps the archived set at 10,000 — but the
   failure direction is deletion of a user's file, so it fails the safe way: rule 4
   simply reclaims nothing that run.

### Rule 3 is live on the email queue alone

Asynq drops a successfully-processed task the instant it succeeds unless the
enqueue sets `asynq.Retention`, in which case the task stays in Redis for that
long in the **completed** set and the janitor reaps it afterwards. Nothing here
set it at first, so `ListCompletedTasks` was always empty and rule 3 was dead
code — which is how the *old* cleanup came to be inverted end to end: its
`Completed` branch could never fire, so the only files it ever deleted were the
archived ones it had to keep.

`enqueueOptions` (`task_enqueue.go`) now attaches
`asynq.Retention(completedTaskRetention)` — **6 hours** — to
`EmailReceiptProcessingQueue`, so a succeeded email upload's attachment and its
`image-` OCR copy are released on the next hourly sweep instead of waiting out
the user-configured window. The duration is a handful of sweep intervals rather
than a day because `EmailProcessTaskPayload` carries the email body **twice**
(`Metadata.Body` + `Metadata.BodyHtml`) and `email.go` copies the metadata per
`groupSettingsId`, so one message consumed by N groups with M attachments retains
N×M copies of a 50-500 KB body. Overshooting degrades benignly — the files just
fall back to waiting out the window.

**It is deliberately NOT on the other two queues**, and "the ingest queues" is the
wrong mental model:

- **Quick scan** already deletes its own file on success — `ReceiptService.QuickScan`
  ends with `os.Remove(params.TempPath)`, reached only after the create-receipt
  transaction commits, and every failure path returns earlier and keeps the file.
  It is 1:1 task-to-file, so by the time such a task completes there is nothing on
  disk for the sweep to classify. Retention there is pure Redis cost.
- **`EmailPollingQueue`** owns no temp files at all.

**Do not "fix" the email case by deleting on success inside
`HandleEmailProcessTask`** instead: one `TempFilePath` fans out to N sibling tasks
and the first to succeed would break the rest. The sweep owns the decision
precisely because it can see every referencing task at once, and retention is what
lets it see them.

**No janitor configuration is needed.** `asynq.Server.Start` starts the janitor
unconditionally and it runs `DeleteExpiredCompletedTasks` per queue every 8s by
default; `asynq_server.go` sets only `Concurrency` and `Queues`, so it takes that
default.

**Retention changed two things outside the sweeper**, both now guarded — see
"Activity source files" below for `rerunnableState` (the rerun endpoint's state
check, which asynq does *not* enforce for us) and `offersSourceFile` (a succeeded
task stops advertising its upload).

### `TempFileRetentionHours`

A System Setting, default **720** (30 days), bounds **24-8760** (1 year), `0`
means unset. It follows the `RefreshTokenValidForHours` machinery exactly — see
"Session lifetime" above for the `*int` command field, `OmittedColumns`
and why an omitted key is *dropped from the UPDATE* rather than copied. Both helpers
cover every pointer-backed field — the lifetimes, this retention window, the server
public URL, the time zone and the secure-cookie toggle — which is why neither is named
for the lifetimes they started out holding.

`tempFileRetention()` is the read-side clamp, the same shape as
`repositories.pdfRasterizationDpi`: **the clamp, not the validator, is the real
safety net**, because it also covers a value that predates the bounds or a
settings row that cannot be read at all.

### Upgrade path

`EmailReceiptImageCleanupQueue` stays declared in `models/queue_names.go` —
persisted `TaskQueueConfigurations` reference it, `QueueName.Value()` whitelists
it, and `UpsertSystemSettingsCommand.Validate` requires a configuration for every
name. The retired `EmailProcessImageCleanUp` **task type** also stays declared and
stays routed in `BuildMux`: the type is a string in Redis, so tasks the old cron
already enqueued outlive the deploy and would otherwise fail as unregistered.
`DeleteAllScheduledTasks` alone is insufficient — it only touches the *scheduled*
set, so `retireEmailReceiptImageCleanupQueue` drains **pending** too.

**The symmetric case — a persisted list MISSING a name — is handled in
`repositories/system_settings.go`, and it used to 400 every settings save.** An
install that last saved before `SystemCleanUpQueue` existed has four
`task_queue_configuration` rows. `GetSystemSettings` substituted defaults only when
the list was **empty**, so it returned four; the desktop form builds its rows from
that response rather than from the `QueueName` enum, so it submitted four; and
`UpsertSystemSettingsCommand.Validate` requires one per queue name, so it rejected
the **whole** body with a 400 — `tempFileRetentionHours` along with everything else.
`asynq_server.go` falls back to the default priority for a missing configuration, so
the task server ran fine and nothing surfaced the gap.

`withMissingQueueConfigurations` now fills any absent name from
`GetDefaultQueueConfigurationMap()` on read (persisted priorities win; ordering
follows `GetQueueNames()`; a row naming a queue this build does not know is dropped,
since `QueueName.Value()` would reject it on the way back down), and
`UpdateSystemSettings`' per-name branch **inserts** a submitted configuration that
has no row instead of no-op'ing its `UPDATE`. Both halves are needed: without the
insert the DB never heals and the submitted priority is silently discarded on every
save. Which names exist is read up front rather than inferred from `RowsAffected`,
which a no-change `UPDATE` reports as `0` on MySQL. The bug predates this feature —
adding a setting worth changing is what made it reachable.

### Testing

The sweeper is deliberately three seams so the rules stay Redis-free:
`listTempFileReferences` (Redis, injectable listers), `classifyTempFile` (**pure**
— no Redis, no filesystem) and `sweepTempDirectory` (`t.TempDir()` + `os.Chtimes`).
`wranglerasynq/main_test.go` starts no Redis and must not have to. Keep it that
way — the precedence table above is only cheap to cover exhaustively because the
classifier is pure.

## Activity source files (preview / download)

A failed quick scan or email upload keeps its image (above), so the user can get
it back and enter the receipt by hand. Two endpoints on the system task router,
both gated on **`group.activities.read`** for the task row's own `GroupId`:

- `GET /systemTask/{id}/sourceFile` — JSON `{ name, encodedImage }`.
- `GET /systemTask/{id}/sourceFile/download` — the original bytes.

**Preview goes through `FileRepository.BuildDisplayImageString`, like every other
browser-facing image response.** That helper is the one canonical
`raw bytes → data URI a browser can render in an <img>` transform: PDFs rasterize,
HEIC transcodes, everything else passes through. Reading the bytes stays the
caller's job — `data/` and `temp/` are separate trust boundaries with separate
containment checks, and `ConvertToJpg` has no file at all — but the *encode* half
is shared, so no surface can get it half-right.

**This shipped wrong and the failure was silent**, which is worth knowing before
touching any of it. The preview originally called `BuildEncodedImageString` on the
raw upload. That looks harmless until you read `GetFileType` (`files.go`), which
relabels PDF bytes as `image/jpeg` **without converting them** — so a quick-scanned
PDF previewed as `data:image/jpeg;base64,JVBERi0…`: a well-formed data URI carrying
`%PDF`, which the browser silently fails to decode into a broken-image icon. No
error, no suspicious mime, nothing in the logs. Quick scan was the exposed case
because `taskSourceFiles.Preview` is always empty for `QUICK_SCAN`, so the handler
always falls back to the raw upload; email escaped it only because
`ImageForOcrPath` was already converted at enqueue.

**Preview still prefers `ImageForOcrPath` when the payload has one; download always
serves `TempFilePath`.** That preference is an optimization *within* the canonical
path, not a divergence: `GetBytesFromImageBytes` on a multi-page PDF runs
`ConvertPdfToJpg` at the configured `PdfDpi` and concatenates every page into one
tall JPEG — seconds of CPU inside a request — and for email that conversion already
exists on disk, passing through untouched. Quick scan has no such copy and pays the
conversion, exactly as the receipt form does. Download is the *original*, under its
original name, because that is the file the user recognises.

**The path comes out of a Redis payload**, so it goes through
`FileRepository.AssertWithinTempDirectory` before anything opens it — see
"Filesystem Access & Path-Traversal Safety" **above**. `resolveActivityFlags`
applies the same rule via `sourcePathUsable`, so the flag and the endpoint cannot
disagree about a path; without that, a flag resolved from a bare `FileExists`
would render a preview control whose request can only fail, and the stat would
double as an existence oracle for arbitrary server paths.

**Never `utils.ReadFile` here.** It returns `(nil, nil)` on any read error, which
would serve an empty image as a success.

`Content-Disposition` is **sanitized and then formatted by `mime.FormatMediaType`**,
never concatenated. `utils.SanitizeFileName` is `filepath.Base`, so it reduces an
email attachment's MIME-header name to a basename but leaves a `"` in it — and a
raw `filename="`+name+`"` then closes the quoted value early, so a client reading
the header back gets `receipt` out of `receipt"final.pdf` and loses the extension.
`FormatMediaType` escapes it, and switches to the RFC 5987 `filename*` form for a
non-ASCII name; the desktop parser reads both (`desktop/src/utils/file.ts`), which
is why changing one side without the other breaks every accented attachment name.

The other five `Content-Disposition` writers are unchanged and are **not** parsed
client-side: `DownloadReceiptImage` is unquoted *and* unsanitized (a latent bug),
and the two report downloads quote without sanitizing. Copy the rest of
`DownloadReceiptImage`'s shape (`ResponseType: ""` so `http.ServeFile` owns
Content-Type, header before serving, `return 0, nil` once streaming begins), not
that line.

### The two flags, and why they differ

- **`hasSourceFile`** — the task type expects an upload **and** it exists. True in
  **any** task state.
- **`canBeRestarted`** — the task is **`Archived`** and every file a rerun reads
  exists.

`canBeRestarted` stays pinned to `Archived`, and **enforcing that is ours, not
asynq's.** `Inspector.RunTask` does *not* refuse a task by state: its Lua script
special-cases only `active` (-1) and `pending` (-2), and every other state —
**`completed` included** — falls into the branch that `ZREM`s the id from its set
and `LPUSH`es it back onto pending (`asynq@v0.25.1/internal/rdb/inspect.go`). That
was harmless only while completed tasks never existed; with the email queue's
retention, a rerun of a succeeded upload would pass `GetTaskInfo`, pass
`RerunSourceFilesPresent` (the files are still on disk) and **create a duplicate
receipt**. So `rerunnableState` (`source_file.go`) is the rule, `CanRerunTask` is
its exported form, and `RerunActivity` refuses anything else with a **400** —
matching exactly what `canBeRestarted` advertises.

`hasSourceFile` is deliberately **not** gated on `Archived`: asynq backs its
retries off exponentially, so a task takes minutes to archive and the user should
not watch an activity sit at FAILED with no way to retrieve their image. It *is*
gated on the task not having **succeeded** (`offersSourceFile`), which is a
different rule: a retained `Completed` email task still has both files on disk
until the next sweep, and without this the preview/download controls would appear
on a succeeded activity for up to an hour and then 404. `ResolveSystemTaskSourceFile`
applies the same predicate — the two read the same task state, so they share one
function rather than two conditions that can drift.

**Email needs both files for a rerun** — `HandleEmailProcessTask` reads
`TempFilePath` for the `FileData` it persists and hands `ImageForOcrPath` to the
OCR/vision pipeline — while `hasSourceFile` keys on `TempFilePath` alone. A
**body-only email has neither and stays rerunnable**, which is why "expects a
file" is tracked apart from "has a file".

### Authorization order, and why it is not the payload's job

Both endpoints, and `RerunActivity`, answer **every** authorization question before
any file-specific one. The sequence is: load the task row → gate on its `GroupId`
with `group.activities.read` → check actor visibility → only then resolve the file
and emit 400/404.

That ordering is the fix for two things.

**Member isolation applies per task, not just per list.** `GetActivitiesForGroups`
filters rows in SQL through `ActivityVisibilityResolver`, so inside an isolated
group a plain member never *sees* a hidden co-member's activity — but they hold
`group.activities.read` there, so naming that activity's id directly returned the
hidden member's upload. `enforceActivityActorVisible` applies
`applyActivityVisibilityDisjunction`'s rule to one row, through the exported
`PermissionService.UserVisibleInGroup`. A nil `RanByUserId` is a system action and
stays visible, exactly as the SQL clause has it — which is **every EMAIL_UPLOAD**,
since email is polled rather than run by a user.

**Resolving the file first leaked its existence.** The group used to come from the
asynq payload via `ResolveActivityGroupId`, which meant a Redis round trip — and a
400/404 written from it — *before* `HandleRequest` ran the gate, so an
unauthorized caller learned whether a task existed, whether its type had a source
file, and whether that file was still on disk. `models.SystemTask` carries
`GroupId` and `RanByUserId` as ordinary columns, populated for both task types, so
nothing about the gate needs Redis. A nil `GroupId` **fails closed**. The
authorization denial and the "no such file" answer are now deliberately
indistinguishable, and the gate is reachable in a handler test without a Redis
instance — which is how it is covered, since the source-file endpoints otherwise
cannot be.

**Two smaller oracles sit a layer above that one, and both are closed the same way.**
A task id naming **no row** used to surface `gorm.ErrRecordNotFound` as a **500**
while a denied one answered 403, so the status distinguished "exists" from "does not
exist" for arbitrary ids; `loadSystemTaskForSourceFile` now maps that one error — and
only that one — to the same 403 and the same `activityAccessDeniedMessage`. And
`RerunActivity`'s "only a quick scan or email upload can be rerun" check ran *before*
the handler was built, so it answered ahead of the gate and told an unauthorized
caller the activity's type; it now runs inside `HandlerFunction`, after
`enforceActivityActorVisible`.

**The body is half the answer, and it is written by three different places.** Closing
the status oracle with wording of this feature's own just moved it: the group gate is
`HandleRequest`, which writes `unauthorizedEntityMessage`
(`handlers/generic_handler.go`), while the row lookup and the actor check write their
own. So `activityAccessDeniedMessage` **is** `unauthorizedEntityMessage` — not a copy
of its text, the constant itself, since a divergence is only ever a bug. Any handler
that denies for a reason of its own has to reuse it. Assert such a fix across **every**
denial path at once — the unknown id against the hidden actor *and* against the
non-member — since one pairing passes while another still leaks.

### Hydration traps

- **`AssociatedSystemTaskId == nil` is NOT a valid "top level" test.**
  `email_process_handler.go` chains every `EMAIL_UPLOAD` task under its
  `EMAIL_READ` parent, so that rule reads false for exactly the rows this feature
  exists for. `SetSystemTaskHasSourceFile` hydrates the rows the page returned and
  never recurses into `ChildSystemTasks`.
- **`GetPagedActivities` must `Omit` every computed field.** GORM infers the
  SELECT list from the `structs.Activity` destination, so it is
  `.Omit("can_be_restarted", "has_source_file")`. `asynq_task_id` is a real column
  and needs none — it is selected so the flags resolve without a per-row query.
- **Both hydrators live in `wranglerasynq`, not `repositories`** —
  `wranglerasynq` imports `repositories`, so a loader there needing
  `GetAsynqInspector` is an import cycle. The precedent to cite is
  `models.ReportTemplate.AllowedActions` (a `gorm:"-"` field the list handler
  fills per row), not `LoadSettingsProjections` (pure-DB, lives in the repo).
- **The system-task flag is resolved per caller.** That table is app-scoped
  (`app.system-tasks.read`) and lists groups the caller may not belong to, so the
  handler passes a `canReadGroup` predicate; without it the table would offer
  buttons that 403.
- **Nothing here fails a request.** A task that aged out of Redis, an unreadable
  payload or a failed stat leaves both flags false; the first *other* lookup error
  logs once and stops, rather than paying a dial timeout per remaining row.

**A known duplication, accepted:** `CreateSystemTasksFromMetadata` creates up to
two rows of the same type sharing one `AsynqTaskId` when a fallback processing
setting ran, so such a failure already shows two Rerun buttons and now shows two
button pairs. Deduping would change rerun's existing behaviour.

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

## System task filtering

`GET`-by-POST `/systemTask/getPagedSystemTasks` accepts a `filter` on
`GetSystemTaskCommand`: a `SystemTaskPagedRequestFilter` of four `PagedRequestField`s — `type`,
`ranBy`, `startedAt`, `endedAt` — reusing the receipt filter's `FilterOperation` enum and wire shape,
so the desktop renders both dialogs from one row component.

- **`BaseRepository.BuildFilterQuery`** is the shared operation→WHERE translator, promoted from
  `ReceiptRepository.buildFilterQuery` so the receipt filter and the system task filter cannot drift
  on what an operation means. `fieldName` is interpolated into the clause, so it MUST stay a
  hardcoded column literal supplied by the caller — never a request value.
- **Predicates are applied before `Count`** (`buildSystemTaskFilterQuery`, called from
  `GetPagedSystemTasks`), mirroring `GetPagedActivities`' visibility disjunction, so `totalCount`
  describes the filtered set rather than the whole table.
- **Every unwrap is a comma-ok assertion.** Values arrive as `interface{}` off the request body; a
  wrong-typed value is treated as "field not set". This is deliberately unlike the receipt builder's
  bare `.(string)` / `.([]interface{})` assertions, which panic on a malformed body — and it is why
  there is no `initSystemTaskFilterValues` analogue seeding non-nil defaults.
- **`ranBy` does not go through the shared helper**, because `ran_by_user_id` is nullable and the
  rows with no user are exactly the ones the table labels "System". The desktop submits
  `repositories.SystemRanByUserId` (`-1`, negative so it can never collide with a real id — the
  `OWN_PAID_RECEIPTS_OPTION_ID` convention) for those, and `applyRanByFilter` splits the sentinel out
  of the id list: sentinel only ⇒ `IS NULL`, ids only ⇒ `IN ?`, both ⇒
  `IS NULL OR ran_by_user_id IN ?`. Ids decode as `float64` through `encoding/json`, hence `toInt64`.
- **`started_at` / `ended_at` are compared as whole calendar days in the app time zone**, through
  the shared `BaseRepository.BuildDayFilterQuery` (see "App time zone" below), which the receipt
  filter uses too. These carry a time of day, so a raw comparison to the datepicker's midnight would
  make `EQUALS` never match and `BETWEEN` drop everything after midnight on the end day.
  `EQUALS d` ⇒ `[start(d), start(d+1))`, `GREATER_THAN d` ⇒ `>= start(d+1)`, `LESS_THAN d` ⇒
  `< start(d)`, `BETWEEN [a,b]` ⇒ `[start(a), start(b+1))`, `WITHIN_CURRENT_MONTH` ⇒ month start
  through the end of the app zone's today.
  - **The wire format is a bare calendar day, `yyyy-MM-dd`**, which the desktop sends
    (`toDateWireFilter`). An RFC 3339 value is still accepted and read as the day **as written**.
    The day used to resolve in the server process's zone (`time.Local`); it is now the app time
    zone setting, so the server's own zone plays no part.
  - **`ended_at` is nullable, so any filter on it excludes tasks that are still running.** That is
    the correct reading of "ended before X"; it is not special-cased.
- **The two child-only types stay excluded.** `filteredSystemTaskTypes` (`CHAT_COMPLETION`,
  `OCR_PROCESSING`) is applied before the filter, so a `type` filter narrows within the top-level
  rows. The desktop's Type picker omits them for the same reason; keep
  `CHILD_ONLY_SYSTEM_TASK_TYPES` in `desktop/src/constants/system-task-type-options.ts` in sync with
  the Go list, which `TestGetPagedSystemTasksExcludesChildTaskTypes` pins.
- **`RECEIPT_UPLOADED` is listed only when it has no parent.** Manual creates and duplicates record
  a parentless row, and the admin page is the only place those surface. Quick scan and email record
  theirs as a child of the QUICK_SCAN / EMAIL_UPLOAD task, and the table already shows it by
  expanding that row, so a `Not(type = RECEIPT_UPLOADED AND associated_system_task_id IS NOT NULL)`
  keeps them from being listed twice. A parent id is the right test for this type alone — an
  EMAIL_UPLOAD also has a parent (see "Hydration traps") and is still a top-level row. The Type
  picker offers "Receipt Uploaded", and it returns only the parentless rows.
  - **`CreateReceipt` records its own failure only when `createSystemTask` is true.** That flag
    means "this call owns the upload task": the manual form passes `true`, so a failed create now
    shows as a FAILED Receipt Uploaded row. Quick scan and email pass `false` because
    `CreateReceiptUploadedSystemTask` already records the failure as a child. The guard used to be
    inverted, which recorded nothing for a manual failure and a second, parentless row for every
    quick scan and email failure. Rows of that second kind written before the fix still list, as
    Failed; email's carry `ran_by_user_id = 0` rather than NULL, so the "System" ran-by filter does
    not match them. Pinned by `TestCreateReceiptRecordsFailedUploadTaskWhenItOwnsTheTask` and
    `TestCreateReceiptLeavesFailedUploadTaskToTheCaller`.
  - **`DuplicateReceipt` is all or nothing.** The new receipt and its image files are created in
    one `db.Transaction`. `copyDuplicateImages` returns every path it wrote or tried to write, and
    `DuplicateReceipt` removes them whenever the transaction returns an error, so a failed commit is
    covered as well as a failed copy. Each path is tracked **before** its write, because a write
    that fails partway can leave a truncated file. A failed duplicate therefore leaves no receipt, no
    `FileData` and no files. Its task records FAILED with the error and no receipt link:
    `ReceiptId` / `AssociatedEntityId` are set only once the transaction commits. Three details are
    load-bearing:
    - **`err` is a named result**, read by the deferred `CreateSystemTaskFromError`. It used to be a
      plain local that the image loop shadowed with `:=`, so a failed copy returned a 500 while the
      task said SUCCEEDED.
    - **The copy runs on the transaction** (`NewFileRepository(tx)`): `BuildFilePath` looks the new
      receipt's group up by id, and the row is not visible on another connection yet.
    - **It copies the stored bytes** (`utils.ReadDataFile`), never `GetBytesForFileData`, which is
      the display conversion. That rasterized a PDF or transcoded HEIC to JPEG and saved it under the
      original name and file type, so the duplicate's "receipt.pdf" downloaded as a JPEG.
    Pinned by `services/duplicate_receipt_test.go`.

The handler is unchanged: `GetSystemTasks` already passes the whole command through and the
`app.system-tasks.read` gate still applies. Note this endpoint has **no** member-isolation filtering
(unlike `GetPagedActivities`) — it is admin-only by that permission.

**Tests:** the bound values live in `repositories/app_time_zone_test.go`
(`TestBuildDayFilterQuery*`), asserted off a `DryRun` statement rather than row counts, because the
test DB is SQLite: it stores timestamps as text and compares them lexically, where a raw bound
happens to sort *after* every row on the same day (`" " < "T"`), so a row count alone cannot tell
the widened path from the broken one. The behavioural cases in `repositories/system_task_test.go` (`...FiltersByType`, `...FiltersByRanBy`,
`...BetweenIncludesTasksLateOnTheEndDay`, `...EqualsMatchesTheWholeDay`,
`...LessThanOnEndedAtExcludesRunningTasks`, `...IgnoresWrongTypedFilterValues`) all assert
`totalCount` alongside the rows. Also `commands/get_system_task_command_test.go` (the wire keys, the
`float64` ids, and an absent `filter` staying zero-valued).

## Receipt update snapshots (`RECEIPT_UPDATED`)

`ReceiptRepository.UpdateReceipt` records a `RECEIPT_UPDATED` system task whose description is
`{"before": "<receipt JSON>", "after": "<receipt JSON>", "version": 2}`. Each side is
`Receipt.ToString()`, so the value is **double-encoded**. The desktop parses it itself and renders a
side-by-side diff (see `desktop/CLAUDE.md` → "Receipt update diff"), so keep the format stable: an old
row must still parse.

**The format is versioned** by `repositories.ReceiptUpdateDescriptionVersion`. A row with **no
`version` key is version 1**: its "before" is incomplete (below). **Version 2** rows have a complete
"before". Bump the constant whenever the snapshot format changes; the repository test pins the
literal `2`, and the desktop's `COMPLETE_RECEIPT_UPDATE_VERSION` must agree.

**Both sides are loaded with `GetFullyLoadedReceiptById`.** `before` used to be serialized from
`currentReceipt`, which only preloads `clause.Associations`, one level deep. Loaded that way, the
snapshot has no item categories, tags or linked items and no custom field definitions, and it lists
linked items as top-level items (`FilterLinkedItemsFromReceiptItems` never ran). Every update would
then diff as a change to all of those. `currentReceipt` itself is left alone, because the update
relies on it (`BeforeUpdateReceipt`, `Model(&currentReceipt)`). The snapshot is one extra read.
Pinned by `TestUpdateReceiptSystemTaskSnapshotsAreLoadedToTheSameDepth`, which fails with the old
loader.

**Version 1 rows are rebuilt on read, never rewritten.** `GetSystemTasks` passes each page through
`SystemTaskService.UpcastReceiptUpdateDescriptions` (`services/receipt_update_history.go`). For every
successful version 1 row it swaps the incomplete "before" for the **nearest earlier complete copy of
the same receipt**: the previous `RECEIPT_UPDATED` row's "after", or the `RECEIPT_UPLOADED` copy
stored when the receipt was created. The form, Quick Scan, email and duplicate all store one. It adds
`"beforeSource": {type, systemTaskId, recordedAt}` and keeps `version: 1`.
- **One query per page** fetches the candidates, below the page's highest version 1 id.
- A receipt is matched on `receipt_id`, or on `associated_entity_type = RECEIPT` plus its id.
  Quick Scan and email uploads carry only the former; old rows may carry only the latter.
- A candidate is used only if it parses and its `id` is that receipt's.
- A row with no usable candidate is left as stored.
- Why it stays version 1: a change that writes no update row, such as a bulk status change, falls
  inside a rebuilt comparison, and a version 2 row's own "before" would not have it.
- Why on read: `resultDescription` is a free-form string, so there is no swagger change or client
  regeneration. Mobile never reads it. And the stored history is never rewritten.
- Only the System Tasks listing does this; `getPagedActivities` doesn't display descriptions.

Tests: `services/receipt_update_history_test.go` covers the matching rules, and
`handlers/system_task_handler_test.go` → `TestGetSystemTasksRebuildsVersionOneReceiptUpdates`
covers the listing.

## Receipt statuses

`models.ReceiptStatus` (`internal/models/receipt_status.go`) is a plain Go `string` type — there is
no DB enum or CHECK constraint anywhere, so **adding a value needs no migration**, only these four
edits, which nothing forces to agree with each other:

1. the `const` block,
2. the `Value()` guard chain (a missing value makes every write fail as a generic 500),
3. `ReceiptStatuses()` — the single membership list, which drives the server-side validations of a
   caller-supplied status: `handlers/receipts.go` `BulkReceiptStatusUpdate`,
   `commands/update_group_receipt_settings_command.go` `isValidReceiptStatus` (which covers both the
   quick-scan default and the receipt summary's breakdown set), and
   `repositories/group_receipt_settings.go` `canonicalStatusOrder`, which orders a group's configured
   summary statuses by this list rather than alphabetically,
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

## User Preferences

Per-user settings, stored on `models.UserPrefernces` (note the misspelling - it is the repo-wide
spelling of the type). There is **no service layer and no Upsert command**: `PUT /userPreferences`
unmarshals the request body straight into the model, and the swagger `UserPreferences` schema is
both the request and the response, so one schema edit covers both directions. The user id comes from
the JWT, never the body. Gated by `app.user-preferences.read` / `.update`, both in the Legacy User
set.

The whole object also rides on **AppData** (`services/auth.go`), which is how the desktop and mobile
read it without a second request.

**`UpdateUserPreferences` copies the request onto the stored row one field at a time** - a new field
that is not added to that block silently never persists: no compile error, and nothing else in the
suite fails. `repositories/user_preferences_test.go` round-trips each boolean specifically to catch
that; extend it when adding a field. The write itself is `Select("*").Updates(&struct)` inside a
transaction, so GORM emits the column (including a `false`, which the struct form of `Updates` would
skip) as soon as the model field exists.

Because the body is unmarshalled with `encoding/json` and nothing sets `DisallowUnknownFields`, a
**removed** field sent by an already-released mobile build is ignored rather than rejected, and an
**omitted** field decodes to its zero value - so for a bool, "omitted" and "explicitly false" reach
the repository identically. Both are pinned in `models/user_preferences_test.go`.

Current fields: the three quick-scan defaults, `userShortcuts`, and `closeChipSelectOnSelect` (a
desktop-only behavior flag for the multi-select chip pickers - see `desktop/CLAUDE.md` -> "The chip
picker's close-on-select preference"; mobile carries the generated field but never reads it).

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
  on the server, and `resolveQuickScanFieldConfig` on both clients:
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

## Role-required receipt fields & single-call create

A group role can require its members to keep **at least one comment** and/or **at least one image**
on the group's receipts: `GroupRoleDefinition.RequireReceiptComment` / `RequireReceiptImage`
(`not null;default:false`, so AutoMigrate adds them and existing roles are unchanged — no migration).
Carried on `UpsertRoleCommand` / `RoleView` / swagger `Role` + `UpsertRoleCommand` as
`requireReceiptComment` / `requireReceiptImage`; **group scope only** (`Validate` rejects them on APP
under the shared `grants` key, like `seesAllMembers`). Persisted by
`RoleRepository.SetGroupRoleReceiptRequirements` — a separate map-form `Updates`, for the same
transposable-bools reason as `SetGroupRoleIndividualGrantConfig` — called inside `CreateRole` /
`UpdateRole`'s transactions; `GetAllRoles` and `groupRoleToView` read them back.

**Never read the raw flags to decide anything — resolve them.**
`ReceiptService.ResolveReceiptRequirements(userId, groupId)` (`services/receipt_requirements.go`)
returns `structs.ReceiptRequirements{commentRequired, imageRequired}` with the waivers applied:

- **comment** waived when the group sets `HideComments` **or** the caller lacks
  `group.comments.create` (the same pair `IsQuickScanCommentShown` + quick scan's permission check
  use — a caller who can't comment must never be required to);
- **image** waived when the group sets `HideImages` (no permission waiver: images ride the create,
  which `group.receipts.create` covers);
- a non-member, a member with no role, and the synthetic **All** group require nothing.

The flags come from `RoleRepository.GetMemberReceiptRequirementFlags` — **one** join over
`group_members` → `group_role_definitions` → `groups` for any number of groups (All group filtered in
Go, since `is_all_group` is nullable). The common case (no flag set) therefore costs one query; the
settings and the permission are read only when a flag is on. The batched
`ResolveReceiptRequirementsForGroups` builds **`AppData.groupReceiptRequirements`**
(`map[groupId]ReceiptRequirements`) from the groups' preloaded settings and the loop's already
resolved group permissions, still one query. Only groups where something is required are present
(absent ⇒ nothing required); the map is never `null` (`{}` when empty). It is deliberately **not** in
AppData's swagger `required` list, so a newer mobile build can still parse an older server's payload.

### `POST /receipt/withFiles` (`createReceiptWithFiles`)

Create is **one multipart call** carrying the receipt, its comments and its images, so the server
sees the images at create time (the old flow — JSON create, then one `POST /receiptImage` per image —
never let it) and the create is atomic.

- **Wire shape:** `multipart/form-data` with part **`receipt`** = `UpsertReceiptCommand` JSON (comments
  included) and zero or more **`files`** parts (binary). Swagger declares
  `encoding: receipt: contentType: application/json`. Generated clients differ in how they encode a
  model inside multipart, so `CreateReceiptWithFilesCommand.LoadDataFromRequest` accepts `receipt`
  **either as a plain form value or as a file part** (e.g. `application/json`), in that order. A
  missing/malformed part or a failed `Validate` is a **400** (the legacy JSON create still uses 500).
- **Order:** parse → `group.receipts.create` gate (`HandleRequest`) → **every file through
  `ValidateFileType` before anything is written** (400 keyed `files.<i>`) → the shared
  `enforceReceiptCreate` (grant selection, member visibility, custom-field selection — 403s — then
  the requirement check) → `ReceiptService.CreateReceiptWithFiles` → `writeCreatedReceipt` (the same
  category/tag strip + member masking as `CreateReceipt`). Both create handlers share
  `enforceReceiptCreate` and `writeCreatedReceipt`, so they cannot drift.
- **`CreateReceiptWithFiles`** runs `CreateReceipt(…, createSystemTask=false)` plus
  `CreateReceiptImage` per file in **one `db.Transaction`**. Files are written inside it (the new
  receipt's group is only visible there), so every written path is tracked and removed if the
  transaction fails — the `DuplicateReceipt` pattern. `CreateReceiptImage` now returns the row (with
  its id) alongside a **write** error so a partially written file can be located too. The
  `RECEIPT_UPLOADED` task is recorded afterwards from a deferred `CreateSystemTaskFromError`, like
  `DuplicateReceipt` — `CreateReceipt`'s own task write would run on another connection. The response
  is re-read after commit, so it carries `imageFiles`.
- **Size:** `ParseMultipartForm(MultipartFormMaxSize)` is only the in-memory threshold (larger parts
  spill to temp files, which net/http removes after the request). The real cap is nginx's
  `client_max_body_size 50M` in `docker/default.conf`, which now bounds **all** of a create's images
  together, as it already does for a multi-file quick scan.
- **The old `POST /receipt/` stays** (`deprecated: true` in swagger) for already-released mobile
  builds. It enforces the same requirements with an image count of 0, so while an image is required
  it always 400s — with a `files` message telling the user to **update their app**.

### Where the requirement is enforced

`handlers/receipt_requirement_enforcement.go` — every failure is a **400 validator error keyed
`comments` / `files`** (`WriteValidatorErrorResponse`):

- **Both creates** — comment = any non-blank `command.Comments[].comment`; image = files in the call.
- **`UpdateReceipt`** — against the **target** group (the destination when moving, so a move into a
  requiring group is refused), using the **stored** `currentReceipt.Comments` / `ImageFiles`
  (`GetFullyLoadedReceiptById` preloads both). Edit mode adds images and comments immediately through
  their own endpoints, so the stored receipt is the truth; judging the command would let a client
  claim a comment it never saved.
- **`DeleteComment` / `RemoveReceiptImage`** — refuse removing the **last** one while it is required
  in the receipt's group (`enforceCommentDeleteKeepsRequired` / `enforceImageDeleteKeepsRequired`).
  To swap the only image, upload the new one first.
- **Quick scan** (`ResolveQuickScanFields`) — a role-required comment **shows and requires** the
  field even when the group's quick-scan config leaves it off (the resolver already carries the
  `HideComments` / permission waivers). Quick scan always has an image. Resolved once per group.
- **Out of scope:** bulk status update, duplicate, email ingest (no user) and import.

### The synthetic "All" group is never a destination

**Both creates and Quick Scan reject it with a 400**, as `UpdateReceipt` already did for a move.
A caller's All-group membership carries the default unrestricted role, so its
`group.receipts.create` passes the declarative gate, and the receipt would then live under a role
that sidesteps every real group's grant and visibility controls. One helper,
`isAllGroupDestination` (`handlers/receipts.go`), holds the rule for all four paths; a group that
does not exist is *not* the All group, so the permission check after it still denies that.
- **Creates:** checked first inside `enforceReceiptCreate`, so it answers before any 403 check and
  before any file is written (`allGroupCreateMessage`).
- **Quick Scan:** checked per file before `ResolveQuickScanFields` and before anything is enqueued;
  a 400 validator error keyed `files.<i>.groupId`, matching that resolver's keys.
- Neither client can send it in normal use (both pickers exclude the All group), so this closes a
  crafted-request hole rather than changing a flow. Email ingest and import are out of scope: their
  destination is configuration, not caller input.

Tests: `commands/upsert_role_command_test.go` (APP rejected), `repositories/roles_test.go`
(`TestGroupRoleReceiptRequirementsRoundTrip` incl. toggle-off via `GetAllRoles`,
`TestGetMemberReceiptRequirementFlags`), `services/roles_test.go` (service round-trip),
`services/receipt_requirements_test.go` (resolver matrix incl. batched parity, nothing-required
cases, AppData `{}`, `CreateReceiptWithFiles` happy path + a mid-transaction failure leaving no
receipt/FileData/comment/file and a FAILED task, quick-scan role comment), and
`handlers/receipt_requirement_enforcement_test.go` (both `receipt` encodings, invalid file type
writes nothing, both creates, update incl. a move, last comment/image delete, quick scan).

That handler file also pins the new endpoint on its own, because the old one shares its checks and
would otherwise hide a regression in the new one. Each case was verified to **fail** when its
production line is broken:
- **Zero images** — the common case: no requirement, and a comment-only requirement, both create a
  receipt with no `files` part (fails if the image rule fires when not required).
- **The shared 403 chain** — disallowed category, new category without `app.categories.create`, a
  non-visible payer, a custom field without `app.custom-fields.read`, each with an image attached
  and asserting nothing was written (`assertNothingWritten`: no receipt/FileData/comment rows, no
  file in the group directory), plus a granted-category positive control. These fail if
  `CreateReceiptWithFiles` stops calling `enforceReceiptCreate`.
- **Invalid bodies** — malformed JSON as a form value or as a JSON file part, an empty `receipt`, a
  non-multipart body, and a receipt failing `Validate` (400 keyed `name`), all writing nothing.
  These fail if a parse error stops being a 400.

`writeCreatedReceipt`'s category/tag strip and member masking are **not** targeted there: on a fresh
create the caller can only attach what it may see and is itself the creator, so there is nothing for
them to hide.

**The partial-write clean-up has a test hook.** `CreateReceiptImage` writes through the unexported
`writeReceiptImageFile`, which `repositories.SetReceiptImageWriterForTests` swaps (returning a
restore func). A real filesystem will not fail halfway through a file on demand, and that is the one
case where a row exists and a truncated file sits on disk: `CreateReceiptWithFiles` tracks the path
from the row id so the rollback removes it. `TestCreateReceiptWithFiles_PartialWriteIsRemoved` writes
half the second image and then fails; it was checked to fail when a failed write's path is not
tracked. Nothing else reassigns the variable.

**Parsing has direct unit tests** in `commands/create_receipt_with_files_command_test.go` (no DB):
both encodings, the documented precedence (a non-empty form value wins over the file part; an empty
one falls back to it), every error, and `files` keeping order, names and exact bytes, with an absent
`files` giving an empty, non-nil slice. The precedence case was checked to fail when the two reads
are swapped. The All-group rejections are pinned by `TestCreateReceiptWithFiles_AllGroupRejected`,
`TestCreateReceipt_LegacyEndpointAllGroupRejected` and `TestQuickScanHandlerAllGroupRejected`, each
with a real-group positive control.

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
  - **The `associatedGroup` / `associatedApiKeys` list filters are default-deny.** `GetPagedGroups`
    only permission-gates the literal `ALL` (via `app.groups.read`) and the repository only scoped the
    literal `MINE` to the caller — so any *other* value (empty or bogus) fell through both and returned
    every group in the system (same shape for `GetPagedApiKeys` / `app.api-keys.read-any`). Fixed on
    two layers: the command `Validate` rejects anything that is not exactly `MINE`/`ALL` (400), and the
    repository's filter is now `if ALL {…} else {scope to caller}` so an unexpected value fails closed
    to the caller's own rows. Tests: `commands/paged_group_request_command_test.go`,
    `commands/paged_api_key_request_command_test.go`,
    `repositories/api_key_test.go` (`TestGetPagedApiKeys_UnknownFilterFailsClosedToOwnKeys`).
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

## Receipt Summary

A block of totals under the receipts table, aggregated over the **whole current filter result set**
rather than the visible page: a receipt count and amount total overall, then the same figures for
each status the group has configured. Configuration lives on `GroupReceiptSettings` and applies to
every member — it is not a per-user preference.

**Four settings, two join tables.** `ReceiptSummaryEnabled` is a plain column (off by
default, so existing installs are unchanged) and so is `ReceiptSummaryPosition` (see below).
`ReceiptSummaryCustomFieldIds` and
`ReceiptSummaryStatuses` are `gorm:"-"` projections over
`GroupReceiptSettingsSummaryCustomField` and `GroupReceiptSettingsSummaryStatus`, both keyed on
**GroupId** for the same reason `GroupReceiptSettingsCustomField` is (a lazily created settings row
still has `ID == 0` on the call that created it).

- **Not a `purpose` discriminator on the existing defaults join.** That table's PK is
  `{GroupId, CustomFieldId}`, so a discriminator has to join the key — a PK change `AutoMigrate`
  does not perform on a deployed install, and `replaceGroupDefaultCustomFields` deletes
  `WHERE group_id = ?` unscoped, so saving the defaults would silently wipe the summary selection.
  `TestUpdateGroupReceiptSettingsKeepsDefaultAndSummaryFieldSetsSeparate` is the regression guard.
- **Not a delimited or JSON column for the statuses.** The house pattern for "a parent owns a set of
  enum-ish strings" is row-per-value (`GroupRolePermission`), and a per-row column keeps
  `ReceiptStatus.Value()`'s `driver.Valuer` rejection at the DB boundary for free.
- **`size:32` on the status column is load-bearing.** A Go string maps to an unbounded TEXT, and
  MySQL/MariaDB reject that in a key — without it the table creates fine on SQLite and fails
  AutoMigrate on MySQL.

**`LoadDefaultCustomFieldIds` is now `LoadSettingsProjections`** (and `…ForGroups`), and loads **all
three** projections. The rename was deliberate: it makes the compiler find every call site, and
loading them together means no caller can hydrate one and miss another — which would emit a `null`
where swagger promises an array, the failure that only shows up on an already-released mobile build.
The `[]`-never-`null` rule covers all three.

**The command's three fields are pointers**, `nil` == leave unchanged, each for its own reason:
the custom field ids because the desktop omits that key for an admin without `app.custom-fields.read`;
the statuses because a plain slice cannot tell "omitted" from the legitimate "clear every status";
and the bool because a plain one unmarshals as `false` for any client that does not send the key and
would silently switch a configured summary off.

**The permission gate is deliberately asymmetric.** The CURRENCY field selection joins the existing
`app.custom-fields.read` 403 because it names catalog entries. `receiptSummaryEnabled` and
`receiptSummaryStatuses` stay **outside** it — neither reads the catalog, and gating them would lock
an admin without that permission out of the feature entirely. That is the difference from
`applyDefaultCustomFieldsOnIngest`, which *is* gated because it decides what the (to such a caller
invisible) default set does. Validation: unknown id → 400, **non-CURRENCY id → 400** (only
`CurrencyValue` is summed, so a TEXT field would total `0.00` forever and read as data rather than
misconfiguration), invalid status → 400 via the existing `isValidReceiptStatus`.

### Where the block renders — `ReceiptSummaryPosition`

`TOP` or `BOTTOM`, per group, defaulting to **BOTTOM** — where the summary rendered before the
setting existed, so an install that never touches it is unchanged. Modelled on
`CurrencySymbolPosition` (`models/receipt_summary_position.go`): a Go string type with `Scan` /
`Value`, no DB enum and no CHECK constraint, and a plain column carrying `gorm:"default:BOTTOM"` —
`AutoMigrate` adds it and backfills existing rows with the default on all three engines, so there
is **no migration**.

- **It rides on the summary RESPONSE, not just on the settings.** `ReceiptSummary.Position` is
  what both clients render from. The same argument as `Enabled`: a client's cached
  `groupReceiptSettings` is stale the moment an admin changes the configuration, and placement is
  configuration. Reading it from the response is what stops the block rendering in the old place
  until the next AppData refresh.
- **The server never emits `""`.** A closed Dart `EnumClass` throws on an unrecognized wire value
  and fails the WHOLE payload — `GroupReceiptSettings` rides on AppData, i.e. on **login**, which
  is exactly the two documented `Permission` outages. Empty is genuinely reachable (`loadSettings`
  maps a missing settings row to a *zero* `GroupReceiptSettings`), so `OrDefault()` normalizes it
  at both emit points: on the summary response, and in `LoadSettingsProjections` — the one batched
  hydrator every settings read passes through. **The swagger enum carries TOP and BOTTOM only** —
  no `""` member, matching the `CurrencySymbolPosition` precedent beside it — and the generated
  Dart enum's `fallback: true` sits on `BOTTOM`, so a client meeting a value added *later* lands on
  the same placement `OrDefault()` would have sent. Client and server agree by construction, and an
  empty position is not expressible on the write side at all.
- **The command field is a pointer**, like the other three: a non-pointer would unmarshal to `""`
  for any caller that omits the key, and the repository's assignment would blank a configured
  position. It must be assigned inside the `if command.X != nil` block in
  `UpdateGroupReceiptSettings` — the write is `Select("*")`, so a field missing from that block is
  actively zeroed, silently.
- **A `nil` pointer OMITS the column from the UPDATE** rather than writing back the value read at
  load time. `UpdateGroupReceiptSettings` builds an `omittedColumns` list from the three nil
  pointers and passes it to `Select("*").Omit(...)`, mirroring
  `SystemSettingsRepository.UpdateSystemSettings` (`repositories/system_settings.go`). Writing the
  loaded value back costs two things: a concurrent admin's change is clobbered by a value read
  before it landed, and for an **enum** the stored value goes back through `Value()` — so a
  position this build does not recognize (a newer release's member, seen after a downgrade) fails
  an otherwise unrelated settings save with a 500. `Omit` also preserves that value for the trip
  back up, where `OrDefault()` already keeps it off the wire.
  `TestUpdateGroupReceiptSettingsToleratesAnUnknownStoredPosition` pins it.
- **It needs no permission of its own.** Like `ReceiptSummaryEnabled` and `ReceiptSummaryStatuses`
  and unlike `ReceiptSummaryCustomFieldIds`, it reads no catalog — gating it on
  `app.custom-fields.read` would leave such an admin able to turn the summary on but not to say
  where it goes. `TestUpdateGroupReceiptSettingsAllowsSummaryPositionWithoutCustomFieldPermission`
  pins that.
- **Validated in the command, not at the DB boundary.** `Validate()` routes it through its own
  `Value()` (the `pie_chart_data_command.go` style), which turns what the DB layer would surface as
  a generic 500 into a field-level 400.

### `POST /api/receipt/group/{groupId}/summary`

Gated on **`group.receipts.read`** — the same permission as the table it sits under. Not a new
permission (the endpoint returns only aggregates of rows the caller can already page through, so a
separate gate would just produce a table whose own totals 403), not `group.widgets.read` (this is not
a dashboard widget), and **no `app.custom-fields.read` gate**: that permission covers the catalog
endpoints and `enforceReceiptCustomFieldSelection`, which explicitly lets a non-holder read the
values of fields already on a receipt, and `FULL_RECEIPT_ASSOCIATIONS` already ships those names to
any receipt reader.

`ReceiptSummaryCommand` carries the filter and an optional `ConfigurationGroupId`, **not** the field
or status list — the configuration is the group's and applies to everyone, so a client must not be
able to add a column or opt out. `ConfigurationGroupId` exists for the synthetic "All" group, which
spans several groups and has no meaningful settings of its own.

**Borrowing a configuration is the All group's privilege alone.** A real group must use its own, or a
member could render group A's receipts under group B's statuses and currency fields — overriding what
A's admin configured, and switching on a summary A has turned off, which is exactly the invariant
above. `resolveConfigurationGroupId` therefore tests `IsAllGroup(uintGroupId)` first
(`ErrConfigurationGroupNotAllGroup` → **400**, a malformed request rather than an access failure), and
only then authorizes the named group with `HasGroupPermissions`
(`ErrConfigurationGroupForbidden` → **403**), or it becomes a way to enumerate another group's
configured field names. **That order matters**: rejecting on the request's shape first means the
response cannot depend on whether the caller can read the named group, so the endpoint can never be
used to probe for another group's existence — pinned by
`TestReceiptSummary_ConfigurationGroupRejectedBeforeAccessCheck`, which asserts an unreadable group and
a nonexistent one come back identically. The desktop is unaffected: it sends the viewed group's own id
for a real group, so only the All group ever sends a differing one.

`services/receipt_summary.go` follows `pie_chart.go`: `IntersectReceiptFilterWithGrants`, then
`GetPagedReceiptsByGroupId` unpaged, then a Go fold. That one repository call is what supplies the
filter builder, the All-group `group_id IN (...)` fan-out **and** paid-by visibility — so the totals
cannot drift from the rows above them. `MaskReceiptsForMemberVisibility` and
`SubstituteRestrictedCategoriesTags` are not called: the first only masks user references, the second
only matters for category/tag bucketing.

**Go `decimal` fold, not SQL `SUM`.** There is no `SUM` anywhere in this API. `amount` is
`decimal(10,2)`, but SQLite has no decimal type, so `SUM(amount)` returns an IEEE double there and an
exact decimal on Postgres/MySQL — the same endpoint would report different cents on the three
supported engines, on exactly the money this feature exists to report
(`TestReceiptSummary_DecimalPrecision`). A SUM query would also have to re-derive the access controls
above, where the failure mode is a totals row that counts receipts the viewer may not see.

**Rules that are easy to get backwards:**

- Every **configured** status is seeded before the walk, so one matching no receipt still renders as
  a zero row — that is what keeps the block's shape steady as the filter narrows.
- A receipt whose status is **not** configured still counts toward the overall row. It is in the
  filter result, so excluding it would make the total disagree with the table's own count.
- Where a receipt holds **several** values for one custom field, **lowest id wins**, mirroring
  `reporting/receiptsource.addCustomFields`. `custom_field_values` has no unique index on
  `(receipt_id, custom_field_id)` and the association loads without an `ORDER BY`, so preferring
  whichever came back first would let two identical requests return different numbers. A row with no
  `CurrencyValue` never wins, so an empty low-id row cannot hide a real one.
- Columns are built from the configured id order, never by ranging the map — a summary whose columns
  reshuffle between requests is unreadable.
- `enabled: false` comes back at a normal **200** with zeroed rows, and returns **before any receipt
  query**. That is both the cost floor for every install that has not opted in and the reason a
  client with stale group settings renders nothing rather than an error.

**Money crosses the wire as a string** (`decimal.Decimal` marshals quoted), matching `Receipt.amount`
and `CustomFieldValue.currencyValue`. `PieChartDataPoint`'s `float64` is the outlier, forced on it by
the charting library.

**Cost.** This loads the filtered set unpaged — the same bargain `PieChartService` and
`ReportDataService` already make, but the receipts table is a hotter screen than the dashboard. It is
bounded on four sides: a group that has not enabled the summary never reaches a query, the desktop
does not re-request on paging or sorting, the desktop skips the request entirely when no configured
group resolves, and `CustomFields` is preloaded only when a field is configured. If it ever needs
more, the next step is a `SUM` fast path behind the same service signature — not caching.

## Custom fields on the receipts list (columns & sorting)

The desktop receipts table can show a column per custom field and sort on it. Two things on this side
make that work; both have a sharp edge.

### The list response always carries custom field values *with their definitions*

`GetPagedReceiptsForGroup` (`handlers/receipts.go`) preloads
`constants.CUSTOM_FIELD_ASSOCIATIONS` — the value, its `CustomField`, and that field's `Options` —
unconditionally, not only under `fullReceipts`. `FULL_RECEIPT_ASSOCIATIONS` is composed from the same
constant so the two lists cannot drift.

**Loading the values without their definitions breaks the mobile app.**
`models.CustomFieldValue.CustomField` is a non-pointer struct with no `omitempty`, so it serializes
even when unloaded — as `{"id":0,"name":"","type":""}`. Mobile calls this exact endpoint
(`getReceiptsForGroup`) and deserializes each row into a typed `Receipt`; its generated
`CustomFieldType` is a closed enum with no empty member, and the `one_of` `AnyOfSerializer` swallows
the failure rather than reporting it, so every row of the receipts list silently collapses. Same
failure mode as the `aggFunc` omitempty fix — see `mobile/CLAUDE.md` → "Serialization contract".
`TestPagedReceiptCustomFieldsAlwaysCarryTheirDefinition` asserts the response never contains
`"type":""`.

**That guard is duplicated at the handler on purpose.** The repository test hands the association
list to the repository itself, so it stays green if `GetPagedReceiptsForGroup` ever stops choosing
`CUSTOM_FIELD_ASSOCIATIONS` — the one line that actually decides. `handlers/receipts_test.go` →
`TestPagedReceiptsCarryCustomFieldDefinitionsWithoutFullReceipts` drives the handler with
`fullReceipts: false` and asserts on the **raw response bytes**, which is where `"type":""` exists at
all. Verified to fail both ways the line can regress: preloading nothing, and preloading the values
without their definitions.

This is not a new disclosure: the single-receipt endpoints already returned these values to any
holder of `group.receipts.read`, and `MaskReceiptsForMemberVisibility` already masks their created-by.

### Sorting by `custom_<id>`

`orderBy` accepts the reporting engine's own key shape, parsed by
`receiptsource.ParseCustomFieldKey` (digits only — `custom_1_month` and `custom_abc` are rejected, and
a malformed value still errors out, which is the guard keeping `orderBy` out of the SQL it is
concatenated into). One key vocabulary for reports and the table.

**The parse is 32-bit on purpose.** The id it yields is a `models.CustomField` id, i.e. a `uint` —
which is 32 bits on a 32-bit build, where parsing at 64 bits and narrowing would not merely overflow
but name a *different* field (`custom_4294967297` would resolve to field `1`). No id reaches 2^32, so
capping only rejects keys that cannot name a real field, and the caller then treats them like any
other malformed key. Flagged by CodeQL as an incorrect integer conversion.

**The sort DIRECTION is never concatenated on a custom-field path.** Both fallback branches hand it to
`BaseRepository.Sort`, which renders `ASC`/`DESC` from a `Desc bool` on `clause.OrderByColumn` — the
same convention the other twelve repositories use — and the main ordering path picks its keyword from
literals, because `clause.OrderBy{Expression}` leaves `Columns`/`Desc` unavailable (see below).
`GetPagedReceiptsByGroupId` does validate the direction up front with `commands.IsValidSortDirection`,
but that check sits ~350 lines upstream, so the safety is deliberately restated at the point the SQL is
built rather than inherited from one call path. CodeQL flags the concatenated form as a query built
from user-controlled sources; the remaining concatenation at the built-in-column branch is left alone,
its `orderBy` and direction both being allow-listed against literals by `isTrustedValue`.

- **A correlated subquery, never a join.** `custom_field_values` has no unique index on
  `(receipt_id, custom_field_id)`, so a join multiplies receipt rows and corrupts both the total count
  and pagination. `orderByCustomField` builds the subquery as an ordinary `*gorm.DB` and passes it as
  a bind var — gorm expands a `*gorm.DB` var into its own built SQL (`gorm/statement.go`), so this
  stays in the query builder.
- **Which duplicate wins matches reporting**: lowest id among the values that actually resolve. Hence
  the `IS NOT NULL` clause, and hence a SELECT **inner**-joins `custom_field_options` — an
  unresolvable option id is skipped, not preferred. Drop either and the table sorts a receipt as "no
  value" while a report shows one for it.
- **`currency_value` is a TEXT column, so CURRENCY sorts through a cast.**
  `CustomFieldValue.CurrencyValue` is a `*decimal.Decimal` with **no `gorm:"type:"` tag**, unlike
  `Receipt.Amount` (`gorm:"type:decimal(10,2)"`). gorm types an untagged `driver.Valuer` by what
  `Value()` returns, and `decimal.Decimal.Value()` returns a string — so a bare sort is lexicographic
  (`"100" < "20"`). `CAST(... AS DECIMAL(20,6))` is ANSI and needs no dialect branching: MySQL and
  Postgres take `DECIMAL(p,s)` directly, SQLite gives that type name NUMERIC affinity. **Retyping the
  column is the better fix** and would remove the cast entirely (gorm's Postgres migrator does emit
  the required `USING ?::?`), but it is a migration over existing data and belongs in its own change.
  The same trap applies to any future untagged decimal field.
- **The direction and the `receipts.id` tiebreaker live inside the one `clause.Expr`.**
  `clause.OrderBy` builds its `Expression` *instead of* its `Columns` and `Desc`, and a second
  `Order()` call replaces the clause rather than appending to it — silently. The tiebreaker is not
  cosmetic: a BOOLEAN or SELECT field has a handful of distinct values, and without a unique last term
  `LIMIT`/`OFFSET` paging repeats and skips rows.
- **A deleted custom field falls back to `created_at` rather than erroring.** Clients persist their
  sort, and `handlers/receipts.go` maps every repository error to a 500, so erroring would make the
  list fail to load at all for anyone holding a stale sort. Both fallbacks (the unknown field and the
  unsortable type) go through `defaultReceiptOrder`, which **carries the same `receipts.id`
  tiebreaker** as the custom-field path — `created_at` is not unique, and receipts created in one
  batch share a timestamp, so without it `LIMIT`/`OFFSET` paging repeats and skips rows. That helper
  can chain the tiebreaker onto `BaseRepository.Sort` because both clauses are **column**-based and
  gorm appends them; the expression form above cannot, which is why it builds one `clause.Expr`.
- **NULL ordering is engine-dependent** and deliberately not normalised: SQLite and MySQL sort NULL
  first, Postgres last, so receipts with no value land at opposite ends. `NULLS LAST` is not portable
  (MySQL 8 and MariaDB reject it), and normalising it would mean evaluating the subquery twice. This
  matches the existing nullable built-in column, `resolved_date`.
- **The subquery is served by a composite index.** `CustomFieldValue` declares
  `idx_custom_field_value_lookup` over `(receipt_id, custom_field_id)` — without it the subquery is a
  full scan of `custom_field_values` **per candidate receipt** on SQLite and Postgres (MySQL gets one
  from the FK). It is two columns, not three: the `ORDER BY id LIMIT 1` would nominally like `id`
  trailing, but `ID` lives on the embedded `BaseModel` and cannot be tagged without overriding the
  field, and the two-column seek already narrows to roughly one row. Declared **only** as a model tag —
  there is no hand-written schema per engine, so AutoMigrate creates it everywhere.
- `POST /api/export/{groupId}` deserializes the same `ReceiptPagedRequestCommand` and calls the same
  repository method, so it inherits both behaviours.
- **Tests**: `repositories/receipt_custom_field_sort_test.go` — one per type (CURRENCY with values
  that sort differently as text than as numbers), value-less receipts still listed with a matching
  count, the duplicate/lowest-non-null rule, stable paging across equal values, the fallback
  tiebreaker (asserted on the generated SQL — a tie is broken by whatever order the engine happens to
  return, so paging real rows would pass either way), the index existing after migration, both fallbacks
  (unknown id, and an unsortable type — an empty one, the only unknown `CustomFieldType.Value()` lets
  through) asserted in **both** directions so a dropped direction cannot pass, the malformed-key
  rejection, and the `"type":""` guard. `receiptsource_test.go` → `TestParseCustomFieldKey` covers the
  parser directly, including the 2^32 boundary and the round trip against `CustomFieldKey`. The suite is **SQLite only**, so
  the cross-engine NULL ordering is not covered there.

### The first comment (`first_comment`)

The desktop table's **Comment** column shows each receipt's first comment and sorts on it. The paged
list only preloads `Comments` under `fullReceipts`, so the value rides a separate projection.

- **"First" is `firstCommentOrder`** (`repositories/comments.go`): `created_at`, then `id` — comments
  written in one insert share a timestamp. The sort and the display both use that one constant, so the
  column can never sort by one comment while showing another. Replies cannot be created through the
  API, so the list is flat and no `comment_id` filter is applied.
- **`Receipt.FirstComment`** is a `*string` with `gorm:"-"` and `omitempty`: the handler fills it
  after `MaskReceiptsForMemberVisibility`, through `PermissionService.LoadFirstVisibleComments`. That
  is **one** `GetCommentsForReceiptIds` query per page, plus the per-group visibility the masker already
  resolves. Every other endpoint leaves it nil, so the key is absent there, not `null`.
- **`orderBy = first_comment`** (`constants.FIRST_COMMENT_ORDER_BY`) takes its own branch ahead of
  `isTrustedValue`: `orderByFirstComment` is a correlated subquery, never a join (a receipt has many
  comments, so a join would multiply rows and corrupt the count). It shares `orderBySubquery` with
  the custom-field sort, which is where the literal direction keyword and the `receipts.id DESC`
  tiebreaker live. The tiebreaker matters here too, because most receipts have no comment at all.
- **No new permission.** Reading a receipt already means reading its comments on every other
  surface, so the list's `group.receipts.read` gate covers this.
- **Member isolation applies to both halves.** A comment by an author the caller can't see in that
  receipt's group is dropped from every response, so it must not be shown **or sorted on**. A sort
  on hidden text would order the table by comments the caller never sees, and would leak them one
  comparison at a time. `GetPagedReceiptsByGroupId` therefore takes a sixth argument, a
  `CommentAuthorVisibilityResolver`. The handler and the CSV export pass
  `PermissionService.CommentAuthorVisibilityResolver`; the export passes it because it honours the
  table's sort. Internal callers pass `nil`, the same contract as the paid-by resolver.
  `commentAuthorVisibility` builds a per-group disjunction on `receipts.group_id`, mirroring
  `applyActivityVisibilityDisjunction`, so an All-group page judges each receipt by its own group's
  rules. It adds **no predicate at all** when no group restricts the caller, which covers every
  non-isolated install. An authorless comment stays visible, as it does in `filterComments`.
- **`idx_comment_receipt_id`** on `Comment.ReceiptId` serves both the subquery and the loader. It is
  a model tag only; there was no index on that column before.
- NULL (no comment) ordering and text collation are engine-dependent. That matches `name`,
  `resolved_date` and the custom-field sorts, and is deliberately not normalised.
- **Tests**:
  - `repositories/receipt_first_comment_sort_test.go`: both directions; the earliest comment beating
    a later one that sorts lower; the id tiebreak; hidden authors skipped, with an unrestricted
    contrast; the All-group per-group rule; the SQL shape; no predicate when unrestricted; the index.
  - `services/first_comment_test.go`: the earliest pick, hidden and authorless authors, and the
    resolver.
  - `handlers/receipt_first_comment_test.go`: the key is present or absent on the wire, the sort
    works through the handler, and an isolated member never receives a peer's text, neither as the
    value nor through the sort order. That last test was checked to fail when the handler passes a
    `nil` resolver.

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
source, err := receiptsource.New(customFields, appLocation) // needs CustomField.Options loaded
model, err := reporting.Run(spec, source.Catalog(), source.Rows(receipts), meta)
```

**Caller must preload** `PaidByUser`, `Group`, `Categories`, `Tags`, `CustomFields`. An unloaded
association resolves to no value, which surfaces as a `(None)` bucket — not an error.

**Date period grouping — derived string fields.** `date`, `resolved_date` and `created_at` are
`TypeDate`, so grouping by one buckets on the exact instant (one group per receipt). To group by
calendar period, `receiptsource` also offers derived **string** fields `<base>_day` / `_month` /
`_year` (e.g. `date_month`, `created_at_year`, `resolved_date_day`) — zero-padded ISO, so they sort
chronologically as plain text. **Which zone the parts are cut in depends on the field** (see "App time
zone"): `date` and every DATE custom field are calendar days, read in **UTC**; `created_at` and
`resolved_date` are instants, read in the **app time zone** passed to `receiptsource.New`. The
package never reads the setting itself — the zone is a parameter, so it stays pure. A report groups by one of these instead of the raw date field;
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
mapping bool → `Yes`/`No`, date → `2006-01-02` **in the field's own `FieldRef.Location`** (UTC for a
calendar day, the app zone for an instant — matching the derived `_day` field), currency →
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

**The period's date field (`ReportPeriod.dateField`).** A period covers one receipt date: `date`,
`resolvedDate` or `createdAt`. These are the `ReceiptPagedRequestFilter` JSON keys, and the same fields,
in the same order, as the receipts table's quick date filter, whose list drives the Report Builder's
picker.
- **Mapping and validation both live in `commands`.** `ReceiptDateFilterKeys()` pins the list and
  `(*ReceiptPagedRequestFilter).DateFilterField(key)` returns the matching slot.
  `validatePeriod` accepts a key only if `DateFilterField` resolves it, so the two cannot disagree. A
  bad value is a 400 under the `period` key, reported after the preset checks.
- **`applyPeriod` writes the BETWEEN onto that one slot**, overwriting whatever condition it held.
  The preamble, `Meta.Params["Period"]` and `{{period}}` are unchanged.
  - An unknown value falls back to `Date`. `GenerateReportFromTemplate` and the dashboard render path
    run a stored configuration without re-validating it, so the fallback stays lenient, like
    `resolvePeriodBounds`' default.
  - Only the chosen slot is overwritten, so a builder **Date** filter now ANDs with a period on
    another field. On `date` it is still replaced, as before.
  - A nil `resolved_date` never matches, so a Resolved Date period excludes every unresolved receipt,
    DECLINED included.
- **Empty means `date`** (`ReportPeriod.DateFilterKey()`). That is how templates saved before the field
  existed keep their meaning: no migration, and `CurrentReportConfigurationVersion` stays 1. The Go
  field is `json:"dateField,omitempty"`. Templates are stored with `json.Marshal`, and the tag keeps an
  empty value out of the blob instead of storing `""`.
- **It is a plain `type: string` in swagger, deliberately not an enum.** `ReportPeriod` rides inside
  `ReportTemplate.configuration`, a response the mobile client deserializes. A closed dart-dio enum
  would throw on the first date key added later and fail the whole template payload on every
  already-released build. `TestReportPeriodDateFieldIsAnOpenStringOnTheContract` parses `swagger.yml`
  to hold that.
- **Adding a date key** means `ReceiptDateFilterKeys()`, `DateFilterField`, the swagger description,
  and desktop's `RECEIPT_DATE_FILTER_FIELDS`. `TestReceiptDateFilterKeys` pins the list, and
  `TestReceiptDateFilterKeysAreFilterJsonKeys` checks it against the struct by reflection.
- **The drill-in list is server-side: `POST /api/report/receipts` (`ReportService.Receipts`).** The
  builder's "N receipts" chip opens a list of what the report covers.
  - **Why the server builds it.** The list used to build its own period BETWEEN in the browser, and
    disagreed with the count in three ways:
    - it used the browser's time zone, not the server's (now the app time zone setting);
    - on SQLite, its ISO `…T…` bounds compared as text against the stored `YYYY-MM-DD HH:MM:SS…`,
      dropping first-day receipts;
    - it sent the "report generator" paid-by sentinel (`-1`) as-is, which matches nothing.
  - **How it stays in step with the report.** Both start from `prepareReportFilter` (the paid-by
    sentinel plus `applyPeriod`) and fetch through `ReportDataService.fetchReceipts`, so the list and
    the count are the same query. They differ only in presentation:
    - `Rows` marks hidden categories/tags `(Restricted)`;
    - `Receipts` strips them, masks for member visibility, and preloads
      `CUSTOM_FIELD_ASSOCIATIONS`, like the receipts list.
  - **Gating.** It is gated exactly like `PreviewReport`: `app.reports.read`/`readAll` plus
    `group.reports.read` in every group. It is **not** `group.receipts.read`, so a report reader sees
    the receipts their report already covers.
  - **Ordering and the cap.** The list is merged newest-first across groups and capped at
    `reportReceiptsCap`.
    - **The cap is 100** because it is the repository's page-size ceiling: `BaseRepository.Paginate`
      clamps any larger page.
    - **Each group fetches only its newest 100** as one page, which always contains the newest 100
      overall.
    - **`totalCount` is the sum of the groups' repository counts.** Those are taken after the grant
      intersection and the paid-by WHERE, so it is the true total, not the loaded length. The desktop
      shows a "Showing the newest N of M" notice when the two differ.
- **Tests:**
  - `commands/report_request_command_test.go` and `paged_request_command_test.go`: validation,
    marshalling, and the sync guards above.
  - `services/report_period_date_field_test.go`: DB-backed.
    - One receipt per field, with inclusive bounds, presets, and the AND with a Date filter.
    - For the drill-in:
      - parity with the report on every field;
      - the app zone set to America/Los_Angeles, then UTC, at the May 31/June 1 boundary;
      - the `-1` sentinel;
      - custom-field definitions;
      - merge order;
      - the cap across two groups, which keeps the newest overall and the full total;
      - a limited fetch whose count still excludes hidden payers.
  - `handlers/report_period_date_field_test.go`:
    - preview and template CRUD;
    - the unvalidated render/generate-from-template paths, with a May case proving the fallback lands
      on the receipt date and not the resolved one;
    - `GetReportReceipts`, per field and for its gate.
  - `repositories/report_template_test.go`: the stored blob.

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

**Splitting a receipt equally across categories / tags.** `ReportRequestCommand` carries
`splitCategoriesEqually`, `splitTagsEqually` and `splitExcludedFields` (all `omitempty`, so
`CurrentReportConfigurationVersion` stays `1` and an untouched template stores none of them). With a
box ticked, a receipt's money is divided equally across its distinct categories / tags instead of
counting in full in each — the default fan-out in the table below is untouched.
- **It is a receipt transformation, not an engine change.** `receiptsource.Source.Split`
  (`reporting/receiptsource/split.go`) turns a receipt with N categories and M tags into N×M shallow
  copies, one category and one tag each, every split money value divided across them. A dimension
  that is not split, or that holds at most one value, keeps its whole list.
- **What is divided:** the amount and every CURRENCY custom field not named in
  `splitExcludedFields` (keys `custom_<id>`; anything else is a 400 under that key; a stale id is
  ignored, like `GroupByLabels`). A nil value stays nil; non-currency fields ride along unchanged.
- **Cents are spread, never lost.** `allocateGrid` works in whole units of the value's scale (at
  least cents; a TEXT-stored currency value may carry more) with `big.Int`, never `Div` — so 100 over
  three is 33.34/33.33/33.33 and the shares sum back exactly. It allocates the whole N×M grid **in
  one pass**: splitting by category and then each share by tag skews the tag totals (0.04 over three
  categories × two tags would give one tag 0.03 and the other 0.01). In one pass every row sums to a
  categories-only split and every column to a tags-only split. Leftover units go by ascending
  category/tag id, so the same receipt always splits the same way. Negatives split by magnitude.
- **It runs BEFORE the restricted substitution** (`ReportDataService.RowsWithSplit`: fetch → split →
  `SubstituteRestrictedCategoriesTags` → `Source.Merge` → rows). So a receipt is divided by its
  **true** category count: a visible category's share is the same for every viewer, and
  `(Restricted)` receives exactly the hidden shares. `Merge` then folds copies of one receipt that
  substitution made identical (two hidden categories → two `(Restricted)` copies), keeping one row per
  receipt per bucket. Copies are tracked by their **position in the input**, not their id — unsaved
  receipts all carry id 0.
- **It only applies where the report cuts by the dimension** (`reportSplitOptions`): a grouping
  level, or the aggregate `detail.by`. Anywhere else a ticked box is a no-op (it would only change the
  row count). A currency custom field the report groups or aggregates by is never split — dividing
  it would move receipts between its own buckets.
- **Rows outnumber receipts now**, so `RowsWithSplit` returns `RowsPerReceipt`: `receiptCount` is
  its length, and the preview cap (`capRowsToReceipts`) cuts on whole receipts, never part-way
  through one's shares. `Rows` is unchanged — a wrapper with no split.
- Consequences worth knowing: COUNT stays one per receipt per bucket; AVG/MIN/MAX see the shares; a
  records-mode category label shows the copy's one category; the receipts drill-in still lists full
  amounts; the pie chart is untouched.
- **Tests:** `receiptsource/split_test.go` (allocation incl. a seeded property test of the grid's
  marginals, Split, Merge, isolation, determinism, and through the engine — `mutation-check.sh`
  carries nine split mutations), `services/report_split_test.go` (the true-count restricted case, the
  merge, exclusions, `reportSplitOptions`, the cap, Generate and the dashboard render),
  `handlers/report_split_test.go` (preview, the 400, template round trip, generate-from-template and
  render), and the command/contract tests in `commands/report_request_command_test.go`.

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
| Multi-value fan-out | A receipt with two tags is attributed to **both** buckets in full, so it double-counts, and that double count propagates to the grand total. Intended — matches `services/pie_chart.go`. Two multi-value levels produce a cross product. But only **distinct** values fan out: a receipt tagged `"Alex"` twice is attributed once. The opt-in split (below the table) is done **before** the engine, by splitting receipts, so this rule never changes. |
| Multi-valued measures | Refused (`ErrMeasureIsMultiValued`). Summing a field that resolves to several values would silently read the first and drop the rest. A `Multi` field is still a fine dimension and a fine display label. Note `Multi` is a **producer's declaration**, never checked against the rows: a catalog that lies loses values in `Row.Measure`. |
| Bucket keys | Two values share a bucket key **exactly when `compareValues` finds them equal**. Numbers key on `decimal.String()` (canonical, lossless); dates key on `Unix()` seconds + `Nanosecond()` and **never `UnixNano()`**, which is undefined outside 1678–2262 — a zero `time.Time` and a date in 585 share one. A coarser key merges buckets *and* makes the surviving bucket's value depend on input order. |
| Bucket **values** | A bucket keeps `Value.canonical()`, not whichever member arrived first. Agreeing with `compareValues` is only half the job: dates compare by instant, so one instant in two zones merges, and `Value.Time()` would otherwise hand a renderer an arbitrary one of them (`2026-05-01T00:00:00Z` vs `2026-04-30T19:00:00-05:00` format as different **calendar days**). **Date buckets are emitted in UTC** — that is bucket *identity* and must not change. The zone a day is *named* in is presentation: a producer declares it per field (`FieldRef.Location`, carried to `ColumnDescriptor.Location` / `render.Dimension.Location`) and the renderer applies it. Calendar-period grouping in a zone is done by the producer (`receiptsource`'s `_day`/`_month`/`_year` strings), never by the engine. |
| Unknown enums | `Validate` rejects an `AggFunc` or `DetailMode` outside the known set (`ErrUnknownAggFunc`, `ErrUnknownDetailMode`). Both used to fail silently — an unknown reduction fell through `finalize` and blanked the column; an unknown detail mode skipped the label-column check. Ask "is this aggregate mode?" only through `DetailSpec.isAggregate()`; two spellings of it once disagreed. |
| Enum switch drift | `AggFunc` is switched on in **four** places (`String`, `valid`, `aggFuncFromName`, `accumulator.finalize`) and Go forces none of them to agree. `enums_test.go` pins them to each other exhaustively over all 256 values. **Adding a reduction means updating all four** — `valid()` alone would let `Validate` accept a column `finalize` blanks. `finalize`'s fallthrough stays `Null()`, not a `panic`: nothing in this package panics, the line is unreachable through `Validate`, and drift is a compile-time mistake caught at `go test` time. |
| Formula size | `ParseArithmetic`/`ParseAggregate` refuse source over `maxFormulaLength` (1 KB) **before parsing**. expr's 10k-node cap does *not* bound this: a parenthesis builds no node, so nesting is bounded only by the goroutine stack (~640 B/paren; ~1.6M overflows the default 1 GB), and a Go stack overflow is a fatal error `recover` cannot catch. |
| Field keys | `NewFieldCatalog` rejects a key that is not a plain identifier (`ErrInvalidFieldKey`). `unit-price` would tokenize as a subtraction. |
| Empty dimension | Explicit `(None)` bucket (`IsNone: true`, `Value` null). Never dropped. |
| Ordering | Buckets sort by typed value with `(None)` last; records preserve input order. **Never range a Go map to produce output** (`pie_chart.go` has exactly this bug). |
| Formatting | The engine emits raw typed values. Currency symbols and decimal places are a renderer's job. |
| `GeneratedAt` | An **input** (`MetaInput`), never `time.Now()`. Determinism depends on it. The zone it is printed in (`MetaInput.Location`, the app zone) is an input too; the HTML footer prints it with the zone's abbreviation. |
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
./internal/reporting/mutation-check.sh          # all 63 mutations, engine + receiptsource + render
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
