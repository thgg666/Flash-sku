"""
Flash Sku - Common Admin
公共组件管理后台配置
"""

from django.contrib import admin
from django.utils.html import format_html
from .models import SystemConfig


@admin.register(SystemConfig)
class SystemConfigAdmin(admin.ModelAdmin):
    """系统配置管理"""

    list_display = [
        'key', 'value_preview', 'description',
        'is_active', 'updated_at'
    ]
    list_filter = ['is_active', 'created_at', 'updated_at']
    search_fields = ['key', 'description', 'value']
    list_editable = ['is_active']
    list_per_page = 20

    fieldsets = (
        ('配置信息', {
            'fields': ('key', 'value', 'description')
        }),
        ('状态设置', {
            'fields': ('is_active',)
        }),
    )

    readonly_fields = ['created_at', 'updated_at']

    def value_preview(self, obj):
        """值预览"""
        if len(obj.value) > 50:
            return format_html(
                '<span title="{}">{}</span>',
                obj.value,
                obj.value[:50] + '...'
            )
        return obj.value
    value_preview.short_description = '配置值'

    def get_queryset(self, request):
        """按key排序"""
        return super().get_queryset(request).order_by('key')

    actions = ['activate_configs', 'deactivate_configs']

    def activate_configs(self, request, queryset):
        """激活配置"""
        updated = queryset.update(is_active=True)
        self.message_user(request, f'成功激活 {updated} 个配置项')
    activate_configs.short_description = '激活选中的配置项'

    def deactivate_configs(self, request, queryset):
        """停用配置"""
        updated = queryset.update(is_active=False)
        self.message_user(request, f'成功停用 {updated} 个配置项')
    deactivate_configs.short_description = '停用选中的配置项'
