#ldapmodify -Y EXTERNAL -H ldapi:/// <<EOF
slapadd -n 0 <<EOF
dn: olcDatabase={-1}frontend,cn=config
changetype: modify
replace: olcSizeLimit
olcSizeLimit: 2000
EOF