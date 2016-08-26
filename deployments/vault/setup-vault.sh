#/bin/dumb-init /bin/sh

export VAULT_ADDR=http://127.0.0.1:8200
VAULT_ADDR=http://127.0.0.1:8200 vault audit-enable file path=/root/logs/audit.log
VAULT_ADDR=http://127.0.0.1:8200 vault policy-write myapp /etc/myapp.hcl
VAULT_ADDR=http://127.0.0.1:8200 vault write secret/myapp/db host="mysql" port=3306
VAULT_ADDR=http://127.0.0.1:8200 vault mount mysql
VAULT_ADDR=http://127.0.0.1:8200 vault write mysql/config/connection connection_url="root:password@tcp(mysql:3306)/"
VAULT_ADDR=http://127.0.0.1:8200 vault write mysql/config/lease lease=1m lease_max=10m
VAULT_ADDR=http://127.0.0.1:8200 vault write mysql/roles/readonly \
   sql="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';"
VAULT_ADDR=http://127.0.0.1:8200 vault write mysql/roles/fullaccess \
   sql="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT ALL ON *.* TO '{{name}}'@'%';"
