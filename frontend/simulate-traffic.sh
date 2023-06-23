#!/usr/bin/env bash

set -euo pipefail

if [ -n "${DEBUG-}" ]; then
  set -x
fi

api_host="host.docker.internal"
api_port="8080"
api_endpoint="${api_host}:${api_port}"

# We're simulating the TODO list UI (imaginary Node.js Frontend)
export OTEL_SERVICE_NAME="todo-list-ui"

# Locally our Jaeger UI is not using secure transport
export OTEL_EXPORTER_OTLP_INSECURE=true
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=https://jaeger:4318/v1/traces
export OTEL_CLI_VERBOSE=true

if ! curl -o /dev/null -v "http://${api_endpoint}/ping"; then
  echo "API appears to be unavailable. Aborting simulation"
  exit 1
fi

function start_ui_request() (
  name=$1
  path=$2
  route=$3
  traceparent_carrier=$4
  tracesock=$5

  exec otel-cli span background \
    --sockdir "${tracesock}" \
    --kind 'server' \
    --name "${name}" \
    --attrs "http.request.method=GET,http.route=${route},url.scheme=https,url.path=${path},http.response.status_code=200,server.address=todo.example.com,server.port=443" \
    --tp-carrier "${traceparent_carrier}" \
    --tp-print \
    --skip-pid-check=true \
    --timeout 60 &

  # Allow background server to start
  while [ ! -e "${traceparent_carrier}" ]; do
    echo "Awaiting Start of UI Trace"
    sleep 0.1
    if [ -n "${DEBUG-}" ]; then
      ls -lash "$(dirname "${traceparent_carrier}")"
    fi
  done

  if [ -n "${DEBUG-}" ]; then
    ls -lash "${tracedir}"
  fi
)

function request_user_homepage() (
  pinned_list="$1"

  carrier=$(mktemp -d)/traceparent # traceparent propagation via tempfile
  tracedir=$(mktemp -d)

  start_ui_request 'User Homepage' '/' '/' "${carrier}" "${tracedir}"
  trap 'otel-cli span end --sockdir "${tracedir}"' EXIT

  api_fetch "lists" "${carrier}"
  api_fetch "lists/${pinned_list}" "${carrier}"
  api_fetch "lists/${pinned_list}/items" "${carrier}"
)

function request_todo_list() (
  list="$1"

  carrier=$(mktemp -d)/traceparent # traceparent propagation via tempfile
  tracedir=$(mktemp -d)

  start_ui_request 'TODO List' "/todo/${list}" '/todo/:id' "${carrier}" "${tracedir}"
  trap 'otel-cli span end --sockdir "${tracedir}"' EXIT

  api_fetch "lists/${list}" "${carrier}"
  api_fetch "lists/${list}/items" "${carrier}"
)

function api_fetch() (
  path=$1
  trace_carrier=$2

  url="http://${api_endpoint}/api/v1/${path}"
  echo "Performing HTTP GET to ${url}"

  # Need to create a new carrier, as the exec will overwrite it, and cause future spans to be children of the exec
  # instead of the proper parent
  child_span_carrier=$(mktemp -d)/traceparent
  cp "${trace_carrier}" "${child_span_carrier}"

  otel-cli exec \
    --tp-required \
    --tp-carrier "${child_span_carrier}" \
    --kind 'client' \
    --attrs "http.request.method=GET,server.address=${api_host},server.port=${api_port},url.full=${url}" \
    --name "GET" \
    -- bash -c "curl ${DEBUG+-v} -H \"traceparent: \$TRACEPARENT\" -o /dev/null \"${url}\""
)

request_user_homepage "447"
request_todo_list "235"
request_todo_list "11"

# Need to give the otel-cli time to report the spans and shutdown
sleep 2
