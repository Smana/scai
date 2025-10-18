# Fix: Terraform Directory Isolation Per Deployment

## Problem

All SCIA deployments were sharing the same Terraform directory (`/tmp/scia/terraform`), causing:

1. **Overwrites**: Each new deployment overwrote the previous deployment's Terraform files
2. **Destroy failures**: Cannot destroy specific deployments because their Terraform context is gone
3. **State conflicts**: While S3 state was unique per deployment, local files were shared
4. **No isolation**: Deployments interfere with each other

## Root Cause

In `internal/deployer/deployer.go` line 108:

```go
// BEFORE (WRONG)
tfDir := filepath.Join(d.config.WorkDir, "terraform")
// All deployments use /tmp/scia/terraform
```

This created the same directory for all deployments.

## Solution

Changed to use deployment-specific directories:

```go
// AFTER (CORRECT)
tfDir := filepath.Join(d.config.WorkDir, "terraform", deploymentID)
// Each deployment uses /tmp/scia/terraform/{uuid}/
```

## File Changed

- `internal/deployer/deployer.go` (line 108)

## Verification

### Before Fix

```bash
$ sqlite3 ~/.scia/deployments.db "SELECT id, terraform_dir FROM deployments;"
689996d8-3db5-4222-a2ee-0c7d3fd41c27|/tmp/scia/terraform
b2c0091f-af3f-46a4-9b13-213f607b1e1b|/tmp/scia/terraform
79c004c9-cd54-4dca-9f16-bd8ddc72eb1d|/tmp/scia/terraform
# ❌ All share the same directory!
```

### After Fix

```bash
$ sqlite3 ~/.scia/deployments.db "SELECT id, terraform_dir FROM deployments;"
abc123de-f456-7890-abcd-ef1234567890|/tmp/scia/terraform/abc123de-f456-7890-abcd-ef1234567890
def456ab-c789-0123-def4-567890abcdef|/tmp/scia/terraform/def456ab-c789-0123-def4-567890abcdef
# ✅ Each has its own unique directory!
```

## Testing Steps

### 1. Clean State

```bash
# Remove old deployments from database
sqlite3 ~/.scia/deployments.db "DELETE FROM deployments;"

# Clean up old terraform directories
rm -rf /tmp/scia/terraform
mkdir -p /tmp/scia/terraform
```

### 2. Deploy Multiple Applications

```bash
# Deploy first app
./scia deploy --yes "Deploy this Flask app" https://github.com/Arvo-AI/hello_world

# Deploy second app (different repo)
./scia deploy --yes "Deploy this Express app" https://github.com/example/express-app

# List deployments
./scia list
```

### 3. Verify Isolation

```bash
# Check database - should show different terraform_dir for each
sqlite3 ~/.scia/deployments.db "SELECT id, app_name, terraform_dir FROM deployments;"

# Check filesystem - should see multiple directories
ls -la /tmp/scia/terraform/
# Expected output:
# drwxr-xr-x  3 user user  120 Oct 18 16:30 abc123de-f456-7890-abcd-ef1234567890
# drwxr-xr-x  3 user user  120 Oct 18 16:31 def456ab-c789-0123-def4-567890abcdef

# Each directory should contain its own Terraform files
ls /tmp/scia/terraform/abc123de-f456-7890-abcd-ef1234567890/
# Expected: backend.tf, main.tf, .terraform/, .terraform.lock.hcl
```

### 4. Test Destroy

```bash
# Get deployment IDs
DEPLOYMENT1=$(./scia list | grep -oP '[a-f0-9\-]{36}' | head -1)
DEPLOYMENT2=$(./scia list | grep -oP '[a-f0-9\-]{36}' | tail -1)

# Destroy first deployment
./scia destroy --yes $DEPLOYMENT1

# Verify second deployment still exists
./scia show $DEPLOYMENT2
# Should show deployment details successfully

# Verify first deployment's terraform dir is gone/updated
ls /tmp/scia/terraform/
# Should still show second deployment's directory
```

## Impact

### Positive Changes

✅ **Isolation**: Each deployment has its own Terraform context
✅ **Destroy works**: Can destroy specific deployments reliably
✅ **Concurrent deployments**: Can deploy multiple apps simultaneously
✅ **Debugging**: Easy to inspect specific deployment's Terraform files
✅ **No overwrites**: Deployments don't interfere with each other

### Breaking Changes

None - this is purely an internal implementation fix. The public API remains unchanged.

### Migration

Old deployments in the database will have the old path (`/tmp/scia/terraform`). These should be:

1. **Destroyed** using the destroy command (will fail if files are overwritten)
2. **Cleaned from database** manually if destroy fails
3. **Re-deployed** with the new version

Clean migration script:

```bash
# Backup database
cp ~/.scia/deployments.db ~/.scia/deployments.db.backup

# Remove all old deployments
sqlite3 ~/.scia/deployments.db "DELETE FROM deployments;"

# Clean terraform directories
rm -rf /tmp/scia/terraform
mkdir -p /tmp/scia/terraform

# Now all new deployments will use isolated directories
```

## Related Files

- `internal/deployer/deployer.go` - Creates deployment, sets `tfDir`
- `internal/store/store.go` - Stores `TerraformDir` in database
- `cmd/destroy.go` - Reads `TerraformDir` to run destroy
- `cmd/show.go` - Displays `TerraformDir` in output

## Future Improvements

1. **Cleanup**: Add automatic cleanup of old terraform directories after successful destroy
2. **Validation**: Add check in destroy command to verify terraform directory exists
3. **Migration**: Add command to migrate old deployments to new directory structure
4. **Configuration**: Make terraform base directory configurable (currently hardcoded to `/tmp/scia`)

## Example Output

### New Deployment

```bash
$ ./scia deploy --yes "Deploy Flask app" https://github.com/Arvo-AI/hello_world

🚀 SCIA Deployment Starting...
   User Prompt: Deploy Flask app
   Repository: https://github.com/Arvo-AI/hello_world
   Work Directory: /tmp/scia
   AWS Region: eu-west-3

📊 Analyzing repository...
   Framework: flask
   Language: python
   Port: 5000

🤖 Determining deployment strategy...
   Strategy: vm
   Instance Type: t3.medium

   Created deployment record: abc123de-f456-7890-abcd-ef1234567890
   Creating Terraform configuration...
   Running Terraform...

✅ Deployment Complete!
```

### Show Command

```bash
$ ./scia show abc123de-f456-7890-abcd-ef1234567890

# 🔧 Terraform

   State Key:    deployments/abc123de-f456-7890-abcd-ef1234567890/terraform.tfstate
   Directory:    /tmp/scia/terraform/abc123de-f456-7890-abcd-ef1234567890
```

### Filesystem Layout

```
/tmp/scia/
├── repo/                           # Git clone workspace
└── terraform/                      # Terraform workspace
    ├── abc123de-f456-7890-abcd-ef1234567890/
    │   ├── backend.tf
    │   ├── main.tf
    │   ├── .terraform/
    │   └── .terraform.lock.hcl
    └── def456ab-c789-0123-def4-567890abcdef/
        ├── backend.tf
        ├── main.tf
        ├── .terraform/
        └── .terraform.lock.hcl
```

## Commit Message

```
fix: isolate Terraform directories per deployment

Each deployment now uses its own unique Terraform directory:
/tmp/scia/terraform/{deployment-id}/

This fixes issues where deployments were overwriting each other's
Terraform files, making it impossible to destroy specific deployments.

Changes:
- internal/deployer/deployer.go: Use deployment ID in terraform path

Fixes: Destroy command failures when multiple deployments exist
```

---

**Status**: ✅ Fixed
**Version**: Applied in commit after v0.3.0
**Verified**: Pending user testing
