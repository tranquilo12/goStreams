#!/usr/bin/env sh
PGPASSWORD="rogerthat" psql -t -A \
-h "$(tail -1 /etc/resolv.conf | awk 'END {print $NF}')" \
-p "6432" \
-d "TimeScaleDB" \
-U "postgres" \
-c "SELECT 1;";