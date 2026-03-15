# Contribution Rules

## Merge Request Rule

Every MR must satisfy these rules.

1. Runtime or deploy behavior changes require regression tests in the same MR.
2. User-facing CLI, API, lifecycle, install, or deploy changes require README/docs updates in the same MR.
3. Wrapper, entrypoint, deploy, install, and environment default changes must include a direct regression path.
4. A passing generic test suite is not enough when the changed path is narrower than the suite. Add a focused test for the path you changed.

## Minimum Expectations

- `internal/`, `cmd/`, `cli/`, `cli-src/`, `plugins/`, `install.sh`, `scripts/deploy*`, `.gitlab-ci.yml`
  These are treated as guarded paths.
- If a guarded path changes, the MR must also change at least one test file.
- If a guarded path changes and it affects operator behavior, the MR must also update docs.

## Review Standard

Review should reject MRs that change behavior without proving the changed path.
