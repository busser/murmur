#!/bin/bash
set -euo pipefail

# TAP (Test Anything Protocol) output
echo "1..3"

# Test 1: Check if extension package exists
EXTENSION_ZIP=$(ls dist/murmur-lambda-extension_*_x86_64.zip 2>/dev/null | head -1 || true)
if [ -z "$EXTENSION_ZIP" ]; then
    echo "not ok 1 - Extension package found"
    exit 1
fi
echo "ok 1 - Extension package found"

# Test 2: Run SAM local invoke
cd lambda/test
if OUTPUT=$(sam local invoke TestFunction --no-event \
    --parameter-overrides "ExtensionPackage=../../$EXTENSION_ZIP" \
    2>/dev/null); then
    echo "ok 2 - SAM local invoke successful"
else
    echo "not ok 2 - SAM local invoke successful"
    exit 1
fi

# Test 3: Validate secrets content
if echo "$OUTPUT" | jq -e '.SECRET_ONE == "test-value-1" and .SECRET_TWO == "test-value-2" and .SECRET_THREE == "hello-world-123"' > /dev/null 2>&1; then
    echo "ok 3 - Extension exported expected secrets"
else
    echo "not ok 3 - Extension exported expected secrets"
    echo "# Output: $(echo "$OUTPUT" | jq -c '.' 2>/dev/null || echo "$OUTPUT")"
    exit 1
fi 