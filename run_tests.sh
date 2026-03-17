#!/bin/bash

echo "==============================="
echo "  Running Food Delivery Tests  "
echo "==============================="

PASS=0
FAIL=0

run_tests() {
  SERVICE=$1
  DIR=$2
  TEST_PATH=${3:-"./..."}
  echo ""
  echo "▶ Testing $SERVICE..."
  cd "$DIR" || exit 1

  if go mod tidy > /dev/null 2>&1 && go test "$TEST_PATH" -v -count=1; then
    echo "✅ $SERVICE — All tests passed!"
    PASS=$((PASS + 1))
  else
    echo "❌ $SERVICE — Tests FAILED"
    FAIL=$((FAIL + 1))
  fi
  cd - > /dev/null
}

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

run_tests "Order Service"   "$SCRIPT_DIR/order-service"   "./internal/..."
run_tests "Kitchen Service" "$SCRIPT_DIR/kitchen-service" "./internal/..."
run_tests "Integration Tests" "$SCRIPT_DIR/tests"           "./..."

echo ""
echo "==============================="
echo "  Results: ✅ $PASS passed  ❌ $FAIL failed"
echo "==============================="

[ "$FAIL" -eq 0 ] && exit 0 || exit 1
