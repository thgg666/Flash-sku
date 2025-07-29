"""
Flash Sku Backend - API Serializers
API 序列化器，用于数据序列化和反序列化
"""

from rest_framework import serializers
from apps.products.models import Category, Product
from apps.activities.models import SeckillActivity
from apps.orders.models import Order
from apps.common.models import SystemConfig


class CategorySerializer(serializers.ModelSerializer):
    """商品分类序列化器"""
    
    products_count = serializers.ReadOnlyField(source='get_products_count')
    children = serializers.SerializerMethodField()
    
    class Meta:
        model = Category
        fields = [
            'id', 'name', 'description', 'parent', 'sort_order',
            'is_active', 'products_count', 'children', 'created_at'
        ]
        read_only_fields = ['created_at']
    
    def get_children(self, obj):
        """获取子分类"""
        children = obj.get_children()
        return CategorySerializer(children, many=True, context=self.context).data


class ProductListSerializer(serializers.ModelSerializer):
    """商品列表序列化器（简化版）"""
    
    category_name = serializers.CharField(source='category.name', read_only=True)
    has_seckill = serializers.ReadOnlyField(source='has_active_seckill')
    
    class Meta:
        model = Product
        fields = [
            'id', 'name', 'image_url', 'category_name', 
            'original_price', 'has_seckill', 'is_active'
        ]


class ProductDetailSerializer(serializers.ModelSerializer):
    """商品详情序列化器（完整版）"""
    
    category = CategorySerializer(read_only=True)
    active_activities = serializers.SerializerMethodField()
    
    class Meta:
        model = Product
        fields = [
            'id', 'name', 'description', 'image_url', 'category',
            'original_price', 'is_active', 'sort_order',
            'active_activities', 'created_at', 'updated_at'
        ]
        read_only_fields = ['created_at', 'updated_at']
    
    def get_active_activities(self, obj):
        """获取活跃的秒杀活动"""
        activities = obj.get_active_activities()
        return SeckillActivityListSerializer(activities, many=True, context=self.context).data


class SeckillActivityListSerializer(serializers.ModelSerializer):
    """秒杀活动列表序列化器（简化版）"""
    
    product_name = serializers.CharField(source='product.name', read_only=True)
    product_image = serializers.CharField(source='product.image_url', read_only=True)
    discount_percentage = serializers.ReadOnlyField(source='get_discount_percentage')
    sold_count = serializers.ReadOnlyField(source='get_sold_count')
    is_active = serializers.ReadOnlyField()
    
    class Meta:
        model = SeckillActivity
        fields = [
            'id', 'name', 'product_name', 'product_image',
            'start_time', 'end_time', 'original_price', 'seckill_price',
            'total_stock', 'available_stock', 'max_per_user',
            'status', 'discount_percentage', 'sold_count', 'is_active'
        ]


class SeckillActivityDetailSerializer(serializers.ModelSerializer):
    """秒杀活动详情序列化器（完整版）"""
    
    product = ProductListSerializer(read_only=True)
    discount_percentage = serializers.ReadOnlyField(source='get_discount_percentage')
    sold_count = serializers.ReadOnlyField(source='get_sold_count')
    is_active = serializers.ReadOnlyField()
    
    class Meta:
        model = SeckillActivity
        fields = [
            'id', 'name', 'description', 'product',
            'start_time', 'end_time', 'original_price', 'seckill_price',
            'total_stock', 'available_stock', 'max_per_user',
            'status', 'discount_percentage', 'sold_count', 'is_active',
            'created_at', 'updated_at'
        ]
        read_only_fields = ['created_at', 'updated_at']


class OrderSerializer(serializers.ModelSerializer):
    """订单序列化器"""
    
    user_username = serializers.CharField(source='user.username', read_only=True)
    activity_name = serializers.CharField(source='activity.name', read_only=True)
    remaining_time = serializers.ReadOnlyField(source='get_remaining_time')
    can_pay = serializers.ReadOnlyField()
    can_cancel = serializers.ReadOnlyField()
    is_expired = serializers.ReadOnlyField()
    
    class Meta:
        model = Order
        fields = [
            'id', 'user_username', 'activity_name', 'product_name',
            'seckill_price', 'quantity', 'total_amount', 'status',
            'payment_deadline', 'remaining_time', 'can_pay', 'can_cancel',
            'is_expired', 'created_at', 'updated_at'
        ]
        read_only_fields = [
            'total_amount', 'created_at', 'updated_at',
            'paid_at', 'cancelled_at'
        ]


class SystemConfigSerializer(serializers.ModelSerializer):
    """系统配置序列化器"""
    
    class Meta:
        model = SystemConfig
        fields = ['key', 'value', 'description', 'is_active']
        read_only_fields = ['key']  # 配置键不允许修改
