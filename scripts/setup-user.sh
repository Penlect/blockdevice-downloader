#!/bin/sh
set -e
set -u

USERNAME="${1:-user}"
PASSWD="${2:-user}"

adduser --gecos user \
  --disabled-password \
  --shell /usr/bin/fish \
  $USERNAME
adduser $USERNAME sudo
adduser $USERNAME dialout
echo "$USERNAME:$PASSWD" | chpasswd
