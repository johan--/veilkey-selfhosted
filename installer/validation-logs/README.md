# Validation Logs

이 폴더는 설치/검증 객체 단위로 기록한다.

규칙:
- `validated/`: 실제로 검증을 끝낸 객체
- `pending/`: 아직 남아 있는 객체
- 각 파일은 `.log`
- 각 파일은 아래 필드를 기본으로 가진다.
  - `status`
  - `date`
  - `object`
  - `repo`
  - `command` 또는 `scope`
  - `target`
  - `expected`
  - `observed` 또는 `pending_items`

현재 이 레포에서 남아 있는 pending은 `pending/` 아래 객체 파일만 보면 된다.

