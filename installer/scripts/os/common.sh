#!/usr/bin/env bash
set -euo pipefail

veilkey_detect_os_family() {
  if [[ -n "${VEILKEY_INSTALLER_OS_FAMILY:-}" ]]; then
    printf '%s\n' "${VEILKEY_INSTALLER_OS_FAMILY}"
    return 0
  fi

  if [[ -f /etc/os-release ]]; then
    # shellcheck disable=SC1091
    . /etc/os-release
    case "${ID_LIKE:-$ID}" in
      *debian*|ubuntu|debian)
        printf 'debian\n'
        return 0
        ;;
      *rhel*|*fedora*|centos|rocky|almalinux)
        printf 'rhel\n'
        return 0
        ;;
    esac
  fi

  case "$(uname -s)" in
    Darwin) printf 'darwin\n' ;;
    *) printf 'debian\n' ;;
  esac
}

veilkey_os_prepare_layout() {
  local root="${1:-/}"
  local create_dirs="${2:-1}"
  VEILKEY_OS_SERVICE_DIR="${root%/}${VEILKEY_OS_SERVICE_DIR_SUFFIX}"
  VEILKEY_OS_PROFILE_DIR="${root%/}${VEILKEY_OS_PROFILE_DIR_SUFFIX}"
  VEILKEY_OS_BIN_DIR="${root%/}${VEILKEY_OS_BIN_DIR_SUFFIX}"
  if [[ "${create_dirs}" = "1" ]]; then
    mkdir -p \
      "${VEILKEY_OS_SERVICE_DIR}" \
      "${VEILKEY_OS_PROFILE_DIR}" \
      "${VEILKEY_OS_BIN_DIR}" \
      "${root%/}/opt/veilkey"
  fi
}

veilkey_os_finalize_install() {
  local root="${1:-/}"
  cat > "${root%/}/opt/veilkey/installer/os-layout.env" <<EOF
VEILKEY_OS_FAMILY=${veilkey_os_family}
VEILKEY_OS_SERVICE_DIR=${VEILKEY_OS_SERVICE_DIR}
VEILKEY_OS_PROFILE_DIR=${VEILKEY_OS_PROFILE_DIR}
VEILKEY_OS_BIN_DIR=${VEILKEY_OS_BIN_DIR}
EOF
}
