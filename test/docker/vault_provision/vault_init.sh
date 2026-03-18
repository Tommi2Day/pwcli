#!/bin/sh
WD=$(dirname "$0")
cd "$WD"
# vault container has no bash
VAULT_ADDR=${VAULT_ADDR:-"http://localhost:8200"}
VAULT_TOKEN=$VAULT_DEV_ROOT_TOKEN_ID
export VAULT_ADDR VAULT_TOKEN
export PGHOST=${PGHOST:-"postgresql"}
export PGPORT=${PGPORT:-5432}

echo "run provision using $VAULT_ADDR and Token $VAULT_TOKEN in $WD"
echo "PGHOST: $PGHOST, PGPORT: $PGPORT"

# engines, kv2 already mounted at secret/
vault secrets list
# vault secrets enable -version=2 -path=secret kv

# static secrets
vault kv put /secret/demo/vault pguser=vault pgpassword=vaultpw pghost=$PGHOST pgport=$PGPORT pgdatabase=demo
vault kv put /secret/demo/owner pguser=demo_o pgpassword=ownerpw pghost=$PGHOST pgport=$PGPORT pgdatabase=demo

# database engine
vault secrets enable database;
vault write database/config/demo \
		plugin_name=postgresql-database-plugin \
		allowed_roles="demo*"  \
		connection_url="postgresql://{{username}}:{{password}}@$PGHOST:$PGPORT/demo?sslmode=disable" \
		username="vault" \
		password="vaultpw" \
		host=$PGHOST \
		port=$GPORT \
		db=demo
vault write database/roles/demo-ro \
		db_name=demo \
        default_ttl=3m max_ttl=10m \
		revocation_stements=" REVOKE demo_ro from \"{{name}}\"; DROP owned by \"{{name}}\"; DROP ROLE \"{{name}}\";" \
		creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';  GRANT demo_ro TO \"{{name}}\";"

vault write database/roles/demo-rw \
		db_name=demo \
        default_ttl=3m max_ttl=10m \
		revocation_stements=" REVOKE demo_rw from \"{{name}}\"; DROP owned by \"{{name}}\"; DROP ROLE \"{{name}}\";" \
		creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';  GRANT demo_rw TO \"{{name}}\";"

# default policies
vault policy write admin "$WD/admin_policy.hcl"
vault policy write provisioner "$WD/provisioner_policy.hcl"
vault policy write database "$WD/database_policy.hcl"
vault policy write demo-owner "$WD/demo_owner_policy.hcl"
vault policy write demo-readonly "$WD/demo_readonly_policy.hcl"
vault policy write demo-readwrite "$WD/demo_readwrite_policy.hcl"

vault auth enable userpass
vault write auth/userpass/users/demo_ro password="demo" policies="demo-readonly,provisioner"
vault write auth/userpass/users/demo_rw password="demo" policies="demo-readwrite,provisioner"
vault write auth/userpass/users/demo_o password="demo" policies="demo-owner,provisioner"
