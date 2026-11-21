#!/bin/sh
# Helm mounts the environment's config files at /etc/app; link them next to the binary.
[ -d "/etc/app/" ] && ln -s /etc/app/* .

exec "./app" "$@"
