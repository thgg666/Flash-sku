"""
Flash Sku - Products Admin
商品管理后台配置
"""

from django.contrib import admin
from django.utils.html import format_html
from django.urls import reverse
from django.utils.safestring import mark_safe
from .models import Category, Product


@admin.register(Category)
class CategoryAdmin(admin.ModelAdmin):
    """商品分类管理"""

    list_display = [
        'name', 'parent', 'products_count', 'sort_order',
        'is_active', 'created_at'
    ]
    list_filter = ['is_active', 'parent', 'created_at']
    search_fields = ['name', 'description']
    list_editable = ['sort_order', 'is_active']
    list_per_page = 20

    fieldsets = (
        ('基本信息', {
            'fields': ('name', 'description', 'parent')
        }),
        ('显示设置', {
            'fields': ('sort_order', 'is_active')
        }),
    )

    def products_count(self, obj):
        """显示商品数量"""
        count = obj.get_products_count()
        if count > 0:
            url = reverse('admin:products_product_changelist')
            return format_html(
                '<a href="{}?category__id__exact={}">{} 个商品</a>',
                url, obj.id, count
            )
        return '0 个商品'
    products_count.short_description = '商品数量'

    def get_queryset(self, request):
        """优化查询"""
        return super().get_queryset(request).select_related('parent')


@admin.register(Product)
class ProductAdmin(admin.ModelAdmin):
    """商品管理"""

    list_display = [
        'name', 'category', 'original_price', 'image_preview',
        'is_active', 'sort_order', 'seckill_status', 'created_at'
    ]
    list_filter = [
        'is_active', 'category', 'created_at',
        ('category', admin.RelatedOnlyFieldListFilter),
    ]
    search_fields = ['name', 'description']
    list_editable = ['is_active', 'sort_order']
    list_per_page = 20

    fieldsets = (
        ('基本信息', {
            'fields': ('name', 'description', 'category')
        }),
        ('价格信息', {
            'fields': ('original_price',)
        }),
        ('图片信息', {
            'fields': ('image_url',)
        }),
        ('显示设置', {
            'fields': ('sort_order', 'is_active')
        }),
    )

    readonly_fields = ['created_at', 'updated_at']

    def image_preview(self, obj):
        """图片预览"""
        if obj.image_url:
            return format_html(
                '<img src="{}" style="width: 50px; height: 50px; object-fit: cover;" />',
                obj.image_url
            )
        return '无图片'
    image_preview.short_description = '图片预览'

    def seckill_status(self, obj):
        """秒杀状态"""
        if obj.has_active_seckill():
            return format_html(
                '<span style="color: green;">🔥 进行中</span>'
            )
        activities = obj.seckill_activities.filter(status='pending')
        if activities.exists():
            return format_html(
                '<span style="color: orange;">⏰ 待开始</span>'
            )
        return format_html(
            '<span style="color: gray;">无活动</span>'
        )
    seckill_status.short_description = '秒杀状态'

    def get_queryset(self, request):
        """优化查询"""
        return super().get_queryset(request).select_related('category').prefetch_related('seckill_activities')

    actions = ['make_active', 'make_inactive']

    def make_active(self, request, queryset):
        """批量上架"""
        updated = queryset.update(is_active=True)
        self.message_user(request, f'成功上架 {updated} 个商品')
    make_active.short_description = '批量上架选中的商品'

    def make_inactive(self, request, queryset):
        """批量下架"""
        updated = queryset.update(is_active=False)
        self.message_user(request, f'成功下架 {updated} 个商品')
    make_inactive.short_description = '批量下架选中的商品'
