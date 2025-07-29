"""
Flash Sku - Orders Models
订单管理相关数据模型
"""

import logging
from django.db import models, transaction
from django.core.validators import MinValueValidator
from django.utils import timezone
from django.contrib.auth import get_user_model
from datetime import timedelta

logger = logging.getLogger(__name__)

User = get_user_model()


class Order(models.Model):
    """订单模型"""

    # 订单状态选择
    STATUS_CHOICES = [
        ('pending_payment', '待支付'),
        ('paid', '已支付'),
        ('cancelled', '已取消'),
        ('refunded', '已退款'),
    ]

    user = models.ForeignKey(
        User,
        on_delete=models.CASCADE,
        related_name='orders',
        verbose_name='用户',
        help_text='下单用户'
    )
    activity = models.ForeignKey(
        'activities.SeckillActivity',
        on_delete=models.CASCADE,
        related_name='orders',
        verbose_name='秒杀活动',
        help_text='关联的秒杀活动'
    )
    product_name = models.CharField(
        max_length=200,
        verbose_name='商品名称',
        help_text='下单时的商品名称（快照）'
    )
    seckill_price = models.DecimalField(
        max_digits=10,
        decimal_places=2,
        validators=[MinValueValidator(0)],
        verbose_name='秒杀价格',
        help_text='下单时的秒杀价格（快照）'
    )
    quantity = models.PositiveIntegerField(
        default=1,
        validators=[MinValueValidator(1)],
        verbose_name='购买数量',
        help_text='购买商品数量'
    )
    total_amount = models.DecimalField(
        max_digits=10,
        decimal_places=2,
        validators=[MinValueValidator(0)],
        verbose_name='订单总额',
        help_text='订单总金额'
    )
    status = models.CharField(
        max_length=20,
        choices=STATUS_CHOICES,
        default='pending_payment',
        verbose_name='订单状态',
        help_text='订单当前状态'
    )
    payment_deadline = models.DateTimeField(
        null=True,
        blank=True,
        verbose_name='支付截止时间',
        help_text='订单支付截止时间，超时将自动取消'
    )
    paid_at = models.DateTimeField(
        null=True,
        blank=True,
        verbose_name='支付时间',
        help_text='订单支付完成时间'
    )
    cancelled_at = models.DateTimeField(
        null=True,
        blank=True,
        verbose_name='取消时间',
        help_text='订单取消时间'
    )
    cancel_reason = models.CharField(
        max_length=200,
        blank=True,
        verbose_name='取消原因',
        help_text='订单取消原因'
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
        verbose_name = '订单'
        verbose_name_plural = '订单'
        ordering = ['-created_at']
        constraints = [
            # 防重复下单：同一用户同一活动只能有一个订单
            models.UniqueConstraint(
                fields=['user', 'activity'],
                name='unique_user_activity_order'
            )
        ]
        indexes = [
            models.Index(fields=['user', 'status']),
            models.Index(fields=['status', 'created_at']),
            models.Index(
                fields=['payment_deadline'],
                condition=models.Q(status='pending_payment'),
                name='idx_orders_deadline_pending'
            ),
        ]

    def __str__(self):
        return f"订单#{self.id} - {self.user.username} - {self.product_name}"

    def save(self, *args, **kwargs):
        """保存时自动计算总金额和支付截止时间"""
        # 计算总金额
        self.total_amount = self.seckill_price * self.quantity

        # 设置支付截止时间（新订单且状态为待支付）
        if not self.pk and self.status == 'pending_payment':
            self.payment_deadline = timezone.now() + timedelta(minutes=30)  # 30分钟支付时限

        super().save(*args, **kwargs)

    def can_pay(self):
        """检查订单是否可以支付"""
        return (
            self.status == 'pending_payment' and
            self.payment_deadline and
            timezone.now() < self.payment_deadline
        )

    def can_cancel(self):
        """检查订单是否可以取消"""
        return self.status in ['pending_payment']

    def mark_paid(self):
        """标记订单为已支付"""
        if self.can_pay():
            self.status = 'paid'
            self.paid_at = timezone.now()
            self.payment_deadline = None
            self.save()
            return True
        return False

    def cancel_order(self, reason='用户取消'):
        """
        取消订单并回滚库存

        Args:
            reason: 取消原因

        Returns:
            bool: 是否成功取消
        """
        if not self.can_cancel():
            return False

        try:
            with transaction.atomic():
                # 使用select_for_update锁定活动记录，防止并发问题
                activity = SeckillActivity.objects.select_for_update().get(id=self.activity_id)

                # 更新订单状态
                self.status = 'cancelled'
                self.cancelled_at = timezone.now()
                self.cancel_reason = reason
                self.save()

                # 安全地回滚库存
                original_stock = activity.available_stock
                activity.available_stock += self.quantity

                # 确保库存不超过总库存
                if activity.available_stock > activity.total_stock:
                    activity.available_stock = activity.total_stock

                activity.save()

                # 记录库存变更日志
                logger.info(f"订单取消库存回滚成功 - 订单ID: {self.id}, "
                          f"活动ID: {activity.id}, 回滚数量: {self.quantity}, "
                          f"原库存: {original_stock}, 新库存: {activity.available_stock}")

                return True

        except Exception as e:
            logger.error(f"订单取消库存回滚失败 - 订单ID: {self.id}, 错误: {str(e)}")
            # 如果库存回滚失败，使用异步任务重试
            from .tasks import rollback_stock
            rollback_stock.delay(self.activity_id, self.quantity, self.id)
            return False

    def is_expired(self):
        """检查订单是否已过期"""
        return (
            self.status == 'pending_payment' and
            self.payment_deadline and
            timezone.now() > self.payment_deadline
        )

    def get_remaining_time(self):
        """获取剩余支付时间（秒）"""
        if self.status == 'pending_payment' and self.payment_deadline:
            remaining = self.payment_deadline - timezone.now()
            return max(0, int(remaining.total_seconds()))
        return 0
