"""
用户认证相关中间件
包含JWT认证、用户状态检查等功能
"""

import logging
from django.http import JsonResponse
from django.utils.deprecation import MiddlewareMixin
from django.contrib.auth.models import AnonymousUser
from rest_framework_simplejwt.authentication import JWTAuthentication
from rest_framework_simplejwt.exceptions import InvalidToken, TokenError
from rest_framework.request import Request
from django.conf import settings

logger = logging.getLogger(__name__)


class JWTAuthenticationMiddleware(MiddlewareMixin):
    """
    JWT认证中间件
    在请求处理前验证JWT token并设置用户信息
    """
    
    def __init__(self, get_response):
        self.get_response = get_response
        self.jwt_auth = JWTAuthentication()
        super().__init__(get_response)
    
    def process_request(self, request):
        """
        处理请求，验证JWT token
        """
        # 在测试环境中跳过JWT认证
        if settings.TESTING:
            return None

        # 跳过不需要认证的路径
        skip_paths = [
            '/admin/',
            '/api/auth/',  # 更宽泛的匹配
            '/api/v1/auth/',  # 包含版本前缀
            '/api/docs/',
            '/api/redoc/',
            '/api/schema/',
            '/static/',
            '/media/',
        ]

        # 检查是否需要跳过认证
        for path in skip_paths:
            if request.path.startswith(path):
                return None
        
        # 尝试从请求中提取和验证JWT token
        try:
            # 创建DRF Request对象
            drf_request = Request(request)
            
            # 尝试认证
            auth_result = self.jwt_auth.authenticate(drf_request)
            
            if auth_result is not None:
                user, token = auth_result
                
                # 检查用户状态
                if not user.is_active:
                    return JsonResponse({
                        'error': '账户已被禁用',
                        'code': 'ACCOUNT_DISABLED'
                    }, status=401)
                
                # 检查用户是否被锁定
                if hasattr(user, 'profile') and user.profile.is_account_locked():
                    return JsonResponse({
                        'error': '账户已被锁定',
                        'code': 'ACCOUNT_LOCKED'
                    }, status=401)
                
                # 检查邮箱是否已验证（对于需要邮箱验证的API）
                if self._requires_email_verification(request.path):
                    if hasattr(user, 'profile') and not user.profile.is_email_verified:
                        return JsonResponse({
                            'error': '请先验证邮箱',
                            'code': 'EMAIL_NOT_VERIFIED'
                        }, status=403)
                
                # 设置用户信息
                request.user = user
                request.auth = token
                
                logger.debug(f"JWT authentication successful for user: {user.username}")
            else:
                # 没有提供token或token无效，设置为匿名用户
                request.user = AnonymousUser()
                request.auth = None
                
        except (InvalidToken, TokenError) as e:
            logger.warning(f"JWT authentication failed: {str(e)}")
            return JsonResponse({
                'error': 'Token无效或已过期',
                'code': 'INVALID_TOKEN'
            }, status=401)
        except Exception as e:
            logger.error(f"JWT authentication error: {str(e)}")
            return JsonResponse({
                'error': '认证失败',
                'code': 'AUTH_ERROR'
            }, status=500)
        
        return None
    
    def _requires_email_verification(self, path):
        """
        检查路径是否需要邮箱验证
        """
        # 需要邮箱验证的API路径
        email_required_paths = [
            '/api/orders/',
            '/api/seckill/',
            # 可以根据需要添加更多路径
        ]
        
        for required_path in email_required_paths:
            if path.startswith(required_path):
                return True
        
        return False


class UserActivityMiddleware(MiddlewareMixin):
    """
    用户活动记录中间件
    记录用户的最后活动时间和IP地址
    """
    
    def process_response(self, request, response):
        """
        处理响应，更新用户活动信息
        """
        if hasattr(request, 'user') and request.user.is_authenticated:
            try:
                # 更新最后活动时间
                from django.utils import timezone
                request.user.last_login = timezone.now()
                request.user.save(update_fields=['last_login'])
                
                # 如果有profile，可以记录更多信息
                if hasattr(request.user, 'profile'):
                    # 这里可以记录IP地址、用户代理等信息
                    pass
                    
            except Exception as e:
                logger.error(f"Failed to update user activity: {str(e)}")
        
        return response


class SecurityHeadersMiddleware(MiddlewareMixin):
    """
    安全头中间件
    添加安全相关的HTTP头
    """
    
    def process_response(self, request, response):
        """
        添加安全头
        """
        # 防止点击劫持
        response['X-Frame-Options'] = 'DENY'
        
        # 防止MIME类型嗅探
        response['X-Content-Type-Options'] = 'nosniff'
        
        # XSS保护
        response['X-XSS-Protection'] = '1; mode=block'
        
        # 引用策略
        response['Referrer-Policy'] = 'strict-origin-when-cross-origin'
        
        # 内容安全策略（开发环境相对宽松）
        if settings.DEBUG:
            response['Content-Security-Policy'] = "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: https:;"
        else:
            response['Content-Security-Policy'] = "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:;"
        
        return response


class RateLimitMiddleware(MiddlewareMixin):
    """
    简单的速率限制中间件
    基于IP地址进行基础的速率限制
    """
    
    def __init__(self, get_response):
        self.get_response = get_response
        super().__init__(get_response)
    
    def process_request(self, request):
        """
        检查请求速率
        """
        # 获取客户端IP
        ip = self._get_client_ip(request)
        
        # 对于敏感API进行额外的速率限制
        sensitive_paths = [
            '/api/auth/login/',
            '/api/auth/register/',
        ]
        
        for path in sensitive_paths:
            if request.path == path:
                if self._is_rate_limited(ip, path):
                    return JsonResponse({
                        'error': '请求过于频繁，请稍后再试',
                        'code': 'RATE_LIMITED'
                    }, status=429)
        
        return None
    
    def _get_client_ip(self, request):
        """
        获取客户端真实IP地址
        """
        x_forwarded_for = request.META.get('HTTP_X_FORWARDED_FOR')
        if x_forwarded_for:
            ip = x_forwarded_for.split(',')[0].strip()
        else:
            ip = request.META.get('REMOTE_ADDR')
        return ip
    
    def _is_rate_limited(self, ip, path):
        """
        检查是否超过速率限制
        """
        from django.core.cache import cache
        import time
        
        # 构建缓存键
        cache_key = f"rate_limit:{ip}:{path}"
        
        # 获取当前时间窗口内的请求次数
        current_time = int(time.time())
        window_start = current_time - 60  # 1分钟窗口
        
        # 获取请求历史
        requests = cache.get(cache_key, [])
        
        # 过滤掉窗口外的请求
        requests = [req_time for req_time in requests if req_time > window_start]
        
        # 检查是否超过限制
        if len(requests) >= 10:  # 每分钟最多10次请求
            return True
        
        # 记录当前请求
        requests.append(current_time)
        cache.set(cache_key, requests, 60)
        
        return False
