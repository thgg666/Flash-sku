"""
Flash Sku - Common Models
公共组件和基础模型
"""

from django.db import models
from django.utils import timezone


class TimeStampedModel(models.Model):
    """时间戳基础模型"""

    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name='创建时间'
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name='更新时间'
    )

    class Meta:
        abstract = True


class SoftDeleteModel(models.Model):
    """软删除基础模型"""

    is_deleted = models.BooleanField(
        default=False,
        verbose_name='是否删除'
    )
    deleted_at = models.DateTimeField(
        null=True,
        blank=True,
        verbose_name='删除时间'
    )

    class Meta:
        abstract = True

    def delete(self, using=None, keep_parents=False):
        """软删除"""
        self.is_deleted = True
        self.deleted_at = timezone.now()
        self.save()

    def hard_delete(self, using=None, keep_parents=False):
        """硬删除"""
        super().delete(using=using, keep_parents=keep_parents)


class BaseModel(TimeStampedModel, SoftDeleteModel):
    """基础模型，包含时间戳和软删除功能"""

    class Meta:
        abstract = True


class SystemConfig(models.Model):
    """系统配置模型"""

    key = models.CharField(
        max_length=100,
        unique=True,
        verbose_name='配置键',
        help_text='配置项的唯一标识'
    )
    value = models.TextField(
        verbose_name='配置值',
        help_text='配置项的值'
    )
    description = models.CharField(
        max_length=200,
        blank=True,
        verbose_name='配置描述',
        help_text='配置项的说明'
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name='是否启用'
    )
    created_at = models.DateTimeField(
        auto_now_add=True,
        verbose_name='创建时间'
    )
    updated_at = models.DateTimeField(
        auto_now=True,
        verbose_name='更新时间'
    )

    class Meta:
        verbose_name = '系统配置'
        verbose_name_plural = '系统配置'
        ordering = ['key']

    def __str__(self):
        return f"{self.key}: {self.value}"

    @classmethod
    def get_config(cls, key, default=None):
        """获取配置值"""
        try:
            config = cls.objects.get(key=key, is_active=True)
            return config.value
        except cls.DoesNotExist:
            return default

    @classmethod
    def set_config(cls, key, value, description=''):
        """设置配置值"""
        config, created = cls.objects.get_or_create(
            key=key,
            defaults={'value': value, 'description': description}
        )
        if not created:
            config.value = value
            if description:
                config.description = description
            config.save()
        return config
