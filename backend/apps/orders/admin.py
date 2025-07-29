"""
Flash Sku - Orders Admin
订单管理后台配置
"""

from django.contrib import admin
from django.utils.html import format_html
from django.urls import reverse
from django.utils import timezone
from .models import Order


@admin.register(Order)
class OrderAdmin(admin.ModelAdmin):
    """订单管理"""

    list_display = [
        'id', 'user', 'product_name', 'status_display',
        'total_amount', 'payment_info', 'created_at'
    ]
    list_filter = [
        'status', 'created_at', 'paid_at',
        ('user', admin.RelatedOnlyFieldListFilter),
        ('activity', admin.RelatedOnlyFieldListFilter),
    ]
    search_fields = [
        'id', 'user__username', 'user__email',
        'product_name', 'activity__name'
    ]
    list_per_page = 20

    fieldsets = (
        ('订单信息', {
            'fields': ('user', 'activity', 'product_name')
        }),
        ('价格信息', {
            'fields': ('seckill_price', 'quantity', 'total_amount')
        }),
        ('状态信息', {
            'fields': ('status', 'payment_deadline')
        }),
        ('时间记录', {
            'fields': ('paid_at', 'cancelled_at', 'cancel_reason'),
            'classes': ('collapse',)
        }),
    )

    readonly_fields = [
        'total_amount', 'created_at', 'updated_at',
        'paid_at', 'cancelled_at'
    ]

    def status_display(self, obj):
        """状态显示"""
        status_colors = {
            'pending_payment': 'orange',
            'paid': 'green',
            'cancelled': 'red',
            'refunded': 'purple',
        }
        status_icons = {
            'pending_payment': '💰',
            'paid': '✅',
            'cancelled': '❌',
            'refunded': '↩️',
        }
        color = status_colors.get(obj.status, 'black')
        icon = status_icons.get(obj.status, '❓')

        # 检查是否过期
        if obj.is_expired():
            return format_html(
                '<span style="color: red;">{} {} (已过期)</span>',
                icon, obj.get_status_display()
            )

        return format_html(
            '<span style="color: {};">{} {}</span>',
            color, icon, obj.get_status_display()
        )
    status_display.short_description = '订单状态'

    def payment_info(self, obj):
        """支付信息"""
        if obj.status == 'pending_payment':
            if obj.payment_deadline:
                remaining = obj.get_remaining_time()
                if remaining > 0:
                    minutes = remaining // 60
                    seconds = remaining % 60
                    return format_html(
                        '<div style="color: orange;">剩余: {}分{}秒<br/>截止: {}</div>',
                        minutes, seconds,
                        obj.payment_deadline.strftime('%H:%M:%S')
                    )
                else:
                    return format_html(
                        '<div style="color: red;">已过期<br/>{}</div>',
                        obj.payment_deadline.strftime('%m-%d %H:%M')
                    )
            return '待支付'
        elif obj.status == 'paid':
            return format_html(
                '<div style="color: green;">已支付<br/>{}</div>',
                obj.paid_at.strftime('%m-%d %H:%M') if obj.paid_at else '未知'
            )
        elif obj.status == 'cancelled':
            return format_html(
                '<div style="color: red;">已取消<br/>{}</div>',
                obj.cancelled_at.strftime('%m-%d %H:%M') if obj.cancelled_at else '未知'
            )
        return '-'
    payment_info.short_description = '支付信息'

    def get_queryset(self, request):
        """优化查询"""
        return super().get_queryset(request).select_related(
            'user', 'activity', 'activity__product'
        )

    actions = ['mark_as_paid', 'cancel_orders', 'export_orders']

    def mark_as_paid(self, request, queryset):
        """标记为已支付"""
        updated = 0
        for order in queryset.filter(status='pending_payment'):
            if order.mark_paid():
                updated += 1
        self.message_user(request, f'成功标记 {updated} 个订单为已支付')
    mark_as_paid.short_description = '标记选中订单为已支付'

    def cancel_orders(self, request, queryset):
        """取消订单"""
        updated = 0
        for order in queryset.filter(status='pending_payment'):
            if order.cancel_order('管理员取消'):
                updated += 1
        self.message_user(request, f'成功取消 {updated} 个订单')
    cancel_orders.short_description = '取消选中的订单'

    def export_orders(self, request, queryset):
        """导出订单"""
        # 这里可以实现订单导出功能
        count = queryset.count()
        self.message_user(request, f'导出功能开发中，选中了 {count} 个订单')
    export_orders.short_description = '导出选中的订单'
