#!/bin/sh
# postrm.sh — post-removal script for .deb packages
if [ "$1" = "purge" ]; then
    rm -rf /etc/kiko/
fi
exit 0
