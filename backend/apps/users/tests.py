"""
用户认证相关测试
包含注册、登录、邮箱验证等功能的测试
"""

import json
from django.test import TestCase, Client
from django.urls import reverse
from django.contrib.auth.models import User
from django.core.cache import cache
from django.core import mail
from rest_framework.test import APITestCase, APIClient
from rest_framework import status
from rest_framework_simplejwt.tokens import RefreshToken
from .models import UserProfile
from .utils import generate_verification_code


class UserRegistrationTestCase(APITestCase):
    """用户注册测试"""

    def setUp(self):
        self.client = APIClient()
        self.register_url = reverse('users:register')
        self.captcha_url = reverse('users:captcha')

    def test_get_captcha(self):
        """测试获取验证码"""
        response = self.client.get(self.captcha_url)
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertEqual(response['Content-Type'], 'image/png')
        self.assertIn('X-Captcha-Key', response)

    def test_register_success(self):
        """测试成功注册"""
        # 先获取验证码
        captcha_response = self.client.get(self.captcha_url)
        captcha_key = captcha_response['X-Captcha-Key']

        # 从缓存中获取验证码（模拟用户输入）
        cached_code = cache.get(f'captcha:{captcha_key}')

        register_data = {
            'username': 'testuser',
            'email': 'test@example.com',
            'password': 'TestPass123',
            'password_confirm': 'TestPass123',
            'captcha_code': cached_code,
            'captcha_key': captcha_key
        }

        response = self.client.post(self.register_url, register_data)
        self.assertEqual(response.status_code, status.HTTP_201_CREATED)

        # 检查用户是否创建
        self.assertTrue(User.objects.filter(username='testuser').exists())

        # 检查UserProfile是否创建
        user = User.objects.get(username='testuser')
        self.assertTrue(hasattr(user, 'profile'))

        # 检查是否发送了验证邮件
        self.assertEqual(len(mail.outbox), 1)

    def test_register_duplicate_username(self):
        """测试重复用户名注册"""
        # 创建已存在的用户
        User.objects.create_user(username='testuser', email='existing@example.com')

        # 获取验证码
        captcha_response = self.client.get(self.captcha_url)
        captcha_key = captcha_response['X-Captcha-Key']
        cached_code = cache.get(f'captcha:{captcha_key}')

        register_data = {
            'username': 'testuser',
            'email': 'test@example.com',
            'password': 'TestPass123',
            'password_confirm': 'TestPass123',
            'captcha_code': cached_code,
            'captcha_key': captcha_key
        }

        response = self.client.post(self.register_url, register_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('用户名已被使用', str(response.data))

    def test_register_duplicate_email(self):
        """测试重复邮箱注册"""
        # 创建已存在的用户
        User.objects.create_user(username='existing', email='test@example.com')

        # 获取验证码
        captcha_response = self.client.get(self.captcha_url)
        captcha_key = captcha_response['X-Captcha-Key']
        cached_code = cache.get(f'captcha:{captcha_key}')

        register_data = {
            'username': 'testuser',
            'email': 'test@example.com',
            'password': 'TestPass123',
            'password_confirm': 'TestPass123',
            'captcha_code': cached_code,
            'captcha_key': captcha_key
        }

        response = self.client.post(self.register_url, register_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('邮箱已被注册', str(response.data))

    def test_register_weak_password(self):
        """测试弱密码注册"""
        # 获取验证码
        captcha_response = self.client.get(self.captcha_url)
        captcha_key = captcha_response['X-Captcha-Key']
        cached_code = cache.get(f'captcha:{captcha_key}')

        register_data = {
            'username': 'testuser',
            'email': 'test@example.com',
            'password': '123456',  # 弱密码
            'password_confirm': '123456',
            'captcha_code': cached_code,
            'captcha_key': captcha_key
        }

        response = self.client.post(self.register_url, register_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)

    def test_register_password_mismatch(self):
        """测试密码不匹配"""
        # 获取验证码
        captcha_response = self.client.get(self.captcha_url)
        captcha_key = captcha_response['X-Captcha-Key']
        cached_code = cache.get(f'captcha:{captcha_key}')

        register_data = {
            'username': 'testuser',
            'email': 'test@example.com',
            'password': 'TestPass123',
            'password_confirm': 'DifferentPass123',
            'captcha_code': cached_code,
            'captcha_key': captcha_key
        }

        response = self.client.post(self.register_url, register_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('两次输入的密码不一致', str(response.data))

    def test_register_invalid_captcha(self):
        """测试无效验证码"""
        register_data = {
            'username': 'testuser',
            'email': 'test@example.com',
            'password': 'TestPass123',
            'password_confirm': 'TestPass123',
            'captcha_code': 'WRONG',
            'captcha_key': 'invalid-key'
        }

        response = self.client.post(self.register_url, register_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('验证码错误', str(response.data))


class UserLoginTestCase(APITestCase):
    """用户登录测试"""

    def setUp(self):
        self.client = APIClient()
        self.login_url = reverse('users:login')
        self.captcha_url = reverse('users:captcha')

        # 创建测试用户
        self.user = User.objects.create_user(
            username='testuser',
            email='test@example.com',
            password='TestPass123'
        )

    def test_login_success_with_username(self):
        """测试用户名登录成功"""
        login_data = {
            'username': 'testuser',
            'password': 'TestPass123'
        }

        response = self.client.post(self.login_url, login_data)
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertIn('access_token', response.data)
        self.assertIn('refresh_token', response.data)

    def test_login_success_with_email(self):
        """测试邮箱登录成功"""
        login_data = {
            'username': 'test@example.com',
            'password': 'TestPass123'
        }

        response = self.client.post(self.login_url, login_data)
        self.assertEqual(response.status_code, status.HTTP_200_OK)
        self.assertIn('access_token', response.data)
        self.assertIn('refresh_token', response.data)

    def test_login_wrong_password(self):
        """测试错误密码"""
        login_data = {
            'username': 'testuser',
            'password': 'WrongPassword'
        }

        response = self.client.post(self.login_url, login_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('用户名或密码错误', str(response.data))

    def test_login_nonexistent_user(self):
        """测试不存在的用户"""
        login_data = {
            'username': 'nonexistent',
            'password': 'TestPass123'
        }

        response = self.client.post(self.login_url, login_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('用户名或密码错误', str(response.data))

    def test_login_inactive_user(self):
        """测试被禁用的用户"""
        self.user.is_active = False
        self.user.save()

        login_data = {
            'username': 'testuser',
            'password': 'TestPass123'
        }

        response = self.client.post(self.login_url, login_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('账户已被禁用', str(response.data))


class EmailVerificationTestCase(APITestCase):
    """邮箱验证测试"""

    def setUp(self):
        self.client = APIClient()
        self.activate_url = reverse('users:email_verification')

        # 创建测试用户
        self.user = User.objects.create_user(
            username='testuser',
            email='test@example.com',
            password='TestPass123'
        )

    def test_email_verification_success(self):
        """测试邮箱验证成功"""
        # 生成验证令牌
        token = self.user.profile.generate_email_verification_token()

        verify_data = {
            'email': 'test@example.com',
            'token': token
        }

        response = self.client.post(self.activate_url, verify_data)
        self.assertEqual(response.status_code, status.HTTP_200_OK)

        # 检查邮箱是否已验证
        self.user.profile.refresh_from_db()
        self.assertTrue(self.user.profile.is_email_verified)

    def test_email_verification_invalid_token(self):
        """测试无效令牌"""
        verify_data = {
            'email': 'test@example.com',
            'token': 'invalid-token'
        }

        response = self.client.post(self.activate_url, verify_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('验证令牌无效', str(response.data))

    def test_email_verification_nonexistent_user(self):
        """测试不存在的用户"""
        verify_data = {
            'email': 'nonexistent@example.com',
            'token': 'some-token'
        }

        response = self.client.post(self.activate_url, verify_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('用户不存在', str(response.data))

    def test_email_verification_already_verified(self):
        """测试已验证的邮箱"""
        # 先验证邮箱
        self.user.profile.verify_email()

        # 生成新令牌
        token = self.user.profile.generate_email_verification_token()

        verify_data = {
            'email': 'test@example.com',
            'token': token
        }

        response = self.client.post(self.activate_url, verify_data)
        self.assertEqual(response.status_code, status.HTTP_400_BAD_REQUEST)
        self.assertIn('邮箱已经验证过了', str(response.data))


class UserProfileModelTestCase(TestCase):
    """用户资料模型测试"""

    def setUp(self):
        self.user = User.objects.create_user(
            username='testuser',
            email='test@example.com',
            password='TestPass123'
        )

    def test_user_profile_creation(self):
        """测试用户资料自动创建"""
        self.assertTrue(hasattr(self.user, 'profile'))
        self.assertIsInstance(self.user.profile, UserProfile)

    def test_email_verification_token_generation(self):
        """测试邮箱验证令牌生成"""
        token = self.user.profile.generate_email_verification_token()
        self.assertIsNotNone(token)
        self.assertEqual(self.user.profile.email_verification_token, token)
        self.assertIsNotNone(self.user.profile.email_verification_token_created)

    def test_email_verification_token_validation(self):
        """测试邮箱验证令牌验证"""
        token = self.user.profile.generate_email_verification_token()

        # 有效令牌
        self.assertTrue(self.user.profile.is_email_verification_token_valid(token))

        # 无效令牌
        self.assertFalse(self.user.profile.is_email_verification_token_valid('invalid-token'))

    def test_email_verification(self):
        """测试邮箱验证"""
        self.assertFalse(self.user.profile.is_email_verified)

        self.user.profile.verify_email()

        self.assertTrue(self.user.profile.is_email_verified)
        self.assertIsNone(self.user.profile.email_verification_token)

    def test_account_lock_status(self):
        """测试账户锁定状态"""
        # 默认未锁定
        self.assertFalse(self.user.profile.is_account_locked())

        # 设置为被封禁
        self.user.profile.is_banned = True
        self.user.profile.save()

        self.assertTrue(self.user.profile.is_account_locked())

    def test_login_failure_tracking(self):
        """测试登录失败跟踪"""
        # 初始状态
        self.assertEqual(self.user.profile.login_failure_count, 0)

        # 增加失败次数
        self.user.profile.increment_login_failure()
        self.assertEqual(self.user.profile.login_failure_count, 1)

        # 重置失败次数
        self.user.profile.reset_login_failures()
        self.assertEqual(self.user.profile.login_failure_count, 0)
