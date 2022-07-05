#!/bin/sh

# logger
log() {
  echo '[' "$(date +"%Y-%m-%d %T")" ']:' INFO "$@"
}

log Detecting the OS of the machine ...
OS=$(uname -s)
log Detected OS: "$OS"

if [ "$OS" = "Linux" ]; then
  # if in case the user have deleted the weeve-agent.service and did not reload the systemd daemon
  sudo systemctl daemon-reload
fi

WEEVE_AGENT_DIR="$PWD/weeve-agent"

SERVICE_FILE=/lib/systemd/system/weeve-agent.service

if [ "$OS" = "Linux" ]; then
  if RESULT=$(systemctl is-active weeve-agent 2>&1); then
    sudo systemctl stop weeve-agent
    sudo systemctl daemon-reload
    log weeve-agent service stopped
  else
    log weeve-agent service not running
  fi

  if [ -f "$SERVICE_FILE" ]; then
    sudo rm "$SERVICE_FILE"
    log "$SERVICE_FILE" removed
  else
    log "$SERVICE_FILE" doesnt exists
  fi
fi

if [ -d "$WEEVE_AGENT_DIR" ] ; then
  sudo rm -r "$WEEVE_AGENT_DIR"
  log "$WEEVE_AGENT_DIR" removed
else
  log "$WEEVE_AGENT_DIR" doesnt exists
fi

log done