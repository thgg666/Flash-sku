# Flash Sku - Nginx API Gateway Dockerfile
FROM nginx:1.25-alpine

# 安装必要的工具
RUN apk add --no-cache \
    curl \
    wget \
    openssl \
    && rm -rf /var/cache/apk/*

# 创建必要的目录
RUN mkdir -p /var/log/nginx \
    && mkdir -p /var/cache/nginx \
    && mkdir -p /etc/nginx/conf.d

# 复制配置文件
COPY nginx.conf /etc/nginx/nginx.conf

# 创建健康检查脚本
RUN echo '#!/bin/sh' > /usr/local/bin/health-check.sh \
    && echo 'curl -f http://localhost/health || exit 1' >> /usr/local/bin/health-check.sh \
    && chmod +x /usr/local/bin/health-check.sh

# 设置权限
RUN chown -R nginx:nginx /var/log/nginx \
    && chown -R nginx:nginx /var/cache/nginx

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /usr/local/bin/health-check.sh

# 暴露端口
EXPOSE 80

# 启动命令
CMD ["nginx", "-g", "daemon off;"]
