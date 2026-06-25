#!/usr/bin/env bash
set -euo pipefail

# Lab4 API test data preparation script for WSL/Linux.
# It creates an isolated set of users, tasks and a team, then writes the
# cookies and resource IDs consumed by lab4_run_api_tests.sh.

BASE_URL="${BASE_URL:-http://127.0.0.1:8080/api/v1}"
OUT_FILE="${OUT_FILE:-scripts/lab4_api_test_data.json}"
PASSWORD="${PASSWORD:-abc12345}"
STAMP="${STAMP:-$(date +%s)${RANDOM}}"
OWNER="${OWNER:-l4o${STAMP}}"
MEMBER="${MEMBER:-l4m${STAMP}}"
OUTSIDER="${OUTSIDER:-l4x${STAMP}}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing dependency: $1"
    echo "Install it in WSL with: sudo apt install curl jq"
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

require_status() {
  local name="$1"
  local expected="$2"
  if [[ "$RESPONSE_STATUS" != "$expected" ]]; then
    echo "[FAIL] ${name}: status=${RESPONSE_STATUS}, expected=${expected}" >&2
    echo "       response=${RESPONSE_BODY}" >&2
    exit 1
  fi
}

api_register() {
  local username="$1"
  local payload
  payload="$(jq -nc --arg username "$username" --arg password "$PASSWORD" '{username:$username,password:$password}')"
  api_request "POST" "" "/auth/register" "$payload"
  require_status "register ${username}" "201"
}

api_login_cookie() {
  local username="$1"
  local payload response status cookie
  payload="$(jq -nc --arg username "$username" --arg password "$PASSWORD" '{username:$username,password:$password}')"

  response="$(curl -i -sS -X POST "${BASE_URL}/auth/login" \
    -H "Content-Type: application/json" \
    -d "$payload")"
  status="$(printf '%s\n' "$response" | awk 'NR==1 {print $2}')"
  cookie="$(printf '%s\n' "$response" | grep -i '^set-cookie:' | sed -E 's/.*task_session=([^;]+).*/\1/' | head -n 1)"

  if [[ "$status" != "200" || -z "$cookie" ]]; then
    echo "Login failed or task_session cookie was not returned for user: $username" >&2
    printf '%s\n' "$response" >&2
    exit 1
  fi

  printf '%s' "$cookie"
}

api_create_task() {
  local cookie="$1"
  local title="$2"
  local team_id="${3:-}"
  local assignee_id="${4:-}"
  local payload id

  if [[ -n "$team_id" && -n "$assignee_id" ]]; then
    payload="$(jq -nc --arg title "$title" --argjson team_id "$team_id" --argjson assignee_id "$assignee_id" \
      '{title:$title,description:"lab4 api test",status:"TODO",priority:2,team_id:$team_id,assignee_id:$assignee_id}')"
  elif [[ -n "$team_id" ]]; then
    payload="$(jq -nc --arg title "$title" --argjson team_id "$team_id" \
      '{title:$title,description:"lab4 api test",status:"TODO",priority:2,team_id:$team_id}')"
  else
    payload="$(jq -nc --arg title "$title" \
      '{title:$title,description:"lab4 api test",status:"TODO",priority:2}')"
  fi

  api_request "POST" "$cookie" "/tasks" "$payload"
  require_status "create task ${title}" "201"
  id="$(printf '%s' "$RESPONSE_BODY" | jq -r '.data.id // empty')"
  if [[ -z "$id" || "$id" == "null" ]]; then
    echo "Create task did not return an id: $title" >&2
    echo "$RESPONSE_BODY" >&2
    exit 1
  fi
  printf '%s' "$id"
}

api_create_team() {
  local cookie="$1"
  local name="$2"
  local payload id
  payload="$(jq -nc --arg name "$name" '{name:$name}')"
  api_request "POST" "$cookie" "/teams" "$payload"
  require_status "create team ${name}" "201"
  id="$(printf '%s' "$RESPONSE_BODY" | jq -r '.data.team_id // empty')"
  if [[ -z "$id" || "$id" == "null" ]]; then
    echo "Create team did not return team_id: $name" >&2
    echo "$RESPONSE_BODY" >&2
    exit 1
  fi
  printf '%s' "$id"
}

api_invite_member() {
  local owner_cookie="$1"
  local team_id="$2"
  local username="$3"
  local payload member_id
  payload="$(jq -nc --arg username "$username" '{username:$username}')"
  api_request "POST" "$owner_cookie" "/teams/${team_id}/members" "$payload"
  require_status "invite member ${username}" "201"
  member_id="$(printf '%s' "$RESPONSE_BODY" | jq -r '.data.user_id // empty')"
  if [[ -z "$member_id" || "$member_id" == "null" ]]; then
    echo "Invite member did not return user_id: $username" >&2
    echo "$RESPONSE_BODY" >&2
    exit 1
  fi
  printf '%s' "$member_id"
}

api_accept_invitation() {
  local member_cookie="$1"
  local team_id="$2"
  local invitation_id
  api_request "GET" "$member_cookie" "/invitations"
  require_status "list invitations" "200"
  invitation_id="$(printf '%s' "$RESPONSE_BODY" | jq -r --argjson team_id "$team_id" '.data[] | select(.team_id == $team_id) | .id' | head -n 1)"
  if [[ -z "$invitation_id" || "$invitation_id" == "null" ]]; then
    echo "No pending invitation found for team ${team_id}" >&2
    echo "$RESPONSE_BODY" >&2
    exit 1
  fi
  api_request "POST" "$member_cookie" "/invitations/${invitation_id}/accept"
  require_status "accept invitation ${invitation_id}" "200"
}

require_cmd curl
require_cmd jq

mkdir -p "$(dirname "$OUT_FILE")"

echo "Preparing Lab4 API data against: $BASE_URL"
api_register "$OWNER"
api_register "$MEMBER"
api_register "$OUTSIDER"

OWNER_COOKIE="$(api_login_cookie "$OWNER")"
MEMBER_COOKIE="$(api_login_cookie "$MEMBER")"
OUTSIDER_COOKIE="$(api_login_cookie "$OUTSIDER")"

TEAM_ID="$(api_create_team "$OWNER_COOKIE" "lab4_${STAMP}_team")"
MEMBER_ID="$(api_invite_member "$OWNER_COOKIE" "$TEAM_ID" "$MEMBER")"
api_accept_invitation "$MEMBER_COOKIE" "$TEAM_ID"

LOG_TITLE="lab4_${STAMP}_log_task"
NO_CHANGE_TITLE="lab4_${STAMP}_no_change"
DELETE_TITLE="lab4_${STAMP}_delete"
LOG_TASK="$(api_create_task "$OWNER_COOKIE" "$LOG_TITLE")"
NO_CHANGE_TASK="$(api_create_task "$OWNER_COOKIE" "$NO_CHANGE_TITLE")"
DELETE_TASK="$(api_create_task "$OWNER_COOKIE" "$DELETE_TITLE")"
OWNER_OTHER_TASK="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_owner_other")"
MEMBER_PERSONAL_TASK="$(api_create_task "$MEMBER_COOKIE" "lab4_${STAMP}_member_personal")"
DEPENDENCY_PRE="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_dependency_pre")"
DEPENDENCY_SUCC="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_dependency_succ")"
ROLLBACK_PRE="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_rollback_pre")"
ROLLBACK_SUCC="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_rollback_succ")"

TEAM_ASSIGNED_TASK="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_team_assigned" "$TEAM_ID" "$MEMBER_ID")"
TEAM_LOG_TASK="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_team_log" "$TEAM_ID")"
TEAM_LEAVE_TASK="$(api_create_task "$OWNER_COOKIE" "lab4_${STAMP}_team_leave" "$TEAM_ID" "$MEMBER_ID")"

jq -n \
  --arg baseUrl "$BASE_URL" \
  --arg password "$PASSWORD" \
  --arg stamp "$STAMP" \
  --arg owner "$OWNER" \
  --arg ownerCookie "$OWNER_COOKIE" \
  --arg member "$MEMBER" \
  --arg memberCookie "$MEMBER_COOKIE" \
  --arg outsider "$OUTSIDER" \
  --arg outsiderCookie "$OUTSIDER_COOKIE" \
  --argjson teamID "$TEAM_ID" \
  --argjson memberID "$MEMBER_ID" \
  --arg logTitle "$LOG_TITLE" \
  --arg noChangeTitle "$NO_CHANGE_TITLE" \
  --arg deleteTitle "$DELETE_TITLE" \
  --argjson logTask "$LOG_TASK" \
  --argjson noChangeTask "$NO_CHANGE_TASK" \
  --argjson deleteTask "$DELETE_TASK" \
  --argjson ownerOtherTask "$OWNER_OTHER_TASK" \
  --argjson memberPersonalTask "$MEMBER_PERSONAL_TASK" \
  --argjson dependencyPre "$DEPENDENCY_PRE" \
  --argjson dependencySucc "$DEPENDENCY_SUCC" \
  --argjson rollbackPre "$ROLLBACK_PRE" \
  --argjson rollbackSucc "$ROLLBACK_SUCC" \
  --argjson teamAssignedTask "$TEAM_ASSIGNED_TASK" \
  --argjson teamLogTask "$TEAM_LOG_TASK" \
  --argjson teamLeaveTask "$TEAM_LEAVE_TASK" \
  '{
    baseUrl: $baseUrl,
    password: $password,
    stamp: $stamp,
    users: {
      owner: {username: $owner, cookie: $ownerCookie},
      member: {username: $member, id: $memberID, cookie: $memberCookie},
      outsider: {username: $outsider, cookie: $outsiderCookie}
    },
    team: {id: $teamID},
    tasks: {
      logTask: {id: $logTask, title: $logTitle},
      noChangeTask: {id: $noChangeTask, title: $noChangeTitle},
      deleteTask: {id: $deleteTask, title: $deleteTitle},
      ownerOtherTask: {id: $ownerOtherTask},
      memberPersonalTask: {id: $memberPersonalTask},
      dependencyPre: {id: $dependencyPre},
      dependencySucc: {id: $dependencySucc},
      rollbackPre: {id: $rollbackPre},
      rollbackSucc: {id: $rollbackSucc}
    },
    teamTasks: {
      assignedTask: {id: $teamAssignedTask},
      logTask: {id: $teamLogTask},
      leaveTask: {id: $teamLeaveTask}
    }
  }' > "$OUT_FILE"

echo "Prepared data written to: $OUT_FILE"
jq '.' "$OUT_FILE"
