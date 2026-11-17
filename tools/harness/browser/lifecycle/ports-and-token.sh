#!/usr/bin/env bash
# shellcheck shell=bash
# shellcheck disable=SC2154

generate_test_route_token() {
  dd if=/dev/urandom bs=32 count=1 2>/dev/null | base64 | tr '+/' '-_' | tr -d '=\n'
}

prepare_test_route_token() {
  TEST_ROUTE_TOKEN="$(generate_test_route_token)"
  if [[ -z "${TEST_ROUTE_TOKEN}" ]]; then
    echo "failed to generate test route token" >&2
    return 1
  fi
  local previous_umask
  previous_umask="$(umask)"
  umask 077
  printf '%s\n' "${TEST_ROUTE_TOKEN}" >"${TEST_ROUTE_TOKEN_FILE}"
  umask "${previous_umask}"
  chmod 600 "${TEST_ROUTE_TOKEN_FILE}" 2>/dev/null || true
}

validate_port_number() {
  local port="$1"
  local name="$2"

  if [[ ! "${port}" =~ ^[0-9]+$ ]] || (( port < 1 || port > 65535 )); then
    echo "${name} port must be an integer from 1 through 65535, got ${port}" >&2
    return 1
  fi
}

port_lease_root() {
  printf '%s\n' "${CARTULARY_WEB_E2E_PORT_LEASE_ROOT:-${ROOT_DIR}/.cartulary/web-e2e-port-leases}"
}

port_lease_dir() {
  local port="$1"
  printf '%s/port-%s\n' "$(port_lease_root)" "${port}"
}

release_port_lease_dir() {
  local lease_dir="$1"
  local lease_pid=""

  if [[ ! -d "${lease_dir}" ]]; then
    return 0
  fi

  if [[ -f "${lease_dir}/pid" ]]; then
    IFS= read -r lease_pid <"${lease_dir}/pid" || true
  fi
  if [[ "${lease_pid}" == "$$" ]] ||
    { [[ "${lease_pid}" =~ ^[0-9]+$ ]] && ! kill -0 "${lease_pid}" 2>/dev/null; }; then
    rm -rf "${lease_dir}"
    return 0
  fi
  return 1
}

release_port_leases() {
  local lease_dir
  local status=0

  for lease_dir in "${PORT_LEASE_DIRS[@]}"; do
    release_port_lease_dir "${lease_dir}" || status=$?
  done
  PORT_LEASE_DIRS=()
  unset CARTULARY_WEB_E2E_BACKEND_PORT
  unset CARTULARY_WEB_E2E_FRONTEND_PORT
  return "${status}"
}

track_port_lease_dir() {
  local lease_dir="$1"
  local existing

  for existing in "${PORT_LEASE_DIRS[@]}"; do
    if [[ "${existing}" == "${lease_dir}" ]]; then
      return 0
    fi
  done
  PORT_LEASE_DIRS+=("${lease_dir}")
}

transfer_port_lease_for_port() {
  local port="$1"
  local owner_pid="$2"
  local lease_dir
  local lease_pid=""
  local next_pid_file

  if [[ ! "${owner_pid}" =~ ^[0-9]+$ ]] || ! kill -0 "${owner_pid}" 2>/dev/null; then
    echo "cannot transfer port ${port} lease to inactive owner ${owner_pid}" >&2
    return 1
  fi
  lease_dir="$(port_lease_dir "${port}")"
  if [[ -f "${lease_dir}/pid" ]]; then
    IFS= read -r lease_pid <"${lease_dir}/pid" || true
  fi
  if [[ "${lease_pid}" != "$$" ]]; then
    echo "cannot transfer port ${port} lease owned by ${lease_pid:-unknown}" >&2
    return 1
  fi
  next_pid_file="${lease_dir}/pid.next.$$"
  printf '%s\n' "${owner_pid}" >"${next_pid_file}"
  mv -f "${next_pid_file}" "${lease_dir}/pid"
  printf 'owner_process_group_pid=%s\n' "${owner_pid}" >>"${lease_dir}/metadata"
}

adopt_port_lease_for_cleanup() {
  local port="$1"
  local lease_dir

  lease_dir="$(port_lease_dir "${port}")"
  if [[ -d "${lease_dir}" ]]; then
    track_port_lease_dir "${lease_dir}"
  fi
}

reclaim_port_lease_for_replacement() {
	local port="$1"
	local previous_owner="$2"
	local lease_dir
	local lease_pid=""
	local next_pid_file

	lease_dir="$(port_lease_dir "${port}")"
	if [[ ! -f "${lease_dir}/pid" ]]; then
		echo "cannot reclaim missing port ${port} lease" >&2
		return 1
	fi
	IFS= read -r lease_pid <"${lease_dir}/pid" || true
	if [[ "${lease_pid}" != "${previous_owner}" ]] || kill -0 "${lease_pid}" 2>/dev/null; then
		echo "cannot reclaim port ${port} lease from active or mismatched owner ${lease_pid:-unknown}" >&2
		return 1
	fi
	next_pid_file="${lease_dir}/pid.next.$$"
	printf '%s\n' "$$" >"${next_pid_file}"
	mv -f "${next_pid_file}" "${lease_dir}/pid"
	printf 'replacement_controller_pid=%s\n' "$$" >>"${lease_dir}/metadata"
	track_port_lease_dir "${lease_dir}"
}

remove_stale_port_lease() {
  local lease_dir="$1"
  local lease_pid=""

  for _ in $(seq 1 10); do
    if [[ -f "${lease_dir}/pid" ]]; then
      break
    fi
    sleep 0.05
  done
  if [[ -f "${lease_dir}/pid" ]]; then
    IFS= read -r lease_pid <"${lease_dir}/pid" || true
  fi
  if [[ -z "${lease_pid}" ]]; then
    return 1
  fi
  if [[ -n "${lease_pid}" && "${lease_pid}" =~ ^[0-9]+$ ]] && kill -0 "${lease_pid}" 2>/dev/null; then
    return 1
  fi

  rm -rf "${lease_dir}"
}

reserve_port_lease() {
  local port="$1"
  local name="$2"
  local lease_root
  local lease_dir
  local lease_pid=""

  lease_root="$(port_lease_root)"
  lease_dir="$(port_lease_dir "${port}")"
  step_secure_mkdir "${lease_root}"

  if [[ -n "${CARTULARY_BROWSER_ACQUISITION_FILE:-}" ]]; then
    "${NODE_BIN:-node}" "${ROOT_DIR}/tools/harness/browser/browser-acquisition.mjs" \
      reserve-port "${CARTULARY_BROWSER_ACQUISITION_FILE}" "${lease_dir}" "$$" || return $?
    track_port_lease_dir "${lease_dir}"
    return 0
  fi

  if ! mkdir "${lease_dir}" 2>/dev/null; then
    if [[ -f "${lease_dir}/pid" ]]; then
      IFS= read -r lease_pid <"${lease_dir}/pid" || true
    fi
    if [[ "${lease_pid}" == "$$" ]]; then
      track_port_lease_dir "${lease_dir}"
      return 0
    fi
    remove_stale_port_lease "${lease_dir}" || return 1
    mkdir "${lease_dir}" 2>/dev/null || return 1
  fi
  chmod 700 "${lease_dir}" 2>/dev/null || true
  printf '%s\n' "$$" >"${lease_dir}/pid"
  {
    printf 'name=%s\n' "${name}"
    printf 'port=%s\n' "${port}"
    printf 'target=%s\n' "${CARTULARY_TEST_TARGET:-}"
    printf 'created_at=%s\n' "$(step_now_utc)"
  } >"${lease_dir}/metadata"
  track_port_lease_dir "${lease_dir}"
}

release_port_lease_for_port() {
  local port="$1"
  local lease_dir

  lease_dir="$(port_lease_dir "${port}")"
  release_port_lease_dir "${lease_dir}" || true
  local next_leases=()
  local existing
  for existing in "${PORT_LEASE_DIRS[@]}"; do
    if [[ "${existing}" != "${lease_dir}" ]]; then
      next_leases+=("${existing}")
    fi
  done
  PORT_LEASE_DIRS=("${next_leases[@]}")
}

browser_stage_name() {
  local stage="${CARTULARY_BROWSER_STAGE:-}"

  if [[ -n "${stage}" ]]; then
    printf '%s\n' "${stage}"
    return 0
  fi

  case "${CARTULARY_TEST_TARGET:-}" in
    browser-e2e-stateful)
      printf '%s\n' "stateful"
      ;;
    *)
      printf '%s\n' "webserver-backed"
      ;;
  esac
}

# effective_ephemeral_port_range prints the host's effective ephemeral port
# range as "low high". Browser ports never come from that range.
effective_ephemeral_port_range() {
  local low=""
  local high=""

  if ! read -r low high <"${EPHEMERAL_PORT_RANGE_FILE}" 2>/dev/null ||
    [[ ! "${low}" =~ ^[0-9]+$ || ! "${high}" =~ ^[0-9]+$ ]] || ((low > high)); then
    echo "failure_class=config reason=configuration_error cannot read the host ephemeral port range from ${EPHEMERAL_PORT_RANGE_FILE} (net.ipv4.ip_local_port_range)" >&2
    return 2
  fi
  printf '%s %s\n' "${low}" "${high}"
}

# browser_port_block_base selects the first deterministic candidate block that
# lies wholly outside the effective ephemeral range, or fails admission.
browser_port_block_base() {
  local range
  local low
  local high
  local base
  local end

  range="$(effective_ephemeral_port_range)" || return $?
  read -r low high <<<"${range}"
  for base in ${BROWSER_PORT_BLOCK_CANDIDATES}; do
    end=$((base + 4 * BROWSER_PORT_WINDOW_WIDTH - 1))
    if ((end < low || base > high)); then
      printf '%s\n' "${base}"
      return 0
    fi
  done
  echo "failure_class=config reason=configuration_error no browser port window fits outside the host ephemeral port range ${low}-${high}; adjust net.ipv4.ip_local_port_range" >&2
  return 2
}

# browser_port_window prints the stage-owned window for one stack component.
# The block layout is frontend web-server-backed, frontend stateful, backend
# web-server-backed and backend stateful, so no two windows overlap.
browser_port_window() {
  local name="$1"
  local base
  local offset=0

  base="$(browser_port_block_base)" || return $?
  if [[ "$(browser_stage_name)" == "stateful" ]]; then
    offset=$((offset + BROWSER_PORT_WINDOW_WIDTH))
  fi
  if [[ "${name}" == "backend" ]]; then
    offset=$((offset + 2 * BROWSER_PORT_WINDOW_WIDTH))
  fi
  printf '%s %s\n' "$((base + offset))" "$((base + offset + BROWSER_PORT_WINDOW_WIDTH - 1))"
}

rotated_port_window_candidates() {
  local name="$1"
  local range_start="$2"
  local range_end="$3"
  local span
  local seed_text
  local seed
  local offset
  local candidate
  local i

  span=$((range_end - range_start + 1))
  if (( span <= 0 )); then
    return 0
  fi

  seed_text="${CARTULARY_BROWSER_SESSION_GROUP:-}:${CARTULARY_BROWSER_STAGE:-}:${name}:"
  seed_text+="${CARTULARY_TEST_RUN_ID:-}:${CARTULARY_TEST_TARGET:-}:$$"
  seed="$(printf '%s' "${seed_text}" | cksum | awk '{print $1}')"
  if [[ ! "${seed}" =~ ^[0-9]+$ ]]; then
    seed=0
  fi
  offset=$((seed % span))

  for i in $(seq 0 $((span - 1))); do
    candidate=$((range_start + ((offset + i) % span)))
    printf '%s\n' "${candidate}"
  done
}

claim_available_port() {
  local outvar="$1"
  local name="$2"
  local candidate="$3"
  local configured="$4"
  local -n port_ref="$outvar"
  local lease_dir

  if port_in_use "${candidate}"; then
    if [[ "${configured}" == "1" ]]; then
      echo "${name} port ${candidate} is already in use; choose another CARTULARY_WEB_E2E_*_PORT override" >&2
      ss -ltnp "sport = :${candidate}" >&2 || true
    fi
    return 1
  fi
  if ! reserve_port_lease "${candidate}" "${name}"; then
    if [[ "${configured}" == "1" ]]; then
      echo "${name} port ${candidate} is reserved by another browser e2e startup; choose another CARTULARY_WEB_E2E_*_PORT override" >&2
    fi
    return 1
  fi
  if port_in_use "${candidate}"; then
    lease_dir="$(port_lease_dir "${candidate}")"
    release_port_lease_dir "${lease_dir}" || true
    if [[ "${configured}" == "1" ]]; then
      echo "${name} port ${candidate} became unavailable during browser e2e port allocation; choose another CARTULARY_WEB_E2E_*_PORT override" >&2
      ss -ltnp "sport = :${candidate}" >&2 || true
    fi
    return 1
  fi

  # shellcheck disable=SC2034
  port_ref="${candidate}"
}

port_is_excluded() {
  local candidate="$1"
  local excluded_ports="$2"
  local excluded

  for excluded in ${excluded_ports//,/ }; do
    if [[ -n "${excluded}" && "${candidate}" == "${excluded}" ]]; then
      return 0
    fi
  done
  return 1
}

allocate_available_port() {
  local outvar="$1"
  local name="$2"
  local configured_port="$3"
  local excluded_ports="${4:-}"
  local range
  local low
  local high
  local window
  local range_start
  local range_end
  local candidate

  printf -v "$outvar" '%s' ""

  if [[ -n "${configured_port}" ]]; then
    validate_port_number "${configured_port}" "${name}" || return $?
    range="$(effective_ephemeral_port_range)" || return $?
    read -r low high <<<"${range}"
    if ((configured_port >= low && configured_port <= high)); then
      echo "failure_class=config reason=configuration_error ${name} port override ${configured_port} lies inside the host ephemeral port range ${low}-${high} (net.ipv4.ip_local_port_range)" >&2
      return 2
    fi
    if port_is_excluded "${configured_port}" "${excluded_ports}"; then
      echo "${name} port ${configured_port} must differ from another browser e2e stack port" >&2
      return 1
    fi
    claim_available_port "$outvar" "${name}" "${configured_port}" "1" || return $?
    return 0
  fi

  window="$(browser_port_window "${name}")" || return $?
  read -r range_start range_end <<<"${window}"
  for candidate in $(rotated_port_window_candidates "${name}" "${range_start}" "${range_end}"); do
    if port_is_excluded "${candidate}" "${excluded_ports}"; then
      continue
    fi
    if claim_available_port "$outvar" "${name}" "${candidate}" "0"; then
      return 0
    fi
  done

  echo "failed to allocate an available ${name} port in browser $(browser_stage_name) window ${range_start}-${range_end}" >&2
  return 1
}

resolve_owned_stack_ports() {
  allocate_available_port BACKEND_PORT "backend" "${CARTULARY_WEB_E2E_BACKEND_PORT:-}" "" || return $?
  allocate_available_port FRONTEND_PORT "frontend" "${CARTULARY_WEB_E2E_FRONTEND_PORT:-}" "${BACKEND_PORT}" || return $?

  if [[ "${BACKEND_PORT}" == "${FRONTEND_PORT}" ]]; then
    echo "backend and frontend ports must differ for browser e2e" >&2
    return 1
  fi

  API_ORIGIN="https://127.0.0.1:${BACKEND_PORT}"
  PUBLIC_ORIGIN="https://127.0.0.1:${FRONTEND_PORT}"
  export CARTULARY_WEB_E2E_API_ORIGIN="${API_ORIGIN}"
  export CARTULARY_WEB_E2E_PUBLIC_ORIGIN="${PUBLIC_ORIGIN}"
  export CARTULARY_WEB_E2E_BACKEND_PORT="${BACKEND_PORT}"
  export CARTULARY_WEB_E2E_FRONTEND_PORT="${FRONTEND_PORT}"
}
