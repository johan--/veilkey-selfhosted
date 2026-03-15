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
- `rulebook/`, `scripts/validate_rulebook.py`
  These are also guarded paths because they change central policy behavior.
- If a guarded path changes, the MR must also change at least one test file.
- If a guarded path changes and it affects operator behavior, the MR must also update docs.

## Examples

- acceptable
  - change `rulebook/veilkey-rulebook.toml` + add `tests/test_rulebook_format.sh` update + add README/docs note
- unacceptable
  - change `rulebook/veilkey-rulebook.toml` only
  - change `scripts/validate_rulebook.py` without any test update

## Review Standard

Review should reject MRs that change behavior without proving the changed path.

## Central Rulebook

- Central rulebook lives at `rulebook/veilkey-rulebook.toml`.
- MR guard now consumes the rulebook through `scripts/check_mr_guard.py`.
- `scripts/check-mr-guard.sh` remains the stable shell entrypoint for CI and local use.
- Validate the current format with:

```bash
python3 scripts/validate_rulebook.py --rulebook rulebook/veilkey-rulebook.toml
bash tests/test_rulebook_format.sh
bash tests/test_mr_guard.sh
```
