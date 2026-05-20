#!/bin/bash
set -e

mkdir -p /root/depg/logs/{api-server,game-server,bot-service}

cp infra/logrotate/depg /etc/logrotate.d/depg
chmod 644 /etc/logrotate.d/depg

echo "Logrotate configured. Directories:"
ls -la /root/depg/logs/
