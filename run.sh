#!/bin/sh
set -e
set -u
set -x


test -f /usr/share/OVMF/OVMF_CODE.fd || sudo apt install ovmf
cp /usr/share/OVMF/OVMF_VARS.fd .

kvm -machine q35,smm=on,accel=kvm\
 -global ICH9-LPC.disable_s3=1 \
 -m 1024 \
 -smp 2 \
 -nic user,hostfwd=tcp::2217-:22 \
 -drive if=ide,format=raw,file=blockdevice-downloader_bookworm-amd64.img \
 -drive if=pflash,format=raw,unit=0,file=/usr/share/OVMF/OVMF_CODE.fd,readonly=on \
 -drive if=pflash,format=raw,file=OVMF_VARS.fd

