ldapmodify -Y EXTERNAL -H ldapi:/// <<EOF
dn: olcDatabase={0}config,cn=config
changetype: modify
replace: olcAccess
olcAccess: {0}to * by dn.base="gidNumber=0+uidNumber=1001,cn=peercred,cn=external,cn=auth"  by dn.base="cn=admin,dc=example,dc=local" manage by * none
EOF

ldapmodify -Y EXTERNAL -H ldapi:/// <<EOF
dn: olcDatabase={2}mdb,cn=config
changetype: modify
replace: olcAccess
olcAccess: {0}to attrs=userPassword,shadowLastChange,sshPublicKey by self write by dn.base="cn=admin,dc=example,dc=local" write by anonymous auth by * none
olcAccess: {1}to * by self write by dn.base="cn=admin,dc=example,dc=local" write by * read

EOF


