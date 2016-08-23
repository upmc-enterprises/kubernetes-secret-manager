FROM vault:0.6.1
MAINTAINER Steve Sloka <steve@stevesloka.com>

ADD myapp.hcl /root/myapp.hcl
ADD vault.conf /root/vault.conf

ADD ./startup.sh /root/startup.sh
RUN chmod +x /root/startup.sh

ENTRYPOINT ["/root/startup.sh"]
