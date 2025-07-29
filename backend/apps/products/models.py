"""
Flash Sku - Products Models
商品管理相关数据模型
"""

from django.db import models
from django.core.validators import MinValueValidator
from django.utils import timezone


class Category(models.Model):
    """商品分类模型"""

    name = models.CharField(
        max_length=100,
        verbose_name='分类名称',
        help_text='商品分类的名称'
    )
    description = models.TextField(
        blank=True,
        verbose_name='分类描述',
        help_text='商品分类的详细描述'
    )
    parent = models.ForeignKey(
        'self',
        on_delete=models.CASCADE,
        null=True,
        blank=True,
        related_name='children',
        verbose_name='父分类',
        help_text='上级分类，为空表示顶级分类'
    )
    sort_order = models.PositiveIntegerField(
        default=0,
        verbose_name='排序',
        help_text='分类显示顺序，数字越小越靠前'
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name='是否启用',
        help_text='是否在前端显示此分类'
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
        verbose_name = '商品分类'
        verbose_name_plural = '商品分类'
        ordering = ['sort_order', 'name']
        indexes = [
            models.Index(fields=['parent', 'is_active']),
            models.Index(fields=['sort_order']),
        ]

    def __str__(self):
        return self.name

    def get_full_name(self):
        """获取完整分类名称（包含父分类）"""
        if self.parent:
            return f"{self.parent.get_full_name()} > {self.name}"
        return self.name

    def get_children(self):
        """获取子分类"""
        return self.children.filter(is_active=True).order_by('sort_order')

    def get_products_count(self):
        """获取该分类下的商品数量"""
        return self.products.filter(is_active=True).count()


class Product(models.Model):
    """商品基础信息模型"""

    name = models.CharField(
        max_length=200,
        verbose_name='商品名称',
        help_text='商品的名称'
    )
    description = models.TextField(
        blank=True,
        verbose_name='商品描述',
        help_text='商品的详细描述'
    )
    image_url = models.URLField(
        max_length=500,
        blank=True,
        verbose_name='商品图片',
        help_text='商品主图的URL地址'
    )
    category = models.ForeignKey(
        Category,
        on_delete=models.SET_NULL,
        null=True,
        blank=True,
        related_name='products',
        verbose_name='商品分类',
        help_text='商品所属分类'
    )
    original_price = models.DecimalField(
        max_digits=10,
        decimal_places=2,
        validators=[MinValueValidator(0)],
        verbose_name='原价',
        help_text='商品的原始价格'
    )
    is_active = models.BooleanField(
        default=True,
        verbose_name='是否上架',
        help_text='商品是否在前端显示'
    )
    sort_order = models.PositiveIntegerField(
        default=0,
        verbose_name='排序',
        help_text='商品显示顺序，数字越小越靠前'
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
        verbose_name = '商品'
        verbose_name_plural = '商品'
        ordering = ['-created_at']
        indexes = [
            models.Index(fields=['category', 'is_active']),
            models.Index(fields=['is_active', 'sort_order']),
            models.Index(fields=['created_at']),
        ]

    def __str__(self):
        return self.name

    def get_active_activities(self):
        """获取该商品的活跃秒杀活动"""
        from apps.activities.models import SeckillActivity
        now = timezone.now()
        return self.seckill_activities.filter(
            status='active',
            start_time__lte=now,
            end_time__gte=now
        )

    def has_active_seckill(self):
        """检查是否有进行中的秒杀活动"""
        return self.get_active_activities().exists()
