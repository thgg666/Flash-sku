"""
用户认证相关装饰器
包含权限检查、邮箱验证等装饰器
"""

import functools
import logging
from django.http import JsonResponse
from django.contrib.auth.models import AnonymousUser
from rest_framework.response import Response
from rest_framework import status

logger = logging.getLogger(__name__)


def login_required(view_func=None, *, message="需要登录才能访问"):
    """
    登录验证装饰器
    确保用户已登录
    """
    def decorator(func):
        @functools.wraps(func)
        def wrapper(request, *args, **kwargs):
            if isinstance(request.user, AnonymousUser) or not request.user.is_authenticated:
                return JsonResponse({
                    'error': message,
                    'code': 'LOGIN_REQUIRED'
                }, status=401)
            return func(request, *args, **kwargs)
        return wrapper
    
    if view_func is None:
        return decorator
    else:
        return decorator(view_func)


def email_verified_required(view_func=None, *, message="请先验证邮箱"):
    """
    邮箱验证装饰器
    确保用户邮箱已验证
    """
    def decorator(func):
        @functools.wraps(func)
        def wrapper(request, *args, **kwargs):
            # 首先检查是否登录
            if isinstance(request.user, AnonymousUser) or not request.user.is_authenticated:
                return JsonResponse({
                    'error': '需要登录才能访问',
                    'code': 'LOGIN_REQUIRED'
                }, status=401)
            
            # 检查邮箱是否已验证
            if hasattr(request.user, 'profile'):
                if not request.user.profile.is_email_verified:
                    return JsonResponse({
                        'error': message,
                        'code': 'EMAIL_NOT_VERIFIED'
                    }, status=403)
            else:
                # 如果没有profile，认为邮箱未验证
                return JsonResponse({
                    'error': message,
                    'code': 'EMAIL_NOT_VERIFIED'
                }, status=403)
            
            return func(request, *args, **kwargs)
        return wrapper
    
    if view_func is None:
        return decorator
    else:
        return decorator(view_func)


def account_active_required(view_func=None, *, message="账户已被禁用"):
    """
    账户状态检查装饰器
    确保账户未被禁用或锁定
    """
    def decorator(func):
        @functools.wraps(func)
        def wrapper(request, *args, **kwargs):
            # 首先检查是否登录
            if isinstance(request.user, AnonymousUser) or not request.user.is_authenticated:
                return JsonResponse({
                    'error': '需要登录才能访问',
                    'code': 'LOGIN_REQUIRED'
                }, status=401)
            
            # 检查账户是否激活
            if not request.user.is_active:
                return JsonResponse({
                    'error': message,
                    'code': 'ACCOUNT_DISABLED'
                }, status=403)
            
            # 检查账户是否被锁定
            if hasattr(request.user, 'profile') and request.user.profile.is_account_locked():
                return JsonResponse({
                    'error': '账户已被锁定',
                    'code': 'ACCOUNT_LOCKED'
                }, status=403)
            
            return func(request, *args, **kwargs)
        return wrapper
    
    if view_func is None:
        return decorator
    else:
        return decorator(view_func)


def admin_required(view_func=None, *, message="需要管理员权限"):
    """
    管理员权限装饰器
    确保用户是管理员
    """
    def decorator(func):
        @functools.wraps(func)
        def wrapper(request, *args, **kwargs):
            # 首先检查是否登录
            if isinstance(request.user, AnonymousUser) or not request.user.is_authenticated:
                return JsonResponse({
                    'error': '需要登录才能访问',
                    'code': 'LOGIN_REQUIRED'
                }, status=401)
            
            # 检查是否是管理员
            if not (request.user.is_staff or request.user.is_superuser):
                return JsonResponse({
                    'error': message,
                    'code': 'ADMIN_REQUIRED'
                }, status=403)
            
            return func(request, *args, **kwargs)
        return wrapper
    
    if view_func is None:
        return decorator
    else:
        return decorator(view_func)


def superuser_required(view_func=None, *, message="需要超级管理员权限"):
    """
    超级管理员权限装饰器
    确保用户是超级管理员
    """
    def decorator(func):
        @functools.wraps(func)
        def wrapper(request, *args, **kwargs):
            # 首先检查是否登录
            if isinstance(request.user, AnonymousUser) or not request.user.is_authenticated:
                return JsonResponse({
                    'error': '需要登录才能访问',
                    'code': 'LOGIN_REQUIRED'
                }, status=401)
            
            # 检查是否是超级管理员
            if not request.user.is_superuser:
                return JsonResponse({
                    'error': message,
                    'code': 'SUPERUSER_REQUIRED'
                }, status=403)
            
            return func(request, *args, **kwargs)
        return wrapper
    
    if view_func is None:
        return decorator
    else:
        return decorator(view_func)


# DRF视图类装饰器
class AuthenticationMixin:
    """
    DRF视图认证混入类
    为DRF视图提供认证相关的方法
    """
    
    def check_authentication(self, require_email_verified=False, require_admin=False):
        """
        检查用户认证状态
        
        Args:
            require_email_verified: 是否需要邮箱验证
            require_admin: 是否需要管理员权限
            
        Returns:
            Response: 如果认证失败返回错误响应，否则返回None
        """
        # 检查是否登录
        if not self.request.user.is_authenticated:
            return Response({
                'error': '需要登录才能访问',
                'code': 'LOGIN_REQUIRED'
            }, status=status.HTTP_401_UNAUTHORIZED)
        
        # 检查账户状态
        if not self.request.user.is_active:
            return Response({
                'error': '账户已被禁用',
                'code': 'ACCOUNT_DISABLED'
            }, status=status.HTTP_403_FORBIDDEN)
        
        # 检查账户是否被锁定
        if hasattr(self.request.user, 'profile') and self.request.user.profile.is_account_locked():
            return Response({
                'error': '账户已被锁定',
                'code': 'ACCOUNT_LOCKED'
            }, status=status.HTTP_403_FORBIDDEN)
        
        # 检查邮箱验证
        if require_email_verified:
            if hasattr(self.request.user, 'profile'):
                if not self.request.user.profile.is_email_verified:
                    return Response({
                        'error': '请先验证邮箱',
                        'code': 'EMAIL_NOT_VERIFIED'
                    }, status=status.HTTP_403_FORBIDDEN)
            else:
                return Response({
                    'error': '请先验证邮箱',
                    'code': 'EMAIL_NOT_VERIFIED'
                }, status=status.HTTP_403_FORBIDDEN)
        
        # 检查管理员权限
        if require_admin:
            if not (self.request.user.is_staff or self.request.user.is_superuser):
                return Response({
                    'error': '需要管理员权限',
                    'code': 'ADMIN_REQUIRED'
                }, status=status.HTTP_403_FORBIDDEN)
        
        return None


def api_permission_required(require_email_verified=False, require_admin=False):
    """
    API权限装饰器（用于DRF视图方法）
    
    Args:
        require_email_verified: 是否需要邮箱验证
        require_admin: 是否需要管理员权限
    """
    def decorator(func):
        @functools.wraps(func)
        def wrapper(self, request, *args, **kwargs):
            # 检查认证
            if hasattr(self, 'check_authentication'):
                auth_response = self.check_authentication(
                    require_email_verified=require_email_verified,
                    require_admin=require_admin
                )
                if auth_response is not None:
                    return auth_response
            else:
                # 如果视图没有继承AuthenticationMixin，使用基础检查
                if not request.user.is_authenticated:
                    return Response({
                        'error': '需要登录才能访问',
                        'code': 'LOGIN_REQUIRED'
                    }, status=status.HTTP_401_UNAUTHORIZED)
                
                if require_email_verified:
                    if hasattr(request.user, 'profile') and not request.user.profile.is_email_verified:
                        return Response({
                            'error': '请先验证邮箱',
                            'code': 'EMAIL_NOT_VERIFIED'
                        }, status=status.HTTP_403_FORBIDDEN)
                
                if require_admin:
                    if not (request.user.is_staff or request.user.is_superuser):
                        return Response({
                            'error': '需要管理员权限',
                            'code': 'ADMIN_REQUIRED'
                        }, status=status.HTTP_403_FORBIDDEN)
            
            return func(self, request, *args, **kwargs)
        return wrapper
    return decorator
