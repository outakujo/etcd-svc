#! /bin/bash
if [ ! -f pids ]; then
  exit 1
fi
kill $(cat pids)
