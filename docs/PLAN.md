> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan: goflare-demo — Fix CI Build Failure

## Root Cause

`goflare build` exits with code 1 in 0 seconds — before TinyGo is invoked.

`goflare.RunBuild` calls `cfg.Validate()` which requires `ProjectName` and `AccountID`.
Both are deploy-only fields (used exclusively in Cloudflare API calls in `cloudflare.go`).
The `build` pipeline never reads them. Requiring them for `build` is a bug in
`tinywasm/goflare` — tracked in `goflare/docs/PLAN.md`.

The CI `Build` step has no Cloudflare env vars (correctly — build should not need them):

```yaml
- name: Build
  run: goflare build   # fails: "AccountID is required" before TinyGo runs

- name: Deploy
  env:
    CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
    CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
  run: goflare deploy
```

## Required env vars for this project

The only secrets/vars this project needs are:

| Name | Type | Used by |
|---|---|---|
| `CLOUDFLARE_API_TOKEN` | secret | `goflare deploy` |
| `CLOUDFLARE_ACCOUNT_ID` | secret | `goflare deploy` |
| `D1_DATABASE_ID` | var | e2e tests |

`PROJECT_NAME` is NOT required — it is only needed by `goflare deploy` internally,
and `goflare` can derive it or it can be set in `.env`. It is not a CI concern.

## Fix

The fix belongs in `tinywasm/goflare` (see `goflare/docs/PLAN.md`):
split `Validate()` into `ValidateBuild()` (no credentials) and `ValidateDeploy()`
(requires `AccountID` + `ProjectName`). After goflare publishes v0.2.17+, update
`go.mod` to the new version — no workflow changes needed in this repo.

## Verification

After updating goflare dependency:
- CI `Build` step completes without credentials (TinyGo compiles WASM).
- CI `Deploy` step runs and pages deployment completes.
- `goflare build` works locally without `CLOUDFLARE_ACCOUNT_ID` set.
- `goflare deploy` still fails with a clear error when `AccountID` is missing.
