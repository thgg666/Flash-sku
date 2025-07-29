# Flash Sku Backend - Django Project Package
# 这个文件使 backend 目录成为一个 Python 包

# 导入 Celery 应用以确保在 Django 启动时加载
from .celery import app as celery_app

__all__ = ('celery_app',)