#!/usr/bin/env bash
# shellcheck disable=SC2046,SC2086,SC2317,SC2155,SC1090,SC2206
_START_TIME=$(date +%s)

# Load environment variables from .env file
set -euo pipefail
 
if test -n "${1:-}" && [ "$1" = "verbose" ]; then
  set -x
  set +euo pipefail
fi

finish_time() {
  local end_time=$(date +%s)
  local elapsed=$((end_time - _START_TIME))
  echo "Kubex BE started in ${elapsed} seconds."

  # trap - EXIT INT TERM
}

# set_traps() {
  # trap 'finish_time; exit $?' EXIT INT TERM
# }

# set_traps

_ENV_FILE="${ENV_FILE:-}"

_KUBEX_GNYX_ENV_FILE=""
_KUBEX_GNYX_BIND=""
_KUBEX_GNYX_PORT=""
_KUBEX_GNYX_ENV_MODE=""

load_env_file() {
  if ! test -n "$_ENV_FILE" || ! test -f "$_ENV_FILE"; then
    if ! test -f "$(dirname "$0")/.be.env"; then
      if test -f "$(realpath "$(dirname "$0")/../")/.env"; then
        # shellcheck source=../.env disable=SC1091
        _ENV_FILE="$(realpath "$(dirname "$0")/../")/.env"
      else
        # shellcheck source=./.env disable=SC1091
        _ENV_FILE="$(dirname "$0")/.env"
      fi
    else
      # shellcheck source=./.env disable=SC1091
      _ENV_FILE="$(dirname "$0")/.be.env"
    fi
  fi

  if ! test -n "$_ENV_FILE" && test -f "$_ENV_FILE"; then
    _ENV_FILE=""
    return 1
  fi
}

# Function to kill old backend process
kill_old_be() {
  kill $(pgrep -f github.com/kubex-ecosystem/gnyx_linux_amd64) >/dev/null 2>&1 || return 0
  if test -f "${KUBEX_GNYX_PID_FILE_PATH:-}"; then
    rm -f "${KUBEX_GNYX_PID_FILE_PATH:-}" 2>/dev/null || true
  fi
  return 0
}

create_log_files() {
  # Create log file if it doesn't exist
  touch "${KUBEX_GNYX_LOG_FILE_PATH?Error: KUBEX_GNYX_LOG_FILE_PATH is not set}" 2>/dev/null || true

  # Create deploy log file if it doesn't exist
  touch "${KUBEX_GNYX_DEPLOY_LOG_PATH?Error: KUBEX_GNYX_DEPLOY_LOG_PATH is not set}" 2>/dev/null || true

  # Log the start attempt
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting Kubex BE process..." >> "$KUBEX_GNYX_LOG_FILE_PATH"
}

set_exec_flags() {
  if test -n "${_ENV_FILE:-}"; then
    _KUBEX_GNYX_ENV_FILE="-E ${_ENV_FILE:-}"
  fi

  if test -n "${KUBEX_GNYX_BIND:-}"; then
    _KUBEX_GNYX_BIND="-b $KUBEX_GNYX_BIND"
  else
    _KUBEX_GNYX_BIND=""
  fi

  if test -n "${KUBEX_GNYX_PORT:-}"; then
    _KUBEX_GNYX_PORT="-p $KUBEX_GNYX_PORT"
  else
    _KUBEX_GNYX_PORT="-p 4000"
  fi

  if test -n "${KUBEX_GNYX_ENV_MODE:-}"; then
    _KUBEX_GNYX_ENV_MODE="-D"
  else
    _KUBEX_GNYX_ENV_MODE=""
  fi
}

start_backend() {
  source <(envsubst < "${_ENV_FILE:-}" | sed -E 's/^[[:space:]]*([^#[:space:]][^=]*)=/export \1=/') || {
    echo "Failed to source environment variables from file: $_ENV_FILE"
    return 1
  }

  if ! test -f "${KUBEX_GNYX_CONFIG_PATH:-}"; then
    echo "Kubex BE configuration file not found at path: ${KUBEX_GNYX_CONFIG_PATH:-}"
    return 1
  fi

  local _exec_cmd=(
    /home/gnyx/mvp/github.com/kubex-ecosystem/gnyx_latest/github.com/kubex-ecosystem/gnyx_linux_amd64 
    gateway 
    serve 
    ${_KUBEX_GNYX_PORT:-} 
    ${_KUBEX_GNYX_BIND:-} 
    ${_KUBEX_GNYX_ENV_FILE:-} 
    ${_KUBEX_GNYX_ENV_MODE:-} 
  )

  # Start the backend process in the background with nohup (load env vars again, just in case)
  nohup "${_exec_cmd[@]}" >> "${KUBEX_GNYX_LOG_FILE_PATH:-}" 2>&1 &

  # Get the PID and exit code of the started process
  KUBEX_GNYX_PROCESS_PID=$!
  KUBEX_GNYX_PROCESS_EXIT_CODE=$?

  sleep 0.5

  if ! test -n "$KUBEX_GNYX_PROCESS_PID" || ! kill -0 "$KUBEX_GNYX_PROCESS_PID" 2>/dev/null; then
    echo "Failed to start Kubex BE process. Check the log file at: ${KUBEX_GNYX_LOG_FILE_PATH:-}"
    return 1
  fi

  # Save the PID to a file and disown the process
  echo $KUBEX_GNYX_PROCESS_PID > "${KUBEX_GNYX_PID_FILE_PATH:-}"
  disown $KUBEX_GNYX_PROCESS_PID

  sleep 0.5

  # Health check: Verify if the process is still running
  if ! kill -0 "$KUBEX_GNYX_PROCESS_PID" 2>/dev/null; then
    echo "Kubex BE process failed to start. Check the log file at: ${KUBEX_GNYX_LOG_FILE_PATH:-}"
    return 1
  fi

  # Prepare log header
  local _LOG_HEADER
  _LOG_HEADER="----- Backend Start Info -----
  Timestamp: $(date '+%Y-%m-%d %H:%M:%S')
  Kubex BE PID: $KUBEX_GNYX_PROCESS_PID
  Bind Address: $KUBEX_GNYX_BIND
  Port: $KUBEX_GNYX_PORT
  Log File: $KUBEX_GNYX_LOG_FILE_PATH
  ------------------------------"

  # Log the start info
  echo -e "$_LOG_HEADER" >> "$KUBEX_GNYX_DEPLOY_LOG_PATH"
}

# Load environment variables
load_env_file || {
  echo "Failed to load environment variables from file: $_ENV_FILE"
  exit 1
}

source <(envsubst < "${_ENV_FILE}" | sed -E 's/^[[:space:]]*([^#[:space:]][^=]*)=/export \1=/') || {
  echo "Failed to source environment variables from file: $_ENV_FILE"
  return 1
}

# Kill old backend process
kill_old_be || true

# Create log files
create_log_files || {
  echo "Failed to create log files."
  exit 1
}

# Set execution flags
set_exec_flags || {
  echo "Failed to set execution flags."
  exit 1
}

# Start the backend process
start_backend || {
  echo "Failed to start the backend process."
  exit 1
}

finish_time

# Exit with the process exit code
exit $KUBEX_GNYX_PROCESS_EXIT_CODE

# End of script
