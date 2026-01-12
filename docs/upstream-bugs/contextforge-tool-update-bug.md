# ContextForge Tool Update API Bug

**Status**: Reported
**Severity**: High
**Component**: ContextForge MCP Gateway - Tool Service
**Affects**: terraform-provider-contextforge resource updates
**Discovered**: 2025-11-20

## Summary

The ContextForge Tool Update API endpoint (`PUT /tools/{id}`) fails to update the `name` field (and potentially other fields) in the database, causing the API to return stale/unchanged data after update operations.

## Impact

This bug prevents the Terraform provider from successfully updating tool resources. When a user attempts to change a tool's name or other attributes via Terraform, the update appears to succeed (HTTP 200) but subsequent reads return the old values, causing Terraform to report:

```
Error: Provider produced inconsistent result after apply

When applying changes to contextforge_tool.test, provider
"provider[registry.terraform.io/hashicorp/contextforge]" produced an
unexpected new value: .name: was cty.StringVal("new-name"), but
now cty.StringVal("old-name").
```

### Affected Operations

- Tool name updates
- Tool enabled/disabled state changes
- Potentially other field updates

### Affected Terraform Tests

The following integration tests are skipped due to this bug:

1. `TestAccToolResource_basic` - Cannot verify name updates
2. `TestAccToolResource_complete` - input_schema drift issues
3. `TestAccToolResource_update` - Cannot verify field updates
4. `TestAccToolResource_enabledToggle` - Cannot toggle enabled flag

Tests that still pass:
- `TestAccToolResource_import` - Import by ID works
- `TestAccToolResource_missingRequired` - Validation works

## Root Cause Analysis

### Location

**File**: `mcpgateway/services/tool_service.py`
**Function**: `update_tool()` (lines 1122-1260)

### Bug Description

The `update_tool()` method accepts a `ToolUpdate` schema which includes an optional `name` field (defined in `mcpgateway/schemas.py:761`), but the service method **never updates the `name` field in the database**.

### Code Analysis

**Schema Definition** (`schemas.py:761`):
```python
class ToolUpdate(BaseModelWithConfigDict):
    name: Optional[str] = Field(None, description="Unique name for the tool")
    displayName: Optional[str] = Field(None, description="Display name for the tool")
    custom_name: Optional[str] = Field(None, description="Custom name for the tool")
    # ... other fields ...
```

**Update Method** (`tool_service.py:1176-1213`):
```python
# Check for name change and ensure uniqueness
if tool_update.name and tool_update.name != tool.name:
    # ... validation logic for name conflicts ...
    # ❌ BUG: Missing tool.name = tool_update.name here!

if tool_update.custom_name is not None:
    tool.custom_name = tool_update.custom_name  # ✅ This is updated
if tool_update.displayName is not None:
    tool.display_name = tool_update.displayName  # ✅ This is updated
if tool_update.url is not None:
    tool.url = str(tool_update.url)  # ✅ This is updated
# ... other fields are updated correctly ...

# ❌ BUG: tool.name is never set!
```

The code validates that the new name doesn't conflict with existing tools (lines 1177-1190), but after validation passes, it **forgets to actually update the field**.

### Missing Field Updates

Based on code review, the following fields from `ToolUpdate` schema are **NOT** updated in the database:

1. **`name`** - Never set despite validation logic existing
2. **`enabled`** - Not found in update logic (lines 1192-1225)
3. **`gateway_id`** - Not found in update logic

Fields that ARE correctly updated:
- custom_name, displayName, url, description
- integration_type, request_type, headers
- input_schema, annotations, jsonpath_filter
- visibility, auth, tags

## Reproduction

### Via cURL

```bash
# 1. Create a tool
curl -X POST http://localhost:8000/tools \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": {
      "name": "original-name",
      "description": "Original description",
      "enabled": true
    }
  }'

# Response includes: "id": "abc123", "name": "original-name"

# 2. Update the tool
curl -X PUT http://localhost:8000/tools/abc123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": {
      "name": "updated-name",
      "description": "Updated description",
      "enabled": true
    }
  }'

# Response still shows: "name": "original-name" ❌

# 3. Verify with GET
curl http://localhost:8000/tools/abc123 \
  -H "Authorization: Bearer $TOKEN"

# Still returns: "name": "original-name" ❌
```

### Expected vs Actual

**Expected**: After update, tool name should be "updated-name"
**Actual**: Tool name remains "original-name"

## Suggested Fix

Add missing field updates in `tool_service.py:update_tool()` after line 1190:

```python
# After the name conflict validation block (line ~1190):
if tool_update.name and tool_update.name != tool.name:
    # ... existing validation logic ...
    # ADD THIS LINE:
    tool.name = tool_update.name

# Also add (around line 1213):
if tool_update.enabled is not None:
    tool.enabled = tool_update.enabled

if tool_update.gateway_id is not None:
    tool.gateway_id = tool_update.gateway_id
```

### Complete Patch

```diff
diff --git a/mcpgateway/services/tool_service.py b/mcpgateway/services/tool_service.py
index 1234567..abcdefg 100644
--- a/mcpgateway/services/tool_service.py
+++ b/mcpgateway/services/tool_service.py
@@ -1188,6 +1188,8 @@ class ToolService:
                     ).scalar_one_or_none()
                     if existing_tool:
                         raise ToolNameConflictError(existing_tool.custom_name, enabled=existing_tool.enabled, tool_id=existing_tool.id, visibility=existing_tool.visibility)
+                # UPDATE: Set the new name
+                tool.name = tool_update.name

             if tool_update.custom_name is not None:
                 tool.custom_name = tool_update.custom_name
@@ -1210,6 +1212,12 @@ class ToolService:
                 tool.jsonpath_filter = tool_update.jsonpath_filter
             if tool_update.visibility is not None:
                 tool.visibility = tool_update.visibility
+
+            # UPDATE: Add missing field updates
+            if tool_update.enabled is not None:
+                tool.enabled = tool_update.enabled
+            if tool_update.gateway_id is not None:
+                tool.gateway_id = tool_update.gateway_id

             if tool_update.auth is not None:
                 if tool_update.auth.auth_type is not None:
```

## Workaround

**None available**. The bug requires an upstream fix in ContextForge. Until fixed:

1. Terraform provider must skip update tests for affected fields
2. Users cannot reliably update tool names or enabled states via Terraform
3. Manual updates via UI or direct database access required

## Testing After Fix

After applying the fix, verify with:

```bash
# 1. Run the reproduction steps above - name should update
# 2. Run Terraform provider integration tests:
cd terraform-provider-contextforge
make integration-test-all

# All resource_tool tests should pass:
# - TestAccToolResource_basic
# - TestAccToolResource_complete
# - TestAccToolResource_update
# - TestAccToolResource_enabledToggle
```

## Related Files

### ContextForge Repository
- `mcpgateway/services/tool_service.py:1122-1260` - Bug location
- `mcpgateway/schemas.py:755-776` - ToolUpdate schema definition
- `mcpgateway/main.py:2226-2265` - API endpoint handler

### Terraform Provider Repository
- `internal/provider/resource_tool.go` - Tool resource implementation
- `internal/provider/resource_tool_test.go` - Affected tests (4 skipped)
- `docs/upstream-bugs/contextforge-tool-update-bug.md` - This document

## Additional Notes

### Why the Bug Wasn't Caught

1. **No update tests** in ContextForge test suite for tool name changes
2. **UI may use custom_name** instead of name field, masking the issue
3. **API returns 200 OK** with no error indication

### Similar Bugs?

Should audit other update methods in ContextForge for similar patterns:
- `update_gateway()` - Does it update all fields?
- `update_server()` - Does it update all fields?
- `update_resource()` - Does it update all fields?
- `update_prompt()` - Does it update all fields?

## References

- ContextForge Repository: `~/Dev/external/ibm/mcp-context-forge/`
- Terraform Provider: `~/Dev/leefowlercu/terraform-provider-contextforge/`
- Discovery Issue: Integration test failures during tool resource implementation
