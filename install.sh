#!/usr/bin/env bash

# Copyright The Groundcover Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


: ${BINARY_NAME:="groundcover"}
: ${INSTALL_DIR:="/usr/local/bin"}

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="arm";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='windows';;
  esac
}

# initLatestTag discovers latest version on GitHub releases.
initLatestTag() {
  local latest_release_url="https://api.github.com/repos/groundcover-com/cli/releases/latest"
  LATEST_TAG=$(curl -Ls $latest_release_url | awk -F\" '/tag_name/{print $(NF-1)}')
}

# runs the given command as root (detects if we are root already)
runAsRoot() {
  if [ $EUID -ne 0 -a "$USE_SUDO" = "true" ]; then
    sudo "${@}"
  else
    "${@}"
  fi
}

# verifySupported checks that the os/arch combination is supported for
# binary builds, as well whether or not necessary tools are present.
verifySupported() {
  local supported="darwin-amd64\ndarwin-arm64\nlinux-amd64\nlinux-arm64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    exit 1
  fi
}

# checkHelmInstalledVersion checks which version of cli is installed and
# if it needs to be changed.
checkInstalledVersion() {
  if [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
    local version=$("${INSTALL_DIR}/${BINARY_NAME}" version)
    if [[ "$version" == "${LATEST_TAG#v}" ]]; then
      echo "Groundcover ${version} is already latest"
      return 0
    else
      echo "Groundcover ${LATEST_TAG} is available. Changing from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package.
downloadFile() {
  ARCHIVE_NAME="groundcover_${LATEST_TAG#v}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/groundcover-com/cli/releases/download/${LATEST_TAG}/${ARCHIVE_NAME}"
  TMP_ROOT="$(mktemp -dt groundcover-installer-XXXXXX)"
  ARCHIVE_TMP_PATH="$TMP_ROOT/$ARCHIVE_NAME"
  curl -SsL "$DOWNLOAD_URL" -o "$ARCHIVE_TMP_PATH"
}

# installFile installs the cli binary.
installFile() {
  tar xf "$ARCHIVE_TMP_PATH" -C "$TMP_ROOT"
  BIN_TMP_PATH="$TMP_ROOT/$BINARY_NAME"
  echo "Preparing to install $BINARY_NAME into ${INSTALL_DIR}"
  runAsRoot cp "$BIN_TMP_PATH" "$INSTALL_DIR/$BINARY_NAME"
  echo "$BINARY_NAME installed into $INSTALL_DIR/$BINARY_NAME"
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  CLI="$(command -v $BINARY_NAME)"
  if [ "$?" = "1" ]; then
    echo "$BINARY_NAME not found. Is $INSTALL_DIR on your "'$PATH?'
    exit 1
  fi
  set -e
}


# cleanup temporary files
cleanup() {
  if [[ -d "${TMP_ROOT:-}" ]]; then
    rm -rf "$TMP_ROOT"
  fi
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to install $BINARY_NAME"
    echo -e "\tFor support, go to https://github.com/groundcover-com/cli."
  fi
  cleanup
  exit $result
}

# Execution

#Stop execution on any error
trap "fail_trap" EXIT
set -e

initArch
initOS
initLatestTag
if ! checkInstalledVersion; then
  downloadFile
  installFile
fi
testVersion
cleanup
