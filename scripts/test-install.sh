#!/bin/sh
# test-install.sh — valida install.sh contra una "release" local simulada (spec 002, T012).
#
# Cubre dos casos frente al contrato de instalación:
#   (a) caso feliz  → checksum correcto ⇒ instala y el binario responde.
#   (b) caso tamper → checksum manipulado ⇒ install.sh aborta con código != 0 y NO deja
#                     ningún binario en PREFIX (SC-005).
# Sale != 0 si algún aserto falla; 0 si ambos casos pasan.
set -eu

here="$(CDPATH='' cd "$(dirname "$0")" && pwd)"
root="$(CDPATH='' cd "$here/.." && pwd)"
install_sh="$root/install.sh"

[ -f "$install_sh" ] || {
	printf 'test-install: no encuentro %s\n' "$install_sh" >&2
	exit 1
}

pass() { printf 'PASS: %s\n' "$*"; }
fail() {
	printf 'FAIL: %s\n' "$*" >&2
	exit 1
}

sha256_of() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

# Detectar SO/arch igual que install.sh para nombrar el artefacto de forma coherente.
os="$(uname -s)"
case "$os" in
	Darwin) os="darwin" ;;
	Linux) os="linux" ;;
	*) fail "SO de test no soportado: $os" ;;
esac
arch="$(uname -m)"
case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*) fail "arquitectura de test no soportada: $arch" ;;
esac

ver="9.9.9-test"
stage="$(mktemp -d)"
work="$(mktemp -d)"
trap 'rm -rf "$stage" "$work"' EXIT INT TERM

# --- Construir la release local simulada ------------------------------------
# Binario falso: un script que imprime la versión (basta para probar instalación).
mkdir -p "$stage/pkg"
cat >"$stage/pkg/permea" <<EOF
#!/bin/sh
echo "$ver"
EOF
chmod +x "$stage/pkg/permea"

archive="permea_${ver}_${os}_${arch}.tar.gz"
checksums="permea_${ver}_checksums.txt"
( cd "$stage/pkg" && tar -czf "$stage/$archive" permea )

good_sum="$(sha256_of "$stage/$archive")"
printf '%s  %s\n' "$good_sum" "$archive" >"$stage/$checksums"

# --- Caso (a): feliz --------------------------------------------------------
prefix_ok="$work/ok/bin"
PERMEA_BASE_URL="file://$stage" PERMEA_VERSION="$ver" PREFIX="$prefix_ok" \
	sh "$install_sh" >/dev/null 2>&1 || fail "caso feliz: install.sh no debería fallar"

[ -x "$prefix_ok/permea" ] || fail "caso feliz: no se instaló el binario en $prefix_ok"
out="$("$prefix_ok/permea" --version 2>/dev/null || "$prefix_ok/permea")"
[ "$out" = "$ver" ] || fail "caso feliz: el binario instalado no responde la versión ($out)"
pass "caso feliz: instala y el binario responde $ver"

# --- Caso (b): tamper (checksum manipulado) ---------------------------------
prefix_bad="$work/bad/bin"
# Checksum con formato válido (64 hex) pero valor equivocado -> debe abortar.
printf '%s  %s\n' "0000000000000000000000000000000000000000000000000000000000000000" "$archive" \
	>"$stage/$checksums"

if PERMEA_BASE_URL="file://$stage" PERMEA_VERSION="$ver" PREFIX="$prefix_bad" \
	sh "$install_sh" >/dev/null 2>&1; then
	fail "caso tamper: install.sh debería haber abortado con código != 0"
fi
if [ -e "$prefix_bad/permea" ]; then
	fail "caso tamper: NO debe quedar binario instalado tras un checksum manipulado"
fi
pass "caso tamper: aborta sin dejar binario (SC-005)"

printf 'OK: ambos casos pasan\n'
