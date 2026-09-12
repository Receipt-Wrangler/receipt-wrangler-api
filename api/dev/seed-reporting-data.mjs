#!/usr/bin/env node
/*
 * seed-reporting-data.mjs — seed realistic, high-volume reporting test data.
 *
 * Creates (idempotently) a dedicated reporting group owned by the admin, a few
 * member users, a realistic set of app-global categories/tags, and a batch of
 * receipts spread across a date range — so the Report Builder can be exercised at
 * volume with lifelike data.
 *
 * Auth is an admin API key sent RAW in the Authorization header (no "Bearer").
 * The key is read from the RW_API_KEY env var and is never written to disk.
 *
 * Usage:
 *   RW_API_KEY='key.1.<id>.<secret>' node api/dev/seed-reporting-data.mjs
 *
 * Env vars (all optional except RW_API_KEY):
 *   RW_API_KEY          (required) admin API key
 *   API_BASE_URL        default http://localhost:8081/api
 *   RECEIPT_COUNT       default 2000   (receipts created THIS run; additive)
 *   NEW_USER_COUNT      default 3      member users to ensure
 *   SEED_GROUP_NAME     default "Reporting Load Test"
 *   START_DATE          default 2026-01-01   (receipt date floor, inclusive)
 *   END_DATE            default 2026-07-14   (receipt date ceiling, inclusive)
 *   RANDOM_SEED         default 42     (reproducible generation)
 *   CONCURRENCY         default 4      (keep modest on SQLite; higher on pg/maria)
 *   SEED_USER_PASSWORD  default "ReportingSeed!23" (dev-only member password)
 *
 * Requires Node 18+ (uses the built-in global fetch). No npm install.
 */

// ---------- config ----------
function envInt(value, fallback) {
  const n = Number.parseInt(value ?? "", 10);
  return Number.isFinite(n) ? n : fallback;
}

const API_BASE_URL = (process.env.API_BASE_URL || "http://localhost:8081/api").replace(/\/+$/, "");
const API_KEY = process.env.RW_API_KEY;
const RECEIPT_COUNT = envInt(process.env.RECEIPT_COUNT, 2000);
const NEW_USER_COUNT = envInt(process.env.NEW_USER_COUNT, 3);
const SEED_GROUP_NAME = process.env.SEED_GROUP_NAME || "Reporting Load Test";
const START_DATE = new Date(`${process.env.START_DATE || "2026-01-01"}T00:00:00Z`);
const END_DATE = new Date(`${process.env.END_DATE || "2026-07-14"}T23:59:59Z`);
const RANDOM_SEED = envInt(process.env.RANDOM_SEED, 42);
const CONCURRENCY = Math.max(1, envInt(process.env.CONCURRENCY, 4));
const SEED_USER_PASSWORD = process.env.SEED_USER_PASSWORD || "ReportingSeed!23";

if (!API_KEY) {
  console.error("RW_API_KEY is required (the admin API key, sent raw in the Authorization header).");
  console.error("Example: RW_API_KEY='key.1.<id>.<secret>' node api/dev/seed-reporting-data.mjs");
  process.exit(1);
}
if (!(START_DATE < END_DATE)) {
  console.error("START_DATE must be before END_DATE.");
  process.exit(1);
}

// ---------- seeded PRNG (mulberry32; Math.random isn't seedable) ----------
let _rngState = RANDOM_SEED >>> 0;
function rand() {
  _rngState = (_rngState + 0x6d2b79f5) | 0;
  let t = _rngState;
  t = Math.imul(t ^ (t >>> 15), t | 1);
  t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
}
function randInt(min, max) {
  return min + Math.floor(rand() * (max - min + 1));
}
function pick(arr) {
  return arr[Math.floor(rand() * arr.length)];
}
function weightedIndex(weights) {
  let r = rand();
  for (let i = 0; i < weights.length; i++) {
    if ((r -= weights[i]) < 0) return i;
  }
  return weights.length - 1;
}
function sampleTags(tags, count) {
  const copy = [...tags];
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(rand() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy.slice(0, count);
}
function randomDateIso() {
  const t = START_DATE.getTime() + Math.floor(rand() * (END_DATE.getTime() - START_DATE.getTime()));
  return new Date(t).toISOString().replace(/\.\d{3}Z$/, "Z"); // RFC3339, second precision
}
function money(min, max) {
  const skewed = Math.pow(rand(), 1.6); // favour the lower end for realism
  return (min + skewed * (max - min)).toFixed(2);
}

// ---------- HTTP ----------
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function api(method, path, body) {
  const url = API_BASE_URL + path;
  let lastError;
  for (let attempt = 0; attempt < 5; attempt++) {
    try {
      const res = await fetch(url, {
        method,
        headers: { Authorization: API_KEY, "Content-Type": "application/json" },
        body: body === undefined ? undefined : JSON.stringify(body),
      });
      const text = await res.text();
      if (res.status >= 500) {
        // Transient (e.g. SQLite write lock) — back off and retry.
        lastError = new Error(`${method} ${path} -> ${res.status}: ${text.slice(0, 160)}`);
        await sleep(200 * 2 ** attempt);
        continue;
      }
      if (res.status >= 400) {
        throw new Error(`${method} ${path} -> ${res.status}: ${text.slice(0, 200)}`);
      }
      return text ? JSON.parse(text) : null;
    } catch (err) {
      lastError = err;
      if (attempt < 4) await sleep(200 * 2 ** attempt);
    }
  }
  throw lastError;
}

async function runPool(items, worker, concurrency) {
  let index = 0;
  const runners = Array.from({ length: Math.min(concurrency, items.length) }, async () => {
    while (index < items.length) {
      const item = items[index++];
      await worker(item);
    }
  });
  await Promise.all(runners);
}

// ---------- realistic catalog ----------
const CATEGORY_CATALOG = [
  { name: "Groceries", description: "Grocery and household purchases", min: 12, max: 320,
    merchants: ["Whole Foods Market", "Trader Joe's", "Costco Wholesale", "Safeway", "Kroger", "Aldi"] },
  { name: "Restaurants", description: "Dining out and takeout", min: 8, max: 180,
    merchants: ["Chipotle", "Starbucks", "Olive Garden", "Panera Bread", "Shake Shack", "The Cheesecake Factory"] },
  { name: "Travel", description: "Flights, hotels, and lodging", min: 85, max: 1650,
    merchants: ["Delta Air Lines", "United Airlines", "Marriott", "Hilton Hotels", "Airbnb", "Enterprise Rent-A-Car"] },
  { name: "Transportation", description: "Rideshare, fuel, and transit", min: 6, max: 110,
    merchants: ["Uber", "Lyft", "Shell", "Chevron", "BP", "Metro Transit"] },
  { name: "Office Supplies", description: "Office and workspace supplies", min: 5, max: 420,
    merchants: ["Staples", "Office Depot", "Amazon Business", "Uline"] },
  { name: "Software & Subscriptions", description: "SaaS and software subscriptions", min: 9, max: 680,
    merchants: ["GitHub", "Amazon Web Services", "Figma", "Notion", "Slack", "Adobe", "Google Workspace"] },
  { name: "Utilities", description: "Power, internet, phone, and water", min: 40, max: 460,
    merchants: ["PG&E", "Comcast Business", "AT&T", "Verizon", "City Water Dept"] },
  { name: "Hardware", description: "Computers, devices, and equipment", min: 45, max: 2800,
    merchants: ["Best Buy", "Apple Store", "Dell", "Newegg", "B&H Photo"] },
  { name: "Meals & Entertainment", description: "Team events and entertainment", min: 15, max: 520,
    merchants: ["AMC Theatres", "Topgolf", "Blue Plate Catering", "Dave & Buster's"] },
  { name: "Healthcare", description: "Pharmacy and medical", min: 10, max: 540,
    merchants: ["CVS Pharmacy", "Walgreens", "Quest Diagnostics", "City Medical Clinic"] },
  { name: "Professional Services", description: "Legal, accounting, and consulting", min: 150, max: 3200,
    merchants: ["Baker & Associates LLP", "Ledger Accounting", "Northwind Consulting"] },
  { name: "Marketing", description: "Advertising and marketing", min: 40, max: 2400,
    merchants: ["Google Ads", "Meta Ads", "Mailchimp", "VistaPrint"] },
];

const TAG_NAMES = [
  "Reimbursable", "Business", "Personal", "Tax Deductible", "Recurring",
  "Client Billable", "Travel", "Home Office", "Urgent", "Reviewed",
];

const MEMBER_NAMES = ["Jordan Rivera", "Casey Morgan", "Taylor Brooks", "Riley Chen", "Avery Patel", "Sam Okafor"];

const STATUSES = ["RESOLVED", "OPEN", "NEEDS_ATTENTION", "DRAFT", "DECLINED"];
const STATUS_WEIGHTS = [0.6, 0.2, 0.1, 0.05, 0.05];
const TAG_COUNT_WEIGHTS = [0.2, 0.4, 0.3, 0.1]; // 0, 1, 2, or 3 tags

// ---------- seed ----------
async function main() {
  console.log(`Seeding reporting data against ${API_BASE_URL}`);

  const appData = await api("GET", "/user/appData");
  const adminId = appData.claims.userId;
  console.log(`Authenticated as ${appData.claims.username} (id ${adminId})`);

  // Roles: a normal APP role for members, a GROUP role granting them access.
  const roles = await api("GET", "/role");
  const appRole =
    roles.find((r) => r.scope === "APP" && r.name === "Legacy User") ||
    roles.find((r) => r.scope === "APP");
  const groupRole =
    roles.find((r) => r.scope === "GROUP" && r.name === "Legacy Editor") ||
    roles.find((r) => r.scope === "GROUP");
  if (!appRole || !groupRole) throw new Error("Could not resolve an APP and a GROUP role from /role.");

  // Users (idempotent by username).
  const usersByName = new Map((appData.users || []).map((u) => [u.username, u.id]));
  const memberIds = [];
  for (let i = 1; i <= NEW_USER_COUNT; i++) {
    const username = `reporting-user-${i}`;
    let id = usersByName.get(username);
    if (id) {
      console.log(`  user ${username} exists (id ${id})`);
    } else {
      const created = await api("POST", "/user", {
        username,
        password: SEED_USER_PASSWORD,
        displayName: MEMBER_NAMES[(i - 1) % MEMBER_NAMES.length],
        appRoleId: appRole.id,
      });
      id = created.id;
      console.log(`  created user ${username} (id ${id})`);
    }
    memberIds.push(id);
  }

  // Group (idempotent by name). Exclude the admin — they are auto-added as owner.
  let group = (appData.groups || []).find((g) => g.name === SEED_GROUP_NAME);
  if (group) {
    console.log(`  group "${SEED_GROUP_NAME}" exists (id ${group.id})`);
  } else {
    group = await api("POST", "/group", {
      name: SEED_GROUP_NAME,
      status: "ACTIVE",
      groupMembers: memberIds.map((userId) => ({ userId, groupRoleId: groupRole.id })),
    });
    console.log(`  created group "${SEED_GROUP_NAME}" (id ${group.id})`);
  }
  const groupId = group.id;
  const paidByPool = [...new Set([adminId, ...memberIds, ...(group.groupMembers || []).map((m) => m.userId)])];

  // Categories (app-global, idempotent by name).
  const catByName = new Map((appData.categories || []).map((c) => [c.name, c.id]));
  const categories = [];
  for (const cat of CATEGORY_CATALOG) {
    let id = catByName.get(cat.name);
    if (!id) {
      const created = await api("POST", "/category", { name: cat.name, description: cat.description });
      id = created.id;
      console.log(`  created category ${cat.name} (id ${id})`);
    }
    categories.push({ ...cat, id });
  }

  // Tags (app-global, idempotent by name).
  const tagByName = new Map((appData.tags || []).map((t) => [t.name, t.id]));
  const tags = [];
  for (const name of TAG_NAMES) {
    let id = tagByName.get(name);
    if (!id) {
      const created = await api("POST", "/tag", { name, description: "" });
      id = created.id;
      console.log(`  created tag ${name} (id ${id})`);
    }
    tags.push({ id, name });
  }

  // Generate all receipt bodies up front (deterministic under the seed), then
  // POST them with bounded concurrency.
  const bodies = Array.from({ length: RECEIPT_COUNT }, () => {
    const cat = pick(categories);
    const merchant = pick(cat.merchants);
    const tagCount = weightedIndex(TAG_COUNT_WEIGHTS);
    return {
      name: rand() < 0.4 ? `${merchant} #${randInt(1000, 9999)}` : merchant,
      amount: money(cat.min, cat.max),
      date: randomDateIso(),
      groupId,
      paidByUserId: pick(paidByPool),
      status: STATUSES[weightedIndex(STATUS_WEIGHTS)],
      categories: [{ id: cat.id, name: cat.name }],
      tags: sampleTags(tags, tagCount).map((t) => ({ id: t.id, name: t.name })),
    };
  });

  console.log(`Creating ${RECEIPT_COUNT} receipts in group ${groupId} (concurrency ${CONCURRENCY})...`);
  let ok = 0;
  let failed = 0;
  await runPool(
    bodies,
    async (body) => {
      try {
        await api("POST", "/receipt", body);
        ok++;
      } catch (err) {
        failed++;
        if (failed <= 5) console.error(`  receipt failed: ${err.message}`);
      }
      const total = ok + failed;
      if (total % 100 === 0) console.log(`  ${total}/${RECEIPT_COUNT} (ok ${ok}, failed ${failed})`);
    },
    CONCURRENCY
  );

  console.log("\nDone.");
  console.log(`  group:      ${SEED_GROUP_NAME} (id ${groupId})`);
  console.log(`  members:    admin ${adminId} + users [${memberIds.join(", ")}]`);
  console.log(`  paid-by:    [${paidByPool.join(", ")}]`);
  console.log(`  categories: ${categories.length}, tags: ${tags.length}`);
  console.log(`  receipts:   ok ${ok}, failed ${failed}`);
}

main().catch((err) => {
  console.error("FATAL:", err.message);
  process.exit(1);
});
