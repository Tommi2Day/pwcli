
#allow read data
path "secret/data/demo/*" {
  capabilities = [ "create","read","update","delete" ]
}
#list
path "secret/demo/" {
  capabilities = [ "list" ]
}

path "database/" {
  capabilities = [ "read" ]
}

path "database/creds/demo-o" {
  capabilities = [ "read" ]
}

