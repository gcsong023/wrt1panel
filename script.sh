#!/bin/bash

command -v wget >/dev/null || { 
  echo "wget not found, please install it and try again ."
  exit 1
}

if [ ! -f "1pctl" ]; then 
  wget https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/1pctl
fi

if [ ! -f "1panel.service" ]; then 
  wget  https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/1panel.service
fi

if [ ! -f "install.sh" ]; then 
  wget https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/install.sh
fi

chmod 755 1pctl install.sh
