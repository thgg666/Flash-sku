"""
用户认证相关序列化器
包含注册、登录、邮箱验证等功能的序列化器
"""

from rest_framework import serializers
from django.contrib.auth.models import User
from django.contrib.auth import authenticate
from django.core.validators import validate_email
from django.core.exceptions import ValidationError
from .models import UserProfile
from .utils import validate_password_strength, is_email_available, is_username_available
import re


class UserRegistrationSerializer(serializers.Serializer):
    """用户注册序列化器"""
    
    username = serializers.CharField(
        max_length=150,
        min_length=3,
        help_text="用户名，3-150个字符"
    )
    email = serializers.EmailField(
        help_text="邮箱地址"
    )
    password = serializers.CharField(
        min_length=8,
        max_length=128,
        write_only=True,
        help_text="密码，至少8位"
    )
    password_confirm = serializers.CharField(
        min_length=8,
        max_length=128,
        write_only=True,
        help_text="确认密码"
    )
    captcha_code = serializers.CharField(
        max_length=10,
        write_only=True,
        help_text="验证码"
    )
    captcha_key = serializers.CharField(
        max_length=100,
        write_only=True,
        help_text="验证码标识"
    )
    
    def validate_username(self, value):
        """验证用户名"""
        # 检查用户名格式
        if not re.match(r'^[a-zA-Z0-9_]+$', value):
            raise serializers.ValidationError("用户名只能包含字母、数字和下划线")
        
        # 检查用户名是否可用
        if not is_username_available(value):
            raise serializers.ValidationError("用户名已被使用")
        
        return value
    
    def validate_email(self, value):
        """验证邮箱"""
        # Django内置邮箱格式验证
        try:
            validate_email(value)
        except ValidationError:
            raise serializers.ValidationError("邮箱格式不正确")
        
        # 检查邮箱是否可用
        if not is_email_available(value):
            raise serializers.ValidationError("邮箱已被注册")
        
        return value
    
    def validate_password(self, value):
        """验证密码强度"""
        is_valid, errors = validate_password_strength(value)
        if not is_valid:
            raise serializers.ValidationError(errors)
        return value
    
    def validate(self, attrs):
        """验证整体数据"""
        # 验证密码确认
        if attrs['password'] != attrs['password_confirm']:
            raise serializers.ValidationError({
                'password_confirm': '两次输入的密码不一致'
            })
        
        # 验证码验证将在视图中处理
        return attrs
    
    def create(self, validated_data):
        """创建用户"""
        # 移除不需要保存的字段
        validated_data.pop('password_confirm')
        validated_data.pop('captcha_code')
        validated_data.pop('captcha_key')
        
        # 创建用户
        user = User.objects.create_user(
            username=validated_data['username'],
            email=validated_data['email'],
            password=validated_data['password'],
            is_active=True  # 用户创建后即激活，但邮箱需要验证
        )
        
        return user


class UserLoginSerializer(serializers.Serializer):
    """用户登录序列化器"""
    
    username = serializers.CharField(
        max_length=150,
        help_text="用户名或邮箱"
    )
    password = serializers.CharField(
        max_length=128,
        write_only=True,
        help_text="密码"
    )
    captcha_code = serializers.CharField(
        max_length=10,
        write_only=True,
        required=False,
        help_text="验证码（登录失败多次后需要）"
    )
    captcha_key = serializers.CharField(
        max_length=100,
        write_only=True,
        required=False,
        help_text="验证码标识"
    )
    
    def validate(self, attrs):
        """验证登录信息"""
        username = attrs.get('username')
        password = attrs.get('password')
        
        if not username or not password:
            raise serializers.ValidationError('用户名和密码不能为空')
        
        # 尝试通过用户名或邮箱查找用户
        user_obj = None
        if '@' in username:
            # 邮箱登录
            try:
                user_obj = User.objects.get(email=username)
            except User.DoesNotExist:
                raise serializers.ValidationError('用户名或密码错误')
        else:
            # 用户名登录
            try:
                user_obj = User.objects.get(username=username)
            except User.DoesNotExist:
                raise serializers.ValidationError('用户名或密码错误')

        # 检查密码
        if not user_obj.check_password(password):
            raise serializers.ValidationError('用户名或密码错误')

        # 检查账户状态
        if not user_obj.is_active:
            raise serializers.ValidationError('账户已被禁用')

        user = user_obj
        
        # 检查账户是否被锁定
        if hasattr(user, 'profile') and user.profile.is_account_locked():
            raise serializers.ValidationError('账户已被锁定，请联系管理员')
        
        attrs['user'] = user
        return attrs


class EmailVerificationSerializer(serializers.Serializer):
    """邮箱验证序列化器"""
    
    email = serializers.EmailField(
        help_text="邮箱地址"
    )
    token = serializers.CharField(
        max_length=100,
        help_text="验证令牌"
    )
    
    def validate(self, attrs):
        """验证邮箱验证信息"""
        email = attrs.get('email')
        token = attrs.get('token')
        
        try:
            user = User.objects.get(email=email)
        except User.DoesNotExist:
            raise serializers.ValidationError('用户不存在')
        
        if not hasattr(user, 'profile'):
            raise serializers.ValidationError('用户资料不存在')
        
        if user.profile.is_email_verified:
            raise serializers.ValidationError('邮箱已经验证过了')
        
        if not user.profile.is_email_verification_token_valid(token):
            raise serializers.ValidationError('验证令牌无效或已过期')
        
        attrs['user'] = user
        return attrs


class UserProfileSerializer(serializers.ModelSerializer):
    """用户资料序列化器"""
    
    username = serializers.CharField(source='user.username', read_only=True)
    email = serializers.EmailField(source='user.email', read_only=True)
    date_joined = serializers.DateTimeField(source='user.date_joined', read_only=True)
    last_login = serializers.DateTimeField(source='user.last_login', read_only=True)
    
    class Meta:
        model = UserProfile
        fields = [
            'username', 'email', 'avatar', 'phone', 'birth_date', 'gender',
            'is_email_verified', 'is_banned', 'date_joined', 'last_login',
            'created_at', 'updated_at'
        ]
        read_only_fields = [
            'username', 'email', 'is_email_verified', 'is_banned',
            'date_joined', 'last_login', 'created_at', 'updated_at'
        ]


class PasswordChangeSerializer(serializers.Serializer):
    """密码修改序列化器"""
    
    old_password = serializers.CharField(
        max_length=128,
        write_only=True,
        help_text="当前密码"
    )
    new_password = serializers.CharField(
        min_length=8,
        max_length=128,
        write_only=True,
        help_text="新密码"
    )
    new_password_confirm = serializers.CharField(
        min_length=8,
        max_length=128,
        write_only=True,
        help_text="确认新密码"
    )
    
    def validate_old_password(self, value):
        """验证当前密码"""
        user = self.context['request'].user
        if not user.check_password(value):
            raise serializers.ValidationError('当前密码错误')
        return value
    
    def validate_new_password(self, value):
        """验证新密码强度"""
        is_valid, errors = validate_password_strength(value)
        if not is_valid:
            raise serializers.ValidationError(errors)
        return value
    
    def validate(self, attrs):
        """验证整体数据"""
        if attrs['new_password'] != attrs['new_password_confirm']:
            raise serializers.ValidationError({
                'new_password_confirm': '两次输入的新密码不一致'
            })
        
        if attrs['old_password'] == attrs['new_password']:
            raise serializers.ValidationError({
                'new_password': '新密码不能与当前密码相同'
            })
        
        return attrs
    
    def save(self):
        """保存新密码"""
        user = self.context['request'].user
        user.set_password(self.validated_data['new_password'])
        user.save()
        
        # 重置登录失败计数
        if hasattr(user, 'profile'):
            user.profile.reset_login_failures()
        
        return user
