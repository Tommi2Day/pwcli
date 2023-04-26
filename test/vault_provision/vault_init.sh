#!/bin/sh

WD=$(dirname "$0")
# vault container has no bash
set -e
VAULT_ADDR="http://localhost:8200"
VAULT_TOKEN=$VAULT_DEV_ROOT_TOKEN_ID
export VAULT_ADDR VAULT_TOKEN

echo "run provision using $VAULT_ADDR and Token $VAULT_TOKEN in $WD"

# engines, kv2 already mounted at secret/
vault secrets list
# vault secrets enable -version=2 -path=secret kv

# default policies
vault policy write admin "$WD/admin_policy.hcl"
vault policy write provisioner "$WD/provisioner_policy.hcl"
