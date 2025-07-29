"""
Flash Sku Backend - API URLs
API 路由配置，包含所有 REST API 端点
"""

from django.urls import path, include
from rest_framework.routers import DefaultRouter
from .views import (
    CategoryViewSet,
    ProductViewSet,
    SeckillActivityViewSet,
    SystemConfigViewSet,
)

# 创建 DRF 路由器
router = DefaultRouter()

# 注册 ViewSets
router.register(r'categories', CategoryViewSet, basename='category')
router.register(r'products', ProductViewSet, basename='product')
router.register(r'activities', SeckillActivityViewSet, basename='activity')
router.register(r'configs', SystemConfigViewSet, basename='config')

urlpatterns = [
    # DRF 路由
    path('', include(router.urls)),

    # 认证相关 API
    path('auth/', include('apps.users.urls')),

    # 订单相关 API
    path('orders/', include('apps.orders.urls')),
]
