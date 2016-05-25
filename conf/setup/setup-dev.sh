#!/bin/sh
# vim: syntax=sh tabstop=8 expandtab shiftwidth=4 softtabstop=4
# Setup script for local development
if [ ! -d ./.git ]; then
    echo "You must push your project from the root, so the hooks run properly. Exiting."
    exit 1
fi
./setup-go.sh
./setup-hooks.sh
ln -s ./conf/vagrant/Vagrantfile ./Vagrantfile
