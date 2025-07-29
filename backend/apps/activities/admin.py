"""
Flash Sku - Activities Admin
秒杀活动管理后台配置
"""

from django.contrib import admin
from django.utils.html import format_html
from django.urls import reverse
from django.utils import timezone
from .models import SeckillActivity


@admin.register(SeckillActivity)
class SeckillActivityAdmin(admin.ModelAdmin):
    """秒杀活动管理"""

    list_display = [
        'name', 'product', 'status_display', 'time_info',
        'price_info', 'stock_info', 'created_at'
    ]
    list_filter = [
        'status', 'start_time', 'end_time', 'created_at',
        ('product', admin.RelatedOnlyFieldListFilter),
    ]
    search_fields = ['name', 'product__name', 'description']
    list_per_page = 20

    fieldsets = (
        ('基本信息', {
            'fields': ('name', 'description', 'product')
        }),
        ('时间设置', {
            'fields': ('start_time', 'end_time')
        }),
        ('价格设置', {
            'fields': ('original_price', 'seckill_price')
        }),
        ('库存设置', {
            'fields': ('total_stock', 'available_stock', 'max_per_user')
        }),
        ('状态信息', {
            'fields': ('status',),
            'classes': ('collapse',)
        }),
    )

    readonly_fields = ['created_at', 'updated_at']

    def status_display(self, obj):
        """状态显示"""
        status_colors = {
            'pending': 'orange',
            'active': 'green',
            'ended': 'gray',
            'cancelled': 'red',
        }
        status_icons = {
            'pending': '⏰',
            'active': '🔥',
            'ended': '✅',
            'cancelled': '❌',
        }
        color = status_colors.get(obj.status, 'black')
        icon = status_icons.get(obj.status, '❓')
        return format_html(
            '<span style="color: {};">{} {}</span>',
            color, icon, obj.get_status_display()
        )
    status_display.short_description = '状态'

    def time_info(self, obj):
        """时间信息"""
        now = timezone.now()
        if obj.start_time > now:
            return format_html(
                '<div>开始: {}<br/>结束: {}</div>',
                obj.start_time.strftime('%m-%d %H:%M'),
                obj.end_time.strftime('%m-%d %H:%M')
            )
        elif obj.end_time > now:
            return format_html(
                '<div style="color: green;">进行中<br/>结束: {}</div>',
                obj.end_time.strftime('%m-%d %H:%M')
            )
        else:
            return format_html(
                '<div style="color: gray;">已结束<br/>{}</div>',
                obj.end_time.strftime('%m-%d %H:%M')
            )
    time_info.short_description = '时间信息'

    def price_info(self, obj):
        """价格信息"""
        discount = obj.get_discount_percentage()
        return format_html(
            '<div>原价: ¥{}<br/>秒杀: <strong style="color: red;">¥{}</strong><br/>折扣: {}%</div>',
            obj.original_price, obj.seckill_price, discount
        )
    price_info.short_description = '价格信息'

    def stock_info(self, obj):
        """库存信息"""
        sold = obj.get_sold_count()
        progress = (sold / obj.total_stock) * 100 if obj.total_stock > 0 else 0

        if obj.available_stock == 0:
            color = 'red'
            status = '售罄'
        elif obj.available_stock < obj.total_stock * 0.2:
            color = 'orange'
            status = '紧张'
        else:
            color = 'green'
            status = '充足'

        return format_html(
            '<div>总量: {}<br/>剩余: <span style="color: {};">{}</span><br/>已售: {} ({:.1f}%)</div>',
            obj.total_stock, color, obj.available_stock, sold, progress
        )
    stock_info.short_description = '库存信息'

    def get_queryset(self, request):
        """优化查询"""
        return super().get_queryset(request).select_related('product')

    actions = ['update_status', 'cancel_activities']

    def update_status(self, request, queryset):
        """更新状态"""
        updated = 0
        for activity in queryset:
            old_status = activity.status
            activity.update_status()
            activity.save()
            if activity.status != old_status:
                updated += 1
        self.message_user(request, f'成功更新 {updated} 个活动状态')
    update_status.short_description = '更新选中活动的状态'

    def cancel_activities(self, request, queryset):
        """取消活动"""
        updated = queryset.filter(status__in=['pending', 'active']).update(status='cancelled')
        self.message_user(request, f'成功取消 {updated} 个活动')
    cancel_activities.short_description = '取消选中的活动'
