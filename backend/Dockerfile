# Flash Sku Backend - Django Dockerfile
# 基于 Python 3.11 的 Django 应用容器

FROM python:3.11-slim

# 设置环境变量
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    DEBIAN_FRONTEND=noninteractive

# 设置工作目录
WORKDIR /app

# 更换为清华镜像源并安装系统依赖
RUN sed -i 's/deb.debian.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apt/sources.list.d/debian.sources && \
    apt-get update && apt-get install -y \
    build-essential \
    libpq-dev \
    gettext \
    && rm -rf /var/lib/apt/lists/*

# 复制依赖文件
COPY requirements.txt .

# 升级pip并配置超时，一次性安装所有依赖
RUN pip install --upgrade pip && \
    pip config set global.timeout 300 && \
    pip install --no-cache-dir -i https://mirrors.aliyun.com/pypi/simple \
    Django==4.2.7 \
    djangorestframework==3.14.0 \
    django-cors-headers==4.3.1 \
    dj-database-url==2.1.0 \
    django-filter==23.5 \
    drf-spectacular==0.26.5 \
    -r requirements.txt

# 复制项目文件
COPY . .

# 创建必要的目录
RUN mkdir -p /app/media /app/static /app/logs

# 设置权限（如果scripts目录和sh文件存在）
RUN if [ -d "/app/scripts" ] && [ -n "$(ls -A /app/scripts/*.sh 2>/dev/null)" ]; then chmod +x /app/scripts/*.sh; fi

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python manage.py check || exit 1

# 默认命令
CMD ["python", "manage.py", "runserver", "0.0.0.0:8000"]