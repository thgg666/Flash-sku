"""
用户认证相关URL配置
"""

from django.urls import path
from rest_framework_simplejwt.views import TokenRefreshView
from . import views

app_name = 'users'

urlpatterns = [
    # 验证码
    path('captcha/', views.CaptchaView.as_view(), name='captcha'),
    
    # 用户注册
    path('register/', views.UserRegistrationView.as_view(), name='register'),
    
    # 用户登录
    path('login/', views.UserLoginView.as_view(), name='login'),
    
    # 邮箱验证
    path('activate/', views.EmailVerificationView.as_view(), name='email_verification'),
    
    # JWT Token刷新
    path('refresh/', TokenRefreshView.as_view(), name='token_refresh'),
]
