#!/usr/bin/env bash
set -euo pipefail

# Lab4 API acceptance script for WSL/Linux.
# Run scripts/prepare_api_data.sh first. This script intentionally mutates
# the prepared resources; rerun the prepare script before another full pass.

DATA_FILE="${DATA_FILE:-scripts/lab4_api_test_data.json}"
LAB_FILTER=""
CASE_FILTER=""
LIST_ONLY=0

usage() {
  cat <<'EOF'
Usage: scripts/run_api_tests.sh [options]

Options:
  --lab LAB        Run all cases for one lab: lab1, lab2, lab3, lab4.
  --case CASE     Run one case group, for example TC08. Accepts comma-separated values.
  --list          List available cases and exit.
  -h, --help      Show this help.

Examples:
  scripts/run_api_tests.sh
  scripts/run_api_tests.sh --lab lab4
  scripts/run_api_tests.sh --case TC08
  scripts/run_api_tests.sh --case TC08,TC14
EOF
}

normalize_case() {
  local value number
  value="$(printf '%s' "$1" | tr '[:lower:]' '[:upper:]')"
  value="${value//[[:space:]]/}"
  if [[ "$value" =~ ^TC([0-9]+)$ ]]; then
    number="${BASH_REMATCH[1]}"
    printf 'TC%02d' "$((10#$number))"
  elif [[ "$value" =~ ^[0-9]+$ ]]; then
    printf 'TC%02d' "$((10#$value))"
  else
    printf '%s' "$value"
  fi
}

case_in_filter() {
  local target="$1"
  local normalized part
  local -a parts
  IFS=',' read -ra parts <<< "$CASE_FILTER"
  for part in "${parts[@]}"; do
    normalized="$(normalize_case "$part")"
    if [[ "$normalized" == "$target" ]]; then
      return 0
    fi
  done
  return 1
}

lab_in_case_labs() {
  local case_labs="$1"
  local lab part
  local -a parts
  lab="$(printf '%s' "$LAB_FILTER" | tr '[:upper:]' '[:lower:]')"
  IFS=',' read -ra parts <<< "$case_labs"
  for part in "${parts[@]}"; do
    part="${part//[[:space:]]/}"
    if [[ "$(printf '%s' "$part" | tr '[:upper:]' '[:lower:]')" == "$lab" ]]; then
      return 0
    fi
  done
  return 1
}

should_run() {
  local case_id="$1"
  local case_labs="$2"
  if [[ -n "$CASE_FILTER" ]] && ! case_in_filter "$case_id"; then
    return 1
  fi
  if [[ -n "$LAB_FILTER" ]] && ! lab_in_case_labs "$case_labs"; then
    return 1
  fi
  return 0
}

list_cases() {
  cat <<'EOF'
TC01 lab1  register success and duplicate register rejection
TC02 lab1  login success and wrong-password rejection
TC03 lab1  personal task CRUD
TC04 lab2  invited member can see team
TC05 lab2  non-member team access rejected
TC06 lab2  member status-only task permission
TC07 lab3  dependency create and duplicate rejection
TC08 lab3  cross-user/cycle dependency rejection
TC09 lab3  dependency status constraints
TC10 lab3,lab4  member leave ownership-change log and permission loss
TC11 lab4  readable task update logs
TC12 lab4  no-op update does not write another log
TC13 lab4  deleted task logs remain queryable
TC14 lab4  log endpoint auth and permission rejection
TC15 lab4  task trash listing and permission rejection
TC16 lab4  assignment notification dismiss
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --lab)
      if [[ $# -lt 2 || "${2:-}" == --* ]]; then
        echo "Missing value for --lab" >&2
        usage >&2
        exit 2
      fi
      LAB_FILTER="${2:-}"
      shift 2
      ;;
    --lab=*)
      LAB_FILTER="${1#*=}"
      shift
      ;;
    --case)
      if [[ $# -lt 2 || "${2:-}" == --* ]]; then
        echo "Missing value for --case" >&2
        usage >&2
        exit 2
      fi
      CASE_FILTER="${2:-}"
      shift 2
      ;;
    --case=*)
      CASE_FILTER="${1#*=}"
      shift
      ;;
    --list)
      LIST_ONLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -n "$LAB_FILTER" ]]; then
  LAB_FILTER="$(printf '%s' "$LAB_FILTER" | tr '[:upper:]' '[:lower:]')"
  LAB_FILTER="${LAB_FILTER//[[:space:]]/}"
  if [[ "$LAB_FILTER" =~ ^[1-4]$ ]]; then
    LAB_FILTER="lab${LAB_FILTER}"
  fi
  case "$LAB_FILTER" in
    lab1|lab2|lab3|lab4) ;;
    *)
      echo "Unknown lab filter: ${LAB_FILTER}" >&2
      usage >&2
      exit 2
      ;;
  esac
fi

if [[ "$LIST_ONLY" -eq 1 ]]; then
  list_cases
  exit 0
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing dependency: $1"
    echo "Install it in WSL with: sudo apt install curl jq"
    exit 1
  fi
}

require_file() {
  if [[ ! -f "$1" ]]; then
    echo "Missing data file: $1"
    echo "Run scripts/prepare_api_data.sh first."
    exit 1
  fi
}

api_request() {
  local method="$1"
  local cookie="$2"
  local path="$3"
  local payload="${4:-}"
  local body_file status
  body_file="$(mktemp)"

  if [[ -n "$payload" && -n "$cookie" ]]; then
    status="$(curl -sS -o "$body_file" -w "%{http_code}" -X "$method" "${BASE_URL}${path}" \
      -H "Content-Type: application/json" \
      -H "Cookie: task_session=${cookie}" \
      -d "$payload")"
  elif [[ -n "$payload" ]]; then
    status="$(curl -sS -o "$body_file" -w "%{http_code}" -X "$method" "${BASE_URL}${path}" \
      -H "Content-Type: application/json" \
      -d "$payload")"
  elif [[ -n "$cookie" ]]; then
    status="$(curl -sS -o "$body_file" -w "%{http_code}" -X "$method" "${BASE_URL}${path}" \
      -H "Content-Type: application/json" \
      -H "Cookie: task_session=${cookie}")"
  else
    status="$(curl -sS -o "$body_file" -w "%{http_code}" -X "$method" "${BASE_URL}${path}" \
      -H "Content-Type: application/json")"
  fi

  RESPONSE_STATUS="$status"
  RESPONSE_BODY="$(cat "$body_file")"
  rm -f "$body_file"
}

json_message() {
  printf '%s' "$1" | jq -r '.message // ""' 2>/dev/null || true
}

json_has_path() {
  local body="$1"
  local jq_expr="$2"
  printf '%s' "$body" | jq -e "$jq_expr" >/dev/null 2>&1
}

status_matches() {
  local actual="$1"
  local expected="$2"
  local part
  local -a parts
  expected="${expected//,/|}"
  IFS='|' read -ra parts <<< "$expected"
  for part in "${parts[@]}"; do
    if [[ "$actual" == "$part" ]]; then
      return 0
    fi
  done
  return 1
}

run_case() {
  local name="$1"
  local expected="$2"
  local method="$3"
  local cookie="$4"
  local path="$5"
  local payload="${6:-}"
  local extra_check="${7:-}"

  TOTAL=$((TOTAL + 1))
  api_request "$method" "$cookie" "$path" "$payload"

  local ok=0
  if status_matches "$RESPONSE_STATUS" "$expected"; then
    ok=1
  fi

  if [[ -n "$extra_check" && "$ok" -eq 1 ]]; then
    if ! json_has_path "$RESPONSE_BODY" "$extra_check"; then
      ok=0
    fi
  fi

  local message
  message="$(json_message "$RESPONSE_BODY")"

  if [[ "$ok" -eq 1 ]]; then
    PASSED=$((PASSED + 1))
    printf '[PASS] %s | status=%s expected=%s | message=%s\n' "$name" "$RESPONSE_STATUS" "$expected" "$message"
  else
    FAILED=$((FAILED + 1))
    printf '[FAIL] %s | status=%s expected=%s | message=%s\n' "$name" "$RESPONSE_STATUS" "$expected" "$message"
    printf '       response=%s\n' "$RESPONSE_BODY"
  fi
}

record_check() {
  local name="$1"
  local ok="$2"
  local detail="${3:-}"

  TOTAL=$((TOTAL + 1))
  if [[ "$ok" == "1" ]]; then
    PASSED=$((PASSED + 1))
    if [[ -n "$detail" ]]; then
      printf '[PASS] %s | %s\n' "$name" "$detail"
    else
      printf '[PASS] %s\n' "$name"
    fi
  else
    FAILED=$((FAILED + 1))
    if [[ -n "$detail" ]]; then
      printf '[FAIL] %s | %s\n' "$name" "$detail"
    else
      printf '[FAIL] %s\n' "$name"
    fi
  fi
}

create_case_task() {
  local title="$1"
  local payload
  payload="$(jq -nc --arg title "$title" '{title:$title,description:"lab4 api case",status:"TODO",priority:2}')"
  run_case "Create personal task ${title}" "201" "POST" "$OWNER_COOKIE" "/tasks" "$payload" '.code == 0 and (.data.id > 0)'
  CASE_TASK_ID="$(printf '%s' "$RESPONSE_BODY" | jq -r '.data.id // empty')"
  if [[ -z "$CASE_TASK_ID" || "$CASE_TASK_ID" == "null" ]]; then
    echo "Cannot continue because task id was not returned." >&2
    exit 1
  fi
}

ensure_dependency_exists() {
  local predecessor_id="$1"
  local successor_id="$2"
  local payload
  payload="$(jq -nc --argjson predecessor_id "$predecessor_id" '{predecessor_id:$predecessor_id}')"
  api_request "POST" "$OWNER_COOKIE" "/tasks/${successor_id}/dependencies" "$payload"
  if [[ "$RESPONSE_STATUS" != "201" && "$RESPONSE_STATUS" != "409" ]]; then
    echo "Failed to prepare dependency ${successor_id} <- ${predecessor_id}: status=${RESPONSE_STATUS}" >&2
    echo "response=${RESPONSE_BODY}" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq
require_file "$DATA_FILE"

BASE_URL="$(jq -r '.baseUrl' "$DATA_FILE")"
PASSWORD="$(jq -r '.password' "$DATA_FILE")"
STAMP="$(jq -r '.stamp' "$DATA_FILE")"
OWNER_USERNAME="$(jq -r '.users.owner.username' "$DATA_FILE")"
OWNER_COOKIE="$(jq -r '.users.owner.cookie' "$DATA_FILE")"
MEMBER_USERNAME="$(jq -r '.users.member.username' "$DATA_FILE")"
MEMBER_COOKIE="$(jq -r '.users.member.cookie' "$DATA_FILE")"
OUTSIDER_COOKIE="$(jq -r '.users.outsider.cookie' "$DATA_FILE")"
TEAM_ID="$(jq -r '.team.id' "$DATA_FILE")"
LOG_TASK="$(jq -r '.tasks.logTask.id' "$DATA_FILE")"
LOG_TITLE="$(jq -r '.tasks.logTask.title' "$DATA_FILE")"
NO_CHANGE_TASK="$(jq -r '.tasks.noChangeTask.id' "$DATA_FILE")"
NO_CHANGE_TITLE="$(jq -r '.tasks.noChangeTask.title' "$DATA_FILE")"
DELETE_TASK="$(jq -r '.tasks.deleteTask.id' "$DATA_FILE")"
MEMBER_PERSONAL_TASK="$(jq -r '.tasks.memberPersonalTask.id' "$DATA_FILE")"
DEPENDENCY_PRE="$(jq -r '.tasks.dependencyPre.id' "$DATA_FILE")"
DEPENDENCY_SUCC="$(jq -r '.tasks.dependencySucc.id' "$DATA_FILE")"
ROLLBACK_PRE="$(jq -r '.tasks.rollbackPre.id' "$DATA_FILE")"
ROLLBACK_SUCC="$(jq -r '.tasks.rollbackSucc.id' "$DATA_FILE")"
TEAM_ASSIGNED_TASK="$(jq -r '.teamTasks.assignedTask.id' "$DATA_FILE")"
TEAM_LOG_TASK="$(jq -r '.teamTasks.logTask.id' "$DATA_FILE")"
TEAM_LEAVE_TASK="$(jq -r '.teamTasks.leaveTask.id' "$DATA_FILE")"
TEAM_DISMISS_TASK="$(jq -r '.teamTasks.dismissTask.id // empty' "$DATA_FILE")"
DISMISS_ASSIGNMENT_REQUEST="$(jq -r '.assignmentRequests.dismissRequest.id // empty' "$DATA_FILE")"

TOTAL=0
PASSED=0
FAILED=0

echo "Running Lab4 API acceptance tests against: $BASE_URL"
if [[ -n "$LAB_FILTER" ]]; then
  echo "Filter: lab=${LAB_FILTER}"
fi
if [[ -n "$CASE_FILTER" ]]; then
  echo "Filter: case=${CASE_FILTER}"
fi
if should_run "TC16" "lab4"; then
  if [[ -z "$TEAM_DISMISS_TASK" || -z "$DISMISS_ASSIGNMENT_REQUEST" ]]; then
    echo "Data file does not contain TC16 assignment request data."
    echo "Run scripts/prepare_api_data.sh again before running TC16 or the full Lab4 suite."
    exit 2
  fi
fi

# TC01 / TC02: auth regression.
if should_run "TC01" "lab1"; then
  CASE_USER="l4c$(date +%s)${RANDOM}"
  REGISTER_PAYLOAD="$(jq -nc --arg username "$CASE_USER" --arg password "$PASSWORD" '{username:$username,password:$password}')"
  run_case "TC01 register success" "201" "POST" "" "/auth/register" "$REGISTER_PAYLOAD" '.code == 0 and .message == "register success"'
  run_case "TC01 duplicate register rejected" "409" "POST" "" "/auth/register" "$REGISTER_PAYLOAD" '.code == 409'
fi

if should_run "TC02" "lab1"; then
  LOGIN_PAYLOAD="$(jq -nc --arg username "$OWNER_USERNAME" --arg password "$PASSWORD" '{username:$username,password:$password}')"
  BAD_LOGIN_PAYLOAD="$(jq -nc --arg username "$OWNER_USERNAME" '{username:$username,password:"wrong123"}')"
  run_case "TC02 login success" "200" "POST" "" "/auth/login" "$LOGIN_PAYLOAD" '.code == 0 and .message == "login success"'
  run_case "TC02 wrong password rejected" "401" "POST" "" "/auth/login" "$BAD_LOGIN_PAYLOAD" '.code == 401'
fi

# TC03: personal task CRUD.
if should_run "TC03" "lab1"; then
  CRUD_TITLE="lab4_${STAMP}_crud"
  create_case_task "$CRUD_TITLE"
  CRUD_TASK="$CASE_TASK_ID"
  run_case "TC03 personal task appears in list" "200" "GET" "$OWNER_COOKIE" "/tasks?page=1&size=100" "" ".data.list | any(.id == ${CRUD_TASK})"
  run_case "TC03 update personal task" "200" "PUT" "$OWNER_COOKIE" "/tasks/${CRUD_TASK}" '{"title":"lab4 crud updated","status":"DOING","priority":3}' '.data.title == "lab4 crud updated" and .data.status == "DOING"'
  run_case "TC03 delete personal task" "200" "DELETE" "$OWNER_COOKIE" "/tasks/${CRUD_TASK}" "" '.code == 0'
fi

# TC04 / TC05: team permission regression.
if should_run "TC04" "lab2"; then
  run_case "TC04 invited member can see team" "200" "GET" "$MEMBER_COOKIE" "/teams" "" ".data.list | any(.team_id == ${TEAM_ID})"
fi
if should_run "TC05" "lab2"; then
  run_case "TC05 non-member cannot list team tasks" "403" "GET" "$OUTSIDER_COOKIE" "/tasks?team_id=${TEAM_ID}" "" '.code == 403'
fi

# TC06: ordinary team member can update only assigned task status.
if should_run "TC06" "lab2"; then
  run_case "TC06 member updates assigned task status" "200" "PUT" "$MEMBER_COOKIE" "/tasks/${TEAM_ASSIGNED_TASK}" '{"status":"DOING"}' '.data.status == "DOING"'
  run_case "TC06 member cannot edit non-status fields" "403" "PUT" "$MEMBER_COOKIE" "/tasks/${TEAM_ASSIGNED_TASK}" '{"title":"member should not edit"}' '.code == 403'
fi

# TC07 / TC08 / TC09: dependency regression.
if should_run "TC07" "lab3"; then
  run_case "TC07 create personal dependency" "201" "POST" "$OWNER_COOKIE" "/tasks/${DEPENDENCY_SUCC}/dependencies" "$(jq -nc --argjson predecessor_id "$DEPENDENCY_PRE" '{predecessor_id:$predecessor_id}')" '.code == 0'
  run_case "TC07 duplicate dependency rejected" "409" "POST" "$OWNER_COOKIE" "/tasks/${DEPENDENCY_SUCC}/dependencies" "$(jq -nc --argjson predecessor_id "$DEPENDENCY_PRE" '{predecessor_id:$predecessor_id}')" '.code == 409'
fi

if should_run "TC08" "lab3"; then
  ensure_dependency_exists "$DEPENDENCY_PRE" "$DEPENDENCY_SUCC"
  run_case "TC08 cross-user dependency rejected" "403|404" "POST" "$OWNER_COOKIE" "/tasks/${DEPENDENCY_SUCC}/dependencies" "$(jq -nc --argjson predecessor_id "$MEMBER_PERSONAL_TASK" '{predecessor_id:$predecessor_id}')" '.code == 403 or .code == 404'
  run_case "TC08 dependency cycle rejected" "400" "POST" "$OWNER_COOKIE" "/tasks/${DEPENDENCY_PRE}/dependencies" "$(jq -nc --argjson predecessor_id "$DEPENDENCY_SUCC" '{predecessor_id:$predecessor_id}')" '.code == 400'
fi

if should_run "TC09" "lab3"; then
  ensure_dependency_exists "$DEPENDENCY_PRE" "$DEPENDENCY_SUCC"
  run_case "TC09 cannot finish successor with unfinished predecessor" "409" "PUT" "$OWNER_COOKIE" "/tasks/${DEPENDENCY_SUCC}" '{"status":"DONE"}' '.data.unresolved_predecessor_ids | length > 0'

  run_case "TC09 mark rollback predecessor done" "200" "PUT" "$OWNER_COOKIE" "/tasks/${ROLLBACK_PRE}" '{"status":"DONE"}' '.data.status == "DONE"'
  run_case "TC09 create dependency for rollback check" "201|409" "POST" "$OWNER_COOKIE" "/tasks/${ROLLBACK_SUCC}/dependencies" "$(jq -nc --argjson predecessor_id "$ROLLBACK_PRE" '{predecessor_id:$predecessor_id}')" '.code == 0 or .code == 409'
  run_case "TC09 mark rollback successor done" "200" "PUT" "$OWNER_COOKIE" "/tasks/${ROLLBACK_SUCC}" '{"status":"DONE"}' '.data.status == "DONE"'
  run_case "TC09 done predecessor cannot rollback while done successor depends on it" "409" "PUT" "$OWNER_COOKIE" "/tasks/${ROLLBACK_PRE}" '{"status":"DOING"}' '.data.done_successor_ids | length > 0'
fi

# TC11 / TC12 / TC13 / TC14: Lab4 activity log behavior.
if should_run "TC11" "lab4"; then
  run_case "TC11 update task creates readable log" "200" "PUT" "$OWNER_COOKIE" "/tasks/${LOG_TASK}" '{"title":"lab4 log updated","status":"DOING","priority":3}' '.data.title == "lab4 log updated"'
  run_case "TC11 logs contain actor, action, content and time in newest-first page" "200" "GET" "$OWNER_COOKIE" "/tasks/${LOG_TASK}/logs?page=1&size=10" "" '.data.list[0].action_type == "UPDATE_TASK_INFO" and (.data.list[0].actor_name | length > 0) and (.data.list[0].content | length > 0) and (.data.list[0].created_at | length > 0) and (.data.list | any(.action_type == "CREATE_TASK"))'
fi

if should_run "TC14" "lab4"; then
  run_case "TC14 missing auth cannot query logs" "401" "GET" "" "/tasks/${LOG_TASK}/logs" "" '.code == 401'
  run_case "TC14 other user cannot query personal logs" "403" "GET" "$MEMBER_COOKIE" "/tasks/${LOG_TASK}/logs" "" '.code == 403'
  run_case "TC14 non-member cannot query team task logs" "403" "GET" "$OUTSIDER_COOKIE" "/tasks/${TEAM_LOG_TASK}/logs" "" '.code == 403'
fi

if should_run "TC12" "lab4"; then
  run_case "TC12 get no-change log total before update" "200" "GET" "$OWNER_COOKIE" "/tasks/${NO_CHANGE_TASK}/logs" "" '.data.total >= 1'
  NO_CHANGE_TOTAL_BEFORE="$(printf '%s' "$RESPONSE_BODY" | jq -r '.data.total')"
  run_case "TC12 update same title succeeds" "200" "PUT" "$OWNER_COOKIE" "/tasks/${NO_CHANGE_TASK}" "$(jq -nc --arg title "$NO_CHANGE_TITLE" '{title:$title}')" '.code == 0'
  api_request "GET" "$OWNER_COOKIE" "/tasks/${NO_CHANGE_TASK}/logs"
  NO_CHANGE_TOTAL_AFTER="$(printf '%s' "$RESPONSE_BODY" | jq -r '.data.total // -1')"
  if [[ "$RESPONSE_STATUS" == "200" && "$NO_CHANGE_TOTAL_BEFORE" == "$NO_CHANGE_TOTAL_AFTER" ]]; then
    record_check "TC12 no actual change does not write another log" "1" "total=${NO_CHANGE_TOTAL_AFTER}"
  else
    record_check "TC12 no actual change does not write another log" "0" "before=${NO_CHANGE_TOTAL_BEFORE}, after=${NO_CHANGE_TOTAL_AFTER}, status=${RESPONSE_STATUS}"
  fi
fi

if should_run "TC13" "lab4"; then
  run_case "TC13 delete task" "200" "DELETE" "$OWNER_COOKIE" "/tasks/${DELETE_TASK}" "" '.code == 0'
  run_case "TC13 deleted task logs remain queryable by owner" "200" "GET" "$OWNER_COOKIE" "/tasks/${DELETE_TASK}/logs" "" '.data.list | any(.action_type == "DELETE_TASK")'
fi

if should_run "TC15" "lab4"; then
  TRASH_TITLE="lab4_${STAMP}_trash_case"
  create_case_task "$TRASH_TITLE"
  TRASH_TASK="$CASE_TASK_ID"
  run_case "TC15 delete task into trash" "200" "DELETE" "$OWNER_COOKIE" "/tasks/${TRASH_TASK}" "" '.code == 0'
  run_case "TC15 personal trash lists deleted task" "200" "GET" "$OWNER_COOKIE" "/tasks/trash?page=1&size=20" "" ".data.list | any(.id == ${TRASH_TASK} and .deleted_at != null)"
  run_case "TC15 invalid trash team id rejected" "400" "GET" "$OWNER_COOKIE" "/tasks/trash?team_id=abc" "" '.code == 400'
  run_case "TC15 outsider cannot list team trash" "403" "GET" "$OUTSIDER_COOKIE" "/tasks/trash?team_id=${TEAM_ID}" "" '.code == 403'
fi

if should_run "TC16" "lab4"; then
  run_case "TC16 assignment request visible before dismiss" "200" "GET" "$MEMBER_COOKIE" "/assignment-requests" "" "([.data[] | select(.id == ${DISMISS_ASSIGNMENT_REQUEST} and .task_id == ${TEAM_DISMISS_TASK})] | length) == 1"
  run_case "TC16 dismiss assignment notification" "200" "POST" "$MEMBER_COOKIE" "/assignment-requests/${DISMISS_ASSIGNMENT_REQUEST}/dismiss" "" ".data.request_id == ${DISMISS_ASSIGNMENT_REQUEST} and (.data.dismiss_at | length > 0)"
  run_case "TC16 dismissed assignment hidden from pending list" "200" "GET" "$MEMBER_COOKIE" "/assignment-requests" "" "([.data[] | select(.id == ${DISMISS_ASSIGNMENT_REQUEST})] | length) == 0"
  run_case "TC16 outsider cannot dismiss member assignment" "403|404" "POST" "$OUTSIDER_COOKIE" "/assignment-requests/${DISMISS_ASSIGNMENT_REQUEST}/dismiss" "" '.code == 403 or .code == 404'
fi

# TC10 is last because the member leaves the team and loses later permissions.
if should_run "TC10" "lab3,lab4"; then
  run_case "TC10 member leaves team" "200" "POST" "$MEMBER_COOKIE" "/teams/${TEAM_ID}/leave" "" '.code == 0'
  run_case "TC10 owner sees member-leave task ownership log" "200" "GET" "$OWNER_COOKIE" "/tasks/${TEAM_LEAVE_TASK}/logs" "" '.data.list | any(.action_type == "TASK_ASSIGNEE_CHANGED_BY_MEMBER_LEAVE")'
  run_case "TC10 departed member cannot query team task logs" "403" "GET" "$MEMBER_COOKIE" "/tasks/${TEAM_LEAVE_TASK}/logs" "" '.code == 403'
fi

if [[ "$TOTAL" -eq 0 ]]; then
  echo "No test cases matched the selected filter." >&2
  exit 2
fi

echo "Summary: total=${TOTAL}, passed=${PASSED}, failed=${FAILED}"

if [[ "$FAILED" -ne 0 ]]; then
  exit 1
fi
