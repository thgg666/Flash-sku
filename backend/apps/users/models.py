from django.db import models
from django.contrib.auth.models import User
from django.db.models.signals import post_save
from django.dispatch import receiver
from django.utils import timezone
import uuid
import logging

logger = logging.getLogger(__name__)


class UserProfile(models.Model):
    """
    用户扩展信息模型
    扩展Django默认User模型，添加秒杀系统所需的额外字段
    """

    user = models.OneToOneField(
        User,
        on_delete=models.CASCADE,
        related_name='profile',
        verbose_name='关联用户'
    )

    # 基础信息
    avatar = models.ImageField(
        upload_to='avatars/%Y/%m/',
        blank=True,
        null=True,
        verbose_name='头像',
        help_text='用户头像图片'
    )

    phone = models.CharField(
        max_length=11,
        blank=True,
        null=True,
        verbose_name='手机号',
        help_text='用户手机号码'
    )

    birth_date = models.DateField(
        blank=True,
        null=True,
        verbose_name='生日',
        help_text='用户生日'
    )

    gender = models.CharField(
        max_length=10,
        choices=[
            ('male', '男'),
            ('female', '女'),
            ('other', '其他'),
        ],
        blank=True,
        null=True,
        verbose_name='性别'
    )

    # 邮箱激活相关
    is_email_verified = models.BooleanField(
        default=False,
        verbose_name='邮箱已验证',
        help_text='用户邮箱是否已通过验证'
    )

    email_verification_token = models.CharField(
        max_length=100,
        blank=True,
        null=True,
        verbose_name='邮箱验证令牌',
        help_text='用于邮箱验证的令牌'
    )

    email_verification_token_created = models.DateTimeField(
        blank=True,
        null=True,
        verbose_name='验证令牌创建时间',
        help_text='邮箱验证令牌的创建时间'
    )

    # 账户状态
    is_banned = models.BooleanField(
        default=False,
        verbose_name='是否被封禁',
        help_text='用户是否被系统封禁'
    )

    ban_reason = models.CharField(
        max_length=200,
        blank=True,
        null=True,
        verbose_name='封禁原因',
        help_text='用户被封禁的原因'
    )

    ban_until = models.DateTimeField(
        blank=True,
        null=True,
        verbose_name='封禁到期时间',
        help_text='用户封禁的到期时间，为空表示永久封禁'
    )

    # 登录相关
    login_failure_count = models.PositiveIntegerField(
        default=0,
        verbose_name='登录失败次数',
        help_text='连续登录失败的次数'
    )

    last_login_failure_time = models.DateTimeField(
        blank=True,
        null=True,
        verbose_name='最后登录失败时间',
        help_text='最后一次登录失败的时间'
    )

    # 时间戳
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name='创建时间'
    )

    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name='更新时间'
    )

    class Meta:
        verbose_name = '用户资料'
        verbose_name_plural = '用户资料'
        db_table = 'user_profiles'

    def __str__(self):
        return f"{self.user.username} - Profile"

    def generate_email_verification_token(self):
        """生成邮箱验证令牌"""
        self.email_verification_token = str(uuid.uuid4())
        self.email_verification_token_created = timezone.now()
        self.save()
        return self.email_verification_token

    def is_email_verification_token_valid(self, token, hours=24):
        """检查邮箱验证令牌是否有效"""
        if not self.email_verification_token or self.email_verification_token != token:
            return False

        if not self.email_verification_token_created:
            return False

        # 检查令牌是否过期（默认24小时）
        expiry_time = self.email_verification_token_created + timezone.timedelta(hours=hours)
        return timezone.now() <= expiry_time

    def verify_email(self):
        """验证邮箱"""
        self.is_email_verified = True
        self.email_verification_token = None
        self.email_verification_token_created = None
        self.save()
        logger.info(f"User {self.user.username} email verified successfully")

    def is_account_locked(self):
        """检查账户是否被锁定"""
        if self.is_banned:
            if self.ban_until is None:  # 永久封禁
                return True
            return timezone.now() < self.ban_until
        return False

    def reset_login_failures(self):
        """重置登录失败计数"""
        self.login_failure_count = 0
        self.last_login_failure_time = None
        self.save()

    def increment_login_failure(self):
        """增加登录失败计数"""
        self.login_failure_count += 1
        self.last_login_failure_time = timezone.now()
        self.save()


@receiver(post_save, sender=User)
def create_user_profile(sender, instance, created, **kwargs):
    """
    当创建User时自动创建UserProfile
    """
    if created:
        UserProfile.objects.create(user=instance)
        logger.info(f"UserProfile created for user: {instance.username}")


@receiver(post_save, sender=User)
def save_user_profile(sender, instance, **kwargs):
    """
    当保存User时确保UserProfile也被保存
    """
    if hasattr(instance, 'profile'):
        instance.profile.save()
    else:
        # 如果由于某种原因profile不存在，创建一个
        UserProfile.objects.create(user=instance)
        logger.warning(f"UserProfile was missing for user {instance.username}, created new one")
