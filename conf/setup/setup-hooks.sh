#!/bin/sh
# vim: syntax=sh tabstop=8 expandtab shiftwidth=4 softtabstop=4
# Setup git hooks for local development
if [ ! -d ./.git ]; then
    echo "You must run this script from the git root! Exiting."
    exit 1
fi

if [ ! -h ../../.git/hooks/commit-msg ]; then
    ln -s ../../conf/git/commit-msg .git/hooks/commit-msg
fi

if [ ! -h ../../.git/hooks/pre-commit ]; then
    ln -s ../../conf/git/pre-commit .git/hooks/pre-commit
fi

if [ ! -h ../../.git/hooks/pre-push ]; then
    ln -s ../../conf/git/pre-push .git/hooks/pre-push
fi
