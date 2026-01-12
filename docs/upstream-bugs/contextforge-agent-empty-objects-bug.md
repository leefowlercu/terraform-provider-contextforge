# ContextForge Agent Service Empty Objects Bug

**Status**: Reported
**Severity**: Medium
**Component**: ContextForge MCP Gateway - A2A Agent Service
**Affects**: terraform-provider-contextforge agent resource
**Discovered**: 2025-11-21

## Summary

The ContextForge A2A Agent Service returns empty JSON objects (`{}`) for `config` and `capabilities` fields instead of `null` when these fields are not provided, and uses a generic update method that causes computed fields to appear as changed. This behavior causes Terraform drift detection issues and integration test failures.

## Impact

This bug prevents the Terraform provider from reliably managing agent resources without detecting spurious drift. When a user creates an agent without specifying `config` or `capabilities`, the API returns empty objects that differ from the Terraform state (null), causing Terraform to report inconsistencies.

```
Error: Provider produced inconsistent result after apply

When applying changes to contextforge_agent.test, provider
"provider[registry.terraform.io/hashicorp/contextforge]" produced an
unexpected new value: .config: was null, but now cty.ObjectVal({}).
```

### Affected Operations

- Agent creation without config/capabilities (drift detected)
- Agent updates (computed fields show as changed)
- Agent imports (all computed fields marked as "known after apply")

### Affected Terraform Tests

The following integration tests are skipped due to this bug:

1. `TestAccAgentResource_basic` - Config drift on create
2. `TestAccAgentResource_withOptionalFields` - Config/capabilities drift
3. `TestAccAgentResource_update` - Computed fields show as changed
4. `TestAccAgentResource_import` - All fields marked as "known after apply"

Tests that still pass:
- `TestAccAgentResource_missingRequired` - Validation works

## Root Cause Analysis

### Issue 1: Empty Objects Instead of Null

#### Location 1: Schema Definition

**File**: `mcpgateway/schemas.py`
**Lines**: 3579-3580 (A2AAgentCreate), 3723-3724 (A2AAgentUpdate)

```python
class A2AAgentCreate(BaseModelWithConfigDict):
    # ... other fields ...
    capabilities: Dict[str, Any] = Field(
        default_factory=dict,  # ❌ Always returns {} instead of null
        description="Agent capabilities and features"
    )
    config: Dict[str, Any] = Field(
        default_factory=dict,  # ❌ Always returns {} instead of null
        description="Agent-specific configuration parameters"
    )
```

#### Location 2: Database Model

**File**: `mcpgateway/db.py`
**Lines**: 2545-2548

```python
class DbA2AAgent(Base):
    # ... other fields ...
    capabilities: Mapped[Dict[str, Any]] = mapped_column(
        JSON,
        default=dict  # ❌ Stores {} in database when not provided
    )
    config: Mapped[Dict[str, Any]] = mapped_column(
        JSON,
        default=dict  # ❌ Stores {} in database when not provided
    )
```

#### The Problem

1. **Schema**: `default_factory=dict` means Pydantic creates an empty dict `{}` when the field is not provided
2. **Database**: `default=dict` means SQLAlchemy stores `{}` when no value is given
3. **API Response**: Agent GET/Create/Update always return:
   ```json
   {
     "capabilities": {},
     "config": {}
   }
   ```
4. **Terraform State**: When user doesn't provide these fields, state has `config = null`
5. **Drift Detection**: Terraform sees null ≠ {} → reports inconsistency

#### Comparison with Working Services

**Gateway Service** (`mcpgateway/db.py:2420`):
```python
capabilities: Mapped[Dict[str, Any]] = mapped_column(JSON)
# ✅ No default specified - nullable by default, returns null when not provided
```

**Result**: Gateway service works correctly with Terraform because it returns `null` for unprovided fields, matching Terraform's expectation.

### Issue 2: Generic Update Method

#### Location

**File**: `mcpgateway/services/a2a_service.py`
**Function**: `update_agent()` (lines 369-433)
**Specific Code**: Lines 411-415

```python
# Generic update approach
update_data = agent_data.model_dump(exclude_unset=True)
for field, value in update_data.items():
    if hasattr(agent, field):
        setattr(agent, field, value)  # ❌ Generic assignment, less control
```

#### The Problem

1. **Generic Loop**: Uses `setattr` to update all fields dynamically
2. **Lack of Explicit Control**: Doesn't distinguish between user-settable and computed fields
3. **Unpredictable Behavior**: After update, computed fields may appear changed even if they weren't modified

#### Comparison with Working Services

**Gateway Service** (`mcpgateway/services/gateway_service.py:1133-1165`):
```python
# Explicit field-by-field updates
if gateway_update.name is not None:
    gateway.name = gateway_update.name  # ✅ Explicit control
    gateway.slug = slugify(gateway_update.name)
if gateway_update.description is not None:
    gateway.description = gateway_update.description  # ✅ Explicit
if gateway_update.url is not None:
    gateway.url = str(gateway_update.url)  # ✅ Explicit
# ... more explicit field checks ...
```

**Server Service** (`mcpgateway/services/server_service.py:737-780`):
```python
# Explicit field-by-field updates
if server_update.name is not None:
    server.name = server_update.name  # ✅ Explicit control
if server_update.description is not None:
    server.description = server_update.description  # ✅ Explicit
# ... more explicit field checks ...
```

**Result**: Gateway and Server services use explicit field-by-field assignments, providing predictable behavior and preventing accidental updates to computed fields.

### Comparison Summary

| Aspect | Gateway Service | Server Service | A2A Agent Service | Issue? |
|--------|----------------|----------------|-------------------|--------|
| **Config field** | No config field | No config field | `default_factory=dict` | ✅ Returns `{}` not `null` |
| **Capabilities** | `nullable=True` (no default) | No capabilities | `default_factory=dict` | ✅ Returns `{}` not `null` |
| **Update method** | Explicit field checks | Explicit field checks | Generic `setattr` loop | ⚠️ Less control |
| **Computed fields** | Well separated | Well separated | Mixed with updates | ⚠️ Can show as changed |
| **Test results** | All tests pass | All tests pass | 4 tests skipped | ✅ Confirmed issues |

## Reproduction

### Via cURL

```bash
# 1. Create an agent WITHOUT config or capabilities
curl -X POST http://localhost:8000/a2a \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": {
      "name": "test-agent",
      "endpoint_url": "http://localhost:9000/agent",
      "description": "Test agent"
    }
  }'

# Response includes:
# "capabilities": {},  ❌ Should be null
# "config": {}         ❌ Should be null

# 2. Get the agent
curl http://localhost:8000/a2a/{agent-id} \
  -H "Authorization: Bearer $TOKEN"

# Still returns:
# "capabilities": {},  ❌
# "config": {}         ❌
```

### Via Terraform

```hcl
resource "contextforge_agent" "test" {
  name         = "test-agent"
  endpoint_url = "http://localhost:9000/agent"
  description  = "Test agent"
  # Note: config and capabilities NOT provided
}
```

**Result**: After `terraform apply`, running `terraform plan` shows:
```
~ resource "contextforge_agent" "test" {
    ~ config = {} -> null
    # Terraform detects drift because API returned {} but state expects null
}
```

### Expected vs Actual

**Expected**:
- API returns `"config": null` when not provided
- API returns `"capabilities": null` when not provided
- Terraform state matches API response (no drift)

**Actual**:
- API returns `"config": {}` when not provided
- API returns `"capabilities": {}` when not provided
- Terraform detects drift (null ≠ {})

## Suggested Fix

### Fix 1: Change Schema and Database Defaults (Preferred)

This fix makes Agent service consistent with Gateway service behavior.

#### File: `mcpgateway/schemas.py`

**Lines 3579-3580** (A2AAgentCreate):
```diff
-capabilities: Dict[str, Any] = Field(default_factory=dict, description="Agent capabilities and features")
-config: Dict[str, Any] = Field(default_factory=dict, description="Agent-specific configuration parameters")
+capabilities: Optional[Dict[str, Any]] = Field(None, description="Agent capabilities and features")
+config: Optional[Dict[str, Any]] = Field(None, description="Agent-specific configuration parameters")
```

**Lines 3723-3724** (A2AAgentUpdate):
```diff
-capabilities: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Agent capabilities and features")
-config: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Agent-specific configuration parameters")
+capabilities: Optional[Dict[str, Any]] = Field(None, description="Agent capabilities and features")
+config: Optional[Dict[str, Any]] = Field(None, description="Agent-specific configuration parameters")
```

#### File: `mcpgateway/db.py`

**Lines 2545-2548**:
```diff
-capabilities: Mapped[Dict[str, Any]] = mapped_column(JSON, default=dict)
-config: Mapped[Dict[str, Any]] = mapped_column(JSON, default=dict)
+capabilities: Mapped[Optional[Dict[str, Any]]] = mapped_column(JSON, nullable=True)
+config: Mapped[Optional[Dict[str, Any]]] = mapped_column(JSON, nullable=True)
```

**Impact**:
- Agent service behavior matches Gateway service
- API returns `null` instead of `{}` for unprovided fields
- No Terraform drift detected
- Existing agents in database: May need migration to convert `{}` → `null`

### Fix 2: Make Update Method Explicit (Recommended)

Replace generic `setattr` loop with explicit field-by-field assignments to match Gateway/Server pattern.

#### File: `mcpgateway/services/a2a_service.py`

**Lines 411-427** (replace entire generic loop):
```diff
-# Update fields
-update_data = agent_data.model_dump(exclude_unset=True)
-for field, value in update_data.items():
-    if hasattr(agent, field):
-        setattr(agent, field, value)
+# Update fields explicitly (matching Gateway/Server pattern)
+if agent_data.name is not None:
+    agent.name = agent_data.name
+if agent_data.description is not None:
+    agent.description = agent_data.description
+if agent_data.endpoint_url is not None:
+    agent.endpoint_url = agent_data.endpoint_url
+if agent_data.agent_type is not None:
+    agent.agent_type = agent_data.agent_type
+if agent_data.protocol_version is not None:
+    agent.protocol_version = agent_data.protocol_version
+if agent_data.capabilities is not None:
+    agent.capabilities = agent_data.capabilities
+if agent_data.config is not None:
+    agent.config = agent_data.config
+if agent_data.auth_type is not None:
+    agent.auth_type = agent_data.auth_type
+if agent_data.auth_value is not None:
+    agent.auth_value = agent_data.auth_value
+if agent_data.tags is not None:
+    agent.tags = agent_data.tags
+if agent_data.team_id is not None:
+    agent.team_id = agent_data.team_id
+if agent_data.owner_email is not None:
+    agent.owner_email = agent_data.owner_email
+if agent_data.visibility is not None:
+    agent.visibility = agent_data.visibility
```

**Impact**:
- More explicit control over field updates
- Prevents accidental updates to computed fields
- Matches proven Gateway/Server patterns
- More maintainable and predictable

### Complete Patch

```diff
diff --git a/mcpgateway/schemas.py b/mcpgateway/schemas.py
index 1234567..abcdefg 100644
--- a/mcpgateway/schemas.py
+++ b/mcpgateway/schemas.py
@@ -3577,8 +3577,8 @@ class A2AAgentCreate(BaseModelWithConfigDict):
     endpoint_url: str = Field(..., description="Agent endpoint URL")
     agent_type: str = Field(default="generic", description="Type of the agent")
     protocol_version: str = Field(default="1.0", description="A2A protocol version")
-    capabilities: Dict[str, Any] = Field(default_factory=dict, description="Agent capabilities and features")
-    config: Dict[str, Any] = Field(default_factory=dict, description="Agent-specific configuration parameters")
+    capabilities: Optional[Dict[str, Any]] = Field(None, description="Agent capabilities and features")
+    config: Optional[Dict[str, Any]] = Field(None, description="Agent-specific configuration parameters")
     auth_type: Optional[str] = Field(None, description="Authentication type")
     auth_value: Optional[str] = Field(None, description="Authentication value")
     tags: List[str] = Field(default_factory=list, description="Agent tags")
@@ -3721,8 +3721,8 @@ class A2AAgentUpdate(BaseModelWithConfigDict):
     endpoint_url: Optional[str] = Field(None, description="Agent endpoint URL")
     agent_type: Optional[str] = Field(None, description="Type of the agent")
     protocol_version: Optional[str] = Field(None, description="A2A protocol version")
-    capabilities: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Agent capabilities and features")
-    config: Optional[Dict[str, Any]] = Field(default_factory=dict, description="Agent-specific configuration parameters")
+    capabilities: Optional[Dict[str, Any]] = Field(None, description="Agent capabilities and features")
+    config: Optional[Dict[str, Any]] = Field(None, description="Agent-specific configuration parameters")
     auth_type: Optional[str] = Field(None, description="Authentication type")
     auth_value: Optional[str] = Field(None, description="Authentication value")
     tags: Optional[List[str]] = Field(None, description="Agent tags")

diff --git a/mcpgateway/db.py b/mcpgateway/db.py
index 2345678..bcdefgh 100644
--- a/mcpgateway/db.py
+++ b/mcpgateway/db.py
@@ -2543,8 +2543,8 @@ class DbA2AAgent(Base):
     endpoint_url: Mapped[str] = mapped_column(String, nullable=False)
     agent_type: Mapped[str] = mapped_column(String, default="generic")
     protocol_version: Mapped[str] = mapped_column(String, default="1.0")
-    capabilities: Mapped[Dict[str, Any]] = mapped_column(JSON, default=dict)
-    config: Mapped[Dict[str, Any]] = mapped_column(JSON, default=dict)
+    capabilities: Mapped[Optional[Dict[str, Any]]] = mapped_column(JSON, nullable=True)
+    config: Mapped[Optional[Dict[str, Any]]] = mapped_column(JSON, nullable=True)
     auth_type: Mapped[Optional[str]] = mapped_column(String)
     auth_value: Mapped[Optional[str]] = mapped_column(Text)
     tags: Mapped[List[str]] = mapped_column(JSON, default=list)

diff --git a/mcpgateway/services/a2a_service.py b/mcpgateway/services/a2a_service.py
index 3456789..cdefghi 100644
--- a/mcpgateway/services/a2a_service.py
+++ b/mcpgateway/services/a2a_service.py
@@ -408,11 +408,30 @@ class A2AService:
                 raise A2AAgentNotFoundError(agent_id)

-            # Update fields
-            update_data = agent_data.model_dump(exclude_unset=True)
-            for field, value in update_data.items():
-                if hasattr(agent, field):
-                    setattr(agent, field, value)
+            # Update fields explicitly (matching Gateway/Server pattern)
+            if agent_data.name is not None:
+                agent.name = agent_data.name
+            if agent_data.description is not None:
+                agent.description = agent_data.description
+            if agent_data.endpoint_url is not None:
+                agent.endpoint_url = agent_data.endpoint_url
+            if agent_data.agent_type is not None:
+                agent.agent_type = agent_data.agent_type
+            if agent_data.protocol_version is not None:
+                agent.protocol_version = agent_data.protocol_version
+            if agent_data.capabilities is not None:
+                agent.capabilities = agent_data.capabilities
+            if agent_data.config is not None:
+                agent.config = agent_data.config
+            if agent_data.auth_type is not None:
+                agent.auth_type = agent_data.auth_type
+            if agent_data.auth_value is not None:
+                agent.auth_value = agent_data.auth_value
+            if agent_data.tags is not None:
+                agent.tags = agent_data.tags
+            if agent_data.team_id is not None:
+                agent.team_id = agent_data.team_id
+            if agent_data.visibility is not None:
+                agent.visibility = agent_data.visibility

             # Update metadata
             if modified_by:
```

## Workaround

### Provider-Side Workaround (Temporary)

If upstream fixes are not immediately available, the Terraform provider can work around this by normalizing empty objects to null:

**File**: `internal/provider/resource_agent.go`
**Function**: `mapAgentToState()` (lines 650-663)

```go
// Map config (map[string]any -> types.Dynamic)
if agent.Config != nil && len(agent.Config) > 0 {  // ← Check if empty
    configValue, err := tfconv.ConvertMapToObjectValue(ctx, agent.Config)
    if err != nil {
        diags.AddError(...)
        return
    }
    data.Config = types.DynamicValue(configValue)
} else {
    data.Config = types.DynamicNull()  // ← Normalize empty {} to null
}

// Same for capabilities
if agent.Capabilities != nil && len(agent.Capabilities) > 0 {
    capabilitiesValue, err := tfconv.ConvertMapToObjectValue(ctx, agent.Capabilities)
    if err != nil {
        diags.AddError(...)
        return
    }
    data.Capabilities = types.DynamicValue(capabilitiesValue)
} else {
    data.Capabilities = types.DynamicNull()
}
```

**Limitations**:
- Doesn't fix root cause
- Can't distinguish between user-provided `{}` and API-defaulted `{}`
- Still shows drift if user explicitly sets `config = {}`

## Testing After Fix

After applying both fixes, verify with:

### 1. Direct API Testing

```bash
# Test 1: Create agent without config/capabilities
curl -X POST http://localhost:8000/a2a \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": {
      "name": "test-agent",
      "endpoint_url": "http://localhost:9000/agent"
    }
  }'

# Expected: Response should have:
# "capabilities": null  ✅
# "config": null        ✅

# Test 2: Update agent
curl -X PUT http://localhost:8000/a2a/{agent-id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "agent": {
      "name": "updated-agent"
    }
  }'

# Expected: Only name should change, computed fields stable ✅
```

### 2. Terraform Provider Integration Tests

```bash
cd terraform-provider-contextforge
make integration-test-all

# All agent resource tests should pass:
# - TestAccAgentResource_basic ✅
# - TestAccAgentResource_withOptionalFields ✅
# - TestAccAgentResource_update ✅
# - TestAccAgentResource_import ✅
# - TestAccAgentResource_missingRequired ✅
```

### 3. Manual Terraform Testing

```hcl
# 1. Create without config
resource "contextforge_agent" "test" {
  name         = "test-agent"
  endpoint_url = "http://localhost:9000/agent"
}

# Run: terraform apply
# Then: terraform plan
# Expected: No changes, no drift ✅

# 2. Add config
resource "contextforge_agent" "test" {
  name         = "test-agent"
  endpoint_url = "http://localhost:9000/agent"
  config = {
    timeout = 30
  }
}

# Run: terraform apply
# Then: terraform plan
# Expected: No changes ✅
```

## Related Files

### ContextForge Repository (~/Dev/external/ibm/mcp-context-forge/)
- `mcpgateway/services/a2a_service.py:369-433` - update_agent() method with generic loop
- `mcpgateway/services/a2a_service.py:637-707` - _db_to_schema() conversion method
- `mcpgateway/schemas.py:3579-3580` - A2AAgentCreate schema (capabilities, config)
- `mcpgateway/schemas.py:3723-3724` - A2AAgentUpdate schema (capabilities, config)
- `mcpgateway/db.py:2545-2548` - DbA2AAgent model (capabilities, config)
- `mcpgateway/main.py:3456-3489` - PUT /a2a/{id} endpoint handler

### Terraform Provider Repository (~/Dev/leefowlercu/terraform-provider-contextforge/)
- `internal/provider/resource_agent.go` - Agent resource implementation
- `internal/provider/resource_agent.go:650-663` - mapAgentToState() config/capabilities handling
- `internal/provider/resource_agent_test.go` - Affected tests (4 skipped)
- `docs/upstream-bugs/contextforge-agent-empty-objects-bug.md` - This document

## Additional Notes

### Why the Bug Wasn't Caught

1. **No Terraform integration tests** in ContextForge repository
2. **UI may handle empty objects differently** - JavaScript treats `{}` and `null` more interchangeably
3. **No comparison with Gateway service pattern** - Gateway works correctly but Agent wasn't reviewed against it
4. **API returns 200 OK** - No error indication, looks successful

### Comparison with Tool Service Bug

**Tool Service Bug** (`contextforge-tool-update-bug.md`):
- **Type**: Fields validated but not assigned (logic error)
- **Impact**: Updates silently fail
- **Severity**: High (data loss)
- **Workaround**: None

**Agent Service Bug** (this document):
- **Type**: Inconsistent null handling (design pattern)
- **Impact**: Terraform drift detection
- **Severity**: Medium (workarounds exist)
- **Workaround**: Provider-side normalization

### Gateway Service as Reference Implementation

The Gateway service demonstrates the correct pattern:
- Uses `nullable=True` for optional JSON fields
- Returns `null` when fields not provided
- Uses explicit field-by-field updates
- All Terraform tests pass

**Recommendation**: Align Agent service with Gateway service patterns.

## References

- ContextForge Repository: `~/Dev/external/ibm/mcp-context-forge/`
- Terraform Provider: `~/Dev/leefowlercu/terraform-provider-contextforge/`
- Related Bug Report: `docs/upstream-bugs/contextforge-tool-update-bug.md`
- Discovery Issue: Integration test failures during agent resource implementation (2025-11-21)
- Working Reference: Gateway service implementation (all tests pass)
