#!/bin/bash

# 检查系统是否已经安装了wget
if ! command -v wget >/dev/null; then
    # 检测系统类型
    if command -v apt-get >/dev/null; then
        # Debian/Ubuntu based systems
        echo "Installing wget via apt..."
        sudo apt-get update && sudo apt-get install -y wget
    elif command -v yum >/dev/null; then
        # CentOS/RHEL based systems (assuming older version without dnf)
        echo "Installing wget via yum..."
        sudo yum install -y wget
    elif command -v dnf >/dev/null; then
        # Fedora/CentOS 8 and later
        echo "Installing wget via dnf..."
        sudo dnf install -y wget
    else
        echo "Unable to identify the package manager. Please install wget manually."
        exit 1
    fi

    # 检查安装后是否成功
    if ! command -v wget >/dev/null; then
        echo "Failed to install wget. please install it and try again ."
        exit 1
    fi
fi

if [ ! -f "1pctl" ]; then 
  wget https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/1pctl
fi

if [ ! -f "1panel.service" ]; then 
  wget  https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/1panel.service
fi

if [ ! -f "install.sh" ]; then 
  wget https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/install.sh
fi

if [ ! -f "1paneld" ]; then 
  wget https://raw.githubusercontent.com/gcsong023/wrt_installer/wrt_1panel/etc/init.d/1paneld
fi
chmod 755 1pctl install.sh
