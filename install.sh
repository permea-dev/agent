#!/bin/sh
# install.sh — instalador de `permea` para macOS y Linux (spec 002).
#
# Es el canal PRINCIPAL de instalación en Linux (el cask de Homebrew es solo macOS).
# Detecta SO/arquitectura, descarga el artefacto correcto de la release, VERIFICA su
# SHA256 ANTES de extraer y coloca el binario en PREFIX. Aborta sin instalar nada ante
# SO/arch no soportado, fallo de descarga o checksum que no coincide (SC-005).
#
# Uso:
#   curl -fsSL https://raw.githubusercontent.com/bfgnet/agente_permea/main/install.sh | sh
#   PERMEA_VERSION=v1.4.0 PREFIX="$HOME/.local/bin" sh install.sh
#
# Variables:
#   PERMEA_VERSION   etiqueta a instalar (vX.Y.Z); por defecto, la última release.
#   PREFIX           destino del binario; por defecto /usr/local/bin (o ~/.local/bin
#                    si no hay permisos de escritura).
#   PERMEA_BASE_URL  override de la URL base de descarga (para pruebas locales);
#                    requiere PERMEA_VERSION. Los ficheros se buscan en
#                    "$PERMEA_BASE_URL/<artefacto>" y "$PERMEA_BASE_URL/<checksums>".
set -eu

REPO="bfgnet/agente_permea"

err() { printf 'install.sh: %s\n' "$*" >&2; }
info() { printf 'install.sh: %s\n' "$*" >&2; }

have() { command -v "$1" >/dev/null 2>&1; }

download() { # <url> <dest>
	if have curl; then
		curl -fsSL "$1" -o "$2"
	elif have wget; then
		wget -qO "$2" "$1"
	else
		err "necesito 'curl' o 'wget' para descargar"
		exit 1
	fi
}

sha256_of() { # <file> -> imprime el hash hex
	if have sha256sum; then
		sha256sum "$1" | awk '{print $1}'
	elif have shasum; then
		shasum -a 256 "$1" | awk '{print $1}'
	else
		err "necesito 'sha256sum' o 'shasum' para verificar la integridad"
		exit 1
	fi
}

latest_tag() {
	# Sigue el redirect de /releases/latest y extrae la etiqueta del destino.
	if ! have curl; then
		err "sin 'curl' no puedo resolver la última versión; define PERMEA_VERSION"
		exit 1
	fi
	curl -fsSLI -o /dev/null -w '%{url_effective}' \
		"https://github.com/${REPO}/releases/latest" | sed -e 's#.*/tag/##'
}

# --- 1. Detección de SO y arquitectura --------------------------------------
os="$(uname -s)"
case "$os" in
	Darwin) os="darwin" ;;
	Linux) os="linux" ;;
	*)
		err "sistema operativo no soportado: '$os' (en Windows usa Scoop)"
		exit 1
		;;
esac

arch="$(uname -m)"
case "$arch" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*)
		err "arquitectura no soportada: '$arch'"
		exit 1
		;;
esac

# --- 2. Resolución de versión y URLs ----------------------------------------
if [ -n "${PERMEA_VERSION:-}" ]; then
	tag="$PERMEA_VERSION"
	case "$tag" in
		v*) : ;;
		*) tag="v$tag" ;;
	esac
elif [ -n "${PERMEA_BASE_URL:-}" ]; then
	err "define PERMEA_VERSION cuando uses PERMEA_BASE_URL"
	exit 1
else
	tag="$(latest_tag)"
fi

if [ -z "$tag" ]; then
	err "no pude determinar la versión a instalar"
	exit 1
fi

num="${tag#v}" # versión sin el prefijo 'v' (contrato de nombres)
archive="permea_${num}_${os}_${arch}.tar.gz"
checksums="permea_${num}_checksums.txt"

if [ -n "${PERMEA_BASE_URL:-}" ]; then
	base="$PERMEA_BASE_URL"
else
	base="https://github.com/${REPO}/releases/download/${tag}"
fi

# --- 3. Selección de PREFIX --------------------------------------------------
if [ -z "${PREFIX:-}" ]; then
	if [ -w /usr/local/bin ]; then
		PREFIX="/usr/local/bin"
	else
		PREFIX="${HOME}/.local/bin"
	fi
fi

# --- 4. Descarga a un directorio temporal (limpiado al salir) ----------------
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

info "descargando permea ${num} (${os}/${arch})…"
download "${base}/${archive}" "${tmp}/${archive}"
download "${base}/${checksums}" "${tmp}/${checksums}"

# --- 5. Verificación de integridad ANTES de extraer (SC-005) -----------------
expected="$(awk -v f="$archive" '$2 == f {print $1}' "${tmp}/${checksums}")"
if [ -z "$expected" ]; then
	err "el fichero de checksums no contiene una entrada para ${archive}; abortando"
	exit 1
fi
actual="$(sha256_of "${tmp}/${archive}")"
if [ "$expected" != "$actual" ]; then
	err "CHECKSUM NO COINCIDE para ${archive}"
	err "  esperado: ${expected}"
	err "  obtenido: ${actual}"
	err "abortando sin instalar (posible manipulación)"
	exit 1
fi
info "checksum verificado."

# --- 6. Extracción e instalación --------------------------------------------
tar -xzf "${tmp}/${archive}" -C "$tmp"
if [ ! -f "${tmp}/permea" ]; then
	err "no se encontró el binario 'permea' dentro del archivo"
	exit 1
fi

mkdir -p "$PREFIX"
cp "${tmp}/permea" "${PREFIX}/permea"
chmod +x "${PREFIX}/permea"

info "instalado en ${PREFIX}/permea"
case ":${PATH}:" in
	*":${PREFIX}:"*) : ;;
	*) info "nota: ${PREFIX} no está en tu PATH; añádelo para usar 'permea' directamente" ;;
esac
info "verifica con: permea --version"
