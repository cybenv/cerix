#!/bin/zsh

mkdir -p logs
COMPACT="logs/mullvad-all-servers-speed.log"
FULL="logs/mullvad-all-servers-speed-full.log"
BLOCK="$(mktemp)"
trap 'rm -f "$BLOCK"' EXIT

: > "$COMPACT"
: > "$FULL"

AWK_FILTER='
/^== (IdleLatency|Download|Upload|PacketLoss) ==/ { skip=1; next }
/^== Summary ==/                                   { skip=0 }
skip                                               { next }
/^(Relay constraints updated|Connecting to |    Relay:|    Features:|    Visible location:|Measuring |External IPs:)/ { next }
{ print }
'

emit_block() {
  cat "$BLOCK" >> "$FULL"
  awk "$AWK_FILTER" "$BLOCK" >> "$COMPACT"
  : > "$BLOCK"
}

SERVERS=("${(@f)$(mullvad relay list | awk '/-wg-/ {print $1}')}")

echo "mullvad version: $(mullvad --version 2>&1 | head -n1)" | tee -a "$BLOCK"
echo "cloudflare-speed-cli version: $(cloudflare-speed-cli --version 2>&1 | head -n1)" | tee -a "$BLOCK"
echo "Total servers:  ${#SERVERS[@]}" | tee -a "$BLOCK"
echo "Start: $(date)" | tee -a "$BLOCK"
echo "" | tee -a "$BLOCK"
emit_block

for SERVER in "${SERVERS[@]}"; do
  echo "=== $SERVER | $(date) ===" | tee -a "$BLOCK"

  mullvad disconnect >/dev/null 2>&1
  sleep 2

  mullvad relay set location "$SERVER" 2>&1 | tee -a "$BLOCK"
  mullvad connect 2>&1 | tee -a "$BLOCK"

  echo "Connecting to $SERVER..." | tee -a "$BLOCK"

  CONNECTED=0
  for i in {1..45}; do
    STATUS="$(mullvad status 2>&1)"
    echo "$STATUS" | grep -q "Connected" || { sleep 1; continue; }
    echo "$STATUS" | grep -q "Relay: *$SERVER" || { sleep 1; continue; }
    CONNECTED=1
    break
  done

  if [[ $CONNECTED -ne 1 ]]; then
    echo "Error: failed to confirm connection to $SERVER" | tee -a "$BLOCK"
    echo "$STATUS" | tee -a "$BLOCK"
    echo "" | tee -a "$BLOCK"
    emit_block
    continue
  fi

  echo "$STATUS" | tee -a "$BLOCK"

  cloudflare-speed-cli --text 2>&1 | tee -a "$BLOCK"

  echo "" | tee -a "$BLOCK"
  emit_block
  sleep 2
done

echo "End: $(date)" | tee -a "$BLOCK"
emit_block
