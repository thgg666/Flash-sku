"""
用户认证相关工具函数
包含邮箱激活、验证码生成等功能
"""

import logging
import random
import string
from io import BytesIO
from PIL import Image, ImageDraw, ImageFont
from django.core.mail import send_mail
from django.template.loader import render_to_string
from django.conf import settings
from django.utils.html import strip_tags
from django.urls import reverse
from django.contrib.sites.models import Site

logger = logging.getLogger(__name__)


def generate_verification_code(length=4):
    """
    生成随机验证码

    Args:
        length (int): 验证码长度，默认4位

    Returns:
        str: 生成的验证码
    """
    # 使用不容易混淆的字符
    characters = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
    return ''.join(random.choice(characters) for _ in range(length))


def generate_captcha_image(code, width=120, height=40):
    """
    生成验证码图片
    
    Args:
        code (str): 验证码文本
        width (int): 图片宽度
        height (int): 图片高度
        
    Returns:
        BytesIO: 图片数据流
    """
    # 创建图片
    image = Image.new('RGB', (width, height), color='white')
    draw = ImageDraw.Draw(image)
    
    # 尝试加载字体，如果失败则使用默认字体
    try:
        # 在生产环境中，你可能需要指定字体文件的完整路径
        font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', 24)
    except (OSError, IOError):
        # 如果找不到字体文件，使用默认字体
        font = ImageFont.load_default()
    
    # 绘制验证码文本
    try:
        # 新版本Pillow使用textlength
        text_width = draw.textlength(code, font=font)
    except AttributeError:
        # 旧版本Pillow使用textsize
        text_width, _ = draw.textsize(code, font=font)

    text_height = 24  # 估算的文本高度
    x = (width - text_width) // 2
    y = (height - text_height) // 2
    
    # 为每个字符添加随机颜色和位置偏移
    for i, char in enumerate(code):
        char_x = x + i * (text_width // len(code))
        char_y = y + random.randint(-5, 5)
        color = (
            random.randint(0, 100),
            random.randint(0, 100),
            random.randint(0, 100)
        )
        draw.text((char_x, char_y), char, fill=color, font=font)
    
    # 添加干扰线
    for _ in range(5):
        start = (random.randint(0, width), random.randint(0, height))
        end = (random.randint(0, width), random.randint(0, height))
        draw.line([start, end], fill=(random.randint(100, 200), random.randint(100, 200), random.randint(100, 200)))
    
    # 添加干扰点
    for _ in range(50):
        x = random.randint(0, width)
        y = random.randint(0, height)
        draw.point((x, y), fill=(random.randint(100, 200), random.randint(100, 200), random.randint(100, 200)))
    
    # 保存到内存
    img_buffer = BytesIO()
    image.save(img_buffer, format='PNG')
    img_buffer.seek(0)
    
    return img_buffer


def send_verification_email(user, request=None):
    """
    发送邮箱验证邮件
    
    Args:
        user: Django User对象
        request: HTTP请求对象（可选，用于构建完整URL）
        
    Returns:
        bool: 发送是否成功
    """
    try:
        # 生成验证令牌
        token = user.profile.generate_email_verification_token()
        
        # 构建验证链接
        if request:
            domain = request.get_host()
            protocol = 'https' if request.is_secure() else 'http'
        else:
            # 如果没有request对象，使用默认域名
            try:
                site = Site.objects.get_current()
                domain = site.domain
                protocol = 'https' if settings.DEBUG is False else 'http'
            except:
                domain = 'localhost:8000'
                protocol = 'http'
        
        verification_url = f"{protocol}://{domain}/api/auth/activate/?token={token}&email={user.email}"
        
        # 邮件内容
        context = {
            'user': user,
            'verification_url': verification_url,
            'site_name': 'Flash Sku',
            'token': token,
        }
        
        # 渲染邮件模板
        subject = f'{settings.EMAIL_SUBJECT_PREFIX}请验证您的邮箱地址'
        html_message = render_to_string('emails/email_verification.html', context)
        plain_message = strip_tags(html_message)
        
        # 发送邮件
        send_mail(
            subject=subject,
            message=plain_message,
            from_email=settings.DEFAULT_FROM_EMAIL,
            recipient_list=[user.email],
            html_message=html_message,
            fail_silently=False,
        )
        
        logger.info(f"Verification email sent to {user.email}")
        return True
        
    except Exception as e:
        logger.error(f"Failed to send verification email to {user.email}: {str(e)}")
        return False


def send_password_reset_email(user, reset_url):
    """
    发送密码重置邮件
    
    Args:
        user: Django User对象
        reset_url: 密码重置链接
        
    Returns:
        bool: 发送是否成功
    """
    try:
        context = {
            'user': user,
            'reset_url': reset_url,
            'site_name': 'Flash Sku',
        }
        
        subject = f'{settings.EMAIL_SUBJECT_PREFIX}密码重置请求'
        html_message = render_to_string('emails/password_reset.html', context)
        plain_message = strip_tags(html_message)
        
        send_mail(
            subject=subject,
            message=plain_message,
            from_email=settings.DEFAULT_FROM_EMAIL,
            recipient_list=[user.email],
            html_message=html_message,
            fail_silently=False,
        )
        
        logger.info(f"Password reset email sent to {user.email}")
        return True
        
    except Exception as e:
        logger.error(f"Failed to send password reset email to {user.email}: {str(e)}")
        return False


def validate_password_strength(password):
    """
    验证密码强度
    
    Args:
        password (str): 要验证的密码
        
    Returns:
        tuple: (is_valid, error_messages)
    """
    errors = []
    
    if len(password) < 8:
        errors.append("密码长度至少8位")
    
    if not any(c.isupper() for c in password):
        errors.append("密码必须包含至少一个大写字母")
    
    if not any(c.islower() for c in password):
        errors.append("密码必须包含至少一个小写字母")
    
    if not any(c.isdigit() for c in password):
        errors.append("密码必须包含至少一个数字")
    
    # 检查常见弱密码
    weak_passwords = ['12345678', 'password', 'qwerty123', '11111111']
    if password.lower() in weak_passwords:
        errors.append("密码过于简单，请使用更复杂的密码")
    
    return len(errors) == 0, errors


def is_email_available(email, exclude_user=None):
    """
    检查邮箱是否可用
    
    Args:
        email (str): 要检查的邮箱
        exclude_user: 要排除的用户（用于更新时检查）
        
    Returns:
        bool: 邮箱是否可用
    """
    from django.contrib.auth.models import User
    
    queryset = User.objects.filter(email=email)
    if exclude_user:
        queryset = queryset.exclude(id=exclude_user.id)
    
    return not queryset.exists()


def is_username_available(username, exclude_user=None):
    """
    检查用户名是否可用
    
    Args:
        username (str): 要检查的用户名
        exclude_user: 要排除的用户（用于更新时检查）
        
    Returns:
        bool: 用户名是否可用
    """
    from django.contrib.auth.models import User
    
    queryset = User.objects.filter(username=username)
    if exclude_user:
        queryset = queryset.exclude(id=exclude_user.id)
    
    return not queryset.exists()
