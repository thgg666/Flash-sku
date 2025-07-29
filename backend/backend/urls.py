"""
Flash Sku Backend - URL Configuration
主URL配置文件，包含API路由和管理后台
"""
from django.contrib import admin
from django.urls import path, include
from django.conf import settings
from django.conf.urls.static import static
from drf_spectacular.views import (
    SpectacularAPIView,
    SpectacularSwaggerView,
    SpectacularRedocView,
)

# 管理后台配置
admin.site.site_header = 'Flash Sku 管理后台'
admin.site.site_title = 'Flash Sku'
admin.site.index_title = '秒杀系统管理'

urlpatterns = [
    # 管理后台
    path('admin/', admin.site.urls),

    # API 路由
    path('api/v1/', include('api.urls')),

    # API 文档
    path('api/schema/', SpectacularAPIView.as_view(), name='schema'),
    path('api/docs/', SpectacularSwaggerView.as_view(url_name='schema'), name='swagger-ui'),
    path('api/redoc/', SpectacularRedocView.as_view(url_name='schema'), name='redoc'),
]

# 开发环境静态文件服务
if settings.DEBUG:
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
    urlpatterns += static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
