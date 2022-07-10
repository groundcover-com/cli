#!/usr/bin/env bash

# Copyright The groundcover Authors.
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

: "${GITHUB_REPO:="cli"}"
: "${GITHUB_OWNER:="groundcover-com"}"
: "${BINARY_NAME:="groundcover"}"
: "${INSTALL_DIR:="${HOME}/.groundcover/bin"}"

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
  OS=$(uname |tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*|cygwin*) OS='windows';;
  esac
}

# initLatestTag discovers latest version on GitHub releases.
initLatestTag() {
  local latest_release_url="https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/releases/latest"
  LATEST_TAG=$(curl -Ls ${latest_release_url} | awk -F\" '/tag_name/{print $(NF-1)}')
}

# appendShellPath append our install bin directory to PATH on bash, zsh and fish shells
appendShellPath() {
  local bashrc_file="${HOME}/.bashrc"
  if [ -f "${bashrc_file}" ]; then
    local export_path_expression="export PATH=${INSTALL_DIR}:\${PATH}"
    if ! grep -q "${export_path_expression}" "${bashrc_file}"; then
      echo "${export_path_expression}" >> "${bashrc_file}"
      echo "Added ${INSTALL_DIR} to \$PATH in ${bashrc_file}"
    fi    
  fi

  local zshrc_file="${HOME}/.zshrc"
  if [ -f "${zshrc_file}" ]; then
    local export_path_expression="export PATH=${INSTALL_DIR}:\${PATH}"
    if ! grep -q "${export_path_expression}" "${zshrc_file}"; then
      echo "${export_path_expression}" >> "${zshrc_file}"
      echo "Added ${INSTALL_DIR} to \$PATH in ${zshrc_file}"
    fi
  fi

  local fish_config_file="${HOME}/.config/fish/config.fish"
  if [ -f "${fish_config_file}" ]; then
    local export_path_expression="set -U fish_user_paths ${INSTALL_DIR} \$fish_user_paths"
    if ! grep -q "${export_path_expression}" "${fish_config_file}"; then
      echo "${export_path_expression}" >> "${fish_config_file}"
      echo "Added ${INSTALL_DIR} to \$PATH in ${fish_config_file}"
    fi
  fi

  exec "${SHELL}" # Reload shell
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

# checkInstalledVersion checks which version of cli is installed and
# if it needs to be changed.
checkInstalledVersion() {
  if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    local version
    version=$("${INSTALL_DIR}/${BINARY_NAME}" --skip-selfupdate version)
    if [ "${version}" = "${LATEST_TAG#v}" ]; then
      echo "groundcover ${version} is already latest"
      return 0
    else
      echo "groundcover ${LATEST_TAG} is available. Updating from version ${version}."
      return 1
    fi
  else
    return 1
  fi
}

# downloadFile downloads the latest binary package.
downloadFile() {
  ARCHIVE_NAME="${BINARY_NAME}_${LATEST_TAG#v}_${OS}_${ARCH}.tar.gz"
  DOWNLOAD_URL="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/${LATEST_TAG}/${ARCHIVE_NAME}"
  TMP_ROOT="$(mktemp -dt groundcover-installer-XXXXXX)"
  ARCHIVE_TMP_PATH="${TMP_ROOT}/${ARCHIVE_NAME}"
  echo "Downloading ${DOWNLOAD_URL}"
  curl -SsL "${DOWNLOAD_URL}" -o "${ARCHIVE_TMP_PATH}"
}

# installFile installs the cli binary.
installFile() {
  tar xf "${ARCHIVE_TMP_PATH}" -C "${TMP_ROOT}"
  BIN_PATH="${INSTALL_DIR}/${BINARY_NAME}"
  BIN_TMP_PATH="${TMP_ROOT}/${BINARY_NAME}"
  echo "Preparing to install ${BINARY_NAME} into ${INSTALL_DIR}"
  mkdir -p "${INSTALL_DIR}"
  cp "${BIN_TMP_PATH}" "${BIN_PATH}"
  chmod +x "${BIN_PATH}"
  echo "${BINARY_NAME} installed into ${BIN_PATH}"
}

# testVersion tests the installed client to make sure it is working.
testVersion() {
  set +e
  command -v "${BINARY_NAME}" >/dev/null 2>&1
  if [ "$?" = "1" ]; then
    echo "${BINARY_NAME} not found. Is ${INSTALL_DIR} on your PATH"
    exit 1
  fi
  set -e
}


# cleanup temporary files
cleanup() {
  if [ -d "${TMP_ROOT:-}" ]; then
    rm -rf "${TMP_ROOT}"
  fi
}

# fail_trap is executed if an error occurs.
fail_trap() {
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to install ${BINARY_NAME}"
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
appendShellPath
testVersion
cleanup
