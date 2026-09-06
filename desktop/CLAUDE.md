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
  `http-proxy-middleware`, `undici`, `uuid`). When `npm audit` flags a new transitive advisory that
  the toolchain hasn't bumped yet, add/raise the floor here rather than waiting on an upstream release.
- **Exact pins (no caret):** `ngx-bootstrap` (`21.0.1`) and `@playwright/test` (`1.59.1`) are pinned
  because their next minor introduced an incompatibility (ngx-bootstrap dropped `CarouselModule.forRoot()`
  used by `src/carousel/`; Playwright `1.61` needs a newer browser build than the pinned `e2e:install`
  cache). Bump these deliberately — update the consuming code / reinstall browsers in the same change.
- Stay on the Angular `21.2.x` patch line for security fixes; a jump to Angular 22 is a separate,
  breaking upgrade and out of scope for audit hygiene.

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
  It returns early in view mode, without `canManageCustomFields()`, on a falsy group id, and — via
  `groupChangeIsInitialEmission` — on the listener's `startWith()` replay unless the mode is `add`.
  That guard is the **only** thing keeping an edit-mode receipt untouched on load, since `startWith`
  fires at init in every mode. Both it and `autoAppliedCustomFieldIds` are reset at the **top** of
  `initForm()` (which re-runs on every route-data emission) so a stale auto set can never strip a
  saved receipt's own fields.
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

## Session lifetime settings (Hours/Days selector over an hours-based API)

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

**The max tracks the selected unit** (720 hours === 30 days). `listenForDurationUnitChanges(value,
unit)` — one parameterised method called twice — re-applies
`[Validators.required, durationValueValidator(maxForUnit(...))]` whenever the unit flips, the same
shape as `urlValidators()` above (`setValidators` replaces the whole list, so `required` must be
re-supplied each time rather than declared in `initForm`). All four controls are also added to the
**view-mode disable block**, or they stay editable on the read-only page.

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
  the group default. Category/tag pickers (`app-category-autocomplete`/`app-tag-autocomplete`,
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
  hands them to `ReceiptCommentsComponent.addMagicFilledComments()` — add mode collects them for the create
  submit; edit mode POSTs each via `CommentService` because the receipt-**update** path does not persist
  comments (they're individual resources — see `api/CLAUDE.md` → `UpdateReceipt`).
- **Custom fields** reference a field by id only (the magic-fill response carries no field definition), so a
  value whose `customFieldId` isn't in the loaded catalog pool is skipped, and adding one flips its
  manage-fields menu entry to selected via an immutable array replace (zoneless CD).

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
  Both are opt-in and default off, so every other `app-select` call site is unchanged. The already-picked
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
- **Live preview** (`report-preview-panel`): the container debounces the form (~450ms, `switchMap`) into
  `POST /report/preview` and renders the engine's returned HTML in a **sandboxed `<iframe srcdoc>`**
  (`sandbox="allow-same-origin"`, scripts disabled; sized to content on load). The response's
  `receiptCount` drives the chip that opens the receipts drill-in (`report-receipts-dialog`, paged
  receipts across scope with the filter + resolved period). The drill-in is a read-only list → detail
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