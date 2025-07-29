"""
Flash Sku - Orders URLs
订单管理相关URL路由
"""

from django.urls import path
from . import views

app_name = 'orders'

urlpatterns = [
    # 秒杀订单创建
    path('seckill/', views.create_seckill_order_view, name='create_seckill_order'),
    
    # 订单状态查询
    path('status/<str:task_id>/', views.check_order_status, name='check_order_status'),
    
    # 用户订单列表
    path('', views.user_orders, name='user_orders'),
    
    # 取消订单
    path('<int:order_id>/cancel/', views.cancel_order, name='cancel_order'),
]
