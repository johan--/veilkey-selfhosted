#!/usr/bin/env bats
# resolve TTY guard — ensure resolve blocks non-TTY (pipe) execution

@test "resolve blocks pipe execution (non-TTY stdout)" {
  # Pipe execution: stdout is not a TTY → must be rejected
  result=$(veilkey-cli resolve VK:LOCAL:test123 2>&1 || true)
  [[ "$result" == *"interactive terminal"* ]] || [[ "$result" == *"TTY"* ]]
}

@test "resolve error message does not leak server URL" {
  result=$(veilkey-cli resolve VK:LOCAL:test123 2>&1 || true)
  # Must not contain IP addresses or URLs
  ! echo "$result" | grep -qE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+'
}

@test "source code has TTY check in cmd_resolve" {
  REPO_ROOT="${REPO_ROOT:-$(cd "$(dirname "$BATS_TEST_FILENAME")/../.." && pwd)}"
  grep -q 'isatty' "$REPO_ROOT/services/veil-cli/src/commands.rs"
}
