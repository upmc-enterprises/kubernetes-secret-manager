FROM vault:0.6.1
MAINTAINER Steve Sloka <steve@stevesloka.com>

ADD myapp.hcl /etc/myapp.hcl
ADD conf/local.json /vault/config/local.json
ADD ./setup-vault.sh /usr/local/bin/setup-vault.sh
RUN chmod +x /usr/local/bin/setup-vault.sh
