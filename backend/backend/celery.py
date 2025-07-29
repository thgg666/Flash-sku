"""
Flash Sku Backend - Celery Configuration
Celery 异步任务处理配置文件
"""

import os
from celery import Celery

# 设置默认的 Django settings 模块
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'backend.settings')

# 创建 Celery 应用实例
app = Celery('flashsku')

# 使用 Django 的设置文件配置 Celery
app.config_from_object('django.conf:settings', namespace='CELERY')

# 自动发现任务
app.autodiscover_tasks()

# 调试任务
@app.task(bind=True)
def debug_task(self):
    """调试任务，用于测试 Celery 是否正常工作"""
    print(f'Request: {self.request!r}')


# 确保Django应用在Celery启动时被加载
@app.on_after_configure.connect
def setup_periodic_tasks(sender, **kwargs):
    """设置周期性任务"""
    # 这里可以添加周期性任务的配置
    pass
