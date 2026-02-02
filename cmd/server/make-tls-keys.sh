#!/usr/bin/env bash
set -euo pipefail

print_usage() {
  cat <<'EOF'
Usage: make-tls-keys.sh [--prefix PATH] [--name NAME] [--subject SUBJECT] [--force]

Defaults:
  --prefix  ~/.config/poke
  --name    server
  --subject /CN=localhost
EOF
}

fail() {
  printf 'error: %s\n' "$1" >&2
  exit 1
}

prefix="${HOME}/.config/poke"
name="server"
subject="/CN=localhost"
force="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prefix)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --prefix"
      prefix="$1"
      ;;
    --name)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --name"
      name="$1"
      ;;
    --subject)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --subject"
      subject="$1"
      ;;
    --force)
      force="true"
      ;;
    -h|--help)
      print_usage
      exit 0
      ;;
    *)
      print_usage >&2
      fail "unknown argument: $1"
      ;;
  esac
  shift
done

if ! command -v openssl >/dev/null 2>&1; then
  fail "openssl not found in PATH"
fi
if ! openssl version >/dev/null 2>&1; then
  fail "openssl is installed but not functional"
fi

if [[ -z "${prefix}" ]]; then
  fail "prefix must not be empty"
fi
if [[ -z "${name}" ]]; then
  fail "name must not be empty"
fi
if [[ -z "${subject}" ]]; then
  fail "subject must not be empty"
fi

# shellcheck disable=SC2088
case "${prefix}" in
  '~'|'~/'*)
    prefix="${HOME}${prefix#\~}"
    ;;
esac

key_path="${prefix}/${name}.key"
csr_path="${prefix}/${name}.csr"
crt_path="${prefix}/${name}.crt"

existing=()
if [[ -e "${key_path}" ]]; then
  existing+=("${key_path}")
fi
if [[ -e "${csr_path}" ]]; then
  existing+=("${csr_path}")
fi
if [[ -e "${crt_path}" ]]; then
  existing+=("${crt_path}")
fi

if [[ ${#existing[@]} -gt 0 ]] && [[ "${force}" != "true" ]]; then
  printf 'error: output file(s) already exist:\n' >&2
  for path in "${existing[@]}"; do
    printf '  - %s\n' "${path}" >&2
  done
  printf 're-run with --force to overwrite\n' >&2
  exit 1
fi

mkdir -p "${prefix}"
chmod 700 "${prefix}"

if [[ "${force}" == "true" ]]; then
  rm -f "${key_path}" "${csr_path}" "${crt_path}"
fi

# Generate a self-signed certificate for local TLS usage.
openssl genrsa -out "${key_path}" 2048
openssl req -new -key "${key_path}" -out "${csr_path}" -subj "${subject}"
openssl x509 -req -in "${csr_path}" -signkey "${key_path}" -out "${crt_path}" -days 365 -sha256

chmod 600 "${key_path}"
chmod 644 "${csr_path}" "${crt_path}"

printf 'wrote key: %s\n' "${key_path}"
printf 'wrote csr: %s\n' "${csr_path}"
printf 'wrote crt: %s\n' "${crt_path}"
