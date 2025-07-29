"""
用户认证相关视图
包含注册、登录、邮箱验证等功能
"""

import logging
from django.shortcuts import render
from django.contrib.auth.models import User
from django.contrib.auth import login
from django.core.cache import cache
from django.http import HttpResponse
from django.utils import timezone
from rest_framework import status, permissions
from rest_framework.decorators import api_view, permission_classes
from rest_framework.response import Response
from rest_framework.views import APIView
from rest_framework_simplejwt.tokens import RefreshToken
from rest_framework_simplejwt.views import TokenObtainPairView, TokenRefreshView
from django_ratelimit.decorators import ratelimit
from django.utils.decorators import method_decorator
from django.views.decorators.csrf import csrf_exempt
from django.conf import settings
from drf_spectacular.utils import extend_schema, OpenApiParameter, OpenApiResponse
from drf_spectacular.openapi import OpenApiTypes

from .serializers import (
    UserRegistrationSerializer,
    UserLoginSerializer,
    EmailVerificationSerializer,
    UserProfileSerializer,
    PasswordChangeSerializer
)
from .utils import generate_verification_code, generate_captcha_image, send_verification_email
from .models import UserProfile

logger = logging.getLogger(__name__)


class CaptchaView(APIView):
    """
    验证码生成API
    GET /api/auth/captcha/
    """
    permission_classes = [permissions.AllowAny]

    @extend_schema(
        summary="获取图片验证码",
        description="生成图片验证码，用于用户注册和登录验证",
        responses={
            200: OpenApiResponse(
                description="验证码图片",
                response=OpenApiTypes.BINARY,
                examples={
                    'image/png': {
                        'summary': '验证码图片',
                        'description': '返回PNG格式的验证码图片，同时在响应头中包含验证码标识'
                    }
                }
            ),
            429: OpenApiResponse(description="请求过于频繁"),
            500: OpenApiResponse(description="验证码生成失败")
        },
        parameters=[
            OpenApiParameter(
                name='X-Captcha-Key',
                type=OpenApiTypes.STR,
                location=OpenApiParameter.HEADER,
                description='验证码唯一标识（响应头）',
                response=True
            )
        ]
    )
    def get(self, request):
        """生成验证码图片"""
        try:
            # 生成验证码
            code = generate_verification_code()

            # 生成唯一标识
            import uuid
            captcha_key = str(uuid.uuid4())

            # 将验证码存储到缓存中，5分钟过期
            cache.set(f'captcha:{captcha_key}', code.upper(), 300)

            # 生成验证码图片
            img_buffer = generate_captcha_image(code)

            # 返回图片和标识
            response = HttpResponse(img_buffer.getvalue(), content_type='image/png')
            response['X-Captcha-Key'] = captcha_key

            logger.info(f"Generated captcha for key: {captcha_key}")
            return response

        except Exception as e:
            logger.error(f"Failed to generate captcha: {str(e)}")
            return Response(
                {'error': '验证码生成失败'},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR
            )


class UserRegistrationView(APIView):
    """
    用户注册API
    POST /api/auth/register/
    """
    permission_classes = [permissions.AllowAny]

    @extend_schema(
        summary="用户注册",
        description="用户注册接口，需要提供用户名、邮箱、密码和验证码",
        request=UserRegistrationSerializer,
        responses={
            201: OpenApiResponse(
                description="注册成功",
                examples={
                    'application/json': {
                        'message': '注册成功',
                        'user_id': 1,
                        'username': 'testuser',
                        'email': 'test@example.com',
                        'email_verification_sent': True
                    }
                }
            ),
            400: OpenApiResponse(description="注册信息有误或验证码错误"),
            429: OpenApiResponse(description="请求过于频繁"),
            500: OpenApiResponse(description="注册失败")
        }
    )
    def post(self, request):
        """用户注册"""
        serializer = UserRegistrationSerializer(data=request.data)

        if not serializer.is_valid():
            return Response(
                {'error': '注册信息有误', 'details': serializer.errors},
                status=status.HTTP_400_BAD_REQUEST
            )

        # 验证验证码
        captcha_key = serializer.validated_data.get('captcha_key')
        captcha_code = serializer.validated_data.get('captcha_code')

        cached_code = cache.get(f'captcha:{captcha_key}')
        if not cached_code or cached_code.upper() != captcha_code.upper():
            return Response(
                {'error': '验证码错误或已过期'},
                status=status.HTTP_400_BAD_REQUEST
            )

        try:
            # 创建用户
            user = serializer.save()

            # 发送验证邮件
            email_sent = send_verification_email(user, request)

            # 删除已使用的验证码
            cache.delete(f'captcha:{captcha_key}')

            logger.info(f"User registered successfully: {user.username}")

            return Response({
                'message': '注册成功',
                'user_id': user.id,
                'username': user.username,
                'email': user.email,
                'email_verification_sent': email_sent
            }, status=status.HTTP_201_CREATED)

        except Exception as e:
            logger.error(f"Registration failed: {str(e)}")
            return Response(
                {'error': '注册失败，请稍后重试'},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR
            )


class UserLoginView(APIView):
    """
    用户登录API
    POST /api/auth/login/
    """
    permission_classes = [permissions.AllowAny]

    @extend_schema(
        summary="用户登录",
        description="用户登录接口，支持用户名或邮箱登录，返回JWT访问令牌",
        request=UserLoginSerializer,
        responses={
            200: OpenApiResponse(
                description="登录成功",
                examples={
                    'application/json': {
                        'message': '登录成功',
                        'access_token': 'eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...',
                        'refresh_token': 'eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...',
                        'user': {
                            'id': 1,
                            'username': 'testuser',
                            'email': 'test@example.com',
                            'is_email_verified': True
                        }
                    }
                }
            ),
            400: OpenApiResponse(description="登录信息有误或需要验证码"),
            429: OpenApiResponse(description="请求过于频繁"),
            500: OpenApiResponse(description="登录失败")
        }
    )
    def post(self, request):
        """用户登录"""
        serializer = UserLoginSerializer(data=request.data)

        if not serializer.is_valid():
            return Response(
                {'error': '登录信息有误', 'details': serializer.errors},
                status=status.HTTP_400_BAD_REQUEST
            )

        user = serializer.validated_data['user']

        # 检查是否需要验证码
        login_failures = getattr(user.profile, 'login_failure_count', 0) if hasattr(user, 'profile') else 0
        if login_failures >= 3:
            captcha_key = request.data.get('captcha_key')
            captcha_code = request.data.get('captcha_code')

            if not captcha_key or not captcha_code:
                return Response(
                    {'error': '登录失败次数过多，请输入验证码', 'require_captcha': True},
                    status=status.HTTP_400_BAD_REQUEST
                )

            cached_code = cache.get(f'captcha:{captcha_key}')
            if not cached_code or cached_code.upper() != captcha_code.upper():
                return Response(
                    {'error': '验证码错误或已过期', 'require_captcha': True},
                    status=status.HTTP_400_BAD_REQUEST
                )

            # 删除已使用的验证码
            cache.delete(f'captcha:{captcha_key}')

        try:
            # 生成JWT令牌
            refresh = RefreshToken.for_user(user)
            access_token = refresh.access_token

            # 更新最后登录时间
            user.last_login = timezone.now()
            user.save()

            # 重置登录失败计数
            if hasattr(user, 'profile'):
                user.profile.reset_login_failures()

            logger.info(f"User logged in successfully: {user.username}")

            return Response({
                'message': '登录成功',
                'access_token': str(access_token),
                'refresh_token': str(refresh),
                'user': {
                    'id': user.id,
                    'username': user.username,
                    'email': user.email,
                    'is_email_verified': getattr(user.profile, 'is_email_verified', False) if hasattr(user, 'profile') else False
                }
            }, status=status.HTTP_200_OK)

        except Exception as e:
            logger.error(f"Login failed for user {user.username}: {str(e)}")
            return Response(
                {'error': '登录失败，请稍后重试'},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR
            )


class EmailVerificationView(APIView):
    """
    邮箱验证API
    POST /api/auth/activate/
    """
    permission_classes = [permissions.AllowAny]

    @extend_schema(
        summary="邮箱验证",
        description="验证用户邮箱地址，激活账户功能",
        request=EmailVerificationSerializer,
        responses={
            200: OpenApiResponse(
                description="邮箱验证成功",
                examples={
                    'application/json': {
                        'message': '邮箱验证成功',
                        'user': {
                            'id': 1,
                            'username': 'testuser',
                            'email': 'test@example.com',
                            'is_email_verified': True
                        }
                    }
                }
            ),
            400: OpenApiResponse(description="验证信息有误"),
            500: OpenApiResponse(description="邮箱验证失败")
        }
    )
    def post(self, request):
        """验证邮箱"""
        serializer = EmailVerificationSerializer(data=request.data)

        if not serializer.is_valid():
            return Response(
                {'error': '验证信息有误', 'details': serializer.errors},
                status=status.HTTP_400_BAD_REQUEST
            )

        try:
            user = serializer.validated_data['user']

            # 验证邮箱
            user.profile.verify_email()

            logger.info(f"Email verified successfully for user: {user.username}")

            return Response({
                'message': '邮箱验证成功',
                'user': {
                    'id': user.id,
                    'username': user.username,
                    'email': user.email,
                    'is_email_verified': True
                }
            }, status=status.HTTP_200_OK)

        except Exception as e:
            logger.error(f"Email verification failed: {str(e)}")
            return Response(
                {'error': '邮箱验证失败，请稍后重试'},
                status=status.HTTP_500_INTERNAL_SERVER_ERROR
            )

    def get(self, request):
        """通过GET请求验证邮箱（用于邮件链接）"""
        email = request.GET.get('email')
        token = request.GET.get('token')

        if not email or not token:
            return Response(
                {'error': '缺少必要参数'},
                status=status.HTTP_400_BAD_REQUEST
            )

        serializer = EmailVerificationSerializer(data={'email': email, 'token': token})

        if not serializer.is_valid():
            return Response(
                {'error': '验证信息有误', 'details': serializer.errors},
                status=status.HTTP_400_BAD_REQUEST
            )

        try:
            user = serializer.validated_data['user']
            user.profile.verify_email()

            logger.info(f"Email verified successfully via GET for user: {user.username}")

            # 返回HTML页面或重定向到前端
            return HttpResponse("""
                <html>
                <head><title>邮箱验证成功</title></head>
                <body>
                    <h1>邮箱验证成功！</h1>
                    <p>您的邮箱已成功验证，现在可以正常使用所有功能了。</p>
                    <p><a href="/">返回首页</a></p>
                </body>
                </html>
            """)

        except Exception as e:
            logger.error(f"Email verification failed via GET: {str(e)}")
            return HttpResponse("""
                <html>
                <head><title>邮箱验证失败</title></head>
                <body>
                    <h1>邮箱验证失败</h1>
                    <p>验证链接无效或已过期，请重新申请验证邮件。</p>
                    <p><a href="/">返回首页</a></p>
                </body>
                </html>
            """)
