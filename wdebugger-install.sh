#!/bin/bash

installServer(){
  if [ ! -d /home/wdebugger ];then
    useradd wdebugger
    mkdir -p /home/wdebugger
    chown -R wdebugger:wdebugger /home/wdebugger
  fi
  mkdir -p /home/wdebugger/srv/
  cp -rf * /home/wdebugger/srv/
  if [ ! -f /etc/systemd/system/wdebugger.service ];then
    cp -f wdebugger.service /etc/systemd/system/
  fi
  mkdir -p /etc/wdebugger
  if [ ! -f /etc/wdebugger/wdebugger.json ];then
    cp -f default-wdebugger.json /etc/wdebugger/wdebugger.json
  fi
  systemctl enable wdebugger.service
}

case "$1" in
  -i)
    installServer
    ;;
  *)
    echo "Usage: ./wdebugger-install.sh -i"
    ;;
esac