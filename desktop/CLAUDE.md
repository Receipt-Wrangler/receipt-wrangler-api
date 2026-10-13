# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Core Development
- `npm start` - Start development server with proxy configuration (serves on localhost:4200, proxies /api to localhost:8081)
- `npm run build` - Build production application
- `npm run watch` - Build in watch mode for development
- `npm test` - Run unit tests with coverage
- `npm test:ci` - Run tests in CI mode with ChromeHeadless
- `npm run e2e` - Run Playwright end-to-end tests (see **E2E Testing** below)
- `npm run e2e:ui` - Run Playwright tests in interactive UI mode
- `npm run e2e:install` - Install Playwright browser binaries (one-time setup)

### Build Configuration
- Production builds go to `dist/receipt-wrangler/`
- Development server uses proxy configuration in `proxy.conf.json` to route API calls to backend
- Angular CLI configuration in `angular.json`

### Running in the Claude Code Web/Cloud Sandbox

> Playbook for booting the desktop app in the Claude Code web (cloud) sandbox from a **fresh session**.
> Everything is ephemeral — re-run these each session. See `api/CLAUDE.md` →
> "Running in the Claude Code Web/Cloud Sandbox" for the backend, and the root `CLAUDE.md` for the
> shared root-cause / run-order notes.

1. **Backend first.** The dev server only *proxies* `/api` → `localhost:8081`; it does not start the
   API. Bring the Go backend up first (see `api/CLAUDE.md`).
2. **`npm install`** — first run, node_modules isn't present.
   - **Lockfile gotcha:** the sandbox's npm (10.9.7) rewrites `package-lock.json`, stripping the
     `"libc"` platform-hint fields from optional platform deps. This is **metadata drift only** — no
     packages/versions change. Do **not** commit it: `git restore desktop/package-lock.json` to keep
     the tree clean.
3. **`npm start`** — serves on `0.0.0.0:4200`, proxying `/api` → `:8081` (`proxy.conf.json`). First
   compile is ~30–60s; ready when the log shows `➜  Local:   http://localhost:4200/`. Verify the proxy:
   `curl localhost:4200/api/featureConfig` → `200`.
4. **Log in as `admin` / `admin`** (the backend's auto-created default admin). Login lands on
   `/dashboard/group/<id>` ("All Dashboards", empty on a fresh DB).

**Driving the UI / screenshots with Playwright in the sandbox.** `@playwright/test` is pinned to
`1.59.1`, which expects Chromium build **1217**, but the sandbox pre-installs build **1194** at
`/opt/pw-browsers/chromium` (`PLAYWRIGHT_BROWSERS_PATH=/opt/pw-browsers`). Do **not** run
`playwright install` / `npm run e2e:install` (the download is blocked and pointless). Instead launch
Chromium by explicit path:
```js
const { chromium } = require('@playwright/test');
const browser = await chromium.launch({
  headless: true,
  executablePath: '/opt/pw-browsers/chromium',   // pre-installed 1194, ignore the 1217 mismatch
});
```
Run a standalone script (one that lives outside `desktop/`) with
`NODE_PATH=/home/user/receipt-wrangler/desktop/node_modules node script.js` so `require('@playwright/test')`
resolves. Scripted login flow (selectors from `e2e/helpers/auth.ts`):
```js
await page.goto('http://localhost:4200/auth/login');
await page.getByLabel('Username').fill('admin');
await page.getByLabel('Password').fill('admin');
await page.getByRole('button', { name: 'Login' }).click();
await page.waitForURL(/\/dashboard\/group\/\d+/);
```

**Running the whole `npm run e2e` suite in the sandbox — bridge the build numbers with symlinks.**
The `executablePath` trick above only helps ad-hoc scripts; `playwright.config.ts` has no
`launchOptions` hook, so the suite resolves the 1217 paths and fails. Bridge them to the installed
build with symlinks (outside the repo, so nothing is committed). Alongside the `chromium` convenience
symlink, the image ships two **versioned directories** — `chromium-1194/` and
`chromium_headless_shell-1194/` — and those are what you point at; confirm the number first, since it
moves with the sandbox image:

```bash
ls -d /opt/pw-browsers/chromium*   # chromium, chromium-1194, chromium_headless_shell-1194
SRC=1194                           # installed build, from the listing above
WANT=1217                          # build @playwright/test 1.59.1 resolves

# full chromium: the installed layout already matches what $WANT expects
ln -sfn /opt/pw-browsers/chromium-$SRC /opt/pw-browsers/chromium-$WANT

# headless shell: the directory AND the binary are named differently, so build the shape by hand
mkdir -p /opt/pw-browsers/chromium_headless_shell-$WANT
ln -sfn /opt/pw-browsers/chromium_headless_shell-$SRC/chrome-linux \
        /opt/pw-browsers/chromium_headless_shell-$WANT/chrome-headless-shell-linux64
ln -sfn headless_shell \
        /opt/pw-browsers/chromium_headless_shell-$SRC/chrome-linux/chrome-headless-shell
touch /opt/pw-browsers/chromium_headless_shell-$WANT/{INSTALLATION_COMPLETE,DEPENDENCIES_VALIDATED}
```

`devices['Desktop Chrome']` runs headless, so it is the **headless shell** path that actually fails
first — symlinking only `chromium-$WANT` is not enough. Then export the four `E2E_*` credential vars
and run `npx playwright test`; leaving `E2E_BASE_URL` at `http://localhost:4200` keeps the config's
`webServer` block active, which reuses an `ng serve` you already have running.

## Code Architecture

### Application Structure
Receipt Wrangler Desktop is an Angular 19 application with modular architecture using:

- **State Management**: NGXS store with persistent storage for application state
- **API Layer**: Auto-generated OpenAPI client in `src/open-api/` (do not manually edit these files)
- **Component Architecture**: Feature modules with lazy-loaded routing
- **UI Framework**: Angular Material + Bootstrap 5 + custom shared components

### Key Architectural Patterns

#### Module Organization
- Feature modules (receipts, dashboard, groups, etc.) with their own routing
- Shared UI components in `src/shared-ui/` for reusable elements
- Lazy-loaded modules for performance optimization
- Centralized store management with NGXS states

#### State Management (NGXS)
- All application state managed through NGXS store
- State persistence configured for key data (auth, user preferences, table states)
- Individual state files for each feature (receipt-table.state.ts, group.state.ts, etc.)
- Actions and state updates follow NGXS patterns

#### Component Structure
- Feature components organized by domain (receipts/, dashboard/, groups/)
- Shared UI components provide consistent design patterns
- Form components use reactive forms with custom validation
- Table components use base table service pattern for pagination and filtering

### Key Directories

#### Core Application
- `src/app/` - Main application module and routing
- `src/store/` - NGXS state management (18+ state files)
- `src/services/` - Application services and business logic
- `src/guards/` - Route guards for authentication and authorization

#### Features
- `src/receipts/` - Receipt management (forms, tables, processing)
- `src/dashboard/` - Customizable dashboard widgets and views
- `src/groups/` - Group management and member administration
- `src/categories/` and `src/tags/` - Receipt organization features
- `src/auth/` - Authentication and user management
- `src/roles/` - Role & permission management (admin-only Manage Roles UI)

#### Shared Infrastructure
- `src/shared-ui/` - 30+ reusable UI components (buttons, forms, tables, dialogs)
- `src/pipes/` - Custom Angular pipes for data transformation
- `src/utils/` - Utility functions and helpers
- `src/open-api/` - Generated API client (auto-generated, do not edit)

### Testing Strategy
- Unit tests use Jasmine/Karma framework
- Code coverage reporting with minimum thresholds
- Tests exclude auto-generated API code (`src/open-api/`)
- CI tests run in headless Chrome

### Development Environment
- Angular CLI 21 with TypeScript 5.9
- Bootstrap 5 + Angular Material for UI components
- NGXS for state management with Redux DevTools integration
- Strict TypeScript configuration with comprehensive compiler options

### Dependency Security & Version Pins

Keep `npm audit` at **0 vulnerabilities**. Two conventions exist specifically to hold that line —
do not undo them without re-checking `npm audit`:
- **`overrides` block in `package.json`** forces patched versions of build-time/dev-only transitive
  deps that the Angular toolchain otherwise pins inside vulnerable ranges (`@babel/core`, `esbuild`,
  `http-proxy-middleware`, `qs`, `undici`, `uuid`). When `npm audit` flags a new transitive advisory that
  the toolchain hasn't bumped yet, add/raise the floor here rather than waiting on an upstream release.
- **Exact pins (no caret):** `ngx-bootstrap` (`21.0.1`) and `@playwright/test` (`1.59.1`) are pinned
  because their next minor introduced an incompatibility (ngx-bootstrap dropped `CarouselModule.forRoot()`
  used by `src/carousel/`; Playwright `1.61` needs a newer browser build than the pinned `e2e:install`
  cache). Bump these deliberately — update the consuming code / reinstall browsers in the same change.
- Stay on the Angular `21.2.x` patch line for security fixes; a jump to Angular 22 is a separate,
  breaking upgrade and out of scope for audit hygiene.

**Never regenerate `package-lock.json` from scratch to fix an audit finding.** Deleting the lockfile
and reinstalling re-resolves *every* package to the newest version its range allows, which breaks CI
in two ways that `npm install` and `npm audit` both report as clean:

1. **`npm ci` refuses the result (`EUSAGE`).** A fresh resolve can *hoist* a package that the
   committed tree deliberately keeps nested. The real case: `pkijs` (via `selfsigned` ←
   `webpack-dev-server`) hard-depends on `@noble/hashes@1.4.0`, while `@exodus/bytes` (via `jsdom`)
   declares an *optional* peer `@noble/hashes@^1.8.0 || ^2.0.0`. The committed tree keeps 1.4.0 at
   `node_modules/pkijs/node_modules/@noble/hashes` and installs nothing at the root. A from-scratch
   install hoisted 1.4.0 to the root, where the optional peer then wants 2.4.0 — so `npm ci` fails
   with `Invalid: lock file's @noble/hashes@1.4.0 does not satisfy @noble/hashes@2.4.0` /
   `Missing: @noble/hashes@1.4.0 from lock file`. `npm install` accepts that lockfile happily; only
   `npm ci` rejects it.
2. **It drifts the tree past CI's Node floor.** CI pins Node **20.19.6** (`ci.yml`, `release.yml`,
   `e2e.yml`). A fresh resolve pulled `jsdom` 29→30, `whatwg-url` 16→17 and `@asamuzakjp/*`, all of
   which declare `node: ^22.x || >=24` and emit `EBADENGINE` on Node 20.

**Do this instead** — update the lockfile incrementally, and when npm refuses to move a
peer-coupled family in place (the Angular packages are one), delete just those entries and let it
re-resolve that subtree only:

```bash
# 1. start from the committed lockfile; edit package.json pins/overrides as needed
# 2. drop the entries npm will not move on its own: the peer-coupled family it
#    ERESOLVEs on, AND every `overrides` target (an override is ignored unless its
#    entry is re-resolved), at any nesting depth
python3 - <<'EOF'
import json, re
NAMESPACES = ('@angular/', '@angular-devkit/', '@ngtools/')   # the peer-coupled family
OVERRIDES  = ('@babel/core', 'esbuild', 'http-proxy-middleware',
              'qs', 'undici', 'uuid')                          # keep in sync with package.json
pat = re.compile(
    r'(^|/)node_modules/(?:' + '|'.join(map(re.escape, NAMESPACES)) + r')[^/]+$'
    r'|(^|/)node_modules/(?:' + '|'.join(map(re.escape, OVERRIDES)) + r')$')
d = json.load(open('package-lock.json'))
gone = [k for k in d['packages'] if pat.search(k)]
for k in gone:
    del d['packages'][k]
json.dump(d, open('package-lock.json', 'w'), indent=2)
open('package-lock.json', 'a').write('\n')
print(f'dropped {len(gone)} entries')
EOF
# 3. rebuild; npm re-resolves only what was dropped and leaves the rest of the tree pinned
rm -rf node_modules && npm install
```

**Always gate on `npm ci`, not `npm install`.** `npm ci` applies a stricter sync check, so a lockfile
that installs fine locally can still fail the build. Run the full gate before pushing any dependency
change — a lockfile can satisfy `npm ci` and still break the app, so the build and tests are part of
it, not an afterthought:

```bash
npm ci                    # real install, exactly what CI runs — must exit 0
npm audit                 # must report 0 vulnerabilities
npm run test:ci           # must pass
npm run build             # must succeed

# Every package's engines.node must admit the Node version CI pins (`node-version`
# in .github/workflows/ci.yml). semver is NOT a direct dependency — it resolves out
# of the tree the `npm ci` above installed, so this step must come last. After
# `npm ci --dry-run`, which writes nothing, it fails with MODULE_NOT_FOUND.
node -e '
const semver = require("semver");
const lock = require("./package-lock.json");
const NODE = "20.19.6";
const bad = Object.entries(lock.packages)
  .filter(([, p]) => p.engines?.node && !semver.satisfies(NODE, p.engines.node))
  .map(([k, p]) => k.replace("node_modules/", "") + "@" + p.version + " needs " + p.engines.node);
console.log(bad.length ? "INCOMPATIBLE:\n  " + bad.join("\n  ") : "engines OK for Node " + NODE);
process.exit(bad.length ? 1 : 0);
'
```

### API Integration
- Backend API proxied through development server
- OpenAPI client generated from backend specification
- API base path configurable through environment
- HTTP interceptors handle authentication and error responses

### Code Conventions
- SCSS for styling with component-scoped styles
- TypeScript strict mode enabled
- Angular style guide followed for component organization
- Lazy loading for feature modules to optimize bundle size

### Use Established Patterns (do not invent one-offs)
New UI MUST reuse the application's established patterns and shared components rather than
inventing a divergent, one-off implementation of something the app already standardizes — **unless
the user explicitly confirms the divergence**. Examples of standards to follow:
- **Form actions:** the floating save bar fixed to the bottom of the page — `<app-form>` (which wraps
  `app-form-button-bar` + `app-submit-button`), or, for bespoke layouts, a plain
  `<form (ngSubmit)="...">` ending in a standalone `<app-form-button-bar [mode]="...">` containing an
  `<app-submit-button>` (see `src/receipts/receipt-form/` for the bespoke-layout precedent). Do NOT
  place Save/Cancel buttons in the page header.
- **Form fields:** `app-input`, `app-textarea`, `app-select`, `app-checkbox`, grouped with
  `app-form-section`; bind via the `formGet` pipe.
- **Field hints:** `app-input` and `app-textarea` bind
  `[subscriptSizing]="hint ? 'dynamic' : 'fixed'"`, and that conditional is load-bearing both ways.
  Material's default `fixed` reserves a **single line** for the subscript and positions the hint
  wrapper **absolutely**, so a hint that wraps to two or more lines escapes its box and paints over
  whatever follows the field — which is exactly what a long hint in a narrow column does (the MCP
  section's "Connector sign-in lasts" hint used to cover the Connector URL block). A field with
  **no** hint keeps `fixed` so its error row stays reserved and it does not jump when a `mat-error`
  appears. Note `app-select` has **no** `hint` input at all — passing `hint="…"` to it silently
  renders nothing.
  - **A hint is in-flow content once the subscript is dynamic**, so it also feeds the field's
    intrinsic width. Inside a `d-flex`, `flex-grow-1` alone leaves `flex-basis: auto`, so a hinted
    column swallows the row and squeezes its neighbour. Base both columns at 0 instead — see
    `.duration-row` in `system-settings-form.component.scss`, which keeps the Session and MCP
    value + unit rows even and aligned with each other.
- **Margins on a shared control's host need `d-block`.** `app-checkbox` / `app-input` have no
  `:host { display: block }` and their SCSS is empty, so a bare `class="mb-3"` on the host is an
  **inline** element's vertical margin and does nothing. Pair it (`class="d-block mb-3"`).
- **Password fields:** `app-input` owns both password affordances as opt-in suffix icon buttons —
  `[showVisibilityEye]="true"` (the eye, `data-testid="password-visibility-toggle"`) and
  `[showGeneratePassword]="true"` (`data-testid="password-generate"`). Switching either flag on
  masks the field — on the initial binding *and* on a later `false -> true` flip, so a field that
  becomes a password field at runtime is never left in plain text. Generate fills the
  control with `generateSecurePassword()` (`src/utils/password.utils.ts` — `crypto.getRandomValues`
  with rejection sampling, one char per class, ambiguous glyphs excluded), reveals it, and copies it
  to the clipboard with a toast via `PasswordGeneratorService`
  (`src/services/password-generator.service.ts`). It is deliberately **admin-sets-someone-else's-
  password only** — the user form, Set Password, and Convert Dummy User dialogs — not sign-up or the
  fields that hold an existing external secret (system email, receipt-processing settings). The
  generate handler is synchronous by design: `type` is a plain `@Input`, so under zoneless CD only
  the click event's CD pass renders the reveal (the clipboard write is a detached side effect).
- **Tables:** `app-table`; **dialogs:** `app-dialog` + `app-dialog-footer`.
- **Confirming an action:** the shared **`ConfirmationDialogComponent`**
  (`src/shared-ui/confirmation-dialog/`). Open it with `matDialog.open(ConfirmationDialogComponent)`,
  set `componentInstance.headerText` / `componentInstance.dialogContent` (two plain `@Input()`s — it
  does **not** take `MAT_DIALOG_DATA`), then act on a truthy `afterClosed()`. Check truthiness, not
  `=== true`: a backdrop click or ESC closes with `undefined`. Every action that writes a record the
  user can't trivially undo gates on it — deletes, and the receipt **Duplicate** in both the
  receipts-table row action and the receipt view header (each creates a real receipt, and the table's
  then navigates away, so a mis-click is easy to miss).
- **Badges:** the shared standalone **`app-badge`** (`src/shared-ui/badge/`) — `<app-badge [text]="..."
  [tone]="...">`, a 9.5px uppercase micro-badge for marking an item in a list. Do NOT hand-roll one;
  see **The shared badge** below for the tones and the two traps.
- **Simple filters:** the segmented `app-filter-bar` (`src/shared-ui/filter-bar/`) — pass `FilterTab[]`
  (`{ value, label, icon?, count? }`) and two-way bind the selected `value`.
- **Breadcrumbs:** `app-breadcrumb` with `BreadcrumbItem[]`.
- **List-page headers & the "add" button:** every list page's header is `app-table-header`
  (`src/shared-ui/table-header/`), which takes `[headerText]` and an optional `[subtitle]` one-line
  description — every list should set a subtitle. For a subtitle that needs rich content (e.g. a link),
  project it via the `[table-header-subtitle]` slot instead of the string input (see
  `role-list`). The component owns its own vertical rhythm via a `:host` margin (a small gap above,
  a larger gap below to separate it from the table), so pages should **not** re-add margins around it.
  The primary create control is the shared **`app-add-button`**
  (`src/shared-ui/add-button/`): pass `[buttonText]="'Add X'"` to render a filled **`+ Add X`** button;
  omit `buttonText` and it stays the compact icon-only **`+`** used for in-form section-header adds
  (Add Item / Share / Widget / shortcut / option). Standardize the verb on **"Add X"**; give every add
  control a `<resource>-add` `data-testid` and a `tooltip`. Do NOT hand-roll a raw `app-button` for a
  list-page add action, and do NOT use a bespoke page-title header.
If a design appears to require a new pattern, confirm with the user before diverging.

### The chip picker's close-on-select preference

`app-autocomlete` in `[multiple]` mode (Categories, Tags, the user/group/icon pickers, the role
form's permission picker, report filters) deliberately **re-opens** its option panel after every
pick, so several values can be chosen in a row. `closeChipSelectOnSelect` on `UserPreferences` flips
that per user; the default is **off**, i.e. the panel stays open exactly as before.

Three things about it are load-bearing:

- **The component reads the preference itself**, from `AuthState.closeChipSelectOnSelect`, rather
  than taking an `@Input`. It is global by design — the setting says "all of these pickers" — and
  threading a flag through the twenty-odd call sites would only create a way for one of them to
  disagree with the rest. The default (`?? false`) lives in the selector, not at the reader.
- **`closePanel()` alone does not close it.** `MatAutocompleteTrigger` opens on `focus`, and
  Material returns focus to the input after a selection, so the panel comes straight back.
  `clearFilterAndClosePanel()` therefore also **blurs** `inputMultiple()`. This is also why the e2e
  asserts the blur *before* asserting the panel is hidden: Material closes the panel on selection by
  itself, so "hidden" alone would pass even if the preference were ignored entirely.
- **Removing a chip is deliberately NOT covered.** `removeOption()` refocuses the input, which opens
  the panel again regardless of the preference. The setting is about selecting, and a removal that
  left the field blurred would make removing several chips take an extra click each.

The panel handling runs inside the existing `setTimeout(…, 0)` in `optionSelected()`, so the two
tests in `autocomlete.component.spec.ts` that call `jest.runAllTimers()` are the only ones that
exercise it at all — every other `optionSelected` case stops at the `FormArray` push. They stub the
trigger (`textarea.component.spec.ts` is the precedent), because `CUSTOM_ELEMENTS_SCHEMA` leaves the
template's real `#auto` trigger unresolvable.

Because the component now injects `Store`, **any spec that instantiates a real `AutocomleteComponent`
needs `AuthState` registered** — `NgxsModule.forRoot([AuthState])`. That covers the base spec, the
`category-autocomplete` / `tag-autocomplete` / `grant-picker` specs, and `system-settings-form`
(which registered only `SystemSettingsState`).

`desktop/e2e/chip-select-close-on-select.spec.ts` is the wire test: the Jest specs drive the
component against a mocked store, so only the e2e proves the checkbox reaches the picker through
`PUT /userPreferences` -> the stored column -> AppData -> `AuthState`. It **provisions its own
account**, because the preference is per-user and global and the suite runs `fullyParallel` —
flipping it on a shared e2e account would change behavior under every spec running as that account
at the same time.

### The shared badge (`app-badge`)

`src/shared-ui/badge/` — a small uppercase badge (`text` + `tone` signal inputs) used to mark an item
in a list. It replaced two hand-rolled copies with identical geometry and **different colours**, which
is exactly the drift a shared component prevents: `app-select`'s option badge was slate, the report
panel's custom badge purple. Both are now purple, so **"Custom" looks the same everywhere** — the
report builder's dropdowns and picked rows, and the receipts table's Configure Columns dialog.

`CUSTOM_FIELD_BADGE` ("Custom") is exported from the component file, so the string has one home.

Three things about it are load-bearing:

- **The class is `.rw-badge`, NOT `.badge`.** Bootstrap is a global stylesheet here
  (`angular.json` builds `bootstrap-scss/bootstrap.scss`) and defines an unscoped `.badge` with its
  own `line-height`, `text-align`, `white-space` and a **white** `color`. A component rule only
  overrides what it *declares*, so everything else leaks — measured as a ~44% height change — and
  Bootstrap's badge is genuinely used elsewhere (`dashboard/pie-chart`), so it can't just be dropped.
  `rw-` matches the existing `.rw-chip` / `.rw-card` convention. **Any new shared component must check
  for a Bootstrap collision on its class names.**
- **Tones are opaque, and that is the point.** Both originals used a translucent fill, and both
  carried a comment that at 9.5px/700 the text is "small text" under WCAG and needs 4.5:1 against its
  *composited* background. A translucent fill makes contrast depend on whatever the badge is dropped
  onto, which a shared component cannot know — so each tone is the pre-composited opaque colour.
  `purple` 6.43:1, `slate` 6.22:1, `blue` 5.90:1, `green` 5.88:1. **`blue`/`green` are darker than the
  report panel's original kind badges**, which never got the contrast pass the custom badge did (they
  measured 2.61:1 and 3.03:1); the row's `.kind-*` **icon chip** keeps its lighter translucent fill,
  so chip and badge in the same row are deliberately not the same shade.
- **`flex: none` lives on `:host`, not the inner span** — the host is the flex item, so on the span it
  silently does nothing. The badge carries **no margin**: the two flex call sites space it with `gap`,
  and only `app-select` (where it follows inline text) adds `margin-left`, in its own stylesheet.

Wiring: registered in `SharedUiModule`'s `imports` + `exports` (the `LoginQrComponent` pattern), which
covers `ReceiptsModule` and `ReportsModule`. **`SelectModule` imports the component directly** —
`SharedUiModule` imports `SelectModule`, so reaching it the other way would be a circular import.

Two call-site rules the e2e suites depend on:
- **Keep the badge inside the `mat-option`'s text content.** `e2e/report-custom-fields.spec.ts`
  matches option accessible names with `^${name} Custom$`, and `e2e/helpers/reports.ts`
  (`addGroupingLevel`) documents the same coupling.
- **Keep it OUT of a `mat-checkbox`'s label.** In the Configure Columns dialog it is a sibling of the
  checkbox, so the checkbox's accessible name stays the bare field name — otherwise every locator that
  picks a column by its field name stops resolving. (`.column-checkbox { flex: 1 }` then pushes the
  badge to the right of the row, which is why it reads as a tidy right-hand column there.)

### Roles & Permissions (Manage Roles)

The admin-only **Manage Roles** feature (`src/roles/` — `role-list`, `role-form`, `role-presets`,
`roles.module`) provides CRUD for the backend's app- and group-scoped roles (see the backend
"Roles & Permissions" section in `api/CLAUDE.md` for the permission model). The Manage Roles routes are
gated by `appPermissionGuard` requiring `app.roles.read` (see **Permission-based UI gating** below).

- Talks to the backend via the **generated** clients in `src/open-api/` — `RoleService` for role
  CRUD and `PermissionService` (`GET /permission`) to load the permission catalog that populates the
  role editor's permission picker. `role-presets.ts` holds the role templates. Never hand-edit
  `src/open-api/` — regenerate it from `swagger.yml` instead.
- Built on existing shared patterns: `app-breadcrumb` and the segmented `app-filter-bar` (see "Use
  Established Patterns" above).
- **Category/tag grants (group roles only):** the role editor shows a "Category & tag access" section
  (gated on `showGrants()` = group scope) with `app-category-autocomplete` / `app-tag-autocomplete`
  (both fed the full pool via `CategoryService.getAllCategories()` / `TagService.getAllTags()` — the
  editor is admin-only). Selecting grants restricts members to those categories/tags; **empty = all**.
  The selections drive `FormArray`s loaded from `role.categoryGrants` / `role.tagGrants` (resolved to
  pool objects by an effect once the pool arrives) and serialized back as id arrays on
  `UpsertRoleCommand` for group scope only. The grant pickers pass `[creatable]="false"` (pick from
  existing, never create). See `api/CLAUDE.md` → "Data model".
- **Shared grant picker (`src/shared-ui/grant-picker/`):** the category/tag grant UI is a standalone
  `app-grant-picker` used by **three** forms — `role-form` (a group role's grants), `user-form` and
  `group-member-form` (an individual member's assignment). It owns the two autocomplete `FormArray`s
  and the pool→id resolution effect, takes `[selectedCategoryIds]`/`[selectedTagIds]` and emits
  `(grantsChange)` with the current ids. An optional `[ceiling]` (`GrantCeiling`) **filters the offered
  pool** and shows a hint naming the constraint — filtering rather than per-option disabling because
  the shared `app-autocomlete` has no per-option disable support and adding one would touch a
  component used app-wide. **It emits nothing while merely seeding itself** (`emitEvent: false`), so a
  host must not treat "no emission" as "empty selection" — `role-form` uses a `linkedSignal`
  (`grantSelection`) defaulting to the loaded role's grants, or an untouched save would wipe them.
  `shared-ui/grant-picker/member-grant-assignment.ts` holds the shared row-building
  (`buildMemberGrantRows`, `ceilingForRole`) and the diff-and-write helper
  (`saveChangedMemberGrants`), so the two member-facing entry points cannot drift.
  - **It stops seeding once the user edits.** The effect's inputs keep changing after mount (the
    pool arrives async; the ceiling changes again when the host's role list resolves), so without
    that guard a late re-run silently discards the user's selection and hands the host back the
    original value.
  - **Each instance passes a unique `inputId`** to the autocompletes. The base `app-autocomlete`
    otherwise derives the input's DOM id from its label, so the user form's N pickers would all
    render `id="categories"` — which breaks `<label for>` association for every field after the
    first and misdirects the base component's `getElementById`-based filter clear to the first
    instance. `app-category-autocomplete` / `app-tag-autocomplete` gained an `inputId` passthrough
    for this.
- **Per-member category/tag assignment (`user-form`, `group-member-form`):** grants hang off a group
  **membership**, so both forms only offer the picker for a membership that already exists on the
  server — `user-form` renders one section per group the **edited** user belongs to (nothing in add
  mode), and `group-member-form` shows it only when editing an already-persisted member. They write
  through the dedicated **`PUT /group/{groupId}/member/{userId}/grants`** endpoint (its own permission,
  `group.members.grants.update`), **not** the group-member upsert — so grants are saved independently
  of the parent form and the two entry points cannot clobber each other via `UpdateGroup`'s wholesale
  roster replace. Only **changed** rows issue a request. The role's own grants supply the picker's
  ceiling; the backend re-validates and 400s an out-of-ceiling id. See `api/CLAUDE.md` →
  "Category/tag grant resolution".
  - **Gated on `group.members.grants.update`, per group.** The endpoint declares **no**
    `OrAppPermissions` bypass, so holding `app.users.update` (user form) or `group.members.update`
    (member dialog) is not enough. `user-form` filters `grantRows()` to the groups the admin holds it
    in — so the section disappears when none qualifies — and `group-member-form` wraps its block in
    `*hasGroupPermission`. Without this both forms render pickers whose saves can only 403.
  - **The "leave empty" guidance is conditional.** `MemberGrantRow` carries
    `requiresIndividualCategories`/`requiresIndividualTags` from the role, and the shared
    `emptySelectionHint(row)` turns them into the right sentence: normally empty means "everything the
    role allows", but under a require-individual role empty means **nothing**. Both forms render that
    helper rather than hard-coded text — the wording is the only thing that tells an admin which rule
    is in force, so it must not drift between the two entry points.
  - **Reporting the outcome:** because the grants are a *second* write that can fail on its own, both
    forms fire their success toast only **after** it lands, and a failed write leaves the dialog open
    (a trailing `catchError` — the interceptor already reports the error) so the admin can correct the
    selection instead of losing it. `user-form` is the only submit in the app with two writes, so this
    is where the general "toast in a `tap` on the write it describes" convention needs stating: the
    message must speak for *all* the writes it covers, not just the first.
- **Receipt requirements (`role-form`, group roles only):** "Require a comment / an image on
  receipts". See **Role-required receipt fields & single-call create** below.
- **Require-individual-assignment toggles (`role-form`, group roles only):** two `app-checkbox`es in
  the grants section bound to `requiresIndividualCategoryGrants` / `requiresIndividualTagGrants`. When
  on, a member of that role with no individual assignment sees **nothing** rather than the role's set,
  so a newly added member is never exposed by default. Default off — existing roles are unchanged.
- **Paid-by visibility (group roles only):** the same group-scoped grants section also shows a
  "Paid-by visibility" picker — a single `app-autocomlete` multi-select over `paidByOptions()` (a
  pinned **"Their own receipts"** sentinel option, id `OWN_PAID_RECEIPTS_OPTION_ID = -1`, followed by
  every user from `UserState.users`). On submit the selections split into `includeOwnPaidReceipts`
  (the sentinel is present) and `paidByUserGrants` (the remaining user ids, sentinel excluded); on
  edit they rehydrate from `role.paidByUserGrants` / `role.includeOwnPaidReceipts` via an effect that
  filters the shared `paidByOptions()` (stable references so the autocomplete excludes selected
  options). Empty = members see every payer's receipts; it restricts which receipts a member can see,
  not what they can edit. See `api/CLAUDE.md` → "Paid-by visibility enforcement".
- **Report template access (group roles only):** the same grants section shows a "Report template access"
  matrix — templates (rows, from `ReportService.getReportTemplateOptions()`, gated on `app.roles.read`) ×
  actions (View/Generate/Edit/Delete/Duplicate columns) of `rw-switch` toggles, plus a per-row "All"
  toggle. State is a `signal<Map<number, Set<string>>>` (immutable replace for zoneless CD, mirroring the
  permissions grid's `Set` pattern — NOT a FormArray). **All-empty = unrestricted**; a template maps to the
  subset of actions the role may perform on it. Hydrates directly from `role.reportTemplateGrants`,
  serializes back for group scope only, resets on `pickType`. See `api/CLAUDE.md` → "Report-template access".
- **Group creation (app roles only):** an app-scope-only **"Group creation"** `rw-card` in
  `role-form` (gated on `showAppOptions()` = `type() === "app"`, the mirror of `showGrants()`) holds a
  single `app-checkbox` — "Don't create a personal group for new users with this role"
  (`data-testid="skip-default-group"`) — bound to `skipDefaultGroupCreation` on the
  `UpsertRoleCommand`. New users normally get a personal "My Receipts" group; turning this on skips it
  for accounts that should only belong to groups an admin adds them to (the virtual "All" group is
  always created, so the dashboard still works). It mirrors `seesAllMembers` exactly but in the
  opposite scope: hydrates on edit, resets in `pickType`, and serializes in the **`else`** branch of
  `submit`'s `if (showGrants())` so it only ships on APP scope. Creation-time only — toggling it never
  changes an existing user's groups. See `api/CLAUDE.md` → "Skipping the personal group per app role".
- **Default roles:** the role-list page shows two `app-select` controls above the filter bar —
  "Default application role" and "Default group role". Each is pre-selected from the role flagged
  `isDefault` for its scope and, on change, calls `RoleService.setDefaultRole(scope, roleId)` then
  reloads (setting one default clears the previous one). The default role per scope is what new
  accounts / group creators receive (see `api/CLAUDE.md` → "Default roles"); the current default
  cannot be deleted. Default rows also carry a "Default" badge next to the System badge.
- **Modern role assignment in authoring forms:** the user add/edit form (`src/user/user-form/`) and
  the group-member add/edit form (`src/group/group-member-form/`) assign **modern roles** — an
  `app-select` of `RoleService.getRoles()` filtered to the `APP` / `GROUP` scope, bound to
  `appRoleId` / `groupRoleId` (not the legacy enums). Add forms pre-select the configured default
  role. Each selector has a Preview icon button (`data-testid="role-preview"`) that opens the shared
  **`RolePreviewDialogComponent`** (`src/roles/role-preview/`, standalone, opened via
  `openRolePreviewDialog(dialog, role)`) — a read-only dialog rendering the role's scope, description
  and permissions (grouped by resource using the `role-presets.ts` helpers). The `getRoles()` calls
  use `catchError` so a non-admin who lacks `app.roles.read` (see **Permission-based UI gating** below)
  gets an empty selector rather than an error.
- **Member-table role display:** the admin user-list (`src/user/user-list/`) and `group-form`
  (`src/group/group-form/`, loaded by every group view including non-admins) resolve each user's
  `appRoleId` / each member's `groupRoleId` to the role **name** via the shared `RoleNamePipe`
  (`src/pipes/role-name.pipe.ts` — `{{ id | roleName : roles() : scope }}`). The pipe matches on
  **id _and_ `PermissionScope`** because app- and group-role ids are independent sequences and can
  collide (e.g. group role `id=1` vs app role `id=1`); callers pass `PermissionScope.App` (user-list)
  / `PermissionScope.Group` (group-form) so the wrong-scope role is never matched. Both load roles with
  `RoleService.getRoles()` wrapped in `catchError`, so a non-admin lacking `app.roles.read` sees a
  blank name rather than a 403-driven logout. There is no longer any legacy `groupRole` enum sync,
  and `group-form` no longer enforces a "keep an owner" rule — the backend dropped the owner concept,
  so group management is governed entirely by `group.*` permissions.
  - **Member controls gate on `group.members.*`:** `group-form`'s Add / Edit / Delete member controls
    render only for holders of `group.members.create` / `.update` / `.delete` (signals resolved from
    `AuthState.hasGroupPermission`), mirroring the backend `UpdateGroup` guard (see `api/CLAUDE.md` →
    "Group member management"). They default to enabled in **create** mode (no group yet; the creator
    becomes owner). This is UI-only; the server re-enforces on every request.
- **Manage Users is a server-paged `app-table`:** the admin user-list (`src/user/user-list/`) follows
  the standard paginated-table pattern (`BaseTableComponent` + `UserTableService` + the NGXS
  `UserTableState`, mirroring the groups/roles list pages) and reads a page at a time from
  `UserService.getPagedUsers` (`POST /user/getPagedUsers`) — it no longer wraps `UserState.users` in a
  client-side `MatTableDataSource`. Server-sortable columns use the DB column names as their
  `matColumnDef` (`username`, `display_name`, `created_at`, `updated_at`); the client-resolved **Role**
  column is non-sortable. CRUD row actions refetch via `getTableData()` after the mutation, and delete /
  bulk-delete still dispatch `RemoveUser` / `RemoveUsers` so the `UserState` cache (consumed by the
  role-editor & report-builder paid-by pickers, `user-autocomplete`, the `user` pipe, etc.) stays in
  sync. `UserState` itself is unchanged and still bootstrap-loaded from AppData for those other consumers.
- **Member isolation (presence privacy).** Two config controls drive the backend member-isolation
  feature (see `api/CLAUDE.md` → "Member isolation"): (1) an **"Isolate members"** `app-checkbox` in the
  `group-form` "Group Details" section, bound to `isolateMembers` on the `UpsertGroupCommand` — an
  isolated group's members can't discover each other; (2) a group-scope-only **"Members with this role
  can see, and be seen by, all members"** `app-checkbox` in `role-form` (a "Member visibility" `rw-card`
  inside the `@if (showGrants())` block), bound to `seesAllMembers` on the `UpsertRoleCommand` (mirrors
  `includeOwnPaidReceipts`; hydrates on edit, resets on type switch, serialized only for GROUP scope).
  E2E coverage for the group-creation card lives in `e2e/skip-default-group.spec.ts` (card present on
  APP / absent on GROUP, reset on type switch, round-trip through save including turning it back off,
  the end-to-end effect on a provisioned account, and the server's 400 on GROUP scope).
  Isolation is resolved **per group** on the backend ("isolated means isolated" — an isolated group hides
  co-members and their settlement/report data regardless of any other group you share; co-members are
  visible only through a shared **non-isolated** group). Everything is enforced **server-side** — the
  desktop simply receives the already-filtered `appData.users` (union name-table) and per-group filtered
  group rosters / receipts / Group-Summary settlement, so no client-side filtering is needed (and mobile
  needs no change). This works because every group-context user picker sources its SET from the group
  roster (`group.groupMembers`) or `roster ∩ flat-list`, using the flat `UserState.users` only to resolve
  a name/avatar by id (see `GroupMemberUserService.getUsersInGroup`). The larger `UserAccessService` /
  `UserView`-retype consolidation is a deferred follow-up.
- **Permission-based UI gating.** The UI gates on the user's effective permissions, mirroring the
  backend's enforcement. Permissions are delivered on **AppData** (`appPermissions: string[]` and
  `groupPermissions: { [groupId]: string[] }`) and stored in `AuthState` via the dedicated
  `SetPermissions` action — dispatched **only** from `setAppData` (`utils/app-data.utill.ts`), never from
  `TokenRefreshService` (whose claims-only `SetAuthState` must not wipe permissions). They refresh on
  login + app-init; the server re-checks real permissions on every request, so the stored set is a UI
  hint (a stale button at worst 403s, handled by the interceptor).
  - **`string[]`, not the `Permission` enum — on purpose.** `swagger.yml` types these two `AppData`
    fields as plain strings rather than `$ref`-ing the `Permission` enum, so the generated
    `appData.ts` yields `Array<string>`. The enum stays the contract for the *catalog*
    (`Role.permissions`, `UpsertRoleCommand.permissions`, `PermissionDescriptor.key` — the role
    editor still gets an exhaustive, type-safe list). The desktop is unaffected either way (TS string
    enums are assignable to `string`), but the **mobile** Dart client is not: a closed built_value
    enum throws on an unknown value and fails the whole `AppData` parse, hard-failing login on
    already-released builds. Also, a granted string may be a **wildcard** (`app.*`), which the
    matcher supports and an enum cannot express. Don't "tighten" these back to `Permission`.
  - **403 handling (`src/interceptors/http-interceptor.ts`).** The backend returns **403 for every
    access denial** (auth *and* permission — it never uses 401). With a still-valid token a 403 is a
    permission denial, so the interceptor surfaces it via a **Forbidden toast** (only for
    user-initiated mutations — non-`GET` — and never in `queueMode`) and **re-throws without logging
    the user out**. It does **not** refresh/retry on 403: token freshness is handled proactively
    elsewhere (15-min timer in `app.component.ts`, app-init, and `auth.guard`), and
    `TokenRefreshService` keeps its own logout-on-refresh-failure path for a truly dead session.
    Background `GET` 403s propagate silently for callers to handle (e.g. the `getRoles` +
    `catchError` reads above).
  - **One toast per error, and the server's `errorMsg` always wins.** `MatSnackBar.open()`
    dismisses whatever is already showing, so two `snackbarService.error(...)` calls in one tick
    means only the *last* is readable. The server-supplied `errorMsg` is the user-friendly message;
    Angular's generic `HttpErrorResponse.message` ("Http failure response for ...: 500 Internal
    Server Error") is a **fallback only** for a 5xx that carries no message at all, so an
    infra-level failure (nginx 502, gateway timeout) is not silent. The two branches must stay
    mutually exclusive — **do not restore them as two independent `if`s.** They were, and the
    second one only stayed harmless because its regex was broken: `new RegExp("5d{2}")` matches the
    literal `5dd`, never `"500"`. Repairing it to `5\d{2}` (commit `4bb640c`, 2026-03-23) activated
    the override, and since nearly every Go handler writes 5xx through `WriteCustomErrorResponse`
    with an `errorMsg`, the useful message was being replaced app-wide — most visibly on a failed
    login, where "Invalid credentials." lasted a few milliseconds. `queueMode` suppresses **both**
    branches, like the 403 one. Pinned by three cases in `http-interceptor.spec.ts`
    (5xx-with-`errorMsg`, 5xx-without, and 5xx-in-queue-mode) and by
    `e2e/auth.spec.ts` → "a wrong password reports it and leaves the form filled in".
  - **Category/tag catalogs:** AppData also carries `groupCategories` / `groupTags` (keyed by group
    id, filtered to the user's grants), stored via `SetGroupCatalog` and read with the
    `AuthState.groupCategories(groupId)` / `groupTags(groupId)` selectors. The **receipt form** and
    **receipts-table filters** source their category/tag options from these per-group catalogs (not
    the global `GET /category` / `GET /tag`, which are now admin-only — the receipt routes no longer
    use the category/tag resolvers). The receipt form's pickers gate `[creatable]` on
    `app.categories.create` / `app.tags.create` so restricted users can only pick from the granted set.
  - **Matcher:** `src/utils/permission.utils.ts` — `matches`/`hasAll`/`hasAny`, a faithful port of the Go
    matcher (`api/internal/permissions/matcher.go`) including wildcard semantics, so UI gating === backend.
  - **Selectors** (`AuthState`): `hasAppPermission(perm)`, `hasAnyAppPermission(perms)`,
    `hasGroupPermission(groupId, perm, orApp = [])` — the group one applies the `orApp` app-scoped
    override first, mirroring the backend `OrAppPermissions` (admin-not-a-member) pattern.
  - **Directives** (`DirectivesModule`, signal/`effect`-driven so they re-render when AppData lands after
    first paint): `*hasAppPermission="Permission.X"` and
    `*hasGroupPermission="{ groupId, permission, orApp? }"`. Components expose the generated `Permission`
    const to reference it in templates.
  - **Route guards** (`src/guards/`): `appPermissionGuard` (`data: { appPermissions: [...] }`, ANY-of)
    and `groupPermissionGuard` (`data: { groupPermission, orAppPermissions?, useRouteGroupId? }`).
    `receiptGuardGuard` is unchanged (server-checked per-receipt access); `system-settings-landing.guard`
    redirects `/system-settings` to the first tab the user can read, and `settings-landing.guard`
    does the same for `/settings` (the avatar-menu "User Settings" link → first readable of
    User Profile `app.account.read` / User Preferences `app.user-preferences.read` / API Keys
    `app.api-keys.read`). The `/settings` shell and each tab route are `appPermissionGuard`-gated on
    those reads, the in-page tabs render conditionally on the same, and the avatar-menu button is
    gated by a `hasAnyAppPermission` signal.
  - **Retired** with this migration: `RoleGuard`, `GroupRoleGuard`, the `*appRole` `RoleDirective`, the
    `groupRole` `GroupRolePipe`, and `GroupUtil.hasGroupAccess`. The group-member legacy-enum bridge
    (`legacyGroupRoleFromRole`) and `AuthState.userRole`/`hasRole` are now **removed** as well, since
    the backend legacy `UserRole`/`GroupRole` enums (and the `userRole`/`groupRole` API fields) are gone.
  - **Behavior note:** create actions for categories/tags/custom-fields now gate on the granular
    `.create` permission, so a normal user (Legacy User holds `.create`) sees the **Add** button;
    **Edit/Delete** stay admin-only (`.update`/`.delete`). **Group creation** follows the same shape:
    the Create-Group FAB on the groups list (`group-table`), the sidebar speed-dial "Add Group"
    button, and the `/groups/create` route guard all gate on `app.groups.create`. Note the
    read/create asymmetry — Legacy User holds `app.groups.create` but **not** `app.groups.read`, so
    they create via the sidebar FAB (the groups-list page itself is `app.groups.read`-gated and off
    limits to them), exactly like categories/tags.
  - **Dashboard CRUD** (`group-dashboards.component.html`): the Add / Edit / Delete dashboard buttons
    gate on `group.dashboards.create` / `.update` / `.delete` via `*hasGroupPermission` (the group id
    comes from a `selectedGroupIdNum` computed). Previously ungated — the buttons rendered for every
    member and 403'd on the backend; now they only render for holders, matching the receipts-table.
  - **Notification delete** (`notification/notification.component.html`): the per-notification delete
    control gates on `app.notifications.delete` via `*hasAppPermission`.
  - **Group row actions** (`group/group-table/group-table.component.html`): edit gates on
    `group.update` OR `app.groups.update-settings`; **delete** gates on `group.delete` OR
    **`app.groups.delete`** (`orApp`), so an admin who switched the filter to **All Groups** can clean
    up a group they aren't a member of. Two things follow from that view being server-paged over
    groups the caller may not belong to: the button's `[disabled]` is a `deleteDisabled()` computed —
    `!canDeleteAnyGroup() && groups().length <= 1`, mirroring the backend `CanDeleteGroup` rule and
    its `app.groups.delete` escape, rather than blanket-disabling every row for a one-group admin —
    and `deleteGroup()` refetches with **`getTableData()`** after a successful delete (the standard
    row-mutation pattern) instead of swapping in `GroupState.groupsWithoutAll`, which would collapse
    the table to the caller's own groups and lose pagination. The `RemoveGroup` dispatch stays, to
    keep the `GroupState` cache in sync; it no-ops for a group the caller isn't in. E2E:
    `e2e/group-delete-any.spec.ts`.


### Reporting a field-keyed 400 (`applyApiErrors`)

The API writes errors in two shapes, and only one of them reaches the user on its own.
`WriteErrorResponse` / `WriteCustomErrorResponse` emit `{ errorMsg }`, which
`interceptors/http-interceptor.ts` toasts **at any status**; `WriteValidatorErrorResponse` marshals a
bare `{ <control>: <message> }` map, which the interceptor has no branch for (it is not a 5xx, so the
generic fallback does not fire either). A 400 of that second shape is therefore **completely silent**
— no toast, no inline error — unless the component that issued the request reports it.

**`applyApiErrors(form, error.error, onUnmatched?)`** (`src/utils/form.utils.ts`) is that report. Call
it from a `catchError` on the save, returning `EMPTY`. It is used by the System Settings form and the
OIDC provider dialog; the three hard-coded `err.error["username"]` reads in `auth-form.component.ts`,
`auth-form.util.ts` and `user-form.component.ts` predate it and are the obvious next callers.

Four things about it are load-bearing, each of which the original (dead, never-called) version got
wrong:

- **A body carrying `errorMsg` is skipped entirely**, because the interceptor already toasted it.
  This is what keeps the two mechanisms from both firing — and a double report *costs* the user the
  message rather than repeating it, since `MatSnackBar.open()` dismisses whatever is showing.
  `http-interceptor.spec.ts` → "stays silent on a 400 that carries no errorMsg" pins the other half
  of that contract; if the interceptor ever gains a validator-body fallback, every call site here
  starts reporting twice.
- **The message is the error's *value*, under the key `server`.** That key is deliberately absent
  from `BaseInputComponent.errorMessages`, so it takes the `typeof value === "string"` branch and
  renders the server's sentence verbatim — the same idiom `durationValueValidator` and
  `providerNameValidator` use.
- **It calls `markAsTouched()`.** Material's default `ErrorStateMatcher` is `invalid && (touched ||
  form.submitted)`, so `setErrors` alone leaves a control the user never focused looking perfectly
  fine.
- **A key matching no control goes to `onUnmatched`**, not to the floor — a silent 400 is the bug
  this exists to fix. It also guards a body that is not a flat object at all (a network failure gives
  `null`, an nginx 502 gives HTML).

**It cannot help a control that renders no errors.** `app-checkbox` has no `mat-error` and no
`mat-form-field` to host one, and its `additionalErrorMessages` input is never read — so a server
error keyed to a checkbox is invisible no matter what. That is why the backend sends the
"configure a server public URL first" refusal as an `errorMsg` instead of keying it to `enabled`; see
`api/CLAUDE.md` → "OIDC (OpenID Connect) login" → Configuration.

### Receipt status colors

Everything visual for a receipt status is the ~20 lines of
`src/shared-ui/status-chip/status-chip.component.scss` plus the `[ngClass]` map in its template. The
options themselves need no code: `RECEIPT_STATUS_OPTIONS` (`src/constants/receipt-status-options.ts`)
derives from `Object.keys(ReceiptStatus)` and `formatStatus` (`src/utils/status.utils.ts`)
snake→title-cases the label, so **a status added to `swagger.yml` reaches the receipt form, the
filters, bulk status update and both "default status" settings on regeneration alone** — only a color
has to be authored.

The convention is a **pale tint carrying the default (dark) `mat-chip` text**:

| Status | Background |
|---|---|
| `NEEDS_ATTENTION` | `variables.$warning-amber` (`#ffe0b2`) |
| `DECLINED` | `map.get(variables.$warn-palette, 100)` (`#f2bfbf`) — the red NEEDS_ATTENTION gave up |
| `OPEN` | `#fffacd` |
| `RESOLVED` | `lightgreen` |
| `DRAFT` | *(no rule — the default chip surface)* |

**Do not set an explicit `color` on these rules.** `.needs-attention` and `.open` used to declare
`color: white`, i.e. white on `#f2bfbf` / `#fffacd` — about 1.6:1 and 1.2:1, an effectively invisible
label. The tints are chosen to be legible under the dark default.

The chip is also driven by an unrelated `customStatusColor` input (`"red" | "green" | "gray" |
"yellow"`) for **system-task** status in the activity list; `'red'` maps to `.declined`, so red still
means failure there. (`'gray'` maps to no class — a pre-existing dead value.)
`status-chip.component.spec.ts` pins the whole status → class map, including that repoint.

### Custom fields

The Manage Custom Fields page (`src/custom-fields/`) is a paged `app-table` plus a single
`app-custom-field-form` **dialog** serving all three modes. It follows the categories/tags shape (one
dialog, table refetches on a truthy `afterClosed()`), with the differences a custom field's data model
forces:

- **The dialog is `FormMode`-driven** (`@Input() mode`, defaulting to `FormMode.add`), not the old
  binary `readonly` flag — passing a `customField` used to make it permanently read-only, so there was
  no way to reach an editable state. Fields bind `[readonly]="mode | inputReadonly"`, and the footer
  renders for every mode but `view`.
- **The Type select is locked outside add mode** (`typeReadonly`, i.e. `mode !== FormMode.add`) — in
  *edit* as well as view. A `CustomFieldValue` lives in a type-specific column, so re-typing would
  mis-column every stored value; the server 400s it too (see `api/CLAUDE.md` →
  `app.custom-fields.update`).
- **A saved option can be renamed but not removed** (`canDeleteOption`): `CustomFieldValue.SelectValue`
  holds an option id, so the per-option delete button renders only for an option added in this session
  (no `id` yet). New options are submitted **without** an id, which the server reads as "append";
  existing ones carry their real id so a rename keeps every receipt's selection resolving. `buildOption`
  therefore stores the **real server id** (or `null`) instead of the old `Math.random()` placeholder,
  and the `@for` tracks `$index` — correct because the inner control is already bound by index.
- **Every option value is required**, via the shared `trimmedRequiredValidator()`
  (`src/validators/text-validators.ts`) attached to the option `value` control in `buildOption()` and
  held as a module-level const (the `quick-scan-dialog` precedent). Reused rather than
  `Validators.required` because the API blank-checks after trimming, so `"   "` would otherwise pass
  the form and come back a 400. It emits the standard `required` key, so `app-input` renders "Value is
  required." with no `additionalErrorMessages`. This is what stops a rename to blank — which would keep
  the option id and leave every receipt that selected it with no label.
- **Table row actions:** an `app-edit-button` (`data-testid="custom-field-edit"`) gated on
  `Permission.AppCustomFieldsUpdate`, beside the existing delete. The **name link** opens the dialog in
  edit mode for an update-holder and read-only for everyone else (`resolveDialogMode`). The actions
  column is gated on `hasAnyAppPermission([...Update, ...Delete])` — gating it on delete alone would
  hide the whole column, and with it the new edit button, from an update-only holder.
- `setColumns()` reads permissions with a one-shot `selectSnapshot` from `ngAfterViewInit`, so a spec
  must dispatch `SetPermissions` **before** the first change-detection pass (in the app the route guard
  has already loaded AppData).

E2E: `e2e/custom-field-edit.spec.ts` (two app roles differing only in **Update Custom Fields**; the
holder renames a field through the dialog and it persists with the type untouched, the type control is
non-editable and saved options expose no delete while an appended one does, the non-holder sees no
control and its direct `PUT /api/customField/:id` 403s, and a type change 400s even for the holder).
`e2e/legacy-user-visibility.spec.ts` pins that Legacy User sees neither edit nor delete.

### Custom fields as receipts-table columns

The **Configure Columns** dialog (`src/receipts/column-configuration-dialog/`) lists every custom
field after the ten built-in columns, and each one can be turned into a sortable table column.

- **`custom_<id>` is the column's `matColumnDef` *and* the `orderBy` sent to the API** — the same key
  the reporting engine uses (`receiptsource.CustomFieldKey`). `src/utils/receipt-table-columns.ts`
  owns that convention (`customFieldColumnDef` / `parseCustomFieldColumnDef`) plus
  `RECEIPT_COLUMN_DISPLAY_NAMES`, which the dialog label and the table header now **share** instead of
  each hard-coding its own literal.
- **Hidden by default.** A newly created custom field is appended to the list unchecked, so it never
  silently widens everyone's table.
- **`mergeCustomFieldColumns` reconciles the persisted configuration with the live catalog**, and both
  the dialog and the table run it. The configuration lives in **localStorage** (`receiptTable` is in
  `ngxsStorageKeys`) and is shared by every account on the browser, so it outlives the catalog it was
  written against: a column for a deleted custom field is dropped, a new field is appended hidden, and
  the persisted order — including a custom field dragged above a built-in — is otherwise preserved.
  `resetToDefaults()` restores the built-in defaults but keeps the custom fields listed, since they
  would only reappear on the next open anyway.
- **`reconcileColumnConfig()` runs in `ngOnInit`, before the first fetch**, and also resets a
  persisted `orderBy` that no longer resolves. That ordering is load-bearing: the sort is persisted
  too, so without it the *first* request asks the API to order by a column that no longer exists.
  (The backend falls back rather than erroring, but the table would then disagree with its own
  headers.)
- **`displayColumns` is derived from the columns that actually resolved**, not from the stored
  configuration. `mat-table` throws on a displayed id it has no definition for, and with custom fields
  a stale id is ordinary rather than hypothetical.
- **The permission gate is the resolver.** `customFieldResolverFn` is wired onto the
  `receipts/group/:groupId` route and already returns `[]` without `app.custom-fields.read`, so such a
  user simply has no custom field columns. No new permission code.
- **An empty catalog means "may not look", NOT "none exist" — and the difference is destructive.**
  Because the configuration is persisted per browser and **not namespaced per account**, treating the
  permission stub as an authoritative empty catalog deletes every `custom_*` entry the configuration
  holds: an administrator's saved layout is gone the moment a colleague without the permission opens
  the page on the same machine. So `mergeCustomFieldColumns` takes a **required** third argument,
  `catalogAvailable`, and keeps the persisted custom columns untouched when it is false (built-ins are
  still healed and order re-derived, neither of which needs a catalog). Both persistence paths pass it
  — `reconcileColumnConfig` **and** the dialog, via `customFieldsAvailable` on its data, since saving
  writes the list straight back and would otherwise undo what reconciliation preserved. The flag comes
  from `canReadCustomFieldCatalog(store)`, exported from the resolver so the resolver's own `of([])`
  branch and its consumers read **one** predicate and cannot drift.
  - Nothing renders badly as a result: `setColumns` builds `allColumns` from the catalog and then
    drops any configured column with no definition, so a preserved-but-unnameable column is simply
    absent from the table rather than showing a raw `custom_5` header. The sort is preserved for the
    same reason — it is reset off the reconciled columns, so keeping the column without the sort would
    still discard half the state.
- **The shared `app-table` hands the whole column to a cell template**
  (`{ element, index, column }`), which is what lets one `#customFieldCell` template serve every
  custom field: `receipts-table` carries the `CustomField` on the column object and the template reads
  it back.
- **`app-custom-field-cell`** (`src/receipts/custom-field-cell/`) is the read-only counterpart of
  `app-custom-field` — the same `ngSwitch` on `CustomFieldType`, without a FormGroup (the table renders
  plain API records). **CURRENCY goes through `customCurrency`**, exactly as the built-in Amount column
  does; SELECT renders the option's text, BOOLEAN renders Yes/No.
  - Where a receipt carries several values for one field, the **lowest id that is actually set** wins —
    the same rule the API sorts by and the reporting engine reads by, so the cell can never disagree
    with the sort order or a report.
  - **Do not name a `@let` after the signal it reads.** `@let value = value()` shadows the component
    member inside its own initializer and fails with "undefined is not a function".
- **Each custom field row carries the shared `app-badge`** reading "Custom"
  (`data-testid="column-config-custom"`), so a field named "Vendor" is not mistaken for a built-in
  column. `ColumnConfigItem.isCustom` comes from the existing `parseCustomFieldColumnDef` helper, and
  is stripped in `saveConfiguration()` — the saved object is persisted to localStorage and read back
  as the column config, so a derived key must not ride along. See "The shared badge" above for why it
  sits outside the checkbox label.
- The receipts list response now always carries custom field values **with their definitions**; see
  `api/CLAUDE.md` → "Custom fields on the receipts list" for why the definitions are not optional.
- **E2E:** `e2e/custom-field-columns.spec.ts` (serial, admin storageState, own seeded group) — the
  fields listed after Resolved Date and unchecked, a CURRENCY cell formatted by the configured currency
  display, numeric sorting on it (amounts chosen so a text sort is visibly wrong), a SELECT sorting by
  option text rather than option id, and the column healing away when its field is deleted. Money is
  asserted with a **separator-tolerant** regex because the currency configuration is a global System
  Setting on the shared CI backend that this spec must not mutate. **Configure Columns** sits inside
  the `⋮` overflow menu, so its helpers go through `openReceiptsOverflowMenu` first.

### The Comment column (`first_comment`)

The tenth built-in column shows the receipt's **first** comment and sorts on it. It is the last
built-in in `DEFAULT_RECEIPT_TABLE_COLUMNS`.

- **Hidden by default, with no merge code.** The default entry is `visible: false`, and
  `mergeCustomFieldColumns` appends a missing built-in with its default's visibility. So a layout
  saved before this column existed gains it **unchecked**, and upgrading widens nobody's table.
- **`first_comment` is both the `matColumnDef` and the `orderBy`**, the same convention as
  `custom_<id>`. `sort()` already sends `sortState.active` verbatim, so sorting needed no client
  code. The header reads `RECEIPT_COLUMN_DISPLAY_NAMES.first_comment` rather than a new literal.
- **The value is `Receipt.firstComment`**, which only the paged list returns. The server has already
  applied member isolation, so a comment the viewer may not see is never in the payload.
  `hideComments` is deliberately **not** applied here; that group setting stays receipt-form-only.
  See `api/CLAUDE.md` → "The first comment".
- The cell (`#firstCommentCell`, `data-testid="receipt-first-comment"`) is one ellipsised line capped
  at 20rem, so a long comment can't stretch the row. `matTooltip` shows the full text, which is why
  `ReceiptsModule` now imports `MatTooltipModule`.
- **E2E:** `e2e/receipt-comment-column.spec.ts` (serial, admin storageState, own seeded group) covers:
  - the column offered unchecked and unbadged;
  - the first comment shown, never a later one that sorts lower;
  - the `first_comment` sort in both directions against the real API.

### Seeding the receipt group

The Group field is a default, not a lock: when there is only one group to pick, both the receipt form
and the Quick Scan dialog pre-select it. See the root `CLAUDE.md` → "Seeding the Group Field" for the
cross-client contract.

- **`GroupState.soleGroupId`** (`src/store/group.state.ts`) is the shared rule, declared as
  `@Selector([GroupState.groupsWithoutAll])` so it counts exactly the set `app-group-autocomplete`
  offers. That selector-array form is new to this repo (everything else uses bare `@Selector()` or
  `createSelector`) but is correct here: with a non-empty dependency list NGXS does **not** prepend
  the container state, so the function receives the groups directly.
- **`GroupState.addTargetGroupId`** is the group a *new* receipt would land in: the selected group,
  or — when that is the "All" group, unset, or **no longer resolvable** — `soleGroupId`. Because it
  resolves off `groupsWithoutAll`, all three of those are one lookup miss rather than three checks.
  `selectedGroupId` is persisted to localStorage and `setAppData` only overwrites a *falsy* one
  (`app-data.utill.ts`), so it genuinely outlives a group the user has left — that stale case is why
  the resolution is a lookup rather than a `Number()` coercion.
- **`receipt-form.component.ts` `initForm()`** seeds `addTargetGroupId ?? ""`. No
  `syncSingleDisplay()` is needed: the seed is written while the parent builds the form, which is
  *before* the child autocomplete's `ngOnInit` seeds its display from the control (unlike Magic Fill,
  which patches after init).
  - **The blank sentinel must stay `""` — never `0`.** `Validators.required` calls Angular's
    `isEmptyInputValue`, which counts only `null`/`undefined` and zero-length string/array as empty,
    so a `0` seed leaves a group-less form **valid** and POSTs `groupId: 0`. `0` is also what
    `app-autocomlete` filters its options by (`_filter` does `value.toString()`), so the dropdown
    would silently show only groups whose name contains a zero, while `!!0` still suppresses the
    clear button the e2e helpers rely on. Pinned by an explicit `form.get("groupId").valid === false`
    assertion in the spec.
- **`setReceiptPermissions()` and the `/receipts/add` route guard both gate on
  `addTargetGroupId ?? Number.parseInt(selectedGroupId)`** — the same group the form seeds. Without
  this a sole-group user whose *All*-group membership lacks `group.receipts.create` is bounced off
  `/receipts/add` even though they can create in their real group, which is where login lands them.
  The `??` tail is load-bearing: a multi-group user on the All dashboard has no add target and must
  keep gating on the All group, because `Number.parseInt("")` is `NaN` and `groupPermissions[NaN]` is
  empty — gating on the seed alone would render the add form read-only. The guard opts in per route
  via `data: { useAddTargetGroupId: true }` (a third mode beside `useRouteGroupId`), so a future
  route that wants "the group being browsed" keeps that by default; `/receipts/add` is the only
  consumer of the plain selected-group branch today.
- **`quick-scan-dialog.component.ts` `fileLoaded()`** puts it *after* the user's
  `quickScanDefaultGroupId`. `configureImages()` already runs at the end of `fileLoaded`, so the
  seeded group's show/require config lands on that image with no interaction.
- **The e2e helpers had to learn the field can already be filled.** A single-select `app-autocomlete`
  binds `[readonly]="readonly || singleOptionSelected()"`, and Material's `_canOpen()` refuses to open
  a readonly input — so an unconditional `click()` + pick simply times out. `selectFirstOption`
  (`e2e/receipts.spec.ts`) now skips a filled field, the empty-form validation case clears it first via
  `clearAutocomplete` (`data-testid="autocomplete-clear"`), and `quick-scan-dialog.spec.ts` routes its
  pick through the shared `selectImageGroup`. This matters on a **fresh** database, where
  `api/dev/seed-e2e-users.sh` leaves `e2e-admin`/`e2e-user` with exactly one group until another spec
  creates one.
- **E2e:** `e2e/single-group-default.spec.ts` provisions its own Legacy User (a new account owns only
  "My Receipts" + "All") and asserts both the receipt form and the Quick Scan dialog arrive pre-filled.

### Per-group default custom fields

A group can declare custom fields that are pre-added to its receipts, configured in a **Default
Custom Fields** `app-form-section` on `src/group/group-receipt-settings/` (between "Settings" and
"Quick Scan") and applied by `src/receipts/receipt-form/`.

- **Settings page.** An `app-autocomlete` `[multiple]="true" [creatable]="false"` over the route's
  `customFields` resolver output (`customFieldResolverFn`, wired onto both `receipt-settings` routes),
  bound to a **`FormArray`** named `defaultCustomFields` — multiple mode calls
  `inputFormControl.push(...)`, so a plain `FormControl([])` throws. It is seeded with the **same
  object instances** as the resolved catalog (`buildDefaultCustomFieldsArray`), because
  `app-autocomlete` filters already-selected options by **reference equality**; rebuilt literals leave
  a selected field in the dropdown and let it be added twice. `[readonly]` is bound **explicitly** —
  `readonly` is a plain `@Input` and is not derived from `inputFormControl.disabled`, so the page's
  view-mode `form.disable()` alone would still let a viewer remove chips. A second `app-checkbox`
  binds `applyDefaultCustomFieldsOnIngest` (quick scan / email-created receipts).
- **Both controls exist only for holders of `app.custom-fields.read`** (added in `initForm()`, the
  whole section `@if`-gated on the same flag), and `submit()` builds the command explicitly —
  destructuring `defaultCustomFields` out and mapping to `defaultCustomFieldIds` only when the
  permission is held. This is load-bearing: `UpdateGroupReceiptSettingsCommand` treats a **missing
  key as "leave unchanged"**, so an admin who cannot see the catalog can never wipe another admin's
  configuration (the `hideComments` bug shape). The FormArray of whole `CustomField` objects must
  never ride the payload.
- **Receipt form "smart swap".** `applyGroupDefaultCustomFields(groupId)` runs **inside**
  `listenForGroupChanges()`'s `tap`, **after** the `selectedGroup.set(group)` write — under zoneless
  CD that signal write is the only change-detection trigger; a FormArray mutation has none of its own.
  It returns early only without `canManageCustomFields()` and on a falsy group id.
- **It applies on load in every mode**, via that listener's `startWith()` replay, which fires at init
  in `add`, `edit` **and** `view`. A group's defaults are meant to read as its built-in receipt
  fields, so a receipt saved before the group was configured picks them up too — blank and read-only
  in view (the template's `@for` already binds `[readonly]="mode | inputReadonly"`), editable in edit,
  and persisted as empty attached values once that edit is saved. `addCustomFieldControl` skips a
  field the form already carries, so a receipt that already has the default is untouched.
  `autoAppliedCustomFieldIds` is reset at the **top** of `initForm()` (which re-runs on every
  route-data emission), which is what makes the removal pass inert on load — a stale auto set could
  otherwise strip a saved receipt's own fields — and it is also why a default applied on load is
  still the swap's to take back on a later group change.
- The swap drops a previously auto-added default only while it is still **empty**
  (`isCustomFieldControlEmpty`: every typed column null-or-`""` **and** `booleanValue` falsy — so a
  BOOLEAN deliberately left `false` counts as empty and is swapped out); anything with a value stays
  and leaves the auto set, becoming user data. `customFieldChanged` deletes the id from the auto set
  in **both** branches, so a default toggled off and back on is user-owned forever after.

**E2E:** `e2e/group-default-custom-fields.spec.ts` (serial, admin storageState) covers the two
delivery paths, which the Jest specs cannot — they inject settings into a mocked store:
- The **settings page** round-trips a set and the ingest toggle through a save, asserted after a
  re-navigation so the values have to come back out of `GET /group/{id}` (`groupResolverFn`), and
  shrinking the set persists too (the command carries the whole id list).
- The **receipt form** walks the swap matrix on `/receipts/add` against `GroupState`, hydrated from
  AppData on every navigation: select group A (defaults A+B) → type into A → add C by hand → switch
  to group B (defaults C) → **only B is dropped**, A keeps its value and C is not duplicated →
  switch back → B returns **blank** and C survives → save. Reaching `/receipts/:id/view` is itself
  the proof the client sent every attached field (an id set that doesn't match the stored one is a
  403 from `enforceReceiptCustomFieldSelection`). Verified to FAIL with the empty-check inverted.
- Deleting a custom field prunes it from the group's stored set.
- A receipt created while its group declared **nothing** shows the field once the group gains it:
  blank on `/receipts/:id/view`, editable on `/edit`, and the value survives a save + re-navigation.
  Reaching the view page after that save is the proof `enforceReceiptCustomFieldSelection` accepted
  the newly attached id.

Three `data-testid`s were added for it, none of them pre-existing: **`autocomplete-clear`** on the
shared `app-autocomlete`'s clear button, **`receipt-group`** on the receipt form's group picker, and
**`receipt-manage-custom-fields`** on the form's "Manage custom fields" menu (which both
`receipts.spec.ts` and `group-default-custom-fields.spec.ts` now use). The first two are load-bearing
rather than convenience — a single-select autocomplete marks its input `readonly` once a value is
chosen, so the group cannot be *changed* without clicking that clear button first, and a spec that
skips it silently asserts against the old group. The third replaced a
`button:has(mat-icon:has-text("list_alt"))` locator. Note
`getByTestId('receipt-manage-custom-fields').getByRole('button')` is a **strict-mode violation**: the
`cdkMenuTrigger` also puts `role="button"` on the `<app-button>` host, so use `.locator('button')`.

### Role-required receipt fields & single-call create

A group role can require its members to keep **at least one comment** and/or **one image** on the
group's receipts. The server resolves and enforces it; see `api/CLAUDE.md` → "Role-required receipt
fields & single-call create" for the contract.

- **Role form:** two `app-checkbox`es in a "Receipt requirements" card inside `@if (showGrants())`
  (`data-testid="require-receipt-comment"` / `"require-receipt-image"`), bound to
  `requireReceiptComment` / `requireReceiptImage`. Wired like `seesAllMembers`: hydrated in
  `loadRole`, reset in `pickType`, serialized only in `submit()`'s `showGrants()` branch (APP scope
  400s them).
- **Store:** `SetPermissions` carries AppData's `groupReceiptRequirements` as an optional third
  argument (dispatched only from `setAppData`). Read it through `AuthState.receiptRequirements(groupId)`
  or `receiptRequirementsFor(map, groupId)` (`src/utils/receipt-requirements.ts`), which default to
  nothing required. The server has already applied the waivers (hidden field, no comment permission),
  so **never re-derive them from group settings**.
- **Create is ONE call.** `createReceipt()` sends `ReceiptService.createReceiptWithFiles(form.value,
  filesToUpload() files)` (`POST /receipt/withFiles`, multipart). The comments ride `form.value`.
  The old `createReceipt` + per-image `uploadReceiptImage` `forkJoin` is gone; the create is atomic,
  so there is no "receipt added, images failed" state. Don't call the deprecated JSON `createReceipt`.
- **Submit guard** mirrors the server, keyed on the form's **current** `groupId` (a `currentGroupId`
  signal written in `listenForGroupChanges`), so it re-evaluates on a group change and checks the
  destination group in edit mode. `isImageMissing()` counts `filesToUpload()` in add mode and
  `images()` in edit mode (the receipt's own `imageFiles` while those load). `isCommentMissing()` reads
  the comments child's `commentCount` signal: `commentsArray` is not reactive, so the child mirrors its
  length on every mutation. That covers queued comments in add mode and the saved comments, updated
  live, in edit mode. A miss shows an inline hint in the section (`receipt-image-required-hint` /
  `receipt-comment-required-hint`) and blocks submit with a snackbar.
- **Last item in edit mode:** the image remove button (`receipt-image-remove`) is disabled by
  `isLastImageLocked()`, and `app-receipt-comments` takes `[preventDeletingLastComment]`, hiding
  delete on the only comment (`isLastCommentLocked`). Add-mode items are unsaved, so they stay
  deletable. To swap the only image, upload the new one first.
- **Quick scan:** `resolveQuickScanFieldConfig` takes `roleRequiresComment`, which shows **and**
  requires the comment even when the group's config leaves it off. It still yields to `hideComments`
  and `group.comments.create`.
- **E2E:** `e2e/receipt-role-requirements.spec.ts` (serial). An admin provisions the role through
  the role form (`createRole`'s new `requireReceiptComment` / `requireReceiptImage` options) and a
  group with e2e-user in it. The spec covers the flags round-tripping, create blocked without a
  comment and then an image, then saving in exactly one `POST /api/receipt/withFiles` with no
  `/receiptImage` upload, edit refusing the last image/comment (UI and server 400), and quick scan
  requiring the comment (UI, plus the server's 400 without one).

## Login QR (mobile app setup)

A QR that deep-links users into the mobile app to set it up. It renders in **two** places — the login
page (`src/auth/sign-up/auth-form.component.*`, shared by login + sign-up) and the **About dialog**
(`src/about/about/about.component.*`), so a user who is already signed in can reach it without logging
out. Both consume the shared standalone **`app-login-qr`** (`src/shared-ui/login-qr/`), which owns the
whole generation path; do not re-implement it at a third call site.

It is **self-contained** — generated locally with the `qrcode` package (no external QR service), from
the derived `featureConfig.loginQrUrl` string the backend composes (see `api/CLAUDE.md` → "Login QR &
mobile deep link"). The component reads `FeatureConfigState.loginQrUrl` via `store.selectSignal`,
regenerates a `data:` URL in an `effect()` (the `qrDataUrl` signal), and its template renders
**nothing** unless that signal is set (`@if (qrDataUrl())`), so the QR appears only when an admin
enabled it and neither call site carries generation-state logic. Generation is async, so the effect
takes an `onCleanup` cancellation flag — a URL change leaves the previous `QRCode.toDataURL` in flight,
and without the guard a late-resolving stale QR could overwrite the current one. Its two inputs are
presentation only: an optional `headerText` (renders the divider row; the login page passes
"Set up the mobile app", About omits it) and a `caption` with the default scan-instruction text. The
`.login-qr*` styles live in the component (encapsulated), **not** in the login page's
`ViewEncapsulation.None` stylesheet.

Availability differs per call site but needs no extra plumbing: the login page is pre-auth and gets
`loginQrUrl` from `GET /featureConfig`, while About rides on the authenticated `GET /appData`
(`setAppData` already dispatches `SetFeatureConfig`). About is the one call site that also gates on the
**setting** — `@if (loginQrUrl())` around its "Mobile App" `app-form-section` — because an
`app-form-section` would otherwise render an empty header when the feature is off. No permission gate:
`loginQrUrl` is a public, pre-auth value. `FeatureConfigState` gained a `loginQrUrl` default (`""`), a
selector, and — the load-bearing bit — the field in the `SetFeatureConfig` `patchState` block (that
block lists each field explicitly, so a new field is silently dropped unless added there; guarded by
`feature-config.state.spec.ts`).

Admins configure it on the System Settings form (`src/system-settings/system-settings-form/`): a
"Mobile App Setup" `app-form-section` with a **Show login QR code** `app-checkbox` (`showLoginQr`) and
a **Mobile Server URL** `app-input` (`mobileServerUrl`), the URL conditionally required when the toggle
is on (`listenForShowLoginQrChanges`, mirroring the MCP section's pattern). Saving refetches the
feature config, so the login QR updates without a reload.

Both URL settings — `mobileServerUrl` and `mcpPublicUrl` — also carry the shared
**`absoluteUrlValidator()`** (`src/validators/url-validators.ts`), a port of the backend's
`isValidAbsoluteUrl`: absolute http(s), non-empty host, no embedded credentials, and whitespace-only
treated as invalid (the backend trims before its own emptiness check, so spaces would otherwise
satisfy `Validators.required` and still be rejected server-side). It applies with the toggle **off**
too, because the backend validates any non-empty URL regardless of the toggle.

It also requires a **literal `http://` / `https://` prefix** before parsing, because `new URL()` is
more lenient than Go's `url.Parse`: it normalizes authority-less forms (`https:host/api`,
`https:/host/api`, and backslash variants) into a valid URL, while `url.Parse` leaves `Host` empty
and the server rejects them — so without the prefix check the form would green-light a value that
400s. The test is **case-insensitive** on purpose (`url.Parse` lowercases the scheme, so
`HTTPS://host` is valid server-side); pinned by the "rejects url forms that the backend rejects" spec
case. Both listener methods
build their list through the shared `urlValidators(required)` helper — `setValidators` replaces the
whole list, so the format check must be re-supplied on every toggle rather than declared in `initForm`.

**E2E:** `e2e/login-qr.spec.ts` (serial, admin `storageState` + a fresh unauthenticated context for the
login page) drives the whole flow — admin enables the toggle + URL in System Settings and it persists,
the QR `<img>` then renders on `/auth/login` with the `featureConfig.loginQrUrl` decoding back to the
configured server URL, the **About dialog** shows the same QR in an admin session (sidebar avatar →
About; the avatar carries `data-testid="sidebar-avatar-menu"` for this), and disabling hides it. It
reverts `showLoginQr` via the admin API in `afterAll` (the setting is global). Component-level specs
live alongside the code: `login-qr.component.spec.ts` (the generation unit cases — empty, generated,
divider-only-with-`headerText`, turned back off, and the stale-generation guard),
`feature-config.state.spec.ts`, `auth-form.component.spec.ts` and `about.component.spec.ts` (each
pinning that its page wires the shared component up), and `system-settings-form.component.spec.ts`.

## OIDC sign-in (external identity providers)

Users can sign in with an external identity provider; administrators register providers in System
Settings. The **backend owns the whole OIDC exchange** (see `api/CLAUDE.md` → "OIDC (OpenID Connect)
login"), so the desktop's job is three things: render the buttons, bootstrap the session after the
redirect, and administer the provider rows.

- **Login buttons** (`src/auth/sign-up/auth-form.component.*`): one `app-button` per
  `FeatureConfigState.oidcProviders`, which `AppInitService` already loads pre-login off the public
  `GET /featureConfig`. The click is
  **`window.location.href = /api/oidc/{name}/login`** — a full-page navigation, not an `HttpClient`
  call: the browser has to follow the redirect chain to the identity provider, and an XHR would just
  fetch its HTML into JavaScript. The URL is built by the exported pure `buildOidcLoginUrl`, and the
  navigation itself goes through a `navigateToUrl` seam so a spec can assert where it goes (jsdom
  refuses to let `window.location` be replaced).
  **The section is gated on `oidcProviders().length`, NOT `*appFeature`** — `FeatureConfigState.hasFeature`
  is a truthiness check and `[]` is truthy, so the directive would render a bare "or" divider on an
  install with no providers configured.
- **`/auth/callback`** (`src/auth/oidc-callback/`) is the piece that is easy to miss.
  `AppInitService.initAppData()` only loads app data when `isLoggedIn || hadSession`, both of which
  read persisted NGXS state out of `localStorage`. A **first-ever** OIDC login has perfectly valid
  `jwt` / `refresh_token` cookies but an **empty store**, so the initializer takes its unauthenticated
  branch and the app boots signed-out while holding a working session. The component calls
  `getAppData()` → `setAppData(...)` → navigates to `GroupState.dashboardLink`.
  It lives under `/auth` deliberately: `AuthGuard` lets an unauthenticated user through any route whose
  URL contains `auth`. **Rejected alternative:** making the initializer always attempt `getAppData()` —
  that adds a guaranteed 403 to every anonymous page load.
  A corollary: because `AuthGuard` also *bounces an already logged-in user away* from those routes, the
  **link** flow must not return through `/auth/callback`; its success target is the profile page.
- **Login error reporting:** the backend redirects failures to `/auth/login?oidcError=<code>` from a
  small fixed vocabulary (it never echoes an identity provider's own text). `auth-form.component.ts`
  maps those to copy in `OIDC_ERROR_MESSAGES`, with a fallback for an unrecognized code.
- **Admin CRUD** — a new **OIDC Providers** tab in System Settings
  (`src/system-settings/oidc-provider-table/` + `oidc-provider-form/`, following the
  `custom-fields` shape: a paged table plus one dialog serving add/view/edit). As with every
  system-settings tab, the entry must be added to **both** `system-settings.component.ts`'s `allTabs`
  **and** `guards/system-settings-landing.guard.ts` **at the same index** — the guard's contract is
  that it redirects in the order the tab bar renders.
  - The dialog shows the **exact redirect URI** to register with the provider (read-only, with a copy
    button). Providers match it exactly, so a near miss fails at their end where this app never sees it.
  - **The name is disabled outside add mode** — it is part of that redirect URI, so a rename would
    break every subsequent login. The server rejects it too.
  - **The client secret renders blank on edit and is omitted from the payload when untouched.** That is
    the client half of the server's `ClientSecret *string` contract (omitted = keep the stored one), and
    it is why the secret is never sent to the browser at all; `hasClientSecret` reports whether one
    exists.
  - The **username-linking** checkbox carries a plain-language warning that it is only safe for a
    provider you control — see the backend doc for why it defaults off.
  - The actions column is gated on `hasAnyAppPermission([Update, Delete])`; gating on delete alone
    would hide the whole column, and with it the edit button, from an update-only holder.
  - **It is the one dialog form in the app that guards against a double submit** — a private
    `isSubmitting` field checked in `submit()` and cleared in `finalize()`. Do not "tidy" it back
    into line with the other eleven, all of which are `take(1) → tap → close` with the button
    live for the whole request. `take(1)` caps one subscription; it does not stop a second click
    starting a second request. The shared `disableWhenProgressBarIsShown` input on
    `app-dialog-footer` looks like the house answer and is not: the only dispatcher of
    `ShowProgressBar` anywhere is Magic Fill (`receipt-form.component.ts`), so the three call
    sites that opt in are protected by nothing. Here it mattered because `oidc_providers.name`
    carries a DB unique index — the duplicate request cannot create a second provider, it just
    fails, and `MatSnackBar.open()` dismisses the success toast to show that error, so the admin
    is left reading a failure for a save that worked. It clears in `finalize` rather than `tap`
    because a failed save leaves the dialog open for the interceptor to report.
  - **The redirect URI's copy button handles an absent clipboard, not just a rejected write.**
    `navigator.clipboard` is `undefined` outside a secure context, so the call throws
    *synchronously* — and plain-http LAN self-hosting is a deployment this project supports. This
    is the one string an admin has to hand to their identity provider, so a dead button gets a
    message rather than silence. It is deliberately **not** the shared `app-copy-button`, which is
    silent on failure by design (pinned by `copy-button.component.spec.ts`); the model here is
    `api-key-form-dialog`, the other hand-rolled `content_copy` button in a dialog, which toasts
    both outcomes.
- **Profile → Connected accounts** (`src/settings/user-profile/`): link and unlink providers from an
  authenticated session — `window.location.href = /api/oidc/link/{name}` again. This is the safe way to
  attach a pre-existing password account, and what makes the server's username-matching toggle
  comfortable to leave off. **Unlink is hidden for a provisioned account's last connection**, mirroring
  the server's lockout guard (the server refuses it either way). A failed load yields no section rather
  than an error, so an older server degrades quietly.
- **`serverPublicUrl`** is a new field in the System Settings form's **Server** section — the
  externally reachable origin the redirect URIs are built from. It carries the shared
  `absoluteUrlValidator()` and is deliberately **never conditionally required** here: a blank value is
  fine on an install with no OIDC providers, and the "required once a provider is enabled" rule is
  enforced where providers are saved, so an admin cannot register one with no redirect URI to publish.
  - **The backend enforces the same rule from the other direction too**, which this form has to
    report: clearing the field while an enabled provider exists is refused with a field-keyed 400
    under `serverPublicUrl`. `submit()` therefore carries a `catchError` → `applyApiErrors` (below),
    placed **before** the success `tap` so the toast and the navigation cannot run, and ahead of the
    feature-config refetch so a failure *there* still reaches the interceptor. Without it the page
    sat there having apparently saved.
- **Specs:** `auth-form.component.spec.ts` (buttons per provider, none when empty, URL escaping,
  `?oidcError=` snackbar), `oidc-callback.component.spec.ts` (bootstrap, failure, and the error-code
  passthrough that must **not** call the API), `oidc-provider-form.component.spec.ts` (name disabled on
  edit, blank secret omitted / typed secret included, scope must contain `openid`, reserved and
  malformed names, the warning renders), `oidc-provider-table.component.spec.ts` (the actions-column
  gating in all three permission shapes), and `feature-config.state.spec.ts` (defaults to `[]`, and
  survives a payload that omits the key). Note the table specs use **`detectChanges(false)`**, the
  existing convention for these tables — `setColumns()` runs in `ngAfterViewInit` and mutates a plain
  property, which trips NG0100 in the dev-mode check-no-changes pass.

## Duration settings (Hours/Days selector over an hours-based API)

The System Settings form exposes two configurable refresh-token lifetimes (see `api/CLAUDE.md` →
"Session lifetime"): a **Session** `app-form-section` with **"Stay signed in for"**
(`refreshTokenValidForHours`) and, inside the existing **MCP Server** section, **"Connector sign-in
lasts"** (`mcpRefreshTokenValidForHours`).

**Hours are the wire format; the unit is presentation only.** Each setting is edited as a pair of
**form-only** controls — `<name>Value` (an `app-input type="number"`) and `<name>Unit` (an
`app-select` over `durationUnits`) — which `submit()` folds back into `<name>Hours` via `toHours(...)`
before `delete`-ing the four helper keys from the payload. The exact-payload assertion in
`system-settings-form.component.spec.ts` fails if any of them leak.

The conversion lives in **`src/utils/duration.utils.ts`** (`splitHours` / `toHours` / `maxForUnit`),
shared by both settings so they cannot drift. `splitHours` renders a value as Days only when it
divides evenly into whole days and is at least 24 (720 → 30 Days, 36 → 36 Hours), and falls back to
the 24h default for the `0`/null the API sends when the setting is unset — so the view page shows
"1 Days", not "0 Hours". Note `toHours` guards on `parsed <= 0`, not just `Number.isFinite`:
`Number(null)` and `Number("")` are both `0`, so an emptied number field would otherwise submit a
zero-length lifetime.

**The bounds track the selected unit** (720 hours === 30 days).
`listenForDurationUnitChanges(value, unit, maxHours, minHours = 1)` — one parameterised method called
once per setting — re-applies
`[Validators.required, durationValueValidator(maxForUnit(...), minForUnit(...))]` whenever the unit
flips, the same shape as `urlValidators()` above (`setValidators` replaces the whole list, so
`required` must be re-supplied each time rather than declared in `initForm`). Every control is also
added to the **view-mode disable block**, or it stays editable on the read-only page.

**A third consumer: `tempFileRetentionHours`** — a "Temporary Files" section holding **"Keep failed
uploads for"** (how long a failed upload's image is kept so the user can rerun, preview or download
it; see `api/CLAUDE.md` → "Temporary file retention & cleanup"). Three things it does differently, all
of them traps:

- **Its bounds are its own** — 24 to 8760 hours. `MAX_TOKEN_LIFETIME_HOURS` is the *token* cap; reusing
  it would cap a year-long retention at 30 days. Hence the bounds became parameters of
  `listenForDurationUnitChanges`, which previously hardcoded that constant *inside* the method (only
  the control names were parameters), so both existing call sites had to be updated.
- **Seed it as `splitHours(settings?.tempFileRetentionHours || 720)`.** `splitHours` falls back to
  `DEFAULT_DURATION_HOURS` (24) for the `0`/null the API sends when unset, which would render this
  setting as "1 Days" instead of its real 30-day default.
- **A floor converts with `minForUnit`, never `maxForUnit`.** `maxForUnit` floors, which turns a
  1-hour minimum into **0 days** and admits a zero-length duration; `minForUnit` rounds up and never
  yields zero. `durationValueValidator` gained an optional `min` for this, so the 24-hour floor is a
  field-level message rather than a bare server 400.

The spec's `.duration-row` count assertion is now **3**, and both exact `toEqual` assertions (on
`form.value` and on the `updateSystemSettings` payload) carry the new control pair.

**E2E:** `e2e/session-lifetime.spec.ts` (serial, admin `storageState`; the setting is **global**, so
`afterAll` re-reads live settings and restores only the captured field). It is the only test that
proves the value survives **model → command → DB → JWT → Set-Cookie**: an admin saves "2 Days", the
API stores `48`, the view page renders it back as "2 Days", and a fresh login's `refresh_token`
cookie expiry lands in the configured window while the `jwt` cookie stays inside 20 minutes.

**`durationValueValidator`** (`src/validators/duration-validators.ts`) replaces
`Validators.min`/`max` here for two reasons, both of which cost real UX:

- **Whole numbers.** The API field is a Go `int`, so a fractional entry (`type="number"` accepts
  decimals) fails `json.Unmarshal` outright and returns an unparseable-body error rather than a
  field-level message.
- **Visible messages.** `BaseInputComponent.errorMessages` only maps `required`/`email`/`duplicate`/
  `min`, so a `Validators.max` failure renders as an **empty** `mat-error` — a red field with no
  explanation. This validator emits its message as the error *value*, which takes that component's
  `typeof value === "string"` path and renders verbatim. Reach for the same trick for any new
  validator whose error key is not in that map.

## Activity source file (preview / download)

A failed quick scan or email upload keeps its image, so the user can retrieve it and enter the receipt
by hand — the endpoints and the retention window are in `api/CLAUDE.md` → "Activity source files".

- **Two places render it.** The dashboard **activity widget**'s `itemMetaTemplate`
  (`activity-source-file-preview` / `-download`, beside the existing rerun button), and an **opt-in
  column** on the shared `app-task-table` behind `[showSourceFileActions]="true"` — only the System
  Settings → System Tasks page passes it, so the component's other host is unchanged. The column id
  goes into `displayedColumns` **before** the `"expand"` push (`expand` stays last, and `mat-table`
  throws on a displayed id with no matching `columns` entry).
- **Both gate on `item.hasSourceFile` and the group permission**, and the server decides the flag —
  on the app-scoped tasks table it is resolved per caller, so a button never appears where the
  endpoint would 403.
- **Gate on `item.groupId`, NOT the widget's `groupId()`.** The activity widget's rerun button had
  this wrong: on the "All" dashboard `groupId()` is the synthetic aggregate, which never appears in
  `AuthState.groupPermissions`, so every control silently vanished for activities the user was
  entitled to act on. Mobile fixed it long ago and documents the reasoning at
  `group_activity_list_item.dart`; desktop never did. Rerun was fixed in the same change, and
  `activity.component.spec.ts` pins it by setting the two ids to different values.
- **`SourceFileViewerDialogComponent`** (`shared-ui/source-file-viewer-dialog/`) wraps
  `app-image-viewer` in `app-dialog`, opened with `DEFAULT_DIALOG_CONFIG` — shaped on
  `DescriptionViewerDialogComponent`, and declared **and** exported in `SharedUiModule`.
- **The download is observed as a `response`, not a body.** The original file name is only on
  `Content-Disposition` and an activity row does not carry it, so
  `filenameFromContentDisposition` (`src/utils/file.ts`) reads it back before handing off to the
  shared `downloadFile`.

**E2E:** `e2e/failed-activity-source-file.spec.ts` (serial, admin storageState). It is the only test
that proves the image actually survived in `temp/` through a failure and that the asynq payload still
resolves to it. Making a quick scan fail is the whole trick: the spec points the **global** AI
provider at an unreachable host rather than assuming the backend has none configured, which is what
makes it work against the shared demo backend as well as a local database. It therefore mutates
global state, runs serially, restores what it found, and leans on the `e2e-shared-backend` job-level
concurrency group. It asserts the preview converts, the download is **byte-identical** to the
uploaded fixture, a non-member 403s on both endpoints, and the tasks table renders the controls.

`e2e/temp-file-retention.spec.ts` covers the setting's own round trip (days in, hours stored, days
rendered back) and the 24-hour floor. **Do not try to e2e the sweeper** — the 24-hour minimum
retention makes it unobservable; its coverage is the Go table tests.

## Signals & Zoneless Change Detection

This application uses Angular's signal-based reactivity model with zoneless change detection (`provideZonelessChangeDetection()`). All new code MUST follow these patterns.

### Signal Primitives — Decision Guide

| Need | Use | NOT |
|------|-----|-----|
| Mutable state | `signal()` | Plain class properties |
| Read-only derived value | `computed()` | `effect()` that copies signals |
| Writable derived state (resets on dependency change, can be overridden) | `linkedSignal()` | `effect()` that sets a signal |
| Sync signal state to imperative/external APIs (DOM, localStorage, canvas, analytics) | `effect()` | — |
| DOM measurement/manipulation after render | `afterRenderEffect()` | `effect()` + `setTimeout` |
| Async data fetching | `resource()` | Manual subscribe + signal set |
| Observable → Signal bridge | `toSignal()` | `subscribe()` + signal set |
| Signal → Observable bridge | `toObservable()` | — |

### signal() — Writable State
- Use for mutable, source-of-truth state in components or services.
- Prefer `signal()` over plain class properties — signals automatically notify Angular's change detection.
- Provide a custom equality function when needed to avoid unnecessary updates.

```typescript
count = signal(0);
items = signal<Item[]>([]);
```

### computed() — Derived State
- Use whenever a value is derived from other signals. Always prefer over `effect()` for derivations.
- Computed signals are lazy (not evaluated until read) and cached (not recalculated until dependencies change).
- Safe to perform expensive operations (e.g., filtering arrays) inside computed.

```typescript
fullName = computed(() => `${this.firstName()} ${this.lastName()}`);
filteredItems = computed(() => this.items().filter(i => i.active));
```

### linkedSignal() — Writable Derived State
- Use when a value normally follows a computation but can be manually overridden.
- Resets to the computed value when dependencies change, but allows `set()`/`update()`.
- Perfect for selections that reset when options change.

```typescript
// Resets to first option when options change, but user can select manually
selectedOption = linkedSignal(() => this.options()[0]);
```

### effect() — Side Effects (Last Resort)
- **NEVER** use `effect()` to derive state or copy signal values between signals. Use `computed()` or `linkedSignal()` instead.
- **ONLY** use for syncing to non-reactive/imperative APIs: logging, localStorage, canvas rendering, third-party UI libraries.
- Effects run during change detection. They do not need `allowSignalWrites` (removed in Angular 19).
- Use `afterRenderEffect()` instead when you need to read DOM properties (offsetWidth, etc.) after rendering.

```typescript
// GOOD: Syncing to localStorage
effect(() => {
  localStorage.setItem('theme', this.theme());
});

// BAD: Deriving state — use computed() instead
effect(() => {
  this.fullName.set(`${this.firstName()} ${this.lastName()}`); // ❌ NEVER DO THIS
});
```

### Signal Inputs — input() and input.required()
- Use `input()` for optional inputs with defaults. Use `input.required()` for required inputs.
- Signal inputs are read-only (`InputSignal`). Template binding syntax `[prop]="value"` is unchanged.
- Use `computed()` to derive values from inputs. Use `effect()` only for imperative side effects triggered by input changes.
- Use `model()` for two-way binding (component modifies a value based on user interaction, e.g., custom form controls).

```typescript
// Required input — no undefined in type
mode = input.required<FormMode>();

// Optional input with default
disabled = input(false);

// Optional input without default
tooltip = input<string>();

// Two-way binding
value = model<string>('');

// Deriving from inputs — use computed, NOT effect
displayText = computed(() => this.mode() === FormMode.Edit ? 'Save' : 'Create');
```

**Replacing ngOnChanges:** Convert input-watching logic from `ngOnChanges` to `computed()` (for derived values) or `effect()` (for imperative side effects like loading data).

```typescript
// Before (ngOnChanges)
ngOnChanges(changes: SimpleChanges) {
  if (changes['groupId']) this.loadData();
}

// After (effect for imperative side effect)
constructor() {
  effect(() => {
    const id = this.groupId();
    if (id) this.loadData(id);
  });
}
```

### Signal Outputs — output()
- Use `output()` instead of `@Output() + EventEmitter`. Template syntax `(event)="handler($event)"` is unchanged.
- Use `outputFromObservable()` when the source is an Observable.

```typescript
clicked = output<MouseEvent>();
// Emit: this.clicked.emit(event);
```

### Signal Queries — viewChild() / viewChildren()
- Use `viewChild()` / `viewChildren()` instead of `@ViewChild` / `@ViewChildren`.
- Access via signal call: `this.paginator()` instead of `this.paginator`.
- Use `viewChild.required()` when the element is guaranteed to exist (not behind `@if`).

```typescript
paginator = viewChild.required(MatPaginator);
optionalEl = viewChild<ElementRef>('myEl');
items = viewChildren(ItemComponent);
```

### RxJS Interop
- **`toSignal(observable)`**: Converts Observable to Signal. Creates a subscription — call once and reuse the signal, never call repeatedly. Automatically unsubscribes on destroy.
  - Provide `initialValue` for Observables that don't emit synchronously.
  - Use `requireSync: true` for BehaviorSubject or other synchronous sources.
- **`toObservable(signal)`**: Converts Signal to Observable. Only emits the latest stabilized value.
- **`takeUntilDestroyed()`**: Replaces `@UntilDestroy()` / `untilDestroyed(this)`. Use in constructor or pass `DestroyRef`.
- **`outputFromObservable()`**: Declares an output from an Observable source.

```typescript
// NGXS selector → signal (preferred pattern)
groups = this.store.selectSignal(GroupState.groups);

// HTTP Observable → signal
data = toSignal(this.http.get<Data>('/api/data'), { initialValue: [] });

// Cleanup subscriptions
constructor() {
  this.someObservable$.pipe(
    takeUntilDestroyed(),
  ).subscribe(val => this.doSomething(val));
}
```

### NGXS State Access
- Use `store.selectSignal()` instead of `@Select` decorator for template-bound state. Returns a `Signal<T>`.
- `store.selectSnapshot()` remains valid for synchronous one-time reads in methods.
- Remove `| async` pipe from templates — use signal reads `()` instead.

```typescript
// Before
@Select(AuthState.isLoggedIn) isLoggedIn!: Observable<boolean>;
// Template: *ngIf="isLoggedIn | async"

// After
isLoggedIn = this.store.selectSignal(AuthState.isLoggedIn);
// Template: @if (isLoggedIn()) { ... }
```

### Zoneless Change Detection Rules
Angular no longer uses zone.js. Change detection is triggered ONLY by:
1. **Signal writes** — `signal.set()`, `signal.update()`, `computed()` recalculation
2. **`ChangeDetectorRef.markForCheck()`** — for non-signal reactive patterns (AsyncPipe calls this automatically)
3. **Template event bindings** — `(click)="handler()"` automatically triggers CD
4. **`ComponentRef.setInput()`** — programmatic input setting

**Key implications:**
- Plain property mutations (`this.foo = 'bar'`) in async callbacks (subscribe, setTimeout, Promise.then) will NOT trigger change detection. Always use signals for state that affects templates.
  - **`finalize()` is an async callback too, and a success path can mask the bug.** The login form
    kept `isLoading` as a plain field reset in `finalize()` for ~5 months after the zoneless
    migration (`d3b4246`, which never touched that file): on success `router.navigate()` happened to
    trigger CD so the reset repainted, while on **failure** nothing did and the spinner stuck
    forever. State only written on a failure path is the easiest kind to miss — check that a
    converted signal is exercised by a test on the *error* branch, not just the happy one.
- `ChangeDetectorRef.detectChanges()` still works but is rarely needed — prefer signals.
- `setTimeout` still works for delays but won't auto-trigger CD. The callback must write to a signal if the template needs updating.
- All `@HostListener` handlers automatically trigger CD (same as template events).

### Testing with Zoneless
- Add `provideZonelessChangeDetection()` to `TestBed.configureTestingModule` providers.
- Prefer `await fixture.whenStable()` over `fixture.detectChanges()` for most realistic test behavior.
- Use `TestBed.flushEffects()` when testing effect-based logic.

## E2E Testing

End-to-end tests live in `e2e/` and use **Playwright**. They drive the real Angular UI against a real Go API. Config is `playwright.config.ts`.

### Running locally

1. **One-time:** install browsers — `npm run e2e:install`. (In the Claude Code web sandbox the browser
   is pre-installed and `e2e:install` is blocked — apply the `chromium-1217` /
   `chromium_headless_shell-1217` symlinks from "Running in the Claude Code Web/Cloud Sandbox" above
   instead.)
2. **One-time:** sign up the two e2e accounts against your local DB. The **first** signup is auto-promoted to admin, so order matters. With the API running, go to `http://localhost:4200/auth/sign-up` and create:
   - Admin first: username `e2e-admin`, password `e2e-admin-password`
   - Then user: username `e2e-user`, password `e2e-user-password`
3. **Every run:** source the dev env script so the `E2E_*` vars are exported:
   ```bash
   cd ../api/dev && source switch-to-sqlite.sh && cd -
   ```
   (`switch-to-mariadb.sh` / `switch-to-postgresql.sh` work the same — all three export the same `E2E_*` defaults.)
4. Start the Go API separately (`cd ../api && go run main.go`). Playwright auto-starts the Angular dev server via its `webServer` config, but it cannot launch the API.
5. Run the tests: `npm run e2e` (or `npm run e2e:ui` for watch-style debugging).

### CI

In CI the same spec files run against the demo URL. GitHub secrets populate the `E2E_*` vars — point `E2E_BASE_URL` at `https://demo.receiptwrangler.io` and supply the secret credentials. When `E2E_BASE_URL` is remote, the config skips the `webServer` block and does not start a local dev server.

**The mobile suite shares that backend.** `.github/workflows/mobile-e2e.yml`'s `android-e2e` job reads
the same `secrets.E2E_BASE_URL`, and both suites mutate **global** System Settings (the login-QR
toggle, the AI-powered-receipts flag). Their workflow-level concurrency groups differ, so the desktop
`e2e` job and the mobile `android-e2e` job additionally share a **job-level** concurrency group
(`e2e-shared-backend`, `cancel-in-progress: false`) that queues one behind the other. The group name
is deliberately **ref-independent** — the backend is a single shared resource, and `mobile-e2e.yml`
also fires on `tech/mobile-e2e` / `workflow_dispatch` while `e2e.yml` fires only on `main`, so a
`${{ github.ref }}`-scoped group would put them in different buckets and let both run at once. Keep
the two groups byte-identical. Any new spec that mutates a global setting relies on that lock — don't
remove it, and prefer client-side interception (below) over server mutation whenever the assertion
allows it.

### Best practices (follow these when adding new e2e tests)

**Locators — prefer `data-testid`; auto-retrying selectors only.**
- **Use `page.getByTestId(...)` as the standard selector.** Icon-only controls (the shared
  `app-add-button` / `app-edit-button` / `app-delete-button` / `app-cancel-button`, filter/menu icon
  buttons, etc.) and any element without a stable accessible name **must** carry a `data-testid`. Name
  it `<resource>-<action>` — e.g. `group-delete`, `comment-delete`, `receipt-duplicate`,
  `add-group-member`, `dialog-submit-button`. The `data-testid` passes through the shared button
  components to the host element, so `getByTestId('comment-delete')` resolves it directly.
- `page.getByRole('button', { name: '...' })` / `page.getByLabel(...)` / `page.getByPlaceholder(...)`
  remain fine for elements that already have a real accessible name (text buttons, labelled inputs).
- **Never** use structural CSS chains (`page.locator('app-receipt-comments app-delete-button')`) or raw
  CSS/XPath (`page.locator('.btn-primary')`) — they're brittle to component-structure refactors. Add a
  `data-testid` to the control instead.

**Assertions — rely on web-first expects, never `waitForTimeout`.**
- Use `await expect(locator).toBeVisible()`, `toHaveText()`, `toHaveURL()`, `toHaveCount()` — they auto-retry until `expect.timeout`.
- Never `await page.waitForTimeout(ms)` — it's a fixed sleep and flakes.
- Prefer `await page.waitForURL(/.../)` or `await page.waitForResponse(...)` for navigation/network waits.

**Isolation — each test gets a fresh `BrowserContext`.**
- No cookies/localStorage/session leak between siblings.
- Do NOT hand-write state-sharing between tests. If two tests need a logged-in session, use Playwright's `storageState` pattern (see below), not module-level globals.

**Auth — reuse login state, don't re-login in every test.**
- Current suite is tiny (login IS the test), so each test logs in via the UI. Fine for now.
- When the suite grows, switch to the **setup project** pattern: a `*.setup.ts` file logs in once and saves `storageState` to `e2e/.auth/<role>.json`; other tests declare `test.use({ storageState: 'e2e/.auth/user.json' })`. Keep `.auth/` git-ignored — it contains session cookies.
- One storageState file per role (admin, user). Never share one login across roles.

**`webServer` — for processes Playwright can launch.**
- The config uses `webServer` to start `npm start` when `E2E_BASE_URL` is localhost, and skips it when the URL is remote. `reuseExistingServer: !process.env.CI` lets local devs keep `ng serve` running between runs.
- Playwright cannot launch the Go API — that's always a separate process.

**Env vars and secrets.**
- Read via `process.env.E2E_*` — never hardcode credentials.
- Local defaults come from `api/dev/switch-to-*.sh`. CI values come from GitHub secrets.
- Never commit `.env` files or `e2e/.auth/` artifacts.

**Parallelism and flake budget.**
- `fullyParallel: true` is on. Tests must not mutate shared server state in ways that collide (same DB row, same uploaded file, same group membership). When you need mutation, create unique data per test (timestamp/UUID in names) and clean up after.
- `retries: 2` in CI, `0` locally — a test that only passes with retries is a bug, not a feature. Fix the root cause.
- `trace: 'on-first-retry'` captures a trace file on the first retry; view with `npx playwright show-trace <file>`. Do not set `trace: 'on'` — too heavy.

**Writing selectors for this app.**
- Forms use a custom `<app-input>` wrapper over `<mat-form-field>`. `page.getByLabel('Username')` resolves through the `<mat-label>` association.
- Submit buttons use `<app-button>` rendering `<button>` with visible text — `page.getByRole('button', { name: '...' })` works directly.
- Error feedback is often a Material snackbar (not inline `<mat-error>`). When asserting errors, locate the snackbar container or its text, not the form.
- **`getByLabel` matches substrings.** On any password field carrying the generate button, plain
  `getByLabel('Password')` resolves *two* elements — the input and the button, whose accessible name
  is "Generate password" — and fails on strict mode. Use `getByLabel('Password', { exact: true })`
  for the Create User / Set Password dialogs (the login form has no generate button, so it is
  unaffected). The generated password is asserted in `generate-password.spec.ts`, which also grants
  `permissions: ['clipboard-read', 'clipboard-write']` in `test.use` — the only clipboard-reading
  spec in the suite.

### Per-member category/tag grant specs

Three specs cover the per-member grant feature (see `api/CLAUDE.md` → "Category/tag grant
resolution"). They use the standard **`e2e-user`** as the restricted member with custom **group**
roles — no custom app role or per-spec `storageState` is needed, because the default app role
(Legacy User) omits `app.categories.read`/`app.tags.read`, so that user does **not** get the admin
grant bypass and is genuinely restrictable.

- **`member-grant-visibility.spec.ts`** — the composed semantics: role-only, member-only, the
  intersection, a role narrowed below an existing assignment (fails closed), clearing, category/tag
  independence, the require-individual toggle both ways, the write-side 403 on an out-of-grant
  category, and that the receipt form's picker offers exactly the effective set. Assertions read
  `apiMemberCatalog` (appData's per-group catalogs) — the same array the desktop pickers render
  from, so it tests the real delivery path rather than a parallel one.
- **`member-grant-security.spec.ts`** — the silent failure modes: a member with
  `group.members.update` but **not** `group.members.grants.update` is denied (with the positive
  contrast), ceiling/existence/non-member rejections, the URL (not the body) identifying the
  membership, and the two lifecycle regressions — a group rename preserving both the assignment and
  its restriction flag, and no grant resurrection when a removed member rejoins. Both lifecycle
  tests were verified to FAIL when their fix is reverted.
- **`member-grant-assignment.spec.ts`** — the authoring UI: one section per group, the ceiling
  narrowing the offered pool plus its hint, the unrestricted case, assignment persisting, an
  untouched save issuing **no** grants request, both add-modes hiding the section, the role form
  rehydrating its grants, and the **"one record, two doors"** check that the group-member dialog and
  the user form edit the same membership.

Helpers added to `e2e/helpers/provisioning.ts`: `apiCreateCategory`/`apiCreateTag` (+ deletes),
`apiSetMemberGrants` (returns the raw response so specs assert 200/400/403/404), `apiMemberCatalog`,
`apiSetGroupRoster`, `apiGetGroupMembers`, plus `categoryGrants`/`tagGrants`/`requiresIndividual*` on
`UpsertRolePayload` and `CreateRoleOptions`.

**Gotchas these specs encode:**
- **Categories/tags are global with a unique name** — every spec mints `uniqueName`-suffixed ones and
  deletes them in `afterAll`, or a re-run's create fails.
- **Teardown order:** group → role → categories/tags (a role can't be deleted while assigned).
- The group roster is only editable on `/groups/:id/details/**edit**`; `/details/view` is read-only.
- The group-member dialog's submit is dispatched (`dispatchEvent('click')`) rather than clicked:
  adding a chip grows the dialog and MatDialog re-centres, so the footer button never satisfies
  Playwright's stability check.

### Permission-gating specs (provisioned roles/users/groups)

Negative-permission coverage needs an account/role that *lacks* the permission under test, which no
seeded account provides — so an admin context **provisions a custom role (and user/group) through the
real UI** in `beforeAll` and **tears it down through the admin API** in `afterAll`. Shared flows live
in `e2e/helpers/provisioning.ts`: `createRole` (role form — type, preset, category toggles, individual
toggle-offs), `createUserWithRole`, `createGroupWithMember`, `uniqueName`, and the API-teardown
helpers `withAdminApi` + `apiDeleteUserByName` / `apiDeleteGroupById` / `apiDeleteRoleByName`.
`createRole` also accepts `skipDefaultGroup` (app roles only — ticks the "Group creation" checkbox).

- **Asserting as a provisioned user without a browser session.** `withApiAsCreds(username, password,
  fn)` opens an `APIRequestContext` for arbitrary credentials — `withApiAs` is now a thin wrapper over
  it for the two `E2E_*` fixture accounts. Use it when the assertion is about server state rather than
  UI, e.g. `apiGroupNames(api)` (the caller's own groups, gated on `app.account.read`) in
  `skip-default-group.spec.ts`, which checks a new account got the "All" group but no personal
  "My Receipts". A role built from the **"Read Only"** preset grants every `*.read` permission,
  including the `app.account.read` those calls need.

- An admin `BrowserContext` (`storageState: 'e2e/.auth/admin.json'`) provisions in `beforeAll`. Tests
  then run **either** as the default e2e-user — for a *group-scoped* member added to a fixture group
  (Angular re-fetches AppData on every navigation, so a membership added after the saved session still
  drives the route guards without re-login) — **or** as a freshly-provisioned *custom user* whose
  session is captured to a git-ignored `e2e/.auth/<name>.json` (wait for the held permission in
  `localStorage.auth` before saving) and `rmSync`-ed in `afterAll`.
- **Teardown is API-based, not UI.** The role-list delete button is disabled while a role is assigned,
  and the UI's *bulk* user-delete dummy-converts a group-owning user (so the role stays assigned and
  never deletable) — UI teardown leaks roles. `withAdminApi` logs in via `request.newContext` (through
  the dev-server `/api` proxy) and `DELETE /api/user/{id}` **hard-deletes** (freeing the app-role) /
  `DELETE /api/group/{id}` frees the group-role assignment, so the role then deletes. **Order:** delete
  the user/group first, then the role; best-effort `try/catch` so a cleanup error doesn't mask the result.
- Reference specs: `system-settings-tab-gating.spec.ts` (custom app role + user + storageState — note:
  its UI teardown leaks roles, the reason the new specs use API teardown), `group-viewer-visibility.spec.ts`
  (group member with a group role), `search-bar-visibility.spec.ts` (no `app.receipts.search` → header
  search bar never renders), `dashboard-read-redirect.spec.ts` (no `group.dashboards.read` →
  `/dashboard/group/:id` redirects to `/receipts/group/:id`; an owner contrast still sees the dashboard),
  `paid-by-visibility.spec.ts` (a group role limited to "their own receipts" → a hidden-payer receipt
  `GET` 403s and is absent from the list, the member's own 200s; uses `withApiAs`/`apiCreateReceipt` and
  the `createRole` `paidByOwn` option in `helpers/provisioning.ts`),
  `dashboard-crud-gating.spec.ts` (a Viewer holding `group.dashboards.read` but not create/update/delete
  sees no Add/Edit/Delete dashboard buttons; owner contrast does),
  `comment-gating.spec.ts` (Receipt-Editor-preset members minus `group.comments.create` / `.delete` →
  no composer / no delete control; uses `apiCreateComment`),
  `receipt-feature-gating.spec.ts` (Quick Scan / Poll Email / Magic Fill controls hidden for a Viewer —
  positive contrast is a `test.fixme` because all three also sit behind the `aiPoweredReceipts` feature
  flag, which is `false` in the dev/CI API),
  `group-delete-any.spec.ts` (two app roles differing only in **Delete Any Group**: the holder deletes
  a group it isn't a member of from the All Groups view and the table stays on that filter; the other
  sees the same group with no `group-delete` action and its direct `DELETE /api/group/:id` **403s**),
  `receipt-action-gating.spec.ts` (a Legacy Viewer sees no duplicate/delete row action, the
  `/receipts/:id/edit` route redirects, and `POST /api/receipt` **403s** via `withApiAs('user')`).
  `legacy-user-visibility.spec.ts` likewise carries **API-403** assertions (`DELETE /api/category|tag/:id`)
  so server enforcement is proven, not just the hidden control. Note: the receipts-table **edit** action
  is not template-gated (only duplicate/delete are); the edit *route* is guarded, so the edit denial is
  asserted at the route level, not button absence. (`receipt-action-gating.spec.ts` is a standalone
  spec rather than an extension of `group-viewer-visibility.spec.ts`, whose serial block has a known
  pre-existing failure — a Legacy User can't load `/groups` — that would skip any test appended to it.)

## App time zone (dates)

One admin-set IANA zone (System Setting `timeZone`, default `UTC`) decides every calendar boundary.
It reaches the desktop on AppData and lives in **`SystemSettingsState.timeZone`** (persisted slice;
the selector falls back to `UTC` for a hydrated slice that predates the key), set by `setAppData`.
All helpers are in `src/utils/app-time-zone.ts`. The rules (root `CLAUDE.md` holds the table):

- **Two kinds of value.** A receipt's `date` and DATE custom fields are **calendar days**, stored as
  midnight UTC and read by their **UTC** day, so no zone ever shifts them. `createdAt`,
  `resolvedDate`, `updatedAt`, task `startedAt`/`endedAt` etc. are **instants**, shown in the app zone.
- **Display: the `appDate` pipe** (`src/pipes/app-date.pipe.ts`, standalone, exported by
  `PipesModule`) replaces Angular's `date` pipe everywhere — do not add a bare `| date`.
  `value | appDate[:format]` is instant mode: `DatePipe` formatting with `offsetAt(value, zone)`
  (DatePipe takes an offset, not an IANA name, and the offset depends on DST at that instant).
  `value | appDate:format:'calendar'` shows the stored UTC day. Format strings are DatePipe's
  (default `mediumDate`); empty/unparseable → `null`, so `(x | appDate) || 'Never'` works.
  - **It is impure on purpose**: it reads the zone signal in `transform`, which the template's
    reactive consumer tracks, so a zone change repaints every date under zoneless CD. A pure pipe
    would keep its cached string. It memoizes on (value, format, mode, zone).
  - Code that builds strings outside a template uses `formatAppDate(...)` (e.g. `SearchResultPipe`).
  - **Filter chips deliberately do not use it**: a filter value is the day the user *picked*, a
    local-midnight `Date`, so the chips keep formatting it on the browser's calendar.
- **Wire rule.** Date **filters** go out as a bare `yyyy-MM-dd` via `toDateWireFilter(filter, keys)`
  (`src/utils/date-wire-filter.ts`), applied **at the request, never in the store** (the datepicker
  would misread a stored bare day as UTC midnight). Callers: `ReceiptFilterService.buildPagedRequestCommand`
  (table, summary, filtered-receipts widget), the pie-chart widget, CSV export, the report builder's
  preview/generate (`toReportRequestCommand`) **and save** (`toReportRequestCommandForSave`), and
  system tasks (`toSystemTaskWireFilter`). The server turns the day into a range in the right zone.
  - **Saved report templates store bare days too.** The server generates a template later and reads
    an instant by its calendar day *as written*; a local midnight east of UTC is written as the
    previous day (`2026-08-31T15:00:00.000Z` for a Tokyo Sep 1), so a stored instant silently covered
    the wrong days. The way back is `fromDateWireFilter(filter, keys)` in `buildReportFormFromCommand`:
    a bare day becomes the datepicker's local midnight, and a **legacy instant** loads as its first
    ten characters — the day as written, which is the day the server generates it for, so the builder
    never disagrees with the report. A re-save rewrites it as a bare day. The round-trip specs
    (`report-form.factory.spec.ts`) pin both, through the save mapper as well as generate. **Receipt dates** go out as midnight UTC of the picked day
  (`2026-09-30T00:00:00.000Z`, `toMidnightUtc`), converted in `ReceiptFormComponent.buildSubmitCommand()`
  on a copy of the form value; on load `calendarDayToLocalDate` turns the stored UTC day into the
  datepicker's local midnight (also used for magic fill, replacing the old `getTimezoneOffset` hack).
- **"Today" is `todayInZone(zone)`**, a local-midnight `Date` of the zone's calendar day: the month
  stepper's seed and shortcuts, the `WITHIN_CURRENT_MONTH` implied range in `app-filter-field`, a new
  receipt's default date, and the report period hint, which also names the zone for instant fields
  ("… on Added At (America/New_York)"; Receipt Date is named alone).
- **The setting** is a "Time Zone" `app-form-section` on the System Settings form (above Currency
  Format): a single-select `app-autocomlete` (`data-testid="system-settings-time-zone"`,
  `[creatable]="false"`, `Validators.required`) over `timeZoneOptions()` — `UTC` first, then
  `Intl.supportedValuesOf('timeZone')` (~420 bare strings; the autocomplete filters and renders plain
  strings itself, so no `{ value, displayValue }` objects are needed). Two traps:
  - **It must pass a `displayWith`** (`timeZoneDisplayWith`). The autocomplete's default returns
    `""`, which would blank the field the moment a zone is picked.
  - Seeded from `originalSystemSettings?.timeZone || "UTC"`, added to the view-mode disable block, read
    through `getRawValue()` like every other control, and dispatched as `SetTimeZone` after a
    successful save (next to `SetCurrencyData`), so the open app repaints in the new zone without a
    reload. Other sessions pick it up from AppData on their next load. Both full-object `toEqual`s in
    the spec carry the key.
- **E2E:** `e2e/app-time-zone.spec.ts` (serial, admin storageState; `timeZone` is **global**, so
  `beforeAll` captures it and `afterAll` restores the captured value). It picks `America/New_York` in
  the typeahead (clearing the current value first — a single-select goes `readonly` once filled),
  asserts `GET /api/systemSettings` stored it, then checks a seeded receipt's **"Added at"** line
  (`appDate:"medium"`) against its real `createdAt` formatted in New York, and the report builder's
  hint naming the zone; switching back to UTC through the API (so the browser learns it from AppData)
  moves both back. `created_at` cannot be set through the API, so the "late Sep 30 Eastern lands in
  last month" case is pinned by the Go report tests instead. Whitespace is normalized before
  comparing, because Angular's CLDR data puts a U+202F before AM/PM.

## Filter dialogs (the shared pieces)

Two tables have a `{ operation, value }` filter dialog — receipts and system tasks — and they are
**one implementation with two field lists**, not two dialogs. Anything new of this shape reuses these
four pieces rather than copying a row template:

- **`app-filter-field`** (`src/shared-ui/filter-field/`, declared *and exported* by `SharedUiModule`)
  renders one row: the value editor for the field's `type` beside its Operation `app-select`,
  switching shape with the selected operation (a two-slot range for `BETWEEN`, a **disabled** implied
  range for `WITHIN_CURRENT_MONTH`, a single editor otherwise). It is deliberately presentational and
  form-agnostic — it reaches into the caller's `parentForm` by `basePath + fieldName`, exactly as the
  receipt filter's local `#filterField` template did before it was extracted.
- **`src/utils/filter-form.ts`** holds the form machinery: `buildFieldFormGroup` (a `FormArray` value
  for list/users fields, because the multi-select autocompletes `push()` onto the control),
  `listenForBetweenOperation` (swaps the value control between a scalar and a two-slot range as the
  operation flips) and `setupAutoOperationSelection` (picks the first operation for a field's type as
  soon as it gains a value, and clears it when the value empties). Every entry point takes a
  `thisContext` for `untilDestroyed`, so **the calling component must carry `@UntilDestroy()`**.
  `buildReceiptFilterForm` and `buildSystemTaskFilterForm` are thin field lists over these.
- **`src/utils/filter-chips.ts`** builds the chip labels (`"<Field> <operation> <value>"`,
  `WITHIN_CURRENT_MONTH` stopping at the operation, `BETWEEN` joined with `" – "`). It is pure: the
  caller injects `formatDate` / `formatCurrency` / `resolveOptionName`. `receipt-filter-chips.ts` and
  `system-task-filter-chips.ts` are wrappers supplying only their own id resolution. There is
  deliberately **no suppression hook**: every active condition gets a chip, because one without a
  chip is one the user can neither see nor clear (see "Every active condition is chipped" below).
- **`FilterField<TKey>` / `FilterFieldType`** (`src/constants/filter-fields.constant.ts`) is the one
  field-metadata type; `ReceiptFilterField` and `SystemTaskFilterField` are aliases of it.
  `isFilterEntryActive` (`src/utils/receipt-filter-entry.ts`) is likewise shared verbatim, which is
  what keeps every Filter badge and its chip row in agreement.

**A chip row needs `MatChipsModule` in the consuming module** — `SharedUiModule` imports it but does
not export it. `MatIconModule` too, for the `cancel` icon inside `matChipRemove`.

## Receipts table filtering

The receipts table (`src/receipts/receipts-table/`) offers three ways into **one** filter —
`ReceiptTableState.filter`, which is persisted to localStorage. The advanced dialog
(`app-receipt-filter`), the month stepper and the filter chips all read and write that single slice,
so they can never disagree.

- **Field metadata is shared — but the labels only partly.** `RECEIPT_FILTER_FIELDS`
  (`src/constants/receipt-filter-fields.constant.ts`) defines each field's key, label and operation
  type, and `OperationsPipe` reads the extracted `FILTER_OPERATION_DISPLAY_VALUES`, so the operation
  wording is genuinely single-sourced. **The field labels are not.** The dialog reads only
  `{ key, type }` from the constant (the shared `setupAutoOperationSelection()`) and **authors its own
  label per row** — each row is an `<app-filter-field label="Receipt Date" fieldName="date"
  type="date">`. So the constant's `label` reaches the chips and the quick-date picker only, and
  **renaming a field means editing both `receipt-filter-fields.constant.ts` and
  `receipt-filter.component.html`** or the dialog row will disagree with the chip it produces.
  (Driving those ten rows from the constant with an `@for` is not the one-liner it looks like: four
  carry an extra `[options]` binding and the Group row sits in its own conditional.)
- **A field's label matches its table column.** `date` is **"Receipt Date"**, not "Date" — the
  column header is `Receipt Date` and the table also shows `Resolved Date` and `Added At`, so a bare
  "Date" left the user guessing which of the three a filter or chip meant.
- **`isFilterEntryActive`** (`src/utils/receipt-filter-entry.ts`) is the shared "does this field
  narrow anything" predicate: any non-empty stringified value, **or** the operation
  `WITHIN_CURRENT_MONTH` (the one operation that carries no value). **Zero counts** — no field
  defaults to `0` (they default to `null`/`[]`), and the API applies an `amount EQUALS 0`, so
  excluding it would leave a filter that narrows the table with no badge and no clearable chip. `ReceiptTableState.numFiltersApplied`
  and the chip builder both call it, so the Filter badge and the chip set always agree.
- **Single-field writes go through `SetReceiptFilterField`**, whose handler spreads a new filter
  object and rebuilds a cleared field from a **fresh** `buildDefaultReceiptFilter()`. That factory
  replaced the shared `defaultReceiptFilter` constant inside `@State` defaults and `ResetReceiptFilter`
  for the same reason: the old code wrote the module-level object straight into state, where a later
  in-place edit would have corrupted the default for the rest of the session.
  `ReceiptsTableComponent.applyFilterField()` is the only caller — it dispatches the action, then
  `SetPage(1)`, then refetches, so narrowing a filter from page 7 can never land on an empty page.
- **`sort()` no longer refreshes the receipt summary** — it calls `getFilteredReceiptsPage()`. See
  "Receipt summary" below; sorting changes the order of the result set, not its membership.
- **Refreshes go through one `switchMap`** (`listenForRefreshRequests()`, wired in the constructor;
  `getFilteredReceipts()` just pushes onto its `Subject`). Each refresh used to be its own
  subscription, so the last *response* won rather than the last *request* — and the quick date
  arrows put those one click apart, so a slow earlier page could repaint the table with a month the
  user had already stepped past. The inner observable carries a `catchError(() => EMPTY)`: an error
  surfacing *through* `switchMap` would complete the outer subscription and silently kill every
  later refresh. `getInitialData()` deliberately stays on its own one-shot subscription, because it
  owns the single `setColumns()` call and that reads `viewChild.required` template refs.

### The month stepper targets one date field

`app-month-stepper` (`src/shared-ui/month-stepper/`, standalone, deliberately presentational) emits
months; `receipts-table` translates them with `src/utils/receipt-date-filter.ts` and writes them to
**whichever date field the control is pointed at** — `date`, `resolvedDate` or `createdAt`.

**Its panel is a CDK overlay, not a `mat-menu` — do not "simplify" it back.** The panel holds a year
pager, a month grid and three shortcuts, and **none of them is a `mat-menu-item`**. Inside a menu
that makes the `FocusKeyManager`'s item list empty, so arrow keys are no-ops, and
`ListKeyManager.onKeydown` turns **Tab** into `tabOut`, which `MatMenu` wires to `closed.emit('tab')`
— so Tab *dismissed* the panel instead of entering it, leaving the whole picker mouse-only
(verified in a browser: the grid is gone after one Tab). It is now
`cdkConnectedOverlay` + `cdkTrapFocus [cdkTrapFocusAutoCapture]` with `role="dialog"`, an
`aria-label`, `(keydown.escape)` and a transparent backdrop; the trap restores focus to the trigger
when the overlay is destroyed. The year pager's old `$event.stopPropagation()` is gone with it — it
only existed because a menu closes on any click inside itself.

The same rule applies to the two house alternatives, which is why neither was used: `cdkMenu` (the
`filtered-stateful-menu` precedent) has the identical empty-key-manager problem, and `ngbPopover`
(the header notifications precedent) hard-codes `role="tooltip"` on `NgbPopoverWindow`'s host, which
is equally wrong for interactive content and cannot be overridden. `e2e/receipt-quick-date-filter.spec.ts`
pins the keyboard path — it fails against a `mat-menu` implementation. A month is written
as `BETWEEN [startOfMonth, endOfMonth]` — the one operation that can express *any* month, which is
why this feature needed no API change. Picking a month **overwrites** whatever that field held.

- **The target field is a chip-shaped `mat-menu` trigger at the right-hand end of the control**
  (`receipts-quick-date-field`), so it reads left to right as "September 2026 … on Receipt Date".
  It offers `RECEIPT_DATE_FILTER_FIELDS` — the `type: "date"` subset of
  `RECEIPT_FILTER_FIELDS`, so the picker, the dialog row and the chip all name a field identically.
  **A `mat-menu` is right here and wrong for the stepper's own panel**: every entry really is a
  `<button mat-menu-item>`, so the `FocusKeyManager` has items. This is not a licence to convert that
  panel back — see the paragraph above.
- **The chosen field is persisted state** (`ReceiptTableInterface.quickDateField` +
  `SetQuickDateField`), because the filter itself is persisted: without it a reload would leave the
  stepper reading `date` while the user's month sat on `resolvedDate`. The **fallback lives in the
  `ReceiptTableState.quickDateField` selector, not only in the `@State` defaults** — defaults never
  run for a state hydrated from localStorage, so every install that filtered before this key existed
  would otherwise read `undefined` and index the filter with it. `ResetReceiptFilter` clears it back
  to `date` alongside the filter.
- **Switching fields is deliberately non-destructive and triggers no refetch.** It changes no
  condition, only which one the stepper describes, so a Date filter set in the dialog survives a move
  to Resolved Date — and is still visible and clearable as its own chip. `SetReceiptFilterData` (the
  sort path) omits the key and `patchState` only touches what it is handed, so sorting cannot reset
  it either.
- **Every active condition is chipped, including the stepper's own.** This reverses the original
  rule ("never show the same condition twice", which omitted `date` while the stepper named that
  month): now that the target field is selectable, the chip row is the only place that says *which*
  date column is filtered, and a condition with no chip is one the user can neither see nor clear
  from there. `buildReceiptFilterChips` therefore no longer takes an `omitKeys` argument. When the
  stepper cannot describe the entry (a partial range, a `GREATER_THAN`, `WITHIN_CURRENT_MONTH`) the
  label reads **"Custom"**, and the chip spells out what it actually is.
- **`monthFromFilterEntry` must accept ISO strings, not just `Date`s.** The filter is persisted and
  NGXS serializes through JSON, so after a reload the entry's value is two ISO strings. It matches
  on calendar fields (day 1, same year+month, `getDaysInMonth` on the end) rather than on the
  serialized text, which also makes it tolerate the local-midnight pair the dialog's datepickers
  write. Get this wrong and the label silently degrades to "Custom" after every refresh —
  `e2e/receipt-quick-date-filter.spec.ts` covers it because the Jest specs cannot.
- **`WITHIN_CURRENT_MONTH` is a no-op on the backend for `date` today.** `BuildGormFilterQuery` guards
  every field with `if Filter.Date.Value != nil` and the dialog stores `value: null` for that
  operation, so it never reaches the query builder — the badge counts it and nothing is filtered.
  Pre-existing; the chip just surfaces it for the first time.
- **Arrow steps from "All time"/"Custom" seed the current month** — the **app time zone's**, via
  the stepper's `[timeZone]` input (see "App time zone") — and then apply the delta, so `‹`
  and `›` never do the same thing.
- **`RECEIPT_DATE_FILTER_FIELDS` has a second consumer: the Report Builder's period "Date field"
  picker.** That one is server-backed, so a new key there needs the API side too. See "Period date
  field" under "Reports (Report Builder)".

### Receipt summary (the totals under the table)

A block of totals below `.table-container`, covering the **whole current filter result set** rather
than the visible page — so it cannot be computed client-side from `PagedData.data`. Configuration is
per-group (Group Receipt Settings) and applies to everyone; see `api/CLAUDE.md` → "Receipt Summary"
for the wire contract.

- **`app-receipt-totals` (`src/receipts/receipt-totals/`), deliberately not named `*summary*`.**
  `app-summary-card` already renders on this very page as "Selected Receipt summary" — the
  who-owes-whom settlement card — and two components called summary on one screen is a trap.
  Standalone, imported directly in `ReceiptsModule` (the `MonthStepperComponent` precedent), and
  purely presentational: it fetches nothing and knows nothing about the filter.
- **A SECOND `Subject` + `switchMap`, not a `forkJoin` with the table.** A `forkJoin` would make the
  table repaint wait on the slower unpaged aggregate on the app's hottest screen; two independent
  `switchMap`s each preserve last-request-wins within themselves, and the transient disagreement
  (table on month N, totals on N-1 for a few hundred ms) is bounded and self-correcting. It mirrors
  `listenForRefreshRequests` exactly, inner `catchError(() => EMPTY)` included — an error through
  `switchMap` would complete the outer subscription and silently kill every later refresh.
- **`getFilteredReceipts()` refreshes both; `getFilteredReceiptsPage()` refreshes only the table.**
  Paging and sorting change neither the filter nor the figures, so refetching an unpaged aggregate
  for them would be the most expensive no-op in the app. The split is arranged so the **safe**
  behaviour is the default: `getFilteredReceipts` keeps its name and all five existing callers, and
  only `sort()` and `updatePageData()` were moved to the page-only variant. A new call site that
  forgets the distinction over-refreshes rather than going stale.
- **Two mutation paths must push `summaryRequested` explicitly**, because they patch the datasource
  in place instead of refetching: `deleteReceipt` (the count changes) and the bulk status update
  (**the status buckets move** — the one that looks like it needs nothing and needs it most).
  `duplicateReceipt` navigates away, so it needs nothing.
- **The "no configured group" guard lives inside the `switchMap`**, before the HTTP call — but it
  only fires on the **All** group when no member group has a summary enabled. A **real** group whose
  summary is off still issues the request and renders nothing off the 200's `enabled: false`. That is
  deliberate: gating on the client's cached `groupReceiptSettings` would render a stale block when an
  admin has just changed the configuration. The cheapness that survives is the server's — the off
  state costs one settings read and no receipt query at all (`api/CLAUDE.md` → "Receipt Summary").
- **The All-group chip row.** `Group.groupReceiptSettings` is required on the generated `Group` and
  AppData hydrates every group's projections, so the client already knows which groups have a summary
  configured — no extra fetch. The chips pick whose *configuration* applies while the *data* still
  spans every group. Changing one dispatches `SetSummaryConfigGroupId` and pushes
  `summaryRequested` only — **not** `getFilteredReceipts()`, since the table is unchanged.
- **The fallback cannot live in a state selector.** Unlike `quickDateField`, resolving it needs the
  user's groups and which have the summary enabled, which the `receiptTable` slice does not know.
  `resolveSummaryConfigGroup` / `summaryConfigGroups` (`src/utils/receipt-summary.ts`) own it, and
  handle all three stale cases on one branch: the persisted group turned its summary off, the user
  left it, or the state was hydrated from localStorage before the key existed (`undefined`). The
  selector returns the raw persisted value with no default, on purpose.
- **Settings form: five `app-checkbox`es, not a multi-select.** `app-select` has no `multiple` input
  and `app-status-select` wraps it as single-select; adding multi-select to a control used app-wide
  is a real regression surface for five fixed options, and the Quick Scan section directly above
  already reads as a checkbox grid. The controls are a nested `FormGroup` bound through
  `form | formGet: 'receiptSummaryStatuses.' + option.value` (`FormGetPipe` delegates to
  `form.get`, which takes dot paths). Each carries a `data-testid` because the labels collide with
  Quick Scan's.
- **Only the currency-field picker is permission-gated.** It sits inside the existing
  `canManageDefaultCustomFields` branch and `submit()` spreads its key conditionally, so an admin
  without `app.custom-fields.read` omits it and cannot wipe a selection they cannot see. The toggle
  and the statuses are ungated — gating them would lock that admin out of the feature.
- **`mat-chip-listbox` selection comes from each option's `[selected]`, not a `[value]` on the
  listbox.** The listbox input only takes effect through a form binding, so a `[value]` there
  renders nothing as selected; the spec asserts the selected class for exactly that reason. The
  clickable element is the inner `button[matchipaction]` — `mat-chip-option`'s own host is
  `role="presentation"`, so clicking it in a test does nothing.
- A configured status matching no receipt still renders, as a **muted zero row**, so the block keeps
  its shape as the filter narrows and a legitimate zero does not read as a bug.
- **E2E: `e2e/receipt-summary.spec.ts`** (serial, admin `storageState`). The Jest specs inject group
  settings into a mocked store, so they prove nothing about the wire; this covers the settings form
  round-tripping through a real `GET /group/{id}`, the figures off a real decimal fold, a filter
  recomputing every row, and the All-group chip pick surviving a reload. Two locator traps it
  documents so the next author does not rediscover them: a chip's clickable, state-carrying element
  is the **inner `role="option"` button** (`mat-chip-option`'s host is `role="presentation"`), and
  the filter dialog's Status row is an `app-autocomlete` whose panel stays open over the footer after
  a pick, so it has to be dismissed or the submit click is intercepted and the test hangs to timeout.
  Dismiss it with **Tab, never Escape**: MatAutocomplete consumes Escape only *while its panel is
  open*, so on the runs where the panel has already closed the same keypress reaches MatDialog and
  closes the whole dialog (`receipt-quick-date-filter.spec.ts` → "closes on Escape without changing
  the filter" asserts that), detaching the submit button mid-click. Tab blurs the input, which closes
  the panel when it is open and is harmless when it is not. Escape is fine on the **settings page**
  version of the same picker — there is no dialog behind it there to swallow the keypress.
- **Placement is a per-group setting, and the block MOVES rather than being cloned.**
  `receipts-table.component.html` holds one `<ng-template #receiptTotals>` with one set of
  bindings, rendered through `*ngTemplateOutlet` at one of two anchors: `TOP` puts it immediately
  after `</app-table-header>` — **above `app-summary-card`**, because that card annotates a row
  *selection* and is transient while these totals describe the whole filter — and `BOTTOM` leaves
  it under `.table-container` where it has always been. Two elements behind separate `@if`s would
  drift and would each need handling in the e2e suite.
  - **`summaryPosition()` reads `summary()?.position`, never `GroupState`.** Same reason the
    `enabled` gate does: the cached `groupReceiptSettings` is stale the moment an admin changes the
    configuration. Nothing flickers on first load, because the component renders nothing until the
    response arrives.
  - **The component owns its own vertical rhythm** via a `position` input and a
    `.receipt-totals--top` modifier that moves the `$spacing-md` to the bottom edge. The receipts
    page adds no margin of its own.
  - **The settings control is an `app-select` outside the `canManageDefaultCustomFields` branch**,
    with the toggle and the status checkboxes — it reads no catalog, and gating it would lock an
    admin without `app.custom-fields.read` out of deciding where their own summary goes.
  - **Placement is asserted as DOCUMENT ORDER, and only in e2e.** The block renders at both
    positions, so "is it visible" passes whichever anchor is wrong. The Jest spec asserts the
    derivation only: `receipts-table.component.spec.ts` never renders the template — every case
    there drives the component class, and `app-table` resolves to a custom element under
    `CUSTOM_ELEMENTS_SCHEMA`, so the `viewChild.required` in `ngAfterViewInit` cannot resolve.
- **The chip row must never be asserted by count or position.** It lists every group the admin
  belongs to whose summary is enabled, and the suite runs `fullyParallel` against a shared backend,
  so a group leaked by a crashed earlier run would break an exact-count assertion — and, sorting
  earlier, would even win the alphabetical auto-select. Assert chips **by testid** and drive the
  switch by clicking; the auto-select rule is pinned deterministically in the Jest specs instead.

### The overflow menu

Quick Scan, Export all receipts, Configure Columns, Poll email(s) and the selection actions live in a
`⋮` `mat-menu` (`data-testid="receipts-overflow-menu"`), always collapsed — not breakpoint-driven.

**Every entry is a plain `<button mat-menu-item>` in `receipts-table.component.html`, never a shared
component that renders one.** `MatMenu` collects items with a **content** query
(`@ContentChildren(MatMenuItem, { descendants: true })`), which does not cross a child component's
**view** boundary — so a `mat-menu-item` rendered inside `app-quick-scan-button`'s own template would
be invisible to the menu and silently skipped by its `FocusKeyManager` (no arrow-key navigation, no
typeahead, no open-focus). `display: contents` does not help. The behaviour still lives in one place:
export calls `ReceiptExportService` directly, and Quick Scan calls the shared
`openQuickScanDialog(matDialog)` (`src/receipts/quick-scan-dialog/open-quick-scan-dialog.ts`), which
`app-quick-scan-button` uses too.

Two consequences for tests and styles:
- **A `mat-menu-item`'s role is `menuitem`, not `button`.** A `getByRole('button', { name: 'Quick Scan' })`
  negative would pass whether or not the entry rendered, so `receipt-feature-gating.spec.ts` asserts
  those absences by `data-testid` **with the menu open** (`openReceiptsOverflowMenu` in
  `e2e/helpers/receipts-table.ts`). Any new negative assertion about a menu entry must do the same.
- **Menu content renders in a CDK overlay**, outside `app-receipts-table`, so the
  `ViewEncapsulation.None` + `app-receipts-table { … }` nesting in the component's SCSS cannot reach
  it — the `N selected` section label is styled at the top level of that file instead.

The chips row is inline markup (`mat-chip-set` / `matChipRemove`) with every label built by the pure
`buildReceiptFilterChips` util, so `ReceiptsModule` must import **`MatChipsModule`** — `SharedUiModule`
imports it but does not export it. It makes no exceptions: see "Every active condition is chipped"
above. An id the caller cannot resolve (a category outside their grants,
a group they have left) renders as the raw id rather than dropping the chip, so a filter that is
actively removing rows is never invisible.

## System tasks table filtering

The System Tasks page (`src/system-settings/system-task-table/`) filters on **Type**, **Ran By**,
**Started At** and **Ended At** through the shared pieces above: a badge-counted Filter button and a
Reset button in the `app-table-header`, a chip row below it, and `app-system-task-filter` as the
dialog. `SystemTaskTableState.filter` is the single slice all three read and write.

- **The filter belongs to the page, not to `app-task-table`.** That shared table is rendered by three
  hosts (this page, `system-email-form`, `receipt-processing-settings-form`), each with its own table
  service, so the filter rides in as one optional input (`[filterProvider]`) rather than widening
  `BaseTableService`. It is a **function**, not the filter value: the page dispatches the filter
  change and calls `getTableData()` in the same synchronous turn, before change detection pushes a
  new input value in, so a value input would send the previous filter and a cleared chip would stay
  applied. The two embedded hosts leave it unbound, the key is omitted from the request, and the
  API's zero-value filter adds no predicates.
- **`TaskTableComponent` refreshes through one `switchMap`** (`listenForRefreshRequests()`, wired in
  the constructor; `getTableData()` just pushes onto its `Subject`), with `catchError(() => EMPTY)`
  on the *inner* observable so an error cannot complete the outer subscription and kill every later
  refresh. Clearing two chips in quick succession is the same last-response-wins race the receipts
  month stepper hit.
- **"Ran By" is a `list` field, not `users`.** `app-user-autocomplete` cannot prepend the pinned
  **"System"** option (`SYSTEM_RAN_BY_OPTION_ID = -1`) that matches the rows with no `ranByUserId` —
  most of the table. Same reason the report builder's paid-by picker is a plain `app-autocomlete`.
  Both types offer the same `CONTAINS`-only operation, so the row is identical either way. The API
  turns the sentinel into an `IS NULL` disjunct (see `api/CLAUDE.md` → "System task filtering").
- **The Type picker omits two types.** `SYSTEM_TASK_TYPE_OPTIONS`
  (`src/constants/system-task-type-options.ts`) drops `CHAT_COMPLETION` and `OCR_PROCESSING`:
  `GetPagedSystemTasks` never returns them as top-level rows (they are children, shown in an
  expanded row), so offering them would be a picker that can only ever return zero rows. Keep
  `CHILD_ONLY_SYSTEM_TASK_TYPES` in sync with `filteredSystemTaskTypes` in the Go repository —
  `TestGetPagedSystemTasksExcludesChildTaskTypes` pins that side.
- **"Receipt Uploaded" is offered, and matches only parentless rows.** A manual create or duplicate
  records one with no parent, and those are top-level rows; a quick scan or email upload's stays
  nested under its parent. The row needs no special rendering: `SystemTaskTypePipe` already labels
  it and the generic description path pretty-prints the receipt JSON. See `api/CLAUDE.md` →
  "System task filtering".
- **The persisted slice predates the filter, so every read must tolerate its absence.**
  `systemTaskTable` was already in both storage-key lists, so an existing session rehydrates with no
  `filter` key at all. `SystemTaskTableState.filter` / `.numFiltersApplied` and the
  `SetSystemTaskFilterField` handler all fall back to `buildDefaultSystemTaskFilter()`; without that
  the page throws for anyone who has ever loaded it before. Covered by a spec case.
- **`buildDefaultSystemTaskFilter()` is a factory, not a shared constant** — the value is written
  straight into state, so handing out one object would let a later in-place edit corrupt the default
  for the rest of the session. Same reasoning as `buildDefaultReceiptFilter`.
- **Every filter write outside the dialog goes through `applyFilterChange()`**, which dispatches the
  action, then `SetPage(1)`, then refetches — narrowing from page 7 can never land on an empty page.
  The dialog's `afterClosed()` does the same.
- **Date fields are timestamps, and the server widens them to whole days.** `EQUALS` means that
  calendar day, `BETWEEN` runs to the end of the last day. See `api/CLAUDE.md` → "System task
  filtering" for the query side.
- **The date fields go on the wire as `YYYY-MM-DD`, normalized by `toSystemTaskWireFilter`** (now a
  thin wrapper over the shared `toDateWireFilter` — see "App time zone")
  (`src/utils/system-task-filter.ts`) where `TaskTableComponent` assembles the request. The
  datepicker writes a local-midnight `Date`, which serializes as an *instant*; the server resolves
  that instant to a day in **its** zone, so a UTC-4 browser picking Sep 22 selects Sep 21 against an
  API in America/Los_Angeles. A calendar day carries no zone to misread.
  - **Normalize at the request, never in the store.** NGXS persists the filter and hands it back to
    the datepicker when the dialog reopens, and Material's `NativeDateAdapter.deserialize` matches a
    bare `YYYY-MM-DD` against its ISO-8601 regex and parses it with `new Date()` — UTC midnight,
    which renders as the *previous* day west of Greenwich. Storing the normalized form just moves
    the off-by-one into the picker.

**E2E:** `e2e/system-task-filter.spec.ts` (serial, admin storageState). It seeds a deterministic row
by creating and deleting an API key (`apiRecordApiKeyDeletedSystemTask` — system tasks are only ever
written as a side effect of real work, so there is no endpoint that creates one), then asserts what
the Jest specs cannot: the server narrows (`totalCount` included, which is what proves the predicates
land before the count), the "System" sentinel matches unattributed rows, a Started At of the task's
own day matches while the previous day does not, and the filter survives a reload. It also creates a
receipt through `POST /api/receipt` in a throwaway group and asserts its Receipt Uploaded task comes
back top-level, with every row the type filter returns parentless.

## Receipt update diff (System Tasks)

An **Updated Receipt** (`RECEIPT_UPDATED`) row stores the receipt before and after the edit. The
description cell summarizes it ("Changed: name, amount", or "No changes") and its open button shows
`app-receipt-update-diff-dialog` (`src/shared-ui/receipt-update-diff-dialog/`): the two receipts as
pretty-printed JSON, before on the left and after on the right, like a split-view diff.

- **The description is double-encoded, which is why it has its own parser.** The API stores
  `{"before":"<receipt JSON>","after":"<receipt JSON>"}`, each side a JSON *string*.
  `PrettyJsonPipe`'s cleanup rewrites every `"` inside a string value to `'`, so the rebuilt text is
  invalid JSON and the pipe falls back to printing the raw escaped string. That was the "busted"
  display. `parseReceiptUpdateDescription` (`src/utils/receipt-update-description.ts`) parses each
  side once more and never throws. A **failed** update stores its plain error text, which returns
  `undefined`, and that row renders through the old `app-pretty-json` path like every other task.
  `RECEIPT_UPDATED` is the only double-encoded type, so the pipe is untouched.
- **`TaskTableComponent.receiptUpdates` parses once per page** (a `computed` over the data source,
  keyed by task id), not per change-detection pass, since each row holds two whole receipts.
- **The summary ignores the record-keeping fields at every depth**: the API's `BaseModel`, i.e.
  `id`, `createdAt`, `updatedAt`, `createdBy` and `createdByString`. An update deletes and recreates
  the receipt's items (linked items included) and custom field values, so they come back with new ids
  and timestamps on *every* save, and it bumps `updatedAt` on the receipt, its categories and its tags.
  Comparing those would list `receiptItems`, `customFields`, `categories` and `tags` on every row
  whose receipt has any. The **diff itself is deliberately raw**, so those lines still show as changed
  there.
- **The diff is hand-written** (`src/utils/line-diff.ts`): it trims the common prefix and suffix, runs
  Myers' O(ND) diff on the rest, and keeps each round's frontier only for the diagonals it can read,
  so memory is O(D²). Removed and added lines in one run are paired side by side as `changed` rows,
  and `inlineChange` marks the part that differs. It avoids a new runtime dependency (and the
  sandbox lockfile drift above). `line-diff.spec.ts` fuzzes it against a reference LCS, so a
  regression that stays *valid* but stops being *minimal* still fails.
- **"All lines" is the default**, and "Changes only" (`collapseUnchanged`, 3 lines of context)
  replaces longer unchanged runs with a "⋯ N unchanged lines" row. A run of one line is shown rather
  than replaced by a marker the same size.
- **The code cell's content is written on one template line.** The column is `white-space: pre-wrap`,
  so template whitespace inside the `<td>` would render as indentation.
- **Older rows are versioned, and the dialog says how far to trust "before".** The API writes
  `version: 2`. A row without the key is version 1, whose own "before" is incomplete (see
  `api/CLAUDE.md` → "Receipt update snapshots"). The listing rebuilds such a row from an earlier
  complete copy when one exists, and names it in `beforeSource`. `receiptUpdateBeforeState` turns
  that into one of three states, and the dialog renders a notice for the last two
  (`data-testid="receipt-diff-version-notice"`, `data-before-state`):
  - `complete`: a version 2 row; no notice.
  - `rebuilt`: an `alert-info` naming the copy, "as saved when it was created" or "by the previous
    update", and its date. It also warns that bulk status changes can fold in.
  - `incomplete`: an `alert-warning` that item categories and tags, shares and custom field names
    can show as changed when they weren't.
  The notice sits **inside** the scroll box, above the table, so it scrolls away rather than
  shrinking the diff; the sticky header still pins to the box's top. Nothing else on the desktop
  branches on the version, because the summary and diff simply read the rebuilt "before".

**E2E:** `e2e/receipt-update-diff.spec.ts` updates a receipt through the real API and asserts the
row summary, the old/new name on the correct sides, and the "Changes only" collapse. The receipt
keeps an unchanged item across the update under test, so the summary assertion also proves the
server-side item recreation is not reported as an edit (it fails if only `updatedAt` is ignored).
It also asserts the stored row carries `version: 2` and that the dialog shows no notice. The current
API cannot write a version 1 row, so that path is covered by the Go service tests and the Jest specs. It narrows the
paged response to its own task with `page.route` + `route.fetch()`, so the row is still the real
server's output while other specs' tasks stay out of the way.

## Quick Scan Configuration

- **Group receipt settings** (`src/group/group-receipt-settings/`) has a **Quick Scan** section: per
  field (paid-by, status, categories, tags, comment) a *Show* + *Require* `app-checkbox`, plus a default
  control for paid-by (`app-select` of Uploader/Specific user + a conditional `app-user-autocomplete`) and
  status (`app-status-select`) shown only when that field is not both shown+required. The component
  mirrors the backend rule as reactive validators (default required unless shown+required) and coerces an
  empty `quickScanDefaultPaidById` to `undefined` on submit. See `api/CLAUDE.md` → "Quick Scan Field
  Configuration".
  - **The comment toggles are DISABLED (greyed out), not cleared, while `hideComments` is on** — hiding
    comments group-wide hides the quick-scan comment too, and the derivation is self-healing: the stored
    values stay put and apply again the moment `hideComments` is unticked. `applyQuickScanDerivedState`
    (run on init **and** from the `valueChanges` subscription) does the enable/disable with
    `{ emitEvent: false }` — it runs *inside* that subscription, and `enable()`/`disable()` emit by
    default, which recurses forever — and returns early outside `FormMode.edit`, because `initForm`
    disables the whole form *after* subscribing and that disable emits (an unguarded `enable()` would
    make two checkboxes editable on the read-only view page).
  - **`submit()` MUST read `form.getRawValue()`, not `form.value`.** A disabled control is omitted from
    `form.value`, so with `hideComments` on the two comment toggles would be sent as `undefined`,
    unmarshal server-side as `false`, and the API's unconditional assignment would **wipe the admin's
    stored configuration**. Guarded by a spec case that submits while they are disabled.
- **Quick scan dialog** (`src/receipts/quick-scan-dialog/`) resolves each image's config from **that
  image's selected group** (`GroupState.getGroupById(...).groupReceiptSettings`) to drive per-image
  field visibility + required validators; hidden paid-by/status are sent empty so the server backfills
  the group default.
  - **Until a group is picked for an image, ONLY its Group field renders.** There is no configuration
    to honour yet, and the old fallback (paid-by/status shown+required) was a guess that flipped the
    field set the moment the user chose a group whose config hides them. The derivation lives in one
    pure helper, **`resolveQuickScanFieldConfig(settings, { hasGroup, canCreateComments })`**
    (`quick-scan-field-config.ts`) — a line-for-line port of mobile's
    `lib/shared/functions/quick_scan_field_config.dart`, and the single source for both the
    `show*(i)` getters and `configureImages()`, which previously each carried their own copy of the
    `?? true` / `?? false` default table.
  - **`hasGroup` is "a group id is chosen", NOT "settings resolved".** `settingsForIndex` returns
    `undefined` for *two* states — no group, and a group id the store doesn't know (a stale
    `quickScanDefaultGroupId`, or AppData not carrying it; pinned by the spec case
    `'should push new image when there user preferences'`, which seeds a group absent from the
    store). Only the first collapses to Group-only.
  - **`configureImages()` gates its CLEARING on `hasGroupAt(i)`, but not its `setRequired`.** Hidden
    *because this group's config hides it* must clear, so the server backfills its configured default
    instead of receiving a stale value (what the `preset paid-by falls off the submission` e2e pins).
    Hidden *because no group is picked yet* must **not**: those values are the caller's
    `quickScanDefault*` prefills, and every `FormArray.push` in `fileLoaded()` emits `valueChanges`,
    so `configureImages()` runs **before `groupIds` is even pushed** — an ungated clear wipes the
    prefills on the first pass and never restores them. Requiredness is recomputed unconditionally
    because a validator left over from a group the user just **cleared** (the autocomplete's X, which
    the e2e helper `selectImageGroup` clicks on every switch) has to come off. Note desktop *can*
    return to the no-group state mid-session this way; mobile's dropdown offers no null option, so
    there it is only ever the initial state.
  - `configForIndex(i)` is resolved per call, never cached per index — `removeImage()` shifts every
    later image down, so an index-keyed cache would hand an image another one's config. Category/tag pickers (`app-category-autocomplete`/`app-tag-autocomplete`,
  `[creatable]="false"`) source options from `AuthState.groupCategories`/`groupTags` and are serialized
  as per-image comma-joined id strings for `quickScanReceipt(...)`. The **comment** is an
  `app-textarea` (`data-testid="quick-scan-comment"`) backed by a scalar per-image `FormControl` in a
  `comments` `FormArray` — already one string per image, so unlike categories/tags it needs no id
  joining. `showComment(i)` ANDs the group's `quickScanCommentEnabled`, `!hideComments`, and the
  caller's **`group.comments.create`** for that image's group (read via `selectSnapshot` and memoized
  per group id — `AuthState.hasGroupPermission()` allocates a fresh selector per call and the getter
  runs every change-detection pass). Without the permission the field is hidden and never required, so
  a member who cannot comment is never locked out of quick scan; the server drops any comment they send.
  - **The comment's two validators both mirror backend semantics, and neither is the stock Angular
    one** (`src/validators/text-validators.ts`, shared): `codePointMaxLengthValidator(500)` counts
    **code points** (`Array.from(v).length`) because `Validators.maxLength` counts UTF-16 code units
    and would reject `"😀".repeat(500)` — 1000 units but 500 runes, which the API accepts;
    `trimmedRequiredValidator()` treats whitespace-only as blank because `QuickScanCommand` **trims**
    each comment on parse, so `"   "` reaches the required check empty and 400s while
    `Validators.required` called it valid. Both emit the standard `maxlength` / `required` error keys
    so the shared inputs' message mapping is unchanged. The length cap is attached where the control
    is **created** (it must survive every show/require recompute); the required one is toggled by
    `setRequired`, whose optional third argument overrides which validator is toggled — pass a
    **module-level singleton**, since `removeValidators` matches by reference identity.
  - **`base-input` has no message for the `maxlength` error**, so an unmapped one renders an *empty*
    `mat-error`. The `app-textarea` passes `[additionalErrorMessages]` for it (same pattern as
    `prompt-form` / `receipt-filter`). Any new non-stock error key needs the same treatment.
  - **Each image's `categories`/`tags` control MUST be a `FormArray`** (`this.formBuilder.array([])`),
    *not* a `FormControl([])`: `app-category-autocomplete`/`app-tag-autocomplete` run in `multiple`
    mode, and the base `app-autocomlete`'s `optionSelected` **pushes** the picked option onto the
    control (`inputFormControl.push(...)`) — exactly as the receipt form's `categories` FormArray does.
    A plain `FormControl` has no `push()`, so a selection throws `push is not a function` and silently
    adds nothing (the picker looks dead). Clear a hidden field with `FormArray.clear()`, not
    `setValue([])` (which throws on a non-empty array). Guarded by `quick-scan-dialog-behavior.spec.ts`
    (picks a category and asserts the submit carries its id).
- **E2e** (`e2e/quick-scan-config.spec.ts`, `e2e/quick-scan-dialog.spec.ts`, `e2e/quick-scan-dialog-behavior.spec.ts`,
  admin storageState): the config page is driven directly (checkboxes carry `data-testid`s
  `quick-scan-<field>-show/-require` because the "Show"/"Require" labels collide across the four field groups).
  The dialog is gated by the `aiPoweredReceipts` feature flag (off in dev/CI); rather than mutate that global
  server state, the specs **intercept `GET /api/user/appData`** (`page.route`, like `stubTokenRefresh`) to flip
  `featureConfig.aiPoweredReceipts` true **and** inject the target group's `groupReceiptSettings` (plus
  `userPreferences.quickScanDefault*` and `groupCategories`/`groupTags` catalogs) — a per-BrowserContext
  client-side stub with no server side effects (the negative `receipt-feature-gating.spec.ts` still sees the
  button absent). The shared injector + a multipart field parser live in **`e2e/helpers/quick-scan.ts`**
  (`injectQuickScanAppData`, `parseMultipartFields`, `openQuickScanDialog`, `selectImageGroup`,
  `uploadQuickScanImages`). The Quick Scan header button is icon-only (tooltip is `aria-describedby`, not the
  a11y name), so it carries `data-testid="receipts-quick-scan"`; the dialog's carousel nav buttons carry
  `data-testid="quick-scan-nav-left/-right"` for the same reason. This appData-injection pattern is the general
  way to e2e any feature-flag-gated UI here.
  - `quick-scan-dialog-behavior.spec.ts` covers the deeper matrix: a **user-preference paid-by preset falls
    off** the form (and the submission) when the group hides paid-by; **switching an image's group re-flips**
    its field set; a **category picked from the catalog** rides the multipart; and **two images on different
    groups** get independent field sets where one image's unmet required field blocks the whole submit. The two
    **submit** tests **mock `POST /api/receipt/quickScan`** (the backend validates each group's *persisted*
    config, which the client-side injection doesn't touch, so a real submit would 400) and assert the exact
    multipart the client builds via `parseMultipartFields` — e.g. a hidden paid-by is sent as the empty
    sentinel. To **change** an already-selected group in a single-select `app-autocomlete`, click its **X clear
    button** first (the input goes `readonly` once a value is chosen); `selectImageGroup` handles this.
  - The comment axis is covered on both sides: `quick-scan-config.spec.ts` asserts the toggles grey out
    (keeping their values) with **Hide Comments** and that the configuration **round-trips through a
    save + reload** (the guard for the API's field-by-field persistence), and
    `quick-scan-dialog-behavior.spec.ts` asserts a required comment blocks submit then rides the
    multipart, and that the field is hidden both without `group.comments.create` and when the group
    hides comments. The permission case uses the new **`groupPermissions`** option on
    `injectQuickScanAppData` — inject a group's effective permissions client-side instead of
    provisioning a custom role.
  - `receipt-feature-gating.spec.ts` now has the **positive** Quick Scan contrast (previously `test.fixme`):
    with the flag injected on, a **Legacy Editor** member (holds `group.receipts.quick-scan`) sees the button
    while the Viewer — same user, same flag — does not.

## Magic Fill (receipt form)

Distinct from Quick Scan (which creates a receipt entirely server-side): the ✨ **Magic Fill** button on the
receipt form (`src/receipts/receipt-form/`, `magicFill()` → `patchMagicValues`) calls
`POST /receiptImage/magicFill`, which returns the backend's fully-parsed `UpsertReceiptCommand`, and patches
it **into the form** for the user to review before saving. `patchMagicValues` ingests the **whole** receipt —
name/amount/date/paid-by/status, categories/tags, `receiptItems` (items **and** shares — items with a
`chargedToUserId` — plus nested `linkedItems` and per-item categories/tags), `customFields`, and `comments` —
reusing the form's existing builders (`buildItemForm`, `buildCustomOptionFormGroup`, and the comments child's
`buildCommentFormGroup`) so `this.form.value` round-trips into `UpsertReceiptCommand` on submit with nothing
dropped. Each field is reported filled (and named in the success toast, via a friendly-label map) **only when
it actually changes the form**, so an empty or unmatched value never claims a phantom fill. Guarded by
`receipt-form.component.spec.ts` → `describe("magicFill — full receipt ingest")`.

Two cross-component seams support this (each with its own focused spec):
- **Paid-by display:** `patchValue` updates the `paidByUserId` control but not the autocomplete's shown text
  (the single-select display is seeded from the control only once on init), so
  `AutocomleteComponent.syncSingleDisplay()` re-seeds it after the patch.
- **Comments:** the `app-receipt-comments` child owns the comments array and is **mode-aware**, so Magic Fill
  hands them to `ReceiptCommentsComponent.addMagicFilledComments()` — add mode collects them for the
  single-call create submit; edit mode POSTs each via `CommentService` because the receipt-**update** path does not persist
  comments (they're individual resources — see `api/CLAUDE.md` → `UpdateReceipt`).
- **Custom fields** reference a field by id only (the magic-fill response carries no field definition), so a
  value whose `customFieldId` isn't in the loaded catalog pool is skipped, and adding one flips its
  manage-fields menu entry to selected via an immutable array replace (zoneless CD).

## Receipt image canvas

The receipt form's inline image is an interactive canvas: four corner handles to pull, drag to pan,
wheel to zoom, double-click to fit. It replaced a viewer whose only controls were two zoom buttons
and a free `cdkDrag` on the `<img>`.

**A first attempt resized the form/image *panes* with a draggable splitter. That was reverted** — it
moved the page layout, when the thing worth manipulating is the image. Don't reintroduce it.

- **`app-image-canvas`** (`src/shared-ui/image-canvas/`, standalone, registered in `SharedUiModule`'s
  `imports` + `exports`) is generic and presentational: it takes a resolved `src` and a `stageHeight`
  and owns the whole interaction. The stage is a **fixed viewport** — the image is clipped by it and
  panned around — so growing the image can never move the surrounding page.
- **Scale and pan are one transform on one element.** The old viewer put `scale()` on a wrapper div
  and let `cdkDrag` translate the `<img>` inside it, so drag distance desynchronised from the cursor
  by a factor of the scale. One matrix is what makes direct manipulation tractable.
- **Nothing in the template reads the viewport.** It is a plain field, not a signal: the image and the
  handle frame are both positioned imperatively, so a 120Hz drag runs **no change detection at all**
  over the receipt form's large, non-`OnPush` template. `pointermove` is likewise a native listener
  rather than a host binding, which would schedule CD per event.
- **Handles sit fully INSIDE the image box, not centred on its corners.** The stage clips with
  `overflow: hidden` and clipped pixels are not hit-testable, so a centred handle loses half its
  target the moment the image meets a stage edge — which at the fit scale is always true of two of
  the four. Each also grabs from an invisible 24px box (`::before`, `inset: -6px`).
- **The image floats on the stage; it is not contained by it.** It may be smaller than the stage,
  sit in a corner, or hang over an edge. The only positional rule is that `MIN_VISIBLE_PX` (48) of it
  stays on the stage, so nothing can be flicked out of reach. This replaced a "centre any axis
  narrower than the stage, and never allow dead space at an edge" rule, and **that rule is what made
  the handles fake**: centring throws away the anchor a corner drag just established, and forbidding
  dead space slides the image rather than pinning the corner you did *not* grab. Do not reintroduce
  either — a containment rule and a real resize handle cannot both exist.
- **The resize floor is a minimum size, not the fit.** `clampScale` used to floor at `fitScale()`,
  and since the image *opens* at the fit, pulling a handle inward was a guaranteed no-op — handles
  could only ever grow the image, which is a zoom button with extra steps. The floor is now
  `minScale()`, the scale leaving the image's shorter side at `MIN_VISIBLE_PX`, below which the four
  12px handles would overlap. `fitScale()` is unchanged and is still what load and double-click use.
- **The sliver outranks the anchor** on the one state where they conflict. With the opposite corner
  already off-stage, holding it while the image shrinks would carry the whole image off the stage —
  so the resize keeps going and lets that invisible anchor drift instead. Constraining the *scale* to
  preserve the anchor is the tempting alternative and it is wrong: it freezes the handle exactly
  where the image is hardest to recover, which is the dead handle this section exists to prevent.
  Pinned by its own spec.
- Two consequences worth knowing. **`fit()` centres the image itself** — committing `x: 0, y: 0` and
  letting the clamp centre it no longer works. And **a stage resize re-clamps the position only**:
  the bounds are stage-relative so a stage that shrank can strand the image, but the scale is the
  user's, and a stage that *grows* no longer re-grows an image that was showing in full.
- Wheel zoom is anchored at the cursor and moves a meaningful amount. The viewer this replaced
  multiplied `deltaY` by `-0.000001`, so one notch changed the scale by 0.0001 — wheel zoom did
  nothing at all.
- **A handle stops `dblclick` itself.** The host binds `(dblclick)="fit()"` and the handles are its
  descendants. `onResizeStart` already calls `preventDefault()` + `stopPropagation()`, but those act
  on the **PointerEvent** — `dblclick` is a separate event and is not suppressed by either, so
  quickly pulling a corner twice used to bubble up and throw away the size just dragged to. Hence
  `(dblclick)="$event.stopPropagation()"` on the handle, pinned by its own spec.

**Wiring.** `app-carousel` gained `stageHeight`, passed to `app-image-viewer`, which renders
`app-image-canvas`. **Both** carousels get it — the inline one and the fullscreen dialog's (see "The
fullscreen dialog runs the same canvas" below); there is no opt-out, and no second rendering path.
The canvas rolled out behind a `directManipulation` flag while only the inline view used it; once
fullscreen was ported, both call sites passed `true`, so the flag and the old path it selected were
deleted rather than left as dead configuration.

**What went with the old viewer.** Its `@else` branch (a `transform: scale()` wrapper around an
`img.viewer-image` with an **unbounded `cdkDrag`**, which could fling the image out of view with no
way back), the carousel's `scale` / `adjustScale()` / `onScroll()` — including the `deltaY *
-0.000001` wheel bug above — and the `[scale]` / `wheel` plumbing between them. `image-viewer` no
longer has a stylesheet; `.viewer-image` was its only rule. `DragDropModule` stays: `cdkDrag` is
still used by the column-configuration dialog and the receipt-form dialog.

**Zoom buttons are per-image.** The header Zoom In/Out delegate through the carousel to the *active
slide's* viewer rather than mutating one `scale` shared by every slide, which is what a canvas
implies. That lookup is `viewChildren(ImageViewerComponent)` indexed by `currentlyShownImageIndex`
— a **slide** index into a **viewer** query, so the two only agree while every slide renders exactly
one viewer. This is why `carousel.component.html` renders `app-image-viewer` **unconditionally** in
both loops: the `*ngIf`s that used to gate it were redundant (the viewer already renders nothing
without a source) and would silently compact the query, pointing the buttons at the wrong slide. Do
not reintroduce them.

**The stage is exactly as tall as the Details pane beside it**, so the two columns move together
like a two-column grid rather than a short image box floating next to a tall form. Bootstrap's
`.row` is already `display: flex` with the initial `align-items: stretch`, so the Images column is
*already* the right height — the work is entirely in making its contents fill it, because every link
between the column and the canvas is `height: auto` and three hosts (`app-carousel`, ngx-bootstrap's
`<carousel>`, `app-image-viewer`) are `display: inline`.

The height is handed down in two places, both **gated on a class**:
- `carousel.component.scss` is global (`ViewEncapsulation.None`), which makes it the right home for
  the ngx-bootstrap links. `.rw-carousel--fill` carries `display: block; height: 100%` down through
  `carousel`, `.carousel.slide`, `.carousel-inner`, `.carousel-item.active`, `.item` and
  `app-image-viewer`. Use `.carousel-item.active`, not `slide`, so the inactive slides ngx-bootstrap
  takes out of flow stay out of it, and keep them `block` — `flex` fights Bootstrap's `float: left`.
- `receipt-form.component.scss` covers its own template under `.rw-images-pane`: a flex column so the
  always-present section header takes its own height and the content takes the rest. One `::ng-deep`
  is unavoidable — `app-form-section` wraps projected content in a **classless** div of its own,
  reachable only as `.form-section-header + div`.

`stageHeight` is bound to `[style.height]`, an inline style, so the receipt form simply passes
`stageHeight="100%"`; no rule on the canvas and no `!important`. The canvas needs no change at all —
its `ResizeObserver` already re-clamps when the stage resizes, which is what makes a layout-derived
height safe, and the Details pane's height genuinely does change live (category/tag chips wrap as you
type).

**Below Bootstrap's `md` the panes stack into one column.** Both sections are `col-12 col-md`, so
the stacking itself is plain Bootstrap rather than a custom media query — without it `.col` holds a
50/50 split all the way down, which goes lopsided around 600px as the form's minimum width wins and
only wraps at phone sizes. Once stacked there is no Details pane beside the stage to take height
from, so `.rw-images-pane__stage` gets an explicit **50vh** under `@media (max-width: 767.98px)`;
the carousel chain resolves its `100%` against that. Leave that media query out and the stacked
canvas collapses to its 2px border.

**`app-upload-image` must not carry `h-100` in that column.** The class was inert while its host was
`display: inline`, but a flex container blockifies its children — at which point
`height: 100% !important` makes a hidden file input swallow the whole column and leaves the stage
with nothing but its own 2px border.

**The fullscreen dialog runs the same canvas.** `#expandedImageTemplate` passes
`class="rw-carousel--fill"` and `stageHeight="100%"` exactly as the inline carousel does, plus a
floating close button — it previously had **no way out but Esc and the backdrop**. The height chain
needs no special handling: MatDialog renders a `TemplateRef` as a `TemplatePortal` whose nodes go
**straight into `.mat-mdc-dialog-surface`** with no wrapper, and that
surface is viewport-tall and padding-free (dialog padding lives on `.mat-mdc-dialog-content`, which
this dialog does not use). Two things there are load-bearing:
- **Do not wrap it in `app-dialog`.** That costs ~110px before the stage starts — an
  always-rendered `<h2>` (~39px even when `headerText` is empty), a `.p-4` wrapper and
  `mat-dialog-content`'s `padding: 20px 24px` — **and caps the body at `max-height: 65vh`**. A
  `height: 100%` stage inside it collapses.
- **Do not add a `maxHeight` to the dialog config.** Leaving it undefined is what puts the CDK on its
  flush-vertical path (`shouldBeFlushVertically` requires no maxHeight), giving the pane the full
  viewport height. Setting one silently re-centres the dialog and shortens it.
`autoFocus: "dialog"` because the close button is the only tabbable control, so the default
first-tabbable focus opens the viewer with a focus ring drawn over the image.

**ngx-bootstrap's carousel controls must stay small over a canvas.** Bootstrap sizes prev/next as
`top: 0; bottom: 0; width: 15%` at `z-index: 1`, as siblings of `.carousel-inner` — so on a
full-height stage they become two full-height columns sitting *above* the canvas in hit-testing.
Navigation still works, but pan, wheel zoom and double-click all die in those strips, and both the
left and right corner handles live exactly there, undoing the inset that keeps them grabbable at a
stage edge. `.rw-carousel--fill` shrinks them to 2.75rem circles centred vertically and clear of
every corner. This bites on **any receipt with 2+ images**, inline as well as fullscreen.

**Gating every rule on a class is not optional.** `#expandedImageTemplate` is declared in
`receipt-form`, so its embedded view carries that component's encapsulation attribute even while it
renders inside the MatDialog overlay — an ungated `app-carousel { height: 100% }` reaches into the
dialog too. It cost a distorted `img.viewer-image` back when the dialog rendered the old viewer;
both carousels now render the canvas, so the immediate symptom is gone, but the reach is not and
neither is the rule.
`receipt-form.component.spec.ts` pins that both carousels — the inline one and the dialog's — carry
the fill class and `stageHeight="100%"`.

**The collapse/expand toggle is gone**, along with the `showLargeImagePreviews` checkbox in User
Preferences that seeded it — the canvas replaced what the toggle was for. **The preference is now
removed end to end**: the Go model field, the `swagger.yml` property, and both generated clients.
Only the physical `user_preferences.show_large_image_previews` column survives on already-migrated
installs — AutoMigrate never drops columns, and the column is nullable with no default, so an orphan
is inert. Two things made the removal safe for already-released mobile builds, and both are worth
knowing before removing any other response field: the generated Dart field was `bool?`, so an
**absent** key never enters the deserializer's `switch` (an explicit `null` would still fail the
`as bool` cast — dropping the Go field guarantees the key is omitted, not nulled), and the API sets
no `DisallowUnknownFields`, so an old client still PUTting the field is ignored rather than 400'd.

`carouselComponent` is also no longer `viewChild.required`: the Zoom/Download/Fullscreen header
buttons render whenever there are images, but the carousel itself is behind `*ngIf="… && showImages"`,
so hiding images and clicking one of them used to throw.


## Reports (Report Builder)

The **Report Builder** (`src/reports/`) is a two-pane screen for building and downloading receipt
reports against the backend reporting engine (see `api/CLAUDE.md` → "Reporting Engine"). The lazy
`ReportsModule` is gated by `appPermissionGuard` on **`app.reports.read` OR `app.reports.readAll`** (the
avatar-menu "Reports" entry gates on the same via a `hasAnyAppPermission([...])` signal, since the
`*hasAppPermission` directive is single-key). Its routes: `/reports` is the **templates list** landing
(below), and the builder lives at `/reports/new` and `/reports/:id/edit` (both `fullHeight`).
**Per-template access is enforced end-to-end** (see `api/CLAUDE.md` → "Report-template access"): the list
is server-filtered to the user's visible templates and each row's action buttons are gated purely on the
server-computed **`element.allowedActions`** (never AND-ed with a client `*hasAppPermission` — that would
wrongly hide a button from an `*All`-only holder). Row **Generate** runs the template by id through
`ReportRunnerService.generateFromTemplateById` → `POST /report/template/{id}/generate` (the enforcing
endpoint); the builder's own ad-hoc generate still gates on `app.reports.generate` + per-group
`group.reports.read`. The in-builder group picker only lists groups where the user holds `group.reports.read`.

- **Builder state** — the *builder* is a single reactive form (`report-form.factory.ts`) plus signals, no
  NGXS (the templates *list* is the module's one NGXS slice, `ReportTemplateTableState` — see "Templates
  list" below);
  generate/preview are one-shot calls through `ReportRunnerService` (mirrors `ReceiptExportService`:
  generate → `Blob` → `downloadFile`). `ReportCatalogService` supplies the dimension/measure dropdown
  options: a built-in engine-key→label constant (`report-catalog.constants.ts`) plus custom fields from
  `CustomFieldService`, keyed `custom_<id>`. `report-command.mapper.ts` maps the form to the generated
  `ReportRequestCommand`.
- **Custom fields in the catalog.** *Every* custom field is a **dimension** whatever its type — the
  engine restricts measuring, not cutting — so a **CURRENCY** field appears in **both** lists (groupable
  *and* summable), and a **DATE** field additionally contributes the three calendar-period dimensions the
  backend derives (`custom_<id>_day|_month|_year`, labelled `"Due Date (Month)"`); grouping by the raw
  date buckets on the exact instant, i.e. one bucket per receipt. Keys and labels mirror
  `receiptsource.CustomFieldPeriodKeys` / `dateFieldRefs` — a mismatch is a 400 from the engine, not a
  silent fallback. Every custom-field entry carries `isCustom: true`, which drives the **"Custom" badge**:
  `toFieldOptions()` (the one mapping all four field dropdowns share) turns it into an option `badge`,
  and the shared **`app-select` gained an optional `optionBadgeKey`** input that draws it beside the
  option text. Because `MatOption.viewValue` is the option's `textContent`, a badged option would read
  "HSTCustom" in the closed select — so a badged select also renders its own `<mat-select-trigger>`.
  Both are opt-in and default off, so every other `app-select` call site is unchanged. The badge itself
  is the shared **`app-badge`** (see "The shared badge" above), as are the picked rows' custom and kind
  badges; `app-select` also takes an `optionBadgeTone` (defaulting to the custom-field purple) so a
  select badging something other than a custom field can say so. The already-picked
  grouping levels and column rows carry the same badge (`isCustom` on `groupByLevels()` / `columnRows()`,
  resolved through the catalog rather than by matching the key's shape, so a user without
  `app.custom-fields.read` sees no badge instead of one on a field the builder cannot name).
  - **E2E:** `e2e/report-custom-fields.spec.ts` (serial, admin storageState) seeds its own group +
    a CURRENCY/DATE/BOOLEAN field trio + four receipts through the admin API, so the live preview
    contains exactly its own data, and covers: the grouping picker offering every custom field badged
    (including the three derived `(Day|Month|Year)` levels), grouping by a **currency** field, a
    boolean rendering **Yes/No**, a date field bucketing by **calendar month**, `SUM` over a custom
    currency measure, and a saved template naming its fields rather than showing `custom_<id>`.
    Teardown deletes the group (cascading the receipts, whose custom field values a field-delete would
    otherwise orphan) then the fields — a leaked field shows up in every other spec's pickers.
    Each assertion was verified to **FAIL** with its feature reverted (the renderer's typed
    formatting, the catalog's currency-as-dimension + period entries, and the list's field-name
    resolver).
    - Two gotchas it encodes: the "Custom" badge is inside the `mat-option`, so an option's accessible
      name reads `"Tip Custom"` — `addGroupingLevel` therefore takes a `string | RegExp` (a plain
      string still matches exactly, which is what every other caller wants). And money is asserted
      with a **separator-tolerant** regex (`/1[,. ]500[.,]50/`) rather than a symbol, because the
      currency configuration is a global System Setting on the shared CI backend that this spec must
      not mutate; the seeded value is `1500.50` precisely so the thousands separator and trailing
      zero prove formatting ran.
- **Period date field.** A **"Date field"** `app-select` sits in Parameters, after "Period covering"
  and its custom Start/End. It chooses which receipt date the period filters on, bound to
  `period.dateField`, and the hint reads "Resolves to … on Added At". See `api/CLAUDE.md` →
  "The period's date field" for the wire contract.
  - **Its options are `RECEIPT_DATE_FILTER_FIELDS` mapped as-is** (`periodDateFieldOptions`), so they
    match the receipts table's quick date filter in content and order by construction.
    `ReportBuilderValue.period.dateField` is typed `ReceiptDateFilterFieldKey`, and the panel spec
    pins the list.
  - **A new report defaults to `DEFAULT_QUICK_DATE_FIELD`, and the mapper always sends the field**, so
    a saved template records which date it covers.
  - **Rehydrating goes through `toReportPeriodDateField`** (`report-period.util.ts`), which falls back
    to the **literal** `LEGACY_REPORT_PERIOD_DATE_FIELD` (`"date"`), not the quick filter's default.
    Templates saved before the picker have no field and always ran on the receipt date; a later
    change to the table's default must not change what they cover. The swagger field is a plain
    `string`, so an unrecognized stored value lands on `"date"` too. Re-saving a legacy template writes
    `dateField: "date"`, so that round trip is intentionally not a fixpoint.
  - **The drill-in never resolves the period itself.** `report-preview-panel.openReceipts()` hands
    the dialog the same `toReportRequestCommand` the preview sends. `report-receipts-dialog` makes
    one `ReportService.getReportReceipts(command)` call (`POST /report/receipts`), and the server
    runs the report's own filter and period.
    - It used to build the BETWEEN here, per group, via `getReceiptsForGroup`. That disagreed with
      the count chip in three ways:
      - it used the browser's time zone, not the server's;
      - on SQLite its ISO bounds compared as text and dropped first-day receipts;
      - the report-generator `-1` paid-by matched nothing.
    - The subtitle still formats the period client-side (display only), resolving "today" in the
      app time zone, and names the field — plus the zone for an instant field.
    - The subtitle's count is the preview's `receiptCount`, falling back to the response's
      `totalCount`, not the list length: the server caps the list at 100.
    - When the list is shorter than its `totalCount`, a `report-receipt-truncated` notice reads
      "Showing the newest N of M receipts", so a partial list is never passed off as the whole report.
  - **E2E:** `e2e/report-period-date-field.spec.ts` (serial, admin storageState, own group).
    - Seeds two receipts dated 2024-01-01, one RESOLVED. The server stamps both "resolved" and "added"
      with now.
    - Asserts the options and order, and the counts per field across a past window and a window from
      2025.
    - The drill-in under Added At lists both receipts, and a saved template reopens on its field.
    - A second block runs the browser in `timezoneId: 'America/Los_Angeles'` against the UTC backend.
      A receipt dated 03:00 UTC on January 1 must appear in the January drill-in, matching the chip,
      and the drill-in shows it as **Jan 1, 2024** (a calendar day, not the browser's Dec 31). The
      browser zone plays no part in either.
    - The count chip reads `receipt_long<N> receipts` (icon ligature flush against the number), so it
      is matched with `(?<!\d)N receipts`.
    - The label contains "field", so a bare `getByLabel('Field')` elsewhere on the builder now
      resolves two elements. `report-grouping-label.spec.ts` scopes its column-picker lookup to the
      dialog for that reason.
- **Live preview** (`report-preview-panel`): the container debounces the form (~450ms, `switchMap`) into
  `POST /report/preview` and renders the engine's returned HTML in a **sandboxed `<iframe srcdoc>`**
  (`sandbox="allow-same-origin"`, scripts disabled; sized to content on load). The response's
  `receiptCount` drives the chip that opens the receipts drill-in (`report-receipts-dialog`, the
  receipts the report covers, listed server-side by `POST /report/receipts` from the preview's own
  command; see "Period date field" above). The drill-in is a read-only list → detail
  inspector: a `selected` signal toggles the list (clickable rows) and a per-receipt breakdown card
  (amount/category/paid-by/tags via the shared `customCurrency`/`name`/`user` pipes + `app-status-chip`);
  "Open full receipt" does `window.open(\`/receipts/${id}/view\`, "_blank")` to view it in a new tab.
- **Filters** (`report-filters`): the design's inline add-a-filter chips, but built on the **shared**
  `buildReceiptFilterForm` (`src/utils/receipt-filter.ts`) and SharedUiModule `OperationsPipe`, so it
  produces the exact `ReceiptPagedRequestFilter` the receipts filter does (same BETWEEN handling) — only
  the presentation differs. Category/tag options are the union of the user's group catalogs.
  - **Visible rows on open-in-builder**: the form always holds every filter field; which rows *show* is a
    local `activeFieldKeys` signal. `addFilter`/`removeFilter` maintain it for edits, and `ngOnInit`
    **seeds it from the hydrated filter** (every field whose stored `operation` is non-empty) — otherwise
    a saved template's filter sits in the form but renders no rows. The value itself relies on the backend
    serializing the filter with lowercase `value`/`tags` keys (see `api/CLAUDE.md` → Report templates).
  - **Dynamic report-generator paid-by (reporting-only)**: the paid-by row is the one place the reporting
    filter diverges from the shared receipts filter — instead of the shared `app-user-autocomplete` it
    uses `app-autocomlete` over `paidByOptions()`, which prepends a pinned **"Whoever generates the
    report"** sentinel (`REPORT_GENERATOR_PAID_BY_ID = -1`, negative so it never collides with a real id)
    ahead of `UserState.users`. The control still stores plain numeric ids (the shared form builder, the command
    mapper, and the round-trip factory are untouched), so a saved template carries the `-1` sentinel and
    the backend resolves it to whoever generates the report — User A running User B's saved report filters
    to User A's own receipts. Mirrors the role editor's `OWN_PAID_RECEIPTS_OPTION_ID` convention; the
    shared receipts filter never offers it. See `api/CLAUDE.md` → "Reporting Engine" (buildModel).
- **Columns** (`report-config-panel` + `column-picker-dialog`): a `FormArray` of columns edited through a
  3-step picker (dimension / aggregate / formula). A column's engine `name` (what formulas reference) is a
  derived identifier kept stable across label edits (`report-column.util.ts`); formula validation is
  lightweight inline feedback — the backend is the authoritative validator (a bad spec → 400, surfaced by
  the interceptor). Grouping levels and columns reorder via up/down (no drag-and-drop).
  - **The section is editable in *both* detail modes.** It used to be replaced by a note in Records mode
    ("columns are the receipt fields") — which was never true: the `columns` FormArray was still posted,
    the user just could not see or change it. In Records mode a detail row *is* one receipt, so the
    engine reads a dimension column straight off that record (`emitRecordRows` →
    `record.Get(column.field)`) and the disabled-dimension rule below does not apply — a column may read
    **any** catalog field with no grouping level for it. Aggregates and formulas are configurable there
    too: each record gets its own accumulator set (so `SUM(amount)` is that receipt's amount and
    `COUNT()` is 1), and they still roll up correctly across subtotal/grand-total rows. A Records-only
    hint (`report-columns-records-note`) explains this above the list; the aggregate-only "aggregate by"
    select (`report-detail-aggregate-by` — it has no `label`, so per the e2e locator rules it carries a
    `data-testid` rather than being matched on its sibling span's copy) is the one control the mode still
    toggles. One shared column set spans both modes — switching modes never rewrites it, it just
    re-derives which columns are disabled.
  - **Aggregate dimension-column rule**: in aggregate mode the engine can only label an (aggregated) row
    by a field it's grouped/aggregated by, so a dimension column is valid only when its `field` is the
    `detail.by` dimension or one of the `groupBy` levels. Rather than error, such a column is **disabled**
    — a derived state (`isDimensionColumnDisabled` in `report-command.mapper.ts`) shown greyed in the
    columns list and **left out of the request** (`enabledReportColumns`), auto-re-enabling when the
    config makes it valid again. `report-builder` blocks preview/generate only if *no* enabled column
    remains. Nothing is removed or auto-changed — and **Save persists every column, including disabled
    ones** (`toReportRequestCommandForSave`, distinct from the enabled-only preview/generate
    `toReportRequestCommand`), so a disabled column round-trips into a reopened template and self-heals
    instead of being silently dropped. The backend applies the same projection when a stored template is
    generated (see `api/CLAUDE.md` → "Report templates"), so a template holding a currently-disabled
    column still generates with it omitted.
  - **Grouping levels are columns too, and can be renamed.** Every grouping level also renders as a
    *leading* column in the report, headed server-side from the field catalog (`buildDimensions`) — it
    is not in the `columns` FormArray and was previously unnameable. Each Grouping row therefore carries
    an edit pencil (`data-testid="report-grouping-edit"`) beside move-up/down/remove that reuses the
    **same** `ColumnPickerDialogComponent` in a locked-field mode: `ColumnPickerDialogData.lockField`
    hides the Back button (the kind step is unreachable — the field is chosen by the grouping level, not
    the dialog), binds the Field `app-select` `[readonly]`, disables the `field` control, and retitles the
    dialog "Grouping column". The panel hands the level in as the dimension column it *is* — a synthetic
    `ReportColumnValue` with an id — so the picker's existing edit path seeds the label and opens straight
    on the dimension step; only `label` is read back.
    - **The form's `groupBy` is a `FormArray<FormGroup>` of `{ field, label }`**, not bare field keys, so
      a heading cannot outlive its level (removing a level drops its rename; reordering carries it along).
      Read the keys with `readGroupByFields()` (`report-form.factory.ts`), never `readStringArray`.
    - **A blank `label` means "use the field catalog's label"**, and that is how a user resets: entering
      the field's own label stores nothing. The mapper emits `groupByLabels` (a `{ [fieldKey]: label }`
      map on `ReportRequestCommand`) **only** when at least one level is renamed, so an untouched report
      maps to byte-identical the command it always did and the round-trip spec still holds. Keyed by
      field, not index-aligned, because grouping keys are unique — reordering can't desynchronize it.
    - The templates list's Grouping summary (`report-template-summary.ts`) prefers the override, so the
      list agrees with the report it describes. See `api/CLAUDE.md` → "Reporting Engine".
    - **E2E:** `e2e/report-grouping-label.spec.ts` — the rename reaches the real rendered report (asserted
      against the preview iframe's `srcdoc`, which is the engine's own HTML), the dialog is locked/Back-less,
      retyping the default resets it, and a saved template round-trips the heading into both its list row
      and a reopened builder. The first two were verified to **FAIL** with the backend override reverted.
- **Save Template**: the generate bar's secondary button (left of Generate) persists the current
  configuration. Its gate and label follow the builder's mode, driven by two inputs from
  `report-builder` (`isEditMode` + `saveButtonPermission`): on the **new** route it **creates** a
  template (`POST /report/template` via `ReportRunnerService.saveTemplate`, gated by
  `Permission.AppReportsCreate`, label "Save Template", toast "Template saved"); on the **edit** route it
  **updates the opened template in place** (`PUT /report/template/{id}` via
  `ReportRunnerService.updateTemplate`, gated by `Permission.AppReportsUpdate`, label "Update Template",
  toast "Template updated"). So a user who can open a template (read) but not update it sees no Save
  action. **Save-as-new is retired** — the list's Duplicate row action covers copying. The template's
  name is the report's own name (no separate dialog), enabled under the same validity as Generate plus a
  non-empty name (`canSaveTemplate`). See `api/CLAUDE.md` → "Report templates".
- **Splitting** (`report-config-panel`, a section after Totals): two `app-checkbox`es,
  "Split amounts equally across categories" / "…across tags", bound to `splitCategoriesEqually` /
  `splitTagsEqually`, plus a "Don't split these currency fields" `app-autocomlete`
  (`optionValueKey="key"`, over `ReportCatalogService.currencyCustomFields`, which `measures()` now
  reuses) shown only while a box is ticked and a currency field exists. See `api/CLAUDE.md` →
  "Splitting a receipt equally across categories / tags" for what the server does with them.
  - **A ticked box that does nothing says so.** `isSplitActive` (`report-command.mapper.ts`) is the
    client twin of the server's `reportSplitOptions`: a split applies only when the report groups by
    that dimension or aggregates by it. `splitState()` drives a note
    (`report-split-categories-inactive` / `report-split-tags-inactive`) otherwise; the split
    controls' `valueChanges` join the `revision` merge so it re-derives.
  - **The mapper emits each flag only when true and the exclusions only when non-empty**, so an
    untouched report still maps to a command byte-identical to the one it always did. Exclusions are kept
    while both boxes are off, so un-ticking and re-ticking does not lose them, and they are
    **deduplicated** there: with `optionValueKey` set, `app-autocomlete`'s reference-equality filter
    never hides an already-picked option, so the same field can be picked twice.
  - **E2E:** `e2e/report-split-equally.spec.ts` (serial, own seeded group, categories, tags and a
    currency field) — the split figures in the real preview, the inactive note clearing once the
    report groups by tag, the picked exclusion riding the preview request, and a template
    round-tripping the flags. The exclusion input is driven with `fill`, not `click`: the empty
    field's floating label intercepts the click.
- **Generate gating**: the generate bar's Generate button is
  `*hasAppPermission="Permission.AppReportsGenerate"`-gated (preview is not — it stays group-scoped),
  matching the endpoint, which now ANDs `app.reports.generate` with the per-group `group.reports.read`.
- **Templates list** (`report-template-list/`, the `/reports` landing): a paged `app-table`
  (`BaseTableComponent` + `ReportTemplateTableService` + the NGXS `ReportTemplateTableState`, mirroring the
  groups/roles list pages) of saved templates. Columns Name (+ column count), Scope, Grouping, Detail,
  Formats, Updated — the JSON-blob-derived ones are non-sortable; only `name`/`updated_at` sort
  server-side. The derived display strings come from a pure `report-template-summary.ts` util, which takes
  its lookups as arguments (group ids → names via `GroupState.groupsWithoutAll`; custom-field keys → names
  via a `FieldLabelResolver` the page builds from `ReportCatalogService`, whose `load()` this page calls
  too — otherwise a stored `custom_7` grouping renders as the raw key). An unresolvable key still falls
  back to itself, which is what a user without `app.custom-fields.read` gets. The list shows names only;
  the Custom badge is builder-only. Row actions carry `data-testid="report-template-<action>"` and
  gate on the matching permission: **generate** (`AppReportsGenerate`, runs the stored config through the
  builder's generate path), **open/edit** (read — routes to `/reports/:id/edit`), **duplicate**
  (`AppReportsDuplicate`), **delete** (`AppReportsDelete`, via `ConfirmationDialogComponent`). The header
  is the shared `app-table-header` (with a subtitle) and an **"Add Report"** `app-add-button`
  (`data-testid="report-template-add"`) that routes to the blank builder; an empty state (with a second
  `report-template-add-empty` add button) shows when there are none.
- **Open in builder (hydration)**: `/reports/:id/edit` uses a `reportTemplateResolver`
  (`GET /report/template/{id}`) to load the template before the builder's form initializer, and
  `buildReportFormFromCommand` (`report-form.factory.ts`) builds the form *seeded from* the stored
  `ReportRequestCommand` — the faithful inverse of `toReportRequestCommand` (round-trip-tested), reusing
  `buildReceiptFilterForm` for the filter. Building from the command in the field initializer (before the
  constructor's preview subscription attaches) means the loaded config previews exactly once. The builder's
  page-bar gains a back-to-list button + a breadcrumb showing the loaded template name.
- **Other divergences from the design** (intentional): the **progress bar + Cancel** are gone
  (generation is synchronous → in-flight spinner, then download); the section-card look is a small local
  `app-report-section` shell so the pattern isn't repeated.
- Structural lists (scope, grouping, columns) mutate the `FormArray` and bump a `revision` signal so the
  `@for`s re-render under zoneless CD (dialog-driven changes run outside a template event). Multi-select
  filter controls (categories/tags/paid-by) are `FormArray`s, per the `app-autocomlete` `.push()` contract.
- **Full-height two-pane frame**: the screen fills the viewport below the app header with the config and
  preview panes scrolling **independently** and the page-bar/generate-bar pinned flush. This is opt-in via
  `data: { fullHeight: true }` on the route — `SidebarComponent` reads the deepest active route's
  `fullHeight` flag and, **only for that route**, drops the shell's `p-4` padding and turns the content
  area into a bounded flex column (`.drawer-content--full-height`); every other route is unaffected. Reuse
  the same flag for any future full-bleed page. The report name appears as the rendered heading when
  Document → Title is left blank (the engine falls back to the report name).

### Report dashboard widget

A **view-only** dashboard widget (`WidgetType.Report`, `src/dashboard/report-widget/`) that pins a saved
report template and renders its HTML inside a **sandboxed `<iframe srcdoc>`** (`sandbox="allow-same-origin"`,
scripts disabled; `bypassSecurityTrustHtml`) — the same technique as the builder preview, but the widget
copies the ~10-line idiom locally rather than reusing the builder-coupled `ReportPreviewPanelComponent`.
The widget stores only `{ reportTemplateId }` in its configuration blob and calls
`ReportRunnerService.renderTemplate(id)` → `POST /report/template/{id}/render` (see `api/CLAUDE.md`), which
renders the **full dataset** (not the capped builder preview) and re-resolves access server-side.

- **Restricted state is backend-driven**: a revoked/deleted template comes back as restricted-notice HTML
  at 200, which the widget renders like any other HTML — there is no special client "restricted" branch
  (only a generic error state for a network failure or a missing `reportTemplateId`). The report renders
  in a height-capped, internally-scrolling stage so a long report doesn't blow out the tile.
- **Download button** (`data-testid="report-widget-download"`): shown **only when** the server-returned
  `allowedActions` include `"generate"` — gated purely on the server result, **never** re-AND-ed with a
  client `*hasAppPermission` (same rule as the templates-list row buttons). It calls
  `ReportRunnerService.downloadTemplateById(id)` (resolve template → `generateFromTemplateById`, the
  enforcing `/report/template/{id}/generate` path).
- **Authoring** (`dashboard-form`): the **Report** widget-type option is filtered out unless the user holds
  `app.reports.read`/`readAll` (`availableWidgetTypeOptions`), and its config is a single template picker
  (`app-select`) whose options come from `ReportService.getReportTemplates` — server-filtered to the
  caller's viewable templates (`reportTemplateOptions`, loaded in `ngOnInit`, `catchError`→empty).

## Testing Requirements

**All new code must have accompanying unit tests.**

Before considering any work complete:

1. Write unit tests for all new components, services, and pipes
2. Use Angular TestBed for component testing
3. Mock services and HTTP calls appropriately
4. Run the full test suite: `npm test`
5. Ensure all tests pass before submitting changes

Tests should cover:

- Component rendering and user interactions
- Component method inputs and outputs
- Service method behavior
- Form validation logic
- Error handling scenarios