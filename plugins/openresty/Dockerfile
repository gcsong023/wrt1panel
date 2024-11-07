# 使用官方Debian作为基础镜像
FROM debian:buster

# 维护者信息
LABEL maintainer="skyloon.cn"

# 更新apt源并安装必要的系统工具和依赖
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl gnupg2 lsb-release && \
    rm -rf /var/lib/apt/lists/*

# 添加OpenResty的GPG密钥并设置仓库
RUN set -ex; \
    curl -fsSL https://openresty.org/package/debian/pubkey.gpg | apt-key add - && \
    echo "deb http://openresty.org/package/debian $(lsb_release -sc) main" > /etc/apt/sources.list.d/openresty.list

# 更新apt源并安装OpenResty
RUN apt-get update && \
    apt-get install -y --no-install-recommends openresty && \
    ln -sf /usr/local/openresty/nginx/sbin/nginx /usr/local/bin/nginx && \
    rm -rf /var/lib/apt/lists/*

# 将Nginx配置文件复制到容器中（这里仅为示例，您可以根据需要自定义）
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf

# 将自定义Lua脚本或其它必要文件复制到相应目录（如需）

# 暴露默认的HTTP和HTTPS端口
EXPOSE 80 443

# 定义容器启动命令
CMD ["nginx", "-g", "daemon off;"]
