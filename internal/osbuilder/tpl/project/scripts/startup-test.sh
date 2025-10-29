#!/bin/bash

API_BASE="http://127.0.0.1:5555/v1"
CT="Content-Type: application/json"

res="$1"
action="$2"
arg="$3"

[ -z "$res" ] || [ -z "$action" ] && {
  echo "Usage: $0 <resource> {create|get|list} [args]"
  exit 1
}

url="$API_BASE/$res"

# Common helper: perform a curl request and extract X-Trace-Id
call_api() {
  local method="$1"
  local full_url="$2"
  local data="$3"

  # Use curl to print response headers (-D -) and body (-s)
  response=$(curl -s -D - -X "$method" "$full_url" -H "$CT" -d "$data")
  trace_id=$(echo "$response" | grep -i '^X-Trace-Id:' | awk '{print $2}' | tr -d '\r')
  body=$(echo "$response" | sed -n '/^\r$/,$p' | tail -n +2)

  echo "X-Trace-Id: ${trace_id:-<none>}"
  echo "-----------------------------"
  echo "$body" | jq .
}

case "$action" in
  create)
    [ -z "$arg" ] && { echo "Example: $0 $res create '{\"name\":\"my-job\"}'"; exit 1; }
    call_api POST "$url" "$arg"
    ;;
  get)
    [ -z "$arg" ] && { echo "Example: $0 $res get 123"; exit 1; }
    call_api GET "$url/$arg" ""
    ;;
  list)
    call_api GET "$url" ""
    ;;
  *)
    echo "Unsupported action: $action (supported: create|get|list)"
    exit 1
    ;;
esac
