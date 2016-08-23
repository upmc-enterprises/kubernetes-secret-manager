#!/bin/bash

vault server -config=/root/vault.json

vault audit-enable file path=/root/logs/audit.log
vault policy-write myapp /root/myapp.hcl
vault write secret/myapp/db host="127.0.0.1" port=3306
vault mount mysql
vault write mysql/config/connection connection_url="root:password@tcp(mysql:3306)/"
vault write mysql/config/lease lease=1h lease_max=24h
vault write mysql/roles/readonly \
     sql="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';"

tail -f /dev/null
