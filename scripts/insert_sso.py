"""Insert SSO provider CREATE TABLE into init SQL."""
import sys

init_file = r"D:\Code\open\ydsz-plane\sql\ydsz-plane-init.sql"
prov_file = r"D:\Code\open\ydsz-plane\sql\sso_providers_create.sql"

content = open(init_file, "r", encoding="utf-8").read()
prov = open(prov_file, "r", encoding="utf-8").read()

old = 'DROP TABLE IF EXISTS "public"."sso_providers";\nCREATE TABLE "public"."sso_sessions"'
new = 'DROP TABLE IF EXISTS "public"."sso_providers";\n' + prov + 'CREATE TABLE "public"."sso_sessions"'

if old in content:
    content = content.replace(old, new)
    open(init_file, "w", encoding="utf-8").write(content)
    print("OK: sso_providers table inserted")
else:
    print("WARN: insertion point not found, trying alternative...")
    # Try with just the DROP TABLE
    old2 = 'DROP TABLE IF EXISTS "public"."sso_providers";'
    if old2 in content:
        content = content.replace(old2, old2 + "\n" + prov)
        open(init_file, "w", encoding="utf-8").write(content)
        print("OK: inserted after DROP TABLE")
    else:
        print("FAIL: Could not find sso_providers DROP TABLE")
        sys.exit(1)

# Clean up the temp file
import os
os.remove(prov_file)
print("Cleaned up temp file")
