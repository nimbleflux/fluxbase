# Changelog

All notable changes to Fluxbase are documented in this file. The server is
versioned with CalVer (YYYY.M.N); the JS/TS SDKs use SemVer.

This file is generated from the conventional-commit history by
[git-cliff](https://github.com/orhun/git-cliff) at release time — do not
hand-edit release sections. See RELEASING.md for the release process.

## v2026.11.1 (2026-11-20)


### Bug fixes


- **postgres:** Upgrade base packages on every build; derive CI defaults from versions.json (#444) @bartcode

- **security:** Clear remaining Security Scan findings, pin nginx 1.31.9 (CVE-2026-123007), fix Fuzz setup (#459) @bartcode

- **runtime:** Exempt loopback family from deny list for loopback self URLs (#462) @bartcode


### Dependencies


- **deps:** Update dependency vite to ^8.3.3 [security] (#446) @renovate[bot]

- **deps:** Update dependency zod to ^4.6.6 [security] (#448) @renovate[bot]

- **deps:** Update dependency astro to ^7.3.8 [security] (#447) @renovate[bot]

- **deps:** Pin golang docker tag to d9c68c2 (#449) @renovate[bot]

- **deps:** Update module golang.org/x/net to v0.62.0 [security] (#451) @renovate[bot]

- **deps:** Update module github.com/go-jose/go-jose/v4 to v4.1.6 [security] (#450) @renovate[bot]

- **deps:** Update module github.com/andybalholm/brotli to v1.2.7 [security] (#452) @renovate[bot]

- **deps:** Update module github.com/klauspost/compress to v1.20.2 [security] (#453) @renovate[bot]

- **deps:** Update module golang.org/x/crypto to v0.59.0 [security] (#455) @renovate[bot]

- **deps:** Update module go.opentelemetry.io/otel/sdk to v1.47.0 [security] (#457) @renovate[bot]

- **deps:** Update dependency astro to ^7.4.7 [security] (#456) @renovate[bot]

- **deps:** Update dependency seroval to >=1.7.7 [security] (#460) @renovate[bot]

- **deps:** Update dependency next to v16 [security] (#461) @renovate[bot]

- **deps:** Update all dependencies (#454) @renovate[bot]

- **deps:** Bump tsup to v9 and node to 26.13.0 (#465) @bartcode


### Features


- **admin:** Migrate @tanstack/react-table to v9 (#464) @bartcode

- **kotlin:** Migrate to Dokka Gradle Plugin v2, bump maven-publish to 0.37.0 (#466) @bartcode


### Miscellaneous


- Update version to 2026.10.2 in deployment files @github-actions[bot]

- **docker:** Pgschema bump script; stop renovate auto-bumps (#463) @bartcode


### Refactoring


- **api:** Extract vector domain into internal/api/vector (#445) @bartcode


## v2026.10.2 (2026-10-13)


### Bug fixes


- **sdk-kotlin:** Encode non-primitive JSON values via generated serializers — fixes #440 (#442) @bartcode

- **sdk-kotlin:** Wire per-request RpcInvokeOptions.timeout — was silently ignored (#443) @bartcode


### Features


- **sdk:** AbortSignal support on rpc.invoke and functions.invoke — fixes #439 (#441) @bartcode


### Miscellaneous


- Update version to 2026.10.1 in deployment files @github-actions[bot]

- **postgres:** Rebuild image for latest patches (18.4-v2) @bartcode

- **postgres:** Pin Helm values to fluxbase-postgres 18.4-v2 @bartcode


## v2026.10.1 (2026-10-13)


### Bug fixes


- **go:** Tenant authz guard, error sanitization, panic-resilient loops, RPC lexical params (#396) @bartcode

- **sdk,admin:** Align TS surface with server contract; harden transport, realtime, auth; cookie sessions (#397) @bartcode

- **sdk:** Admin.jobs.listJobs reads the {jobs,limit,offset} envelope (#410) @bartcode

- **go:** Restrict users_admin mutation routes to instance-level admins (#409) @bartcode

- **helm:** Stabilize externalDatabase.password across upgrades (#413) @bartcode

- **sdk:** Type result-envelope errors as FluxbaseError; emit SUBSCRIBED on server ack — fixes #405 (#417) @bartcode

- **admin:** Enable noUncheckedIndexedAccess and fix all surfaced index-access errors — fixes #388 (#421) @bartcode

- **dx:** Day-one experience — actionable boot-guard errors, working make help, honest make targets — fixes #407 (#416) @bartcode

- **go:** Scope execution-log reads to the owning tenant (#419) @bartcode

- **go:** Bound filter parsing; consolidate identifier validation — fixes #394 (#418) @bartcode

- Three follow-ups for the merged remediation wave (#424) @bartcode

- **ci:** Scope gitleaks allowlist to the contract-drift fixture values (#434) @bartcode

- **release:** Grant attestations:write for binary provenance @bartcode

- **test:** Don't keep idle spare connections in the auth test pool @bartcode

- **release:** Detect half-published Kotlin versions instead of 409ing (#437) @bartcode

- **release:** Complete binary provenance rollout + Kotlin 302 handling @bartcode


### CI


- Cut pipeline wall clock ~40%, cancel superseded runs, tune stage selection (#412) @bartcode

- **release:** Overhaul release pipeline — dry-run, idempotent publishes, generated notes, buildinfo (#436) @bartcode

- **release:** Publish npm SDKs via OIDC trusted publishing, drop NPM_TOKEN (#438) @bartcode


### Dependencies


- **deps:** Pin patched vue, sharp, http-cache-semantics, source-map-js @bartcode

- **deps:** Update dependency smol-toml to ^1.9.0 [security] (#367) @renovate[bot]


### Documentation


- Fix user-breaking documentation drift (env vars, docker-secrets patterns, helm keys, SDK signature) — fixes #406 (#411) @bartcode

- Fix misleading defaults and stale examples (post-alignment audit) — fixes #406 (#415) @bartcode


### Features


- **sdk:** AbortSignal support through the fetch layer; admin queryFn cancellation — fixes #386 (#420) @bartcode

- **ops:** GOMEMLIMIT support, make bench, stdlib load-test tool, memory baseline (#433) @bartcode

- **ai:** RAG quality pack, semantic response cache, MCP notifications, SDK parity + docs overhaul (#435) @bartcode


### Miscellaneous


- Update version to 2026.9.10 in deployment files @github-actions[bot]

- **ci/ops:** Supply-chain hardening, Helm secret stability, backup encryption, docs (#400) @bartcode

- **tooling:** Real ESLint across packages, TS config hardening, coverage gates (#399) @bartcode

- **security:** Make servicekey test fixtures gitleaks-inert (#423) @bartcode

- **lint:** Golangci-lint v2.14 + re-enable staticcheck and full errorlint (#425) @bartcode

- **ci:** Retry transient lychee 503s; ignore contract-drift fixture fingerprints (#432) @bartcode

- Update version to 2026.10.1 in deployment files @github-actions[bot]

- Update version to 2026.10.1 in deployment files @github-actions[bot]

- Update version to 2026.10.1 in deployment files @github-actions[bot]


### Refactoring


- **middleware:** Split clientkey_auth into focused files (#414) @bartcode


### Tests


- **auth:** Repair + add DB-coupled MFA/OAuth integration tests, wire into CI — fixes #313 (#422) @bartcode

- Admin unit tests for the auth/client layer; contract-drift CI job (#426) @bartcode

- Close -race coverage gaps in CI and add goroutine-leak detection (#430) @bartcode

- Add Go fuzz targets for the input-parsing surfaces (#431) @bartcode


## v2026.9.10 (2026-09-28)


### Bug fixes


- **rpc:** Honor @fluxbase:name annotation in sync (#365) @bartcode

- **sdk:** GetUser() falls back to the execution token when no session exists (#366) @bartcode


### Miscellaneous


- Update version to 2026.9.9 in deployment files @github-actions[bot]


## v2026.9.9 (2026-09-27)


### Features


- **storage:** Default_public_buckets — seed + self-heal public-read buckets (#364) @bartcode


### Miscellaneous


- Update version to 2026.9.8 in deployment files @github-actions[bot]


## v2026.9.8 (2026-09-26)


### Bug fixes


- **auth:** Accept legacy stateless keys without a token_type claim (#362) @bartcode


### Miscellaneous


- Update version to 2026.9.7 in deployment files @github-actions[bot]


## v2026.9.7 (2026-09-25)


### Miscellaneous


- Update version to 2026.9.6 in deployment files @github-actions[bot]


## v2026.9.6 (2026-09-25)


### Bug fixes


- **boot:** Stop declarative apply from dropping bootstrap-owned objects on fresh databases (#361) @bartcode


### Miscellaneous


- Update version to 2026.9.5 in deployment files @github-actions[bot]


## v2026.9.4 (2026-09-25)


### Features


- SDK account deletion + chatbot-KB link access_level updates (#360) @bartcode


### Miscellaneous


- Update version to 2026.9.3 in deployment files @github-actions[bot]

- Update version to 2026.9.4 in deployment files @github-actions[bot]


## v2026.9.3 (2026-09-25)


### Dependencies


- **deps:** Update dependency vitest to ^4.1.11 [security] (#352) @renovate[bot]

- **deps:** Update dependency js-yaml to v5 [security] (#350) @renovate[bot]

- **deps:** Update dependency astro to ^7.2.7 [security] (#354) @renovate[bot]

- **deps:** Update module google.golang.org/grpc to v1.83.2 [security] (#353) @renovate[bot]

- **deps:** Update go dependencies (#355) @renovate[bot]

- **deps:** Update dependency smol-toml to ^1.8.0 [security] (#357) @renovate[bot]

- **deps:** Update dependency astro to ^7.2.9 [security] (#356) @renovate[bot]

- **deps:** Update dependency sharp to ^0.35.4 [security] (#351) @renovate[bot]

- **deps:** Update all dependencies (#358) @renovate[bot]


### Miscellaneous


- Update version to 2026.9.2 in deployment files @github-actions[bot]

- **docs:** Move Umami tracker to umami.nimbleflux.nl @bartcode


## v2026.9.2 (2026-09-08)


### Bug fixes


- **sdk-kotlin:** Parse Content-Range count case-insensitively (#349) @bartcode


### Miscellaneous


- Update version to 2026.9.1 in deployment files @github-actions[bot]


## v2026.9.1 (2026-09-05)


### Bug fixes


- **auth:** CAS session token rotation + config-aligned refresh cookie (#338) @bartcode

- **ci:** Resolve a live Groq model at release time (#342) @bartcode

- **rpc:** Serialize native uuid columns as canonical strings (#346) @bartcode

- **sdk-kotlin:** Pass JsonElement RPC bodies through; floor autoRefresh retry at 5s (#347) @bartcode

- Android client compat — RPC body serialization, uuid results, refresh tolerance, auth middleware (#348) @bartcode


### Dependencies


- **deps:** Update module golang.org/x/crypto to v0.55.0 [security] (#343) @renovate[bot]

- **deps:** Update module google.golang.org/grpc to v1.83.1 [security] (#344) @renovate[bot]

- **deps:** Update module golang.org/x/crypto to v0.56.0 [security] (#345) @renovate[bot]


### Documentation


- **sdk-kotlin:** Rewrite Dokka .md links to Starlight routes; fix deep links (#340) @bartcode


### Features


- **auth:** Default access-token lifetime 15m → 1h (#339) @bartcode


### Miscellaneous


- Update version to 2026.8.14 in deployment files @github-actions[bot]

- **auth:** Align refresh-expiry examples with the 90-day sliding default (#341) @bartcode


## v2026.8.14 (2026-08-25)


### Bug fixes


- **sdk-kotlin:** Single-flight refreshSession — concurrent refreshes bricked sessions (#337) @bartcode


### Miscellaneous


- Update version to 2026.8.13 in deployment files @github-actions[bot]


## v2026.8.13 (2026-08-25)


### Features


- **auth:** Default refresh token lifetime 7d -> 90d (sliding) (#336) @bartcode

- **sdk-kotlin:** Real autoRefresh — proactive token refresh (#335) @bartcode


### Miscellaneous


- Update version to 2026.8.12 in deployment files @github-actions[bot]


## v2026.8.12 (2026-08-23)


### Bug fixes


- **cli:** Encode kb query params correctly; add kb export-table (#334) @bartcode


### Miscellaneous


- Update version to 2026.8.11 in deployment files @github-actions[bot]


## v2026.8.11 (2026-08-22)


### Bug fixes


- **helm:** Set DENO_DIR on the server so deno's cache stays on the deno-cache volume (#333) @bartcode


### Miscellaneous


- Update version to 2026.8.10 in deployment files @github-actions[bot]


## v2026.8.10 (2026-08-16)


### Features


- **auth:** Allow private-use app schemes in OAuth redirect URLs (#332) @bartcode


### Miscellaneous


- Update version to 2026.8.9 in deployment files @github-actions[bot]


## v2026.8.9 (2026-08-16)


### Features


- **auth:** Support multiple redirect URLs for OAuth providers (#331) @bartcode


### Miscellaneous


- Update version to 2026.8.8 in deployment files @github-actions[bot]


## v2026.8.8 (2026-08-16)


### Miscellaneous


- Update version to 2026.8.8-rc.7 in deployment files @github-actions[bot]


## v2026.8.8-rc.7 (2026-08-16)


### Features


- **sdk-kotlin:** OAuth methods — providers, authorize URL, code exchange (#330) @bartcode


### Miscellaneous


- Update version to 2026.8.8-rc.6 in deployment files @github-actions[bot]


## v2026.8.8-rc.6 (2026-08-16)


### Bug fixes


- **api:** Fall back to FLUXBASE_ANON_KEY env for /auth/config anon_key (#329) @bartcode


### Miscellaneous


- Update version to 2026.8.8-rc.4 in deployment files @github-actions[bot]

- Update version to 2026.8.8-rc.5 in deployment files @github-actions[bot]


## v2026.8.8-rc.4 (2026-08-16)


### Bug fixes


- **ci:** Arch must be a matrix key in publish-docker-arch (#328) @bartcode


### Miscellaneous


- Update version to 2026.8.8-rc.3 in deployment files @github-actions[bot]


## v2026.8.8-rc.3 (2026-08-16)


### Bug fixes


- **build:** Build SDKs before admin UI in make build (#323) (#324) @bartcode

- **build:** Make `make build` work on macOS + building-from-source docs (#325) @bartcode


### Features


- **docker:** Add lite image variant, fix broken vips build tag, add smoke pipeline (#320) @bartcode

- Fluxbase-kotlin SDK — full TS-parity Kotlin port (#322) @bartcode

- Anon-key discovery via /auth/config + opt-in TLS bypass (Kotlin SDK) (#327) @bartcode


### Miscellaneous


- Update version to 2026.8.7 in deployment files @github-actions[bot]


### Tests


- Coverage wave 1 (pure helpers) + integration test infra fixes (#321) @bartcode


## v2026.8.7 (2026-08-09)


### Bug fixes


- **jobs:** Attach tenant_id to execution logs so getLogs returns them (#319) @bartcode


### Miscellaneous


- Update version to 2026.8.6 in deployment files @github-actions[bot]


## v2026.8.6 (2026-08-08)


### Bug fixes


- **routes:** Generic settings :key shadowed secret/user index routes (#314) @bartcode

- **reliability:** Auto-restart crashed workers and retry DB startup on transient errors (#315) @bartcode


### Miscellaneous


- Update version to 2026.8.5 in deployment files @github-actions[bot]


### Tests


- Pure-security unit tests + hardened coverage gate (#312) @bartcode

- **branching:** Cover pure helpers (14.9% -> 19.6% unit-only) (#316) @bartcode

- **ai:** Cover pure helpers (unit-only 28.0% → 31.1%) (#317) @bartcode


## v2026.8.5 (2026-08-08)


### Bug fixes


- **schema:** Prepend DROP FUNCTION for CREATE [OR REPLACE] FUNCTION (#311) @bartcode


### Miscellaneous


- Update version to 2026.8.4 in deployment files @github-actions[bot]


## v2026.8.4 (2026-08-08)


### Dependencies


- **deps:** Update dependency js-yaml to ^4.3.1 [security] (#309) @renovate[bot]


### Documentation


- Fix broken cross-references and external links @bartcode


### Features


- **sdk:** Allow rpc() callable to accept namespace option (#310) @bartcode


### Miscellaneous


- Update version to 2026.8.3 in deployment files @github-actions[bot]


## v2026.8.3 (2026-08-05)


### Dependencies


- **deps:** Update dependency postcss to ^8.5.23 [security] (#307) @renovate[bot]


### Features


- **settings:** Namespace bulk-fetch, is_public flag, chatbot limit config (#308) @bartcode


### Miscellaneous


- Update version to 2026.8.2 in deployment files @github-actions[bot]


## v2026.8.2 (2026-08-01)


### Bug fixes


- **storage:** Chunked/multipart/S3 uploads don't create RLS-owned storage.objects rows (#306) @bartcode


### Miscellaneous


- Update version to 2026.8.1 in deployment files @github-actions[bot]


## v2026.8.1 (2026-08-01)


### Bug fixes


- **sdk:** Clear stale session when 401 token refresh fails (#305) @bartcode

- **auth:** Degrade to anonymous on expired tokens for optional-auth routes (#304) @bartcode


### Miscellaneous


- Update version to 2026.7.8 in deployment files @github-actions[bot]


## v2026.7.8 (2026-07-30)


### Bug fixes


- **schema:** Prepend DROP VIEW for CREATE OR REPLACE VIEW in MakeSQLIdempotent (#303) @bartcode


### Miscellaneous


- Update version to 2026.7.7 in deployment files @github-actions[bot]


## v2026.7.7 (2026-07-30)


### Bug fixes


- **schema:** Run ensureMissingColumns BEFORE bulk SQL exec in fallback (#302) @bartcode


### Miscellaneous


- Update version to 2026.7.6 in deployment files @github-actions[bot]


## v2026.7.6 (2026-07-30)


### Bug fixes


- **extensions:** Catalog sync ordering + empty userID normalization (#301) @bartcode


### Miscellaneous


- Update version to 2026.7.5 in deployment files @github-actions[bot]


## v2026.7.5 (2026-07-30)


### Features


- **extensions:** Enable out-of-the-box via catalog sync + pg_available fallback (#300) @bartcode


### Miscellaneous


- Update version to 2026.7.4 in deployment files @github-actions[bot]


## v2026.7.4 (2026-07-30)


### Miscellaneous


- Update version to 2026.7.4-rc.6 in deployment files @github-actions[bot]


## v2026.7.4-rc.6 (2026-07-30)


### Bug fixes


- **release:** Anchor changelog on last published release, not last tag (#298) @bartcode

- **schema:** Make direct-apply fallback the PRIMARY apply path for app schemas (#299) @bartcode


### Miscellaneous


- Update version to 2026.7.4-rc.5 in deployment files @github-actions[bot]


## v2026.7.4-rc.5 (2026-07-30)


### Features


- **schema:** .pgschemaignore support for declarative app-schema + bump pgschema to 1.12.1 (#297) @bartcode


### Miscellaneous


- Update version to 2026.7.4-rc.4 in deployment files @github-actions[bot]


## v2026.7.4-rc.4 (2026-07-30)


### Bug fixes


- **schema:** Name-based column extraction for app-schema fallback (#295) @bartcode


### Miscellaneous


- Update version to 2026.7.4-rc.3 in deployment files @github-actions[bot]


## v2026.7.4-rc.3 (2026-07-30)


### Features


- **schema:** Ensure missing columns in app-schema direct fallback (#294) @bartcode


### Miscellaneous


- Update version to 2026.7.4-rc.2 in deployment files @github-actions[bot]


## v2026.7.4-rc.2 (2026-07-30)


### Features


- **schema:** Add direct-apply fallback for app-schema with extension refs (#293) @bartcode


## v2026.7.4-rc.1 (2026-07-30)


### Bug fixes


- **release:** Copy all workspaces in Dockerfile and revert js-yaml to v4 (#290) @bartcode


### Dependencies


- **deps:** Update dependency js-yaml to ^5.2.2 [security] (#291) @renovate[bot]


### Features


- **schema:** Opt-in declarative app-schema for application tables (#292) @bartcode


### Miscellaneous


- Update version to 2026.7.4-rc.1 in deployment files @github-actions[bot]


## v2026.7.3 (2026-07-29)


### Dependencies


- **deps:** Update all dependencies (#285) @renovate[bot]

- **deps:** Update dependency svelte to ^5.56.5 [security] (#289) @renovate[bot]


### Features


- **sdk:** Add SvelteKit SDK, Next.js/Vue scaffolds, and injectable storage adapter (#288) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.25 in deployment files @github-actions[bot]


## v2026.7.3-rc.25 (2026-07-29)


### Bug fixes


- **api:** Service_role bypasses max_total_results pagination caps (#287) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.24 in deployment files @github-actions[bot]


## v2026.7.3-rc.24 (2026-07-29)


### Bug fixes


- **sdk:** QueryBuilder.range() returns a clone for safe pagination (#286) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.23 in deployment files @github-actions[bot]


## v2026.7.3-rc.23 (2026-07-28)


### Bug fixes


- **cli:** Show migration error details on sync failures (#283) @bartcode


### Features


- **ai:** Observability, routing eval harness, custom-tools in supervisor mode (#284) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.22 in deployment files @github-actions[bot]


## v2026.7.3-rc.22 (2026-07-27)


### Bug fixes


- **jobs:** Recover stale running jobs on startup (#280) @bartcode

- **test:** Accept new Dex redirect URL format (#282) @bartcode


### Dependencies


- **deps:** Update dependency postcss to ^8.5.16 [security] (#275) @renovate[bot]

- **deps:** Update dependency astro to ^7.1.0 [security] (#276) @renovate[bot]

- **deps:** Update dependency postcss to ^8.5.18 [security] (#277) @renovate[bot]

- **deps:** Update module github.com/gofiber/fiber/v3 to v3.4.0 [security] (#278) @renovate[bot]

- **deps:** Update module go.opentelemetry.io/otel to v1.44.0 [security] (#279) @renovate[bot]


### Features


- **admin:** Add Anthropic (Claude) to provider type dropdown (#281) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.21 in deployment files @github-actions[bot]


## v2026.7.3-rc.21 (2026-07-23)


### Bug fixes


- **admin:** AI Providers duplicate header + sidebar reorg + remove API Reference (#274) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.20 in deployment files @github-actions[bot]


## v2026.7.3-rc.20 (2026-07-23)


### Bug fixes


- **admin:** Middleware ordering, logo sizing, cache keys, double headers (#273) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.19 in deployment files @github-actions[bot]


## v2026.7.3-rc.19 (2026-07-22)


### Bug fixes


- **rpc:** Use ON CONFLICT upsert to resolve sync race (#272) @bartcode


### Dependencies


- **deps:** Update dependency astro to ^7.0.7 [security] (#271) @renovate[bot]


### Miscellaneous


- Update version to 2026.7.3-rc.18 in deployment files @github-actions[bot]


## v2026.7.3-rc.18 (2026-07-22)


### Bug fixes


- **storage:** Add tenant_id to all INSERT statements with ON CONFLICT (#270) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.17 in deployment files @github-actions[bot]


## v2026.7.3-rc.17 (2026-07-22)


### Bug fixes


- **api:** PUT /settings/custom/* does upsert, not update-only (#269) @bartcode


### Dependencies


- **deps:** Update module google.golang.org/grpc to v1.82.1 [security] (#268) @renovate[bot]


### Miscellaneous


- Update version to 2026.7.3-rc.16 in deployment files @github-actions[bot]


## v2026.7.3-rc.16 (2026-07-21)


### Bug fixes


- **ai:** Page-profile agent fallback uses first allowed agent, not chat (#267) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.15 in deployment files @github-actions[bot]


## v2026.7.3-rc.15 (2026-07-21)


### Features


- **ai:** Provider config fallback for max_tokens and temperature (#266) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.14 in deployment files @github-actions[bot]


## v2026.7.3-rc.14 (2026-07-21)


### Bug fixes


- **integrations:** Cast created_by to text before COALESCE (#265) @bartcode


### Dependencies


- **deps:** Update dependency js-yaml to ^4.3.0 [security] (#264) @renovate[bot]


### Miscellaneous


- Update version to 2026.7.3-rc.13 in deployment files @github-actions[bot]


## v2026.7.3-rc.13 (2026-07-20)


### Bug fixes


- **integrations:** Response_time must be float64 not string (#263) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.12 in deployment files @github-actions[bot]


## v2026.7.3-rc.12 (2026-07-20)


### Bug fixes


- **integrations:** COALESCE nullable string columns to prevent NULL scan crash (#262) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.11 in deployment files @github-actions[bot]


## v2026.7.3-rc.11 (2026-07-20)


### Bug fixes


- **rpc:** Parse multi-line @fluxbase:input and @fluxbase:output JSON specs (#260) @bartcode


### Features


- **ai:** Supervisor-web-triggers annotation for chatbot-specific routing (#261) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.10 in deployment files @github-actions[bot]


### Style


- **ai:** Collapse extra blank lines in agent_prompts.go (gofumpt) (#259) @bartcode


## v2026.7.3-rc.10 (2026-07-19)


### Bug fixes


- **ai:** Make web-enabled suffix principle-based, not phrase-based (#258) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.9 in deployment files @github-actions[bot]


## v2026.7.3-rc.9 (2026-07-19)


### Bug fixes


- **ai:** Tell supervisor when web agent is available per chatbot (#257) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.8 in deployment files @github-actions[bot]


## v2026.7.3-rc.8 (2026-07-19)


### Bug fixes


- **rpc:** Honor optional '?' marker in input_schema when building SQL (#256) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.7 in deployment files @github-actions[bot]


## v2026.7.3-rc.7 (2026-07-19)


### Miscellaneous


- Update version to 2026.7.3-rc.6 in deployment files @github-actions[bot]


### Tests


- **integrations:** Cover env_config + handler helpers to clear 20% coverage gate (#255) @bartcode


## v2026.7.3-rc.6 (2026-07-19)


### Features


- **ai:** Tavily web search as first-class tool integration (#254) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.5 in deployment files @github-actions[bot]


## v2026.7.3-rc.5 (2026-07-19)


### Miscellaneous


- Update version to 2026.7.3-rc.4 in deployment files @github-actions[bot]


## v2026.7.3-rc.4 (2026-07-19)


### Bug fixes


- **ai:** Conversations saved with NULL user_id break listConversations (#253) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.3 in deployment files @github-actions[bot]


## v2026.7.3-rc.3 (2026-07-18)


### Bug fixes


- **ai:** Invoke_rpc nil-deref crash + supervisor router panic recovery (#252) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.2 in deployment files @github-actions[bot]


## v2026.7.3-rc.2 (2026-07-18)


### Bug fixes


- **cli:** Add --namespace flag to `chatbots list` (#250) @bartcode

- **ai+cli:** Supervisor crash, language regression, thought streaming, sync tenant_id (#251) @bartcode


### Miscellaneous


- Update version to 2026.7.3-rc.1 in deployment files @github-actions[bot]


## v2026.7.3-rc.1 (2026-07-18)


### Bug fixes


- **cli:** Display sync errors in chatbot sync output (#248) @bartcode

- **docs:** Wrap autogenerate sidebar items per Starlight 0.39+ schema @bartcode


### Dependencies


- **deps:** Update all dependencies (#247) @renovate[bot]


### Features


- **ai:** Multi-agent supervisor + page-aware chatbots (#249) @bartcode


### Miscellaneous


- Update version to 2026.7.2 in deployment files @github-actions[bot]


## v2026.7.2 (2026-07-16)


### Features


- Service-key revocation audit, export React hooks, widen SDK type, guides, config cleanup (#246) @bartcode


### Miscellaneous


- Update version to 2026.7.2-rc.4 in deployment files @github-actions[bot]


## v2026.7.2-rc.4 (2026-07-15)


### Documentation


- Fix accuracy defects and broken examples across guides (#245) @bartcode


### Miscellaneous


- Update version to 2026.7.2-rc.3 in deployment files @github-actions[bot]


## v2026.7.2-rc.3 (2026-07-13)


### Bug fixes


- Storage signed-download 0 bytes, embedded-worker reaping, default-bucket FK (#244) @bartcode


### Dependencies


- **deps:** Update module github.com/xuri/excelize/v2 to v2.11.0 [security] (#243) @renovate[bot]


### Miscellaneous


- Update version to 2026.7.2-rc.2 in deployment files @github-actions[bot]


## v2026.7.2-rc.2 (2026-07-07)


### Features


- Add ffmpeg to runtime image for video transcoding jobs (#242) @bartcode


### Miscellaneous


- Update version to 2026.7.2-rc.1 in deployment files @github-actions[bot]


## v2026.7.2-rc.1 (2026-07-05)


### Miscellaneous


- Update version to 2026.7.1 in deployment files @github-actions[bot]


## v2026.7.1 (2026-07-05)


### Dependencies


- **deps:** Update dependency pnpm to v11.8.0 [security] (#239) @renovate[bot]

- **deps:** Update module github.com/gofiber/fiber/v3 to v3.3.0 [security] (#240) @renovate[bot]


### Miscellaneous


- Update version to 2026.6.3-rc.1 in deployment files @github-actions[bot]


## v2026.6.3-rc.1 (2026-06-30)


### Dependencies


- **deps:** Update dependency vite to ^8.0.16 [security] (#235) @renovate[bot]

- **deps:** Update dependency astro to ^6.4.6 [security] (#236) @renovate[bot]

- **deps:** Update dependency pnpm to v11 [security] (#237) @renovate[bot]


### Features


- AI prompt caching, native Anthropic provider, per-user quota, KG wiring fixes (#238) @bartcode


### Miscellaneous


- Update version to 2026.6.2 in deployment files @github-actions[bot]


## v2026.6.2 (2026-06-14)


### Bug fixes


- **ci:** Anchor gitignore 'fluxbase' pattern to root to unblock helm chart commits @bartcode

- **docs:** Correct relative links in SDK getting-started guide @bartcode


### Miscellaneous


- Sync source files to published 2026.6.1 release @bartcode


## v2026.6.1 (2026-06-14)


### Bug fixes


- **ci:** Remove invalid file threshold from coverage config (#233) @bartcode

- **devcontainer:** Bootstrap app user + grants on fresh volumes; harden admin SPA against backend downtime @bartcode


### Dependencies


- **deps:** Update module golang.org/x/crypto to v0.52.0 [security] (#228) @renovate[bot]

- **deps:** Update dependency axios to ^1.16.0 [security] (#229) @renovate[bot]

- **deps:** Update dependency esbuild to v0.28.1 [security] (#230) @renovate[bot]

- **deps:** Remove unused gorilla/websocket, track devcontainer feature lock (#234) @bartcode


### Features


- AI bug fixes, SDK KB module, React hooks, admin UI improvements (#231) @bartcode

- DX/UX consistency improvements (#232) @bartcode


## v2026.5.5-rc.2 (2026-05-17)


### Bug fixes


- Remove unused label variable in auth-flows test @bartcode


### Dependencies


- **deps:** Update dependency astro to ^6.2.1 [security] (#226) @renovate[bot]


### Documentation


- Update developer guide for module system and file splits @bartcode


### Features


- **e2e-ui:** Comprehensive Playwright page tests (#227) @bartcode


### Miscellaneous


- Update version to 2026.5.5-rc.1 in deployment files @github-actions[bot]


### Refactoring


- Eliminate Server.Shutdown god-method, remove dead registrations (#220) @bartcode

- Eliminate wiring bridge — modules own their handler groups (#221) @bartcode

- Eliminate sideways mutations, split route adapter (#222) @bartcode

- Split admin route deps, eliminate Server god-struct (#223) @bartcode

- Standardize error responses across handlers and middleware (#224) @bartcode

- God-object decomposition, file splits, facade elimination, SettingsCache extraction (#225) @bartcode


## v2026.5.5-rc.1 (2026-05-08)


### Bug fixes


- Use tenant UUID for user count instead of IS NULL check @bartcode

- **docs:** Update broken links in admin, developer-guide, and distributed-tracing docs (#216) @bartcode


### Dependencies


- **deps:** Update dependency axios to ^1.15.2 [security] (#213) @renovate[bot]


### Miscellaneous


- Update version to 2026.5.4 in deployment files @github-actions[bot]


## v2026.5.4 (2026-05-08)


### Features


- Add user_count and admin_count to tenant detail, assign initial admin to default tenant @bartcode


### Miscellaneous


- Update version to 2026.5.4-rc.23 in deployment files @github-actions[bot]


## v2026.5.4-rc.23 (2026-05-07)


### Bug fixes


- Set owner_id to nil for admin table exports to avoid FK violation @bartcode


### Miscellaneous


- Update version to 2026.5.4-rc.22 in deployment files @github-actions[bot]


## v2026.5.4-rc.22 (2026-05-07)


### Bug fixes


- Read KB ID from request body in table export handler, add Export All button @bartcode


### Miscellaneous


- Update version to 2026.5.4-rc.21 in deployment files @github-actions[bot]


## v2026.5.4-rc.21 (2026-05-07)


### Bug fixes


- Wire admin KB routes for documents, search, graph and switch admin UI to admin endpoints @bartcode


### Miscellaneous


- Update version to 2026.5.4-rc.20 in deployment files @github-actions[bot]


## v2026.5.4-rc.20 (2026-05-07)


### Features


- Enforce chatbot settings, add KB/AI CLI commands, fix settings bug (#214) @bartcode


### Miscellaneous


- Update version to 2026.5.4-rc.19 in deployment files @github-actions[bot]


## v2026.5.4-rc.19 (2026-05-05)


### Bug fixes


- Enable dashboard by default and fix OAuth (#207) @bartcode

- Propagate tenant context in AI chat and realtime WebSocket handlers (#211) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.17 in deployment files @github-actions[bot]

- Update version to 2026.5.1 in deployment files @github-actions[bot]

- Update version to 2026.5.2-rc.1 in deployment files @github-actions[bot]

- Update version to 2026.5.2-rc.2 in deployment files @github-actions[bot]

- Update version to 2026.5.2-rc.3 in deployment files @github-actions[bot]

- Update version to 2026.5.2 in deployment files @github-actions[bot]

- Update version to 2026.5.2-rc.4 in deployment files @github-actions[bot]

- Update version to 2026.5.2 in deployment files @github-actions[bot]

- Update version to 2026.5.3-rc.1 in deployment files @github-actions[bot]

- Update version to 2026.5.3 in deployment files @github-actions[bot]


## v2026.3.7-rc.17 (2026-05-01)


### Miscellaneous


- Update version to 2026.3.7-rc.16 in deployment files @github-actions[bot]


## v2026.3.7-rc.16 (2026-05-01)


### Bug fixes


- Update settings defaults (#206) @bartcode


### Dependencies


- **deps:** Update module github.com/gofiber/fiber/v3 to v3.2.0 [security] (#202) @renovate[bot]


### Miscellaneous


- Update version to 2026.3.7-rc.15 in deployment files @github-actions[bot]


## v2026.3.7-rc.15 (2026-04-30)


### Bug fixes


- Add viper config fallback to settings cache so features default to enabled (#205) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.14 in deployment files @github-actions[bot]


## v2026.3.7-rc.14 (2026-04-30)


### Bug fixes


- **tenancy:** Harden FDW schema import against transient connection failures (#203) @bartcode


### Documentation


- Comprehensive documentation overhaul (#201) @bartcode

- Fix voice, middleware ordering, feature grid, CLI tenant flag, and cleanup (#204) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.13 in deployment files @github-actions[bot]


## v2026.3.7-rc.13 (2026-04-27)


### Bug fixes


- Admin redirect, tenant repair, extension deps, schema graph, instance-level OAuth/SAML (#200) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.12 in deployment files @github-actions[bot]


## v2026.3.7-rc.12 (2026-04-26)


### Bug fixes


- Enforce tenant isolation across all handlers and add repair endpoint (#194) @bartcode

- Resolve tenant isolation bugs. (#195) @bartcode


### Dependencies


- **deps:** Update module github.com/jackc/pgx/v5 to v5.9.2 [security] (#197) @renovate[bot]

- **deps:** Update dependency astro to ^6.1.6 [security] (#196) @renovate[bot]


### Miscellaneous


- Update version to 2026.3.7-rc.10 in deployment files @github-actions[bot]


### Refactoring


- Centralize error handling, validation, scheduling, and storage patterns (#198) @bartcode

- Simplify Go codebase for security and maintainability (#199) @bartcode


## v2026.3.7-rc.10 (2026-04-19)


### Bug fixes


- Wire settings cache to fix email provider switcher not persisting (#193) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.9 in deployment files @github-actions[bot]


## v2026.3.7-rc.9 (2026-04-19)


### Bug fixes


- Reorder jobs routes and fix RPC executions infinite loop (#192) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.8 in deployment files @github-actions[bot]


## v2026.3.7-rc.8 (2026-04-19)


### Miscellaneous


- Update version to 2026.3.7-rc.7 in deployment files @github-actions[bot]


## v2026.3.7-rc.7 (2026-04-17)


### Miscellaneous


- Update version to 2026.3.7-rc.6 in deployment files @github-actions[bot]


## v2026.3.7-rc.6 (2026-04-15)


### Bug fixes


- Ensure routes for secrest are the same as before. @bartcode


### Documentation


- Update docs to match the codebase. @bartcode

- More alignment between docs and code. @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.5 in deployment files @github-actions[bot]


## v2026.3.7-rc.5 (2026-04-15)


### Bug fixes


- Wire tenant middleware into Jobs, Webhooks, RPC, and GraphQL routes (#188) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.4 in deployment files @github-actions[bot]


## v2026.3.7-rc.4 (2026-04-11)


### Bug fixes


- Add e2e tests and tenant_id column. @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.3 in deployment files @github-actions[bot]


## v2026.3.7-rc.3 (2026-04-10)


### Bug fixes


- Resolve edge cases for multi-tenancy. @bartcode


### Dependencies


- **deps:** Update dependency axios to ^1.15.0 [security] (#187) @renovate[bot]


## v2026.3.7-rc.2 (2026-04-09)


### Bug fixes


- Update renovate.json to use a minimum release age. @bartcode

- Resolve issues with declarative schemas. @bartcode

- Update service key middleware. @bartcode


### CI


- Upgrade Helm setup. @bartcode


### Dependencies


- **deps:** Update dependency go to v1.26.1 (#185) @renovate[bot]

- **deps:** Update all dependencies (#180) @renovate[bot]

- **deps:** Update go dependencies (#182) @renovate[bot]


### Features


- Introduce Playwright testing and fix storage handler bug. (#186) @bartcode


### Miscellaneous


- Update version to 2026.3.7-rc.1 in deployment files @github-actions[bot]

- Update version to 2026.3.7-rc.2 in deployment files @github-actions[bot]


## v2026.3.7-rc.1 (2026-04-06)


### Dependencies


- **deps:** Update all dependencies (#175) @renovate[bot]

- **deps:** Update module google.golang.org/grpc to v1.79.3 [security] (#181) @renovate[bot]

- **deps:** Update go dependencies (#174) @renovate[bot]


### Miscellaneous


- Update version to 2026.3.6 in deployment files @github-actions[bot]


## v2026.3.6 (2026-03-13)


### Miscellaneous


- Update version to 2026.3.6-rc.2 in deployment files @github-actions[bot]


## v2026.3.6-rc.2 (2026-03-13)


### Bug fixes


- Update package references. @bartcode


## v2026.3.6-rc.1 (2026-03-13)


### Bug fixes


- Wire configuration for all log backends (#176) @bartcode

- Rename GitHub organisation to nimbleflux. (#177) @bartcode

- Do not follow links to Claude or ChatGPT. @bartcode


### Dependencies


- **deps:** Update module golang.org/x/crypto to v0.49.0 (#172) @renovate[bot]

- **deps:** Update dependency @astrojs/starlight to ^0.38.1 (#171) @renovate[bot]

- **deps:** Update module github.com/olekukonko/tablewriter to v1.1.4 (#173) @renovate[bot]


### Miscellaneous


- Update version to 2026.3.5 in deployment files @github-actions[bot]


## v2026.3.5 (2026-03-11)


### Miscellaneous


- Update version to 2026.3.5-rc.7 in deployment files @github-actions[bot]


## v2026.3.5-rc.7 (2026-03-11)


### Bug fixes


- Include admin package.json for workspace resolution in docs build @bartcode


### Miscellaneous


- Update version to 2026.3.5-rc.6 in deployment files @github-actions[bot]


## v2026.3.5-rc.6 (2026-03-11)


### Bug fixes


- Run bun install in builder stage to set up workspace symlinks for docs build @bartcode


### Miscellaneous


- Update version to 2026.3.5-rc.5 in deployment files @github-actions[bot]


## v2026.3.5-rc.5 (2026-03-11)


### Bug fixes


- Resolve astro command not found in docs Docker build @bartcode


### Miscellaneous


- Update version to 2026.3.5-rc.4 in deployment files @github-actions[bot]


## v2026.3.5-rc.4 (2026-03-11)


### Bug fixes


- Add Vite alias for @fluxbase/sdk-react to resolve ESM module @bartcode


### Miscellaneous


- Update version to 2026.3.5-rc.3 in deployment files @github-actions[bot]


## v2026.3.5-rc.3 (2026-03-11)


### Bug fixes


- Remove redundant file:../sdk from sdk-react devDependencies @bartcode


### Dependencies


- **deps:** Update all dependencies @renovate[bot]


### Miscellaneous


- Update version to 2026.3.5-rc.2 in deployment files @github-actions[bot]


## v2026.3.5-rc.2 (2026-03-11)


### Miscellaneous


- Update version to 2026.3.5-rc.1 in deployment files @github-actions[bot]


## v2026.3.5-rc.1 (2026-03-11)


### Miscellaneous


- Update version to 2026.3.4 in deployment files @github-actions[bot]


## v2026.3.4 (2026-03-11)


### Dependencies


- **deps:** Update module golang.org/x/term to v0.41.0 (#169) @renovate[bot]

- **deps:** Update all dependencies (#168) @renovate[bot]


### Features


- Add getProviderToken endpoint for retrieving OAuth provider tokens (#170) @bartcode


### Miscellaneous


- Update version to 2026.3.3 in deployment files @github-actions[bot]


## v2026.3.3 (2026-03-10)


### Miscellaneous


- Update version to 2026.3.2-rc.8 in deployment files @github-actions[bot]

- Update version to 2026.3.2 in deployment files @github-actions[bot]


## v2026.3.2-rc.8 (2026-03-09)


### Bug fixes


- Downgrade Zod in admin. @bartcode


### Miscellaneous


- Update version to 2026.3.2-rc.7 in deployment files @github-actions[bot]


## v2026.3.2-rc.7 (2026-03-09)


### Documentation


- Make docs work with bun. @bartcode


### Miscellaneous


- Update version to 2026.3.2-rc.6 in deployment files @github-actions[bot]


## v2026.3.2-rc.6 (2026-03-09)


### Bug fixes


- **docs:** Add explicit zod v3 dependency for Starlight compatibility @bartcode


### Miscellaneous


- Update version to 2026.3.2-rc.5 in deployment files @github-actions[bot]


## v2026.3.2-rc.5 (2026-03-09)


### Bug fixes


- **ci:** Refresh dependencies after SDK builds in Dockerfile @bartcode


### Miscellaneous


- Update version to 2026.3.2-rc.4 in deployment files @github-actions[bot]


## v2026.3.2-rc.4 (2026-03-09)


### Bug fixes


- **ci:** Use correct bun.lock filename in release workflow @bartcode


## v2026.3.2-rc.3 (2026-03-09)


### Bug fixes


- **ci:** Use correct bun setup action in release workflow @bartcode


## v2026.3.2-rc.2 (2026-03-09)


### Bug fixes


- Update docs about secrets and allow for fetching a secret by name. (#161) @bartcode

- **ci:** Prevent deletion of production Postgres images in cleanup workflow (#162) @bartcode

- **oauth:** Correct refresh token handling for all providers (#164) @bartcode


### Dependencies


- **deps:** Update dependency react-resizable-panels to ^4.7.2 (#158) @renovate[bot]

- **deps:** Update npm dependencies (#160) @renovate[bot]

- **deps:** Update go dependencies (#159) @renovate[bot]

- **deps:** Update dependency typescript-eslint to ^8.57.0 (#165) @renovate[bot]

- **deps:** Update npm dependencies (#167) @renovate[bot]


### Documentation


- Clarify licensing in docs. (#163) @bartcode


### Miscellaneous


- Update version to 2026.3.2-rc.1 in deployment files @github-actions[bot]

- Update version to 2026.3.2-rc.2 in deployment files @github-actions[bot]

- Migrate from npm to bun (#166) @bartcode


## v2026.3.2-rc.1 (2026-03-07)


### Dependencies


- **deps:** Update clickhouse docker tag to v26.2 (#155) @renovate[bot]

- **deps:** Update npm dependencies (#154) @renovate[bot]

- **deps:** Update go dependencies to v1.42.0 (#156) @renovate[bot]

- **deps:** Update dependency knip to ^5.86.0 (#157) @renovate[bot]


### Miscellaneous


- Update version to 2026.3.1 in deployment files @github-actions[bot]


## v2026.3.1 (2026-03-06)


### Bug fixes


- Use current month when promoting RC from previous month @bartcode


### Documentation


- Remove ARM dependency. @bartcode


### Miscellaneous


- Update version to 2026.2.8 in deployment files @github-actions[bot]


## v2026.2.8 (2026-03-06)


### Bug fixes


- Apply security fixes. (#153) @bartcode


### Dependencies


- **deps:** Update dependency lint-staged to ^16.3.0 (#138) @renovate[bot]

- **deps:** Update module github.com/shirou/gopsutil/v4 to v4.26.2 (#141) @renovate[bot]

- **deps:** Update npm dependencies (#140) @renovate[bot]

- **deps:** Update npm dependencies (#142) @renovate[bot]

- **deps:** Update module github.com/davidbyttow/govips/v2 to v2.17.0 (#143) @renovate[bot]

- **deps:** Update all dependencies (#146) @renovate[bot]

- **deps:** Update npm dependencies (#145) @renovate[bot]

- **deps:** Update go dependencies (#144) @renovate[bot]

- **deps:** Update go to v1.26.1 (#149) @renovate[bot]

- **deps:** Update dependency @types/node to ^24.12.0 (#148) @renovate[bot]

- **deps:** Update all dependencies (#147) @renovate[bot]

- **deps:** Update module google.golang.org/grpc to v1.79.2 (#151) @renovate[bot]

- **deps:** Update dependency dt-sql-parser to ^4.4.2 (#150) @renovate[bot]

- **deps:** Update module github.com/minio/minio-go/v7 to v7.0.99 (#152) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.8-rc.1 in deployment files @github-actions[bot]

- Update version to 2026.2.8 in deployment files @github-actions[bot]


## v2026.2.8-rc.1 (2026-02-28)


### Bug fixes


- Align CLI docs with implementaion. (#139) @bartcode


### Miscellaneous


- Update version to 2026.2.7 in deployment files @github-actions[bot]


## v2026.2.7 (2026-02-28)


### Miscellaneous


- Update version to 2026.2.7-rc.1 in deployment files @github-actions[bot]


## v2026.2.7-rc.1 (2026-02-28)


### Bug fixes


- Add missing weight field in Helm chart. (#135) @bartcode

- Store OAuth session in database and prevent any service role keys being blocked automatically. (#137) @bartcode


### Dependencies


- **deps:** Update npm dependencies (#136) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.6 in deployment files @github-actions[bot]


## v2026.2.6 (2026-02-27)


### Bug fixes


- Prevent concurrent map access race in realtime subscriptions (#132) @bartcode

- Add document-filtering. (#134) @bartcode


### Dependencies


- **deps:** Update dependency axios to ^1.13.6 (#133) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.5 in deployment files @github-actions[bot]


## v2026.2.5 (2026-02-27)


### Bug fixes


- Correct health check display in admin overview (#131) @bartcode

- Check if stable version exists when promoting RC to stable @bartcode


### Dependencies


- **deps:** Update dependency @types/node to ^24.10.15 (#128) @renovate[bot]

- **deps:** Update elasticsearch docker tag to v9.3.1 (#130) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.4 in deployment files @github-actions[bot]

- Update version to 2026.2.4 in deployment files @github-actions[bot]


## v2026.2.4 (2026-02-26)


### Miscellaneous


- Update version to 2026.2.4-rc.2 in deployment files @github-actions[bot]


## v2026.2.4-rc.2 (2026-02-26)


### Bug fixes


- Add min-width to quota table cells @bartcode

- Improve quota table cell layout @bartcode

- Remove progress bar from quota page. @bartcode


### Miscellaneous


- Update version to 2026.2.4-rc.1 in deployment files @github-actions[bot]


## v2026.2.4-rc.1 (2026-02-26)


### Bug fixes


- Improve connection stability and add panic recovery (#127) @bartcode

- Update Postgres image in devcontainer. @bartcode

- Use api client in quotas page for proper auth @bartcode


### Dependencies


- **deps:** Update all dependencies (#129) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.3-rc.24 in deployment files @github-actions[bot]


## v2026.2.3-rc.24 (2026-02-26)


### Bug fixes


- Make long-running jobs work again. (#122) @bartcode

- Update broken links in docs and upgrade to 1.26.0 everywhere. (#124) @bartcode


### CI


- Use numeric comparison during release process. @bartcode


### Dependencies


- **deps:** Update module github.com/gofiber/fiber/v3 to v3.1.0 [security] (#118) @renovate[bot]

- **deps:** Update module github.com/xuri/excelize/v2 to v2.10.1 (#119) @renovate[bot]

- **deps:** Update npm dependencies to ^1.163.2 (#120) @renovate[bot]

- **deps:** Update dependency astro to ^5.18.0 (#121) @renovate[bot]

- **deps:** Update all dependencies (#111) @renovate[bot]

- **deps:** Update dependency @types/node to ^24.10.14 (#125) @renovate[bot]

- **deps:** Update dependency react-day-picker to v9.14.0 (#126) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.3 in deployment files @github-actions[bot]

- Update version to 2026.2.3 in deployment files @github-actions[bot]


## v2026.2.3 (2026-02-24)


### Bug fixes


- Improve reasoning and implement ReAct. (#117) @bartcode


### Dependencies


- **deps:** Update npm dependencies (#113) @renovate[bot]

- **deps:** Update go dependencies (#116) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.3-rc.23 in deployment files @github-actions[bot]


## v2026.2.3-rc.23 (2026-02-23)


### Bug fixes


- Improve KB graph. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.22 in deployment files @github-actions[bot]


## v2026.2.3-rc.22 (2026-02-22)


### Bug fixes


- Allow system owners. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.21 in deployment files @github-actions[bot]


## v2026.2.3-rc.21 (2026-02-22)


### Bug fixes


- More KB fixes. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.20 in deployment files @github-actions[bot]


## v2026.2.3-rc.20 (2026-02-22)


### Bug fixes


- Ensure CreatedBy and OwnerID are set. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.19 in deployment files @github-actions[bot]


## v2026.2.3-rc.19 (2026-02-22)


### Bug fixes


- Always set base ID for KB. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.18 in deployment files @github-actions[bot]


## v2026.2.3-rc.18 (2026-02-22)


### Bug fixes


- Always set owner ID when uploading documents. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.17 in deployment files @github-actions[bot]


## v2026.2.3-rc.17 (2026-02-22)


### Bug fixes


- Add logic to determine owner_id when adding KBs. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.16 in deployment files @github-actions[bot]


## v2026.2.3-rc.16 (2026-02-22)


### Bug fixes


- Resolve encoding bug. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.15 in deployment files @github-actions[bot]


## v2026.2.3-rc.15 (2026-02-22)


### Bug fixes


- Update KB CLI and introduce namespaces. (#115) @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.14 in deployment files @github-actions[bot]


## v2026.2.3-rc.14 (2026-02-22)


### Bug fixes


- Allow type annotations for JSONB. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.13 in deployment files @github-actions[bot]


## v2026.2.3-rc.13 (2026-02-21)


### Miscellaneous


- Update version to 2026.2.3-rc.12 in deployment files @github-actions[bot]


## v2026.2.3-rc.12 (2026-02-21)


### Bug fixes


- Automatically upgrade Go version through Renovate. @bartcode

- Pin to Go 1.25. @bartcode

- Make test more reliable. @bartcode

- Show label on KB page. @bartcode

- Minor fixes. @bartcode

- Make webhook retry test more reliable in CI @bartcode

- Escape changelog content in release workflow @bartcode

- Update Renovate config to properly pin Go version @bartcode

- Ensure CLI binaries are published in release workflow @bartcode


### CI


- Upgrade Go version. @bartcode


### Dependencies


- **deps:** Update go dependencies @renovate[bot]


### Miscellaneous


- Update version to 2026.2.3-rc.11 in deployment files @github-actions[bot]


## v2026.2.3-rc.11 (2026-02-21)


### Bug fixes


- Escape changelog content in release workflow @bartcode


### Dependencies


- **deps:** Update npm dependencies to ^1.161.4 (#112) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.3-rc.10 in deployment files @github-actions[bot]


## v2026.2.3-rc.9 (2026-02-21)


### Dependencies


- **deps:** Update all dependencies (#107) @renovate[bot]

- **deps:** Update npm dependencies (#108) @renovate[bot]

- **deps:** Update dependency react-hook-form to ^7.71.2 (#110) @renovate[bot]

- **deps:** Update go dependencies (#105) @renovate[bot]


### Features


- Add multi-backend logging and knowledge graph support. (#109) @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.6 in deployment files @github-actions[bot]


## v2026.2.3-rc.6 (2026-02-11)


### Miscellaneous


- Update version to 2026.2.3-rc.5 in deployment files @github-actions[bot]


## v2026.2.3-rc.5 (2026-02-11)


### Bug fixes


- Update tests. @bartcode

- Add more time to retries. @bartcode

- Make e2e tests work. @bartcode


### Dependencies


- **deps:** Update all dependencies (#106) @renovate[bot]

- **deps:** Update npm dependencies (#104) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.3-rc.4 in deployment files @github-actions[bot]


## v2026.2.3-rc.4 (2026-02-10)


### Bug fixes


- Exempt static files from rate limiting. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.3 in deployment files @github-actions[bot]


## v2026.2.3-rc.3 (2026-02-10)


### Bug fixes


- Update delete query. @bartcode

- Fix grants on system schema. @bartcode

- Add debug logs about rate limiting. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.2 in deployment files @github-actions[bot]


## v2026.2.3-rc.2 (2026-02-10)


### Bug fixes


- Clean up test resources. @bartcode

- Add cleanup command. @bartcode


### CI


- Simplify release options. @bartcode


### Miscellaneous


- Update version to 2026.2.3-rc.1 in deployment files @github-actions[bot]


## v2026.2.3-rc.1 (2026-02-10)


### Dependencies


- **deps:** Update dependency @fluxbase/sdk to ^2026.2.2 (#103) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.2-rc.5 in deployment files @github-actions[bot]


## v2026.2.2-rc.5 (2026-02-10)


### Bug fixes


- Add many more tests (#99) @bartcode

- Implement more security features. (#100) @bartcode

- Update Renovate settings. @bartcode


### CI


- Use latest stable tag for release notes. @bartcode


### Dependencies


- **deps:** Update node.js to v25.6.0 (#92) @renovate[bot]

- **deps:** Update dependency @types/react to ^19.2.11 (#94) @renovate[bot]

- **deps:** Update dependency go to v1.25.7 (#98) @renovate[bot]

- **deps:** Update dependency axios to v1.13.5 [security] (#101) @renovate[bot]

- **deps:** Update npm dependencies (#97) @renovate[bot]

- **deps:** Update go dependencies (#95) @renovate[bot]

- **deps:** Update npm dependencies to ^9.39.2 (#102) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.2 in deployment files @github-actions[bot]


## v2026.2.2 (2026-02-03)


### Bug fixes


- Update AI summary prompt. @bartcode

- Add retries for the database connection. @bartcode


### Miscellaneous


- Update version to 2026.2.2-rc.4 in deployment files @github-actions[bot]


## v2026.2.2-rc.4 (2026-02-03)


### CI


- Improve prompt for AI summary. @bartcode


## v2026.2.2-rc.3 (2026-02-03)


### Bug fixes


- Use Groq. @bartcode


### CI


- Update input label for release pipeline. @bartcode


### Miscellaneous


- Update version to 2026.2.2-rc.2 in deployment files @github-actions[bot]

- Update version to 2026.2.2-rc.3 in deployment files @github-actions[bot]


## v2026.2.2-rc.2 (2026-02-03)


### CI


- Update AI summary in descriptions. @bartcode


### Miscellaneous


- Update version to 2026.2.2-rc.1 in deployment files @github-actions[bot]


## v2026.2.2-rc.1 (2026-02-03)


### Bug fixes


- Update package.lock.json for sdk-react during pipeline. (#75) @bartcode

- Make panels on SQL editor and Tables page resizeable again (#80) @bartcode

- Improve CI process (#79) @bartcode

- Make dashboard page titles consistent. (#83) @bartcode

- Remove rate-limiting in dashboard for webhooks. @bartcode

- Improve test coverage (#85) @bartcode

- Delete redundant file. @bartcode

- Provide CODECOV_TOKEN. @bartcode

- Update e2e tests. @bartcode


### Dependencies


- **deps:** Update module github.com/mailgun/mailgun-go/v5 to v5.12.0 (#73) @renovate[bot]

- **deps:** Update dependency @fluxbase/sdk to ^2026.2.1 (#74) @renovate[bot]

- **deps:** Update dependency react-resizable-panels to ^4.5.8 (#76) @renovate[bot]

- **deps:** Update dependency jsdom to v28 (#77) @renovate[bot]

- **deps:** Update dependency @vitejs/plugin-react-swc to ^4.2.3 (#81) @renovate[bot]

- **deps:** Update all dependencies (#82) @renovate[bot]

- **deps:** Update npm dependencies (#84) @renovate[bot]

- **deps:** Update dependency deno to v2.6.8 (#87) @renovate[bot]

- **deps:** Update go dependencies (#86) @renovate[bot]

- **deps:** Update module github.com/gofiber/fiber/v2 to v2.52.9 [security] (#90) @renovate[bot]

- **deps:** Update module github.com/gofiber/fiber/v2 to v3 (#88) @renovate[bot]


### Miscellaneous


- Update version to 2026.2.1 in deployment files @github-actions[bot]


## v2026.2.1 (2026-02-01)


### Bug fixes


- Improve database branching and fix CLI. (#25) @bartcode

- Add more tests for Captcha and image transformations. (#26) @bartcode

- Update alignment of security banner. @bartcode

- Update UI. @bartcode

- **deps:** Update module google.golang.org/grpc to v1.64.1 [security] (#38) @renovate[bot]

- Group NPM dependencies. @bartcode

- **deps:** Update go dependencies (#55) @renovate[bot]

- **deps:** Update go dependencies (#66) @renovate[bot]

- **deps:** Update go dependencies (#68) @renovate[bot]

- Group all Renovate dependencies. @bartcode

- Upgrade dependencies. @bartcode

- Be less agressive with the db connection in tests. @bartcode


### Build


- **deps:** Bump github.com/gofiber/fiber/v2 from 2.52.6 to 2.52.9 (#33) @dependabot[bot]

- **deps-dev:** Bump glob from 10.4.5 to 10.5.0 in /sdk-react (#32) @dependabot[bot]

- **deps:** Bump github.com/jackc/pgx/v5 from 5.5.1 to 5.5.4 (#31) @dependabot[bot]

- **deps:** Bump github.com/go-chi/chi/v5 from 5.2.1 to 5.2.2 (#30) @dependabot[bot]

- **deps:** Bump golang.org/x/crypto from 0.43.0 to 0.45.0 (#29) @dependabot[bot]

- **deps:** Bump the npm_and_yarn group across 3 directories with 8 updates (#36) @dependabot[bot]

- **deps:** Bump the go_modules group across 1 directory with 2 updates (#35) @dependabot[bot]

- **deps-dev:** Bump esbuild, @vitest/coverage-v8 and vitest in /sdk (#34) @dependabot[bot]


### Dependencies


- **deps:** Update dependency go to v1.25.6 (#39) @renovate[bot]

- **deps:** Update dependency node to v24 (#64) @renovate[bot]

- **deps:** Update github/codeql-action action to v4 (#63) @renovate[bot]

- **deps:** Update docker/build-push-action action to v6 (#62) @renovate[bot]

- **deps:** Update codecov/codecov-action action to v5 (#60) @renovate[bot]

- **deps:** Update actions/upload-artifact action to v6 (#59) @renovate[bot]

- **deps:** Update actions/setup-node action to v6 (#58) @renovate[bot]

- **deps:** Update actions/setup-go action to v6 (#57) @renovate[bot]

- **deps:** Update actions/checkout action to v6 (#56) @renovate[bot]

- **deps:** Update node.js to v25.5.0 (#50) @renovate[bot]

- **deps:** Update dependency node to v20.20.0 (#45) @renovate[bot]

- **deps:** Update npm dependencies (#67) @renovate[bot]

- **deps:** Update npm dependencies (#69) @renovate[bot]

- **deps:** Update go dependencies (#70) @renovate[bot]

- **deps:** Update go dependencies (#71) @renovate[bot]

- **deps:** Update go dependencies (#72) @renovate[bot]


### Features


- Add visualization tools. (#27) @bartcode


### Miscellaneous


- Update version to 2026.1.22 in deployment files @github-actions[bot]


## v2026.1.22 (2026-01-30)


### Miscellaneous


- Update version to 2026.1.22-rc.10 in deployment files @github-actions[bot]


## v2026.1.22-rc.10 (2026-01-30)


### Bug fixes


- Use dashboard.users for MCP authentication. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.9 in deployment files @github-actions[bot]


## v2026.1.22-rc.9 (2026-01-30)


### Bug fixes


- Update cookie name. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.8 in deployment files @github-actions[bot]


## v2026.1.22-rc.8 (2026-01-30)


### Bug fixes


- Another URL read fix... @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.7 in deployment files @github-actions[bot]


## v2026.1.22-rc.7 (2026-01-30)


### Bug fixes


- Use return_to parameter. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.6 in deployment files @github-actions[bot]


## v2026.1.22-rc.6 (2026-01-30)


### Bug fixes


- Use correct login URL. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.5 in deployment files @github-actions[bot]


## v2026.1.22-rc.5 (2026-01-30)


### Bug fixes


- Add browser redirect. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.4 in deployment files @github-actions[bot]


## v2026.1.22-rc.4 (2026-01-30)


### Bug fixes


- Update cookie name for OAuth. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.3 in deployment files @github-actions[bot]


## v2026.1.22-rc.3 (2026-01-30)


### Bug fixes


- Improve dialogs. @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.2 in deployment files @github-actions[bot]


## v2026.1.22-rc.2 (2026-01-30)


### Bug fixes


- OAuth setup for Claude Desktop. (#24) @bartcode


### Miscellaneous


- Update version to 2026.1.22-rc.1 in deployment files @github-actions[bot]


## v2026.1.22-rc.1 (2026-01-30)


### Bug fixes


- Remove size. @bartcode

- Remove redundant Claude files. (#22) @bartcode

- Use the right URL for Fluxbase docs. @bartcode


### Miscellaneous


- Update version to 2026.1.21 in deployment files @github-actions[bot]


## v2026.1.21 (2026-01-29)


### Bug fixes


- Give the logo a light background in dark mode. @bartcode

- Remove impersonation banner. @bartcode

- Remove unnecessary warnings. @bartcode

- Ignore one more rule. @bartcode


### Miscellaneous


- Update version to 2026.1.20 in deployment files @github-actions[bot]


## v2026.1.20 (2026-01-27)


### Bug fixes


- Use partial unique index on key for app settings. @bartcode


### Miscellaneous


- Update version to 2026.1.19 in deployment files @github-actions[bot]


## v2026.1.19 (2026-01-25)


### Bug fixes


- Increase upload limit for admin endpoints. @bartcode

- Update storage limit defaults. @bartcode

- Delete then update setting. @bartcode


### Miscellaneous


- Update version to 2026.1.18 in deployment files @github-actions[bot]


## v2026.1.18 (2026-01-25)


### Bug fixes


- Make user-settings work again. @bartcode

- Move eslint comment. @bartcode

- Settings key constraint. @bartcode


### Miscellaneous


- Update version to 2026.1.17 in deployment files @github-actions[bot]


## v2026.1.17 (2026-01-25)


### Bug fixes


- Upsert shared modules. @bartcode

- Always show items from the namespace that has at least one RPC procedure. @bartcode


### Miscellaneous


- Update version to 2026.1.16 in deployment files @github-actions[bot]


## v2026.1.16 (2026-01-25)


### Bug fixes


- Increase timeout for bundling. @bartcode


### Miscellaneous


- Update version to 2026.1.15 in deployment files @github-actions[bot]


## v2026.1.15 (2026-01-25)


### Miscellaneous


- Update version to 2026.1.14-rc.8 in deployment files @github-actions[bot]


## v2026.1.14-rc.8 (2026-01-25)


### Bug fixes


- Implement forbiddenTools. @bartcode

- Add renovate config. @bartcode

- Improve structured chat results. @bartcode

- Add privacy policy. @bartcode


### Miscellaneous


- Update version to 2026.1.14-rc.7 in deployment files @github-actions[bot]


## v2026.1.14-rc.7 (2026-01-23)


### Bug fixes


- Enrich intent rules of chatbots. @bartcode


### Miscellaneous


- Update version to 2026.1.14-rc.6 in deployment files @github-actions[bot]


## v2026.1.14-rc.6 (2026-01-23)


### Bug fixes


- Parse MCP annotations correctly. @bartcode


### Miscellaneous


- Update version to 2026.1.14-rc.5 in deployment files @github-actions[bot]


## v2026.1.14-rc.5 (2026-01-23)


### Bug fixes


- Make MCP tools page prettier. @bartcode

- Remove MCP resources; we only need MCP tools. @bartcode

- Increase job size limit. @bartcode


### Miscellaneous


- Update version to 2026.1.14-rc.4 in deployment files @github-actions[bot]


## v2026.1.14-rc.4 (2026-01-23)


### Bug fixes


- Add defaults for syncing MCP tools. @bartcode

- Allow for selecting namespace when using chatbots. @bartcode


### Miscellaneous


- Update version to 2026.1.14-rc.3 in deployment files @github-actions[bot]


## v2026.1.14-rc.3 (2026-01-21)


### Bug fixes


- Add tools dir. @bartcode

- Include package-lock. @bartcode


### Miscellaneous


- Update version to 2026.1.14-rc.2 in deployment files @github-actions[bot]


## v2026.1.14-rc.2 (2026-01-21)


### Bug fixes


- Apply formatting. @bartcode

- Make tests work again. @bartcode

- Solve linting error. @bartcode

- Use different approach for IP. @bartcode


### Documentation


- Add @fluxbase:namespace annotation to edge functions and jobs guides @claude


### Features


- Add custom MCP tools and resources support @claude

- Add full Fluxbase SDK to custom MCP tools @claude

- Add namespace support for custom MCP tools, edge functions, and jobs @claude

- **dashboard:** Auto-select non-empty namespace when default is empty @claude

- Add AI capabilities to MCP tools, edge functions, and jobs @claude

- **dashboard:** Add MCP tools management page @claude

- **api:** Add MCP tools and resources endpoints to OpenAPI spec @claude


### Miscellaneous


- Update version to 2026.1.14-rc.1 in deployment files @github-actions[bot]


### Refactoring


- Use consistent @fluxbase: annotation prefix for MCP tools @claude

- Remove redundant @fluxbase:mcp-tool/resource type annotations @claude

- Simplify MCP annotations - all now optional @claude

- Use consistent handler signature for MCP tools @claude


### Tests


- Add tests for namespace annotation support @claude


### Security


- **api:** Restrict internal AI endpoints to localhost only @claude

- **api:** Restrict internal AI endpoints to localhost only @claude

- **api:** Restrict internal AI endpoints to localhost only @claude


## v2026.1.14-rc.1 (2026-01-21)


### Miscellaneous


- Update version to 2026.1.13-rc.1 in deployment files @github-actions[bot]


## v2026.1.13-rc.1 (2026-01-20)


### Miscellaneous


- Update version to 2026.1.13 in deployment files @github-actions[bot]


## v2026.1.13 (2026-01-19)


### Bug fixes


- Create execution record with execution logs disabled. @bartcode

- Solve issues with e2e tests. @bartcode

- Improve pool configuration. @bartcode


### Miscellaneous


- Update version to 2026.1.12-rc.4 in deployment files @github-actions[bot]


## v2026.1.12-rc.4 (2026-01-19)


### Bug fixes


- Use .ts files for chatbots. @bartcode


### Miscellaneous


- Update version to 2026.1.12-rc.3 in deployment files @github-actions[bot]


## v2026.1.12-rc.3 (2026-01-19)


### Bug fixes


- Also bundle geojson files in the CLI. @bartcode


### Miscellaneous


- Update version to 2026.1.12-rc.2 in deployment files @github-actions[bot]


## v2026.1.12-rc.2 (2026-01-19)


### Bug fixes


- Add .geojson support to bundler. @bartcode


### Miscellaneous


- Update version to 2026.1.12-rc.1 in deployment files @github-actions[bot]


## v2026.1.12-rc.1 (2026-01-19)


### Bug fixes


- Update test files. @bartcode

- Get rid of linting issue. @bartcode

- Install go-test-coverage. @bartcode

- **auth:** Wire up TOTP encryption using global encryption key @claude

- Apply formatting. @bartcode

- Remove work docs. @bartcode

- Solve linting issues. @bartcode

- Make tests work again. @bartcode

- Add migration for TOTP secret length. @bartcode

- Update TestWebhookTriggerOnUserUpdate. @bartcode

- Get rid of flakiness. @bartcode


### Documentation


- Refer to the docs in the README.md. @bartcode

- Add comprehensive BaaS architecture review @claude

- Add comprehensive implementation plan with progress tracking @claude

- Simplify TOTP encryption fix - use existing global key @claude

- Update IMPLEMENTATION_PLAN.md with Phase 3.1 and 3.3 completion status @claude

- Add webhook signature verification docs and Prefer header tests @claude

- Add backup and restore documentation and scripts (Phase 5.1) @claude


### Features


- **auth:** Add per-user TOTP rate limiting @claude

- **middleware:** Add warning for in-memory rate limiting in multi-instance environments @claude

- Complete Phase 1 - Security & Reliability improvements @claude

- Complete Phase 2 items 2.1-2.2 - Schema batching & LISTEN pooling @claude

- Complete Phase 2 item 2.3 - Async message broadcasting @claude

- Complete Phase 2 - Scalability & Performance improvements @claude

- Begin Phase 3.1 - Standardize error response format @claude

- Continue Phase 3.1 - Migrate DDL and storage handlers to use SendError helpers @claude

- Complete Phase 3.1 - All fiber.Map error responses migrated @claude

- Complete Phase 3.3 - Implement graceful job shutdown @claude

- Complete Phase 3.6 - Per-endpoint body size limits @claude

- Complete Phase 3.4 - Idempotency keys for mutations @claude

- Complete Phase 3.5 - OAuth state persistence for multi-instance @claude

- Complete Phase 3.7 - Service key revocation and rotation @claude

- Complete Phase 4 - Developer Experience improvements @claude

- Add tracing helpers for functions and jobs (Phase 5.2) @claude

- Complete Phase 5 - Operations & Polish @claude

- Add service key revocation and rotation support @claude


### Miscellaneous


- Update version to 2026.1.11 in deployment files @github-actions[bot]

- Update version to 2026.1.11-rc.2 in deployment files @github-actions[bot]


## v2026.1.11 (2026-01-09)


### Bug fixes


- Add debug logs for RLS access checks. @bartcode

- More realtime improvements. @bartcode

- Add missing parameter. @bartcode


### Miscellaneous


- Update version to 2026.1.11-rc.1 in deployment files @github-actions[bot]


## v2026.1.11-rc.1 (2026-01-08)


### Bug fixes


- Don't block service roles. @bartcode

- Add full JWT claim. @bartcode


### Miscellaneous


- Update version to 2026.1.10-rc.1 in deployment files @github-actions[bot]


## v2026.1.10-rc.1 (2026-01-08)


### Bug fixes


- Allow empty user IDs. @bartcode

- Convert some messages to debug messages. @bartcode

- Improve Realtime connection. @bartcode


### Miscellaneous


- Update version to 2026.1.10 in deployment files @github-actions[bot]


## v2026.1.10 (2026-01-08)


### Bug fixes


- Use inputPath for deno. @bartcode

- Add token fixes. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.11 in deployment files @github-actions[bot]


## v2026.1.9-rc.11 (2026-01-08)


### Bug fixes


- Use deno bin differently. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.10 in deployment files @github-actions[bot]


## v2026.1.9-rc.10 (2026-01-08)


### Bug fixes


- Add configurable IP allowlist. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.9 in deployment files @github-actions[bot]


## v2026.1.9-rc.9 (2026-01-07)


### Bug fixes


- Use WASM builder again. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.8 in deployment files @github-actions[bot]


## v2026.1.9-rc.8 (2026-01-07)


### Bug fixes


- Copy deno from container. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.7 in deployment files @github-actions[bot]


## v2026.1.9-rc.7 (2026-01-07)


### Bug fixes


- Use different base image. @bartcode

- Use different base image. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.6 in deployment files @github-actions[bot]


## v2026.1.9-rc.6 (2026-01-07)


### Bug fixes


- Only build esbuild-wasm once. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.5 in deployment files @github-actions[bot]


## v2026.1.9-rc.5 (2026-01-07)


### Bug fixes


- Use esbuild-wasm. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.4 in deployment files @github-actions[bot]


## v2026.1.9-rc.4 (2026-01-07)


### Bug fixes


- Add Deno cache dir. @bartcode

- Add libc6. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.3 in deployment files @github-actions[bot]


## v2026.1.9-rc.3 (2026-01-07)


### Bug fixes


- Update Deno dir correctly. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.2 in deployment files @github-actions[bot]


## v2026.1.9-rc.2 (2026-01-07)


### Bug fixes


- Set DENO_DIR correctly. @bartcode


### Miscellaneous


- Update version to 2026.1.9-rc.1 in deployment files @github-actions[bot]


## v2026.1.9-rc.1 (2026-01-06)


### Bug fixes


- Update FK of secrets. (#14) @bartcode

- Allow for multiple subscriptions in one channel. (#15) @bartcode


### Miscellaneous


- Update version to 2026.1.8 in deployment files @github-actions[bot]


## v2026.1.8 (2026-01-06)


### Bug fixes


- Update peer dependencies of sdk-react. (#13) @bartcode


### Features


- Improve chatbots and apply bugfixes (#12) @bartcode


### Miscellaneous


- Update version to 2026.1.7 in deployment files @github-actions[bot]


## v2026.1.7 (2026-01-05)


### Miscellaneous


- Update version to 2026.1.6 in deployment files @github-actions[bot]


## v2026.1.6 (2026-01-05)


### Features


- Improve chatbots and let them use MCP tools. (#11) @bartcode


### Miscellaneous


- Update version to 2026.1.5 in deployment files @github-actions[bot]


## v2026.1.5 (2026-01-05)


### Miscellaneous


- Update version to 2026.1.4 in deployment files @github-actions[bot]


## v2026.1.4 (2026-01-04)


### Bug fixes


- Resolve CI pipeline failures @bartcode

- Add documentation and finalise unfinished features. @bartcode

- Do more config checks when starting the app. @bartcode

- RLS. UI, and many more changes. @bartcode

- Add more functionality to SDK and add more docs. @bartcode

- Remove template ruins. @bartcode

- Implement pre-commit. @bartcode

- More docs and another security audit. @bartcode

- UI fixes after manual testing. @bartcode

- Resolve final warnings. @bartcode

- Update Docker image for docs. @bartcode

- Install dependencies in docs image. @bartcode

- Add Umami tracker. @bartcode

- **docs:** Escape angle bracket in edge functions guide @bartcode

- Allow for user migrations. @bartcode

- Update Docker Compose file. @bartcode

- Update env vars for Docker compose file. @bartcode

- Let nginx know it's behind a reverse proxy. @bartcode

- Make docs more straight to the point. @bartcode

- Remove deprecated code. @bartcode

- Add metadata field to users table and update tests. @bartcode

- Improve impersonation functionality. @bartcode

- Migrations should now work. @bartcode

- Always create _fluxbase schema. @bartcode

- Lots of UI fixes. @bartcode

- Fix failed tests. @bartcode

- Enable signup in database settings for all e2e tests @bartcode

- Start versioning from 0.1.0. @bartcode

- Add custom settings handler. @bartcode

- Add more app settings methods to SDK. @bartcode

- Reset Chart.yaml changes before gh-pages checkout @bartcode

- Escape template variable syntax in JSDoc for MDX @bartcode

- Always create migrations tables if they don't exist yet. @bartcode

- Clearly separate admin user and runtime user for database. @bartcode

- Implement RLS for storage. @bartcode

- Use correct database during CI pipeline. @bartcode

- Update user registration and sign in. @bartcode

- Add e2e test fixes. @bartcode

- Improve Supabase SDK compatibility. @bartcode

- Docs now build correctly again. @bartcode

- Make Fluxbase more comaptible. @bartcode

- More compatibility fixes. @bartcode

- Use FLUXBASE_BASE_URL everywhere. Do not use FLUBXASE_URL. @bartcode

- Set default user migrations path. @bartcode

- Add support for auth.uid(). @bartcode

- Add additional Supabase SQL functions. @bartcode

- Also add auth.role(). @bartcode

- Better organize all migrations. @bartcode

- Don't forget to create roles. @bartcode

- Fix e2e tests. @bartcode

- Fix e2e tests for new unified settings table. @bartcode

- Fix linting issues. @bartcode

- CSRF didn't work. @bartcode

- Grant CREATEROLE permission in CI pipeline. @bartcode

- Docs and async SDK methods. @bartcode

- Add upload progress to storage. @bartcode

- Ensure RLS tables are available. @bartcode

- Improve further Supabase SDK compatibility for Realtime and Tables. @bartcode

- Improve Edge Functions support. @bartcode

- Again fix more compatibility issues. @bartcode

- Add MFA tables. @bartcode

- Add migrations. @bartcode

- Improve functions compatibility significantly and add limits to table requests. @bartcode

- More doc fixes. @bartcode

- Add unsubscribe to Realtime and allow for chaining in QueryBuilder. @bartcode

- Implement 2FA flow and add clearer logging. @bartcode

- Improve TOTP support. @bartcode

- Reduce logs. @bartcode

- Significantly improve auth setup. @bartcode

- Optimize testing suite. @bartcode

- Do not hardcode fluxbase_app user. @bartcode

- Align signUp with Supabase's SDK. @bartcode

- Add PostGIS support. @bartcode

- Only enable PostGIS if needed. @bartcode

- Add PostGIS write support. @bartcode

- Adjust PostGIS tests. @bartcode

- Add tests and fixes for PostGIS support. @bartcode

- Fall back to in memory storage if localStorage isn't available. @bartcode

- Make auth more Supabase-compatible. @bartcode

- Add jwt_claims for user JWT authentication. @bartcode

- Set up migration and edge function sync. @bartcode

- Fix e2e tests and linting. @bartcode

- Listen to 127.0.0.1. @bartcode

- Add Jobs config. @bartcode

- Make e2e test setup fail fast on database errors @bartcode

- Only create jobs handler when jobs are enabled @bartcode

- Use configured database name instead of hardcoded fluxbase_dev @bartcode

- Grant permissions on jobs schema to test users @bartcode

- Improve sync. @bartcode

- Add job logs. @bartcode

- Many fixes to improve jobs. @bartcode

- Permission issues in Fluxbase dashboard. @bartcode

- Implement signed URLs for local storage. @bartcode

- Migrate docs to use Astro. @bartcode

- Update migrations. @bartcode

- Mermaid graphs. @bartcode

- Merge conflicts. @bartcode

- Upgrade security CI permissions. @bartcode

- Versioning fix. @bartcode

- **ci:** Restore Go 1.25 version @bartcode

- Resolve CI pipeline failures @bartcode

- Update Go dependencies. @bartcode

- Implement chatbots and RPC. @bartcode

- Add http handler for chatbot. @bartcode

- Improve impersonation. @bartcode

- Solve issues with impersonation and auth in general. @bartcode

- Test fixes. @bartcode

- Update tests. @bartcode

- Make unittests run successfully. @bartcode

- Set up TOTP for dashboard. @bartcode

- Active sessions page. @bartcode

- Get rid of linting issues. @bartcode

- Add admin_session_handler.go. @bartcode

- RPC functionality. @bartcode

- Update e2e tests. @bartcode

- Make RPC scheduling work. @bartcode

- Add extension management. @bartcode

- Improve scheduling. @bartcode

- Improve scheduling. @bartcode

- Add CLI, and improve vector search. @bartcode

- Linting issues. @bartcode

- Update CI. @bartcode

- Enable pgvector. @bartcode

- Fix more linting issues. @bartcode

- Linting issues. @bartcode

- Improve logging system. @bartcode

- Make e2e tests work. @bartcode

- Add logs route. @bartcode

- Make execution logs work. @bartcode

- Update Docker build. @bartcode

- Add retry mechanism for e2e tests. @bartcode

- Add missing dependency for tesseract. @bartcode

- Improve Helm chart and Docker Compose deployment. @bartcode

- Improve Helm chart and add Fluxbase config options. @bartcode

- Explicitly enable dashboard in the e2e tests. @bartcode

- Add more tests. @bartcode

- Ignore formatting of test files. @bartcode

- Linting issues. @bartcode

- Add support for Gateway API. @bartcode

- Improve Helm deployment. @bartcode

- Rename properties in Helm chart. @bartcode

- Rename properties in Helm chart. @bartcode

- Update flaky test. @bartcode

- Add routeName to gateway and fix Postgres CI issues. @bartcode

- Make secrets better configurable. @bartcode

- Update Postgres releases. @bartcode

- Postgres versioning. @bartcode

- Improve secret references. @bartcode

- Remove unused configmap. @bartcode

- Config decoding issues. @bartcode

- Implement cache invalidation. @bartcode

- Temporarily install cURL. @bartcode

- Install Deno through apk. @bartcode

- Protect metadata endpoints. @bartcode

- JWT key references. @bartcode

- Fix e2e tests. @bartcode

- Enable admin API endpoints. @bartcode

- Extract more API endpoints. @bartcode

- Improve Helm chart. @bartcode

- More Docker Compose fixes. @bartcode

- Fix version updates. @bartcode

- Improve SMTP configuration. @bartcode

- Refer to env vars. @bartcode

- Improve health check. @bartcode

- Enable AI by default. @bartcode

- Improve Deno execution. @bartcode

- Add client settings. @bartcode

- Add ClientSettingsPolicy this time. @bartcode

- Rename targetRefs to targetRef. @bartcode

- Don't provide a list either. @bartcode

- Add timeout for websockets. @bartcode

- Use public URL in dashboard. @bartcode

- Deno unittest fix. @bartcode

- Use the same encryption key everywhere. @bartcode

- Implement rate limits for edge functions. @bartcode

- Update hash in test. @bartcode

- Implement distance parameter. @bartcode

- Linting issues. @bartcode

- Use non-wildcard CORS config by default. @bartcode

- Add debug logs. @bartcode

- Potentially fix syncing. @bartcode

- CLI improvements. @bartcode

- Better handling of provider details. @bartcode

- Unittest fixes. @bartcode

- Improve AI provider settings. @bartcode

- Configure storage. @bartcode

- Add secrets management and improve functions/jobs. @bartcode

- Also publish CLI in pipeline. @bartcode

- Fix unittests. @bartcode

- Fix unit test and add Homebrew tap. @bartcode

- Add installation instructions for CLI. @bartcode

- Increase CLI version automatically. @bartcode

- Add CLI installation instructions. @bartcode

- Use the right versioning in pipeline. @bartcode

- Use right URL for CLI. @bartcode

- Disable CGO. @bartcode

- Don't require config file to be available. @bartcode

- Logic behind syncing. @bartcode

- Add sync command for all. @bartcode

- Implement unified sync command. @bartcode

- Fix profile issue. @bartcode

- Don't automatically add an email provider. @bartcode

- Use SDK for managing email settings. @bartcode

- Improve SMTP configuration. @bartcode

- Issue with max number of results. @bartcode

- Add custom settings encryption. @bartcode

- Add settings secrets command. @bartcode

- Distinguish between user-level and system-level secrets. @bartcode

- Enforce encryption key. @bartcode

- Improve query parser. @bartcode

- Add values for embeddings. @bartcode

- Add decryption endpoint. @bartcode

- RPC improvements and AI embedding improvements. @bartcode

- Infer provider type based on configured API keys. @bartcode

- Handle email verification. @bartcode

- Ensure the view privileges are correct. @bartcode

- Chatbot sync issue in CLI. @bartcode

- Implement SAML, captcha, GraphQL, image transformations. @bartcode

- Add _health endpoint. @bartcode

- Install govips. @bartcode

- Install vips in GH action. @bartcode

- Show more endpoints on REST page. @bartcode

- Always apply. @bartcode

- Don't rate-limit service key. @bartcode

- SAML improvements. @bartcode

- Improve security. @bartcode

- More fixes. @bartcode

- Add more security features. @bartcode

- Improve docs and configurations. @bartcode

- Improve SSO. @bartcode

- Make OAuth work. @bartcode

- Linting issues. @bartcode

- Update e2e tests. @bartcode

- Add more scopes. @bartcode

- Improve Promethes endpoint and client keys. @bartcode

- Enable metrics by default. @bartcode

- Linting issues. @bartcode

- Solve OAuth e2e test issues. @bartcode

- Improve CLI. @bartcode

- Streaming body. @bartcode

- Don't block comments in SQL. @bartcode

- Don't redirect during OAuth login. @bartcode

- Remove redundant mounts. @bartcode

- Update devcontainer. @bartcode

- Apply formatting. @bartcode

- Improve OAuth setup and fix unittests. @bartcode

- Clarify secrets usage better. @bartcode

- Update LICENSE to become AGPLv3. @bartcode

- Use the right layout. @bartcode

- Update OAuth e2e tests. @bartcode

- Store used OAuth provider in local storage. @bartcode

- Add support for redirectTo when resetting passwords. @bartcode

- Add rate limit to password reset. @bartcode

- Throw error if mail isn't configured. @bartcode

- Allow adding variables into chatbots. @bartcode

- Improve CLI. @bartcode

- Always return JSON for OAuth. @bartcode

- Use S3 for storage in devcontainer. @bartcode

- More OAuth improvements. @bartcode

- Change versioning and fix OAuth. @bartcode

- Use new versioning everywhere. @bartcode

- Update versioning correctly in CI. @bartcode

- Better store redirect URL. @bartcode

- Dashboard OAuth handler. @bartcode

- OAuth now works for dashboard. @bartcode

- Add public endpoints for auth settings. @bartcode

- Provide setting for forcing SSO logins. @bartcode

- Add missing field to AuthConfig. @bartcode

- Make embeddings configurable while running. @bartcode

- Add UI for AI providers. @bartcode

- Improve embeddings implementation. @bartcode

- Add missing migration. @bartcode

- Explicit Deno exit code. @bartcode

- Add SDK methods for interacting with user-settings. @bartcode

- Don't use scientific notation in Helm chart. @bartcode


### CI


- Refactor CI pipeline. @bartcode

- Allow for rc releases. @bartcode

- Fix binary build. @bartcode

- Allow for easy local building of docs Docker image. @bartcode

- Allow for specific releases. @bartcode

- Also make Go module conditional. @bartcode

- Optimize build time for docs. @bartcode

- Allow for rc releases of Helm chart. @bartcode

- Add dependency update. @bartcode

- Create gh-pages branch if it doesn't exist. @bartcode

- Fix AMD64 + ARM64 CI build. @bartcode

- Update runner image. @bartcode

- Also try to build docs image in CI pipeline. @bartcode

- Run e2e tests without wrapper script for debugging @bartcode

- Fetch latest tags when determining version. @bartcode

- Hopefully fix CI. @bartcode

- Update SDK version. @bartcode

- Add bump type. @bartcode

- Add 'continue' release type. @bartcode

- Don't publish Postgres image by default. @bartcode

- Use CalVer in CI dropdown options. @bartcode


### Documentation


- Fix broken links. @bartcode

- Improve documentation. @bartcode

- Update Discord/Twitter links. @bartcode

- Another iteration of more compact docs. @bartcode

- Remove more redundant docs. @bartcode

- Reduce amount of docs further. @bartcode

- Fix broken link. @bartcode

- Improve documentation. @bartcode

- Remove pricing info. @bartcode

- Fix GitHub link. @bartcode

- Add docs about rate limiting. @bartcode

- Improve and correct docs. @bartcode

- Add secret management to docs. @bartcode

- Improve secret management in docs. @bartcode

- Align docs with implementation. @bartcode

- More doc improvements. @bartcode

- Update license. @bartcode


### Features


- Starting point for thorough testing and CI development. @bartcode

- Add SQL editor. @bartcode

- Add Helm chart and Docker Compose with PostgreSQL 18 support @bartcode

- Add Supabase-compatible realtime row-level filtering @bartcode

- Auto-update versions in deployment files on release @bartcode


### Miscellaneous


- Update version to 0.0.1-rc.7 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.8 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.9 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.10 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.11 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.12 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.13 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.14 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.15 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.16 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.17 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.18 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.19 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.20 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.21 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.22 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.23 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.24 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.25 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.26 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.27 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.28 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.30 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.31 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.32 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.33 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.34 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.35 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.36 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.37 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.38 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.39 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.40 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.41 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.42 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.43 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.44 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.45 in deployment files @github-actions[bot]

- Update version to 0.1.0-rc.2 in deployment files @github-actions[bot]

- Update version to 0.1.0-rc.3 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.46 in deployment files @github-actions[bot]

- Change license from GPL-3.0 to Elastic License 2.0 @claude

- Update version to 0.0.1-rc.47 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.48 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.49 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.50 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.51 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.52 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.53 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.54 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.55 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.56 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.57 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.58 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.59 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.60 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.61 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.62 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.63 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.64 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.65 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.66 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.67 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.69 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.70 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.71 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.72 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.73 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.74 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.75 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.76 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.77 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.78 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.79 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.80 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.81 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.82 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.83 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.84 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.85 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.86 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.88 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.89 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.90 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.91 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.92 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.93 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.94 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.95 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.96 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.97 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.98 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.99 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.100 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.101 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.102 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.103 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.104 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.105 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.106 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.107 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.108 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.109 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.110 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.111 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.112 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.113 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.114 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.115 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.116 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.117 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.118 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.119 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.120 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.121 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.122 in deployment files @github-actions[bot]

- Update version to 0.0.1-rc.123 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.2 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.3 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.4 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.5 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.6 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.7 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.8 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.9 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.10 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.11 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.12 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.13 in deployment files @github-actions[bot]

- Update version to 2026.1.1-rc.14 in deployment files @github-actions[bot]

- Update version to 2026.1.2 in deployment files @github-actions[bot]

- Update version to 2026.1.3-rc.1 in deployment files @github-actions[bot]


### Refactoring


- Consolidate test structure and remove redundant standalone tests @bartcode

<!-- generated by git-cliff; do not edit release sections by hand -->
