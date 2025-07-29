"""
Flash Sku - Activities Models
秒杀活动相关数据模型
"""

from django.db import models
from django.core.validators import MinValueValidator
from django.utils import timezone
from django.contrib.auth import get_user_model

User = get_user_model()


class SeckillActivity(models.Model):
    """秒杀活动模型"""

    # 活动状态选择
    STATUS_CHOICES = [
        ('pending', '待开始'),
        ('active', '进行中'),
        ('ended', '已结束'),
        ('cancelled', '已取消'),
    ]

    product = models.ForeignKey(
        'products.Product',
        on_delete=models.CASCADE,
        related_name='seckill_activities',
        verbose_name='关联商品',
        help_text='参与秒杀的商品'
    )
    name = models.CharField(
        max_length=200,
        verbose_name='活动名称',
        help_text='秒杀活动的名称'
    )
    description = models.TextField(
        blank=True,
        verbose_name='活动描述',
        help_text='秒杀活动的详细描述'
    )
    start_time = models.DateTimeField(
        verbose_name='开始时间',
        help_text='秒杀活动开始时间'
    )
    end_time = models.DateTimeField(
        verbose_name='结束时间',
        help_text='秒杀活动结束时间'
    )
    original_price = models.DecimalField(
        max_digits=10,
        decimal_places=2,
        validators=[MinValueValidator(0)],
        verbose_name='原价',
        help_text='商品原始价格'
    )
    seckill_price = models.DecimalField(
        max_digits=10,
        decimal_places=2,
        validators=[MinValueValidator(0)],
        verbose_name='秒杀价',
        help_text='秒杀活动价格'
    )
    total_stock = models.PositiveIntegerField(
        validators=[MinValueValidator(1)],
        verbose_name='总库存',
        help_text='秒杀活动总库存数量'
    )
    available_stock = models.PositiveIntegerField(
        verbose_name='可用库存',
        help_text='当前可用库存数量'
    )
    max_per_user = models.PositiveIntegerField(
        default=1,
        validators=[MinValueValidator(1)],
        verbose_name='每人限购',
        help_text='每个用户最多可购买数量'
    )
    status = models.CharField(
        max_length=20,
        choices=STATUS_CHOICES,
        default='pending',
        verbose_name='活动状态',
        help_text='秒杀活动当前状态'
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
        verbose_name = '秒杀活动'
        verbose_name_plural = '秒杀活动'
        ordering = ['-start_time']
        indexes = [
            models.Index(fields=['status', 'start_time']),
            models.Index(fields=['product', 'status']),
            models.Index(fields=['start_time', 'end_time']),
        ]

    def __str__(self):
        return f"{self.name} - {self.product.name}"

    def clean(self):
        """模型验证"""
        from django.core.exceptions import ValidationError

        if self.start_time and self.end_time:
            if self.start_time >= self.end_time:
                raise ValidationError('开始时间必须早于结束时间')

        if self.available_stock is not None and self.total_stock is not None:
            if self.available_stock > self.total_stock:
                raise ValidationError('可用库存不能大于总库存')

        if self.seckill_price and self.original_price:
            if self.seckill_price >= self.original_price:
                raise ValidationError('秒杀价必须低于原价')

    def save(self, *args, **kwargs):
        """保存时自动更新状态"""
        self.full_clean()

        # 如果是新创建的活动，设置可用库存等于总库存
        if not self.pk and not self.available_stock:
            self.available_stock = self.total_stock

        # 自动更新状态
        self.update_status()

        super().save(*args, **kwargs)

    def update_status(self):
        """根据时间自动更新活动状态"""
        now = timezone.now()

        if self.status == 'cancelled':
            return

        if now < self.start_time:
            self.status = 'pending'
        elif self.start_time <= now <= self.end_time:
            if self.available_stock > 0:
                self.status = 'active'
            else:
                self.status = 'ended'
        else:
            self.status = 'ended'

    def is_active(self):
        """检查活动是否正在进行"""
        now = timezone.now()
        return (
            self.status == 'active' and
            self.start_time <= now <= self.end_time and
            self.available_stock > 0
        )

    def get_discount_percentage(self):
        """计算折扣百分比"""
        if self.original_price > 0:
            discount = (self.original_price - self.seckill_price) / self.original_price
            return round(discount * 100, 1)
        return 0

    def get_sold_count(self):
        """获取已售数量"""
        return self.total_stock - self.available_stock
