"""
Flash Sku Backend - API Views
API 视图集，处理 HTTP 请求和响应
"""

from rest_framework import viewsets, filters, status
from rest_framework.decorators import action
from rest_framework.response import Response
from django_filters.rest_framework import DjangoFilterBackend
from drf_spectacular.utils import extend_schema, extend_schema_view

from apps.products.models import Category, Product
from apps.activities.models import SeckillActivity
from apps.orders.models import Order
from apps.common.models import SystemConfig

from .serializers import (
    CategorySerializer,
    ProductListSerializer,
    ProductDetailSerializer,
    SeckillActivityListSerializer,
    SeckillActivityDetailSerializer,
    OrderSerializer,
    SystemConfigSerializer,
)


@extend_schema_view(
    list=extend_schema(
        summary="获取商品分类列表",
        description="获取所有启用的商品分类，支持层级结构"
    ),
    retrieve=extend_schema(
        summary="获取商品分类详情",
        description="获取指定分类的详细信息，包括子分类和商品数量"
    ),
)
class CategoryViewSet(viewsets.ReadOnlyModelViewSet):
    """商品分类视图集（只读）"""
    
    serializer_class = CategorySerializer
    filter_backends = [filters.SearchFilter, filters.OrderingFilter]
    search_fields = ['name', 'description']
    ordering_fields = ['sort_order', 'name', 'created_at']
    ordering = ['sort_order', 'name']
    
    def get_queryset(self):
        """获取查询集"""
        return Category.objects.filter(is_active=True).select_related('parent')


@extend_schema_view(
    list=extend_schema(
        summary="获取商品列表",
        description="获取所有上架商品列表，支持分类筛选和搜索"
    ),
    retrieve=extend_schema(
        summary="获取商品详情",
        description="获取指定商品的详细信息，包括分类和活跃的秒杀活动"
    ),
)
class ProductViewSet(viewsets.ReadOnlyModelViewSet):
    """商品视图集（只读）"""
    
    filter_backends = [DjangoFilterBackend, filters.SearchFilter, filters.OrderingFilter]
    filterset_fields = ['category', 'is_active']
    search_fields = ['name', 'description']
    ordering_fields = ['original_price', 'sort_order', 'created_at']
    ordering = ['-created_at']
    
    def get_queryset(self):
        """获取查询集"""
        return Product.objects.filter(is_active=True).select_related('category')
    
    def get_serializer_class(self):
        """根据动作选择序列化器"""
        if self.action == 'retrieve':
            return ProductDetailSerializer
        return ProductListSerializer
    
    @extend_schema(
        summary="获取商品的秒杀活动",
        description="获取指定商品的所有秒杀活动"
    )
    @action(detail=True, methods=['get'])
    def activities(self, request, pk=None):
        """获取商品的秒杀活动"""
        product = self.get_object()
        activities = product.seckill_activities.filter(
            status__in=['pending', 'active']
        ).order_by('start_time')
        
        serializer = SeckillActivityListSerializer(
            activities, many=True, context={'request': request}
        )
        return Response(serializer.data)


@extend_schema_view(
    list=extend_schema(
        summary="获取秒杀活动列表",
        description="获取所有秒杀活动，支持状态筛选和时间排序"
    ),
    retrieve=extend_schema(
        summary="获取秒杀活动详情",
        description="获取指定秒杀活动的详细信息，包括商品信息和库存状态"
    ),
)
class SeckillActivityViewSet(viewsets.ReadOnlyModelViewSet):
    """秒杀活动视图集（只读）"""
    
    filter_backends = [DjangoFilterBackend, filters.SearchFilter, filters.OrderingFilter]
    filterset_fields = ['status', 'product']
    search_fields = ['name', 'description', 'product__name']
    ordering_fields = ['start_time', 'end_time', 'seckill_price']
    ordering = ['start_time']
    
    def get_queryset(self):
        """获取查询集"""
        return SeckillActivity.objects.select_related('product').exclude(status='cancelled')
    
    def get_serializer_class(self):
        """根据动作选择序列化器"""
        if self.action == 'retrieve':
            return SeckillActivityDetailSerializer
        return SeckillActivityListSerializer
    
    @extend_schema(
        summary="获取进行中的秒杀活动",
        description="获取当前正在进行的秒杀活动列表"
    )
    @action(detail=False, methods=['get'])
    def active(self, request):
        """获取进行中的秒杀活动"""
        activities = self.get_queryset().filter(status='active')
        
        page = self.paginate_queryset(activities)
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        
        serializer = self.get_serializer(activities, many=True)
        return Response(serializer.data)
    
    @extend_schema(
        summary="获取即将开始的秒杀活动",
        description="获取即将开始的秒杀活动列表"
    )
    @action(detail=False, methods=['get'])
    def upcoming(self, request):
        """获取即将开始的秒杀活动"""
        activities = self.get_queryset().filter(status='pending')
        
        page = self.paginate_queryset(activities)
        if page is not None:
            serializer = self.get_serializer(page, many=True)
            return self.get_paginated_response(serializer.data)
        
        serializer = self.get_serializer(activities, many=True)
        return Response(serializer.data)


@extend_schema_view(
    list=extend_schema(
        summary="获取系统配置列表",
        description="获取所有启用的系统配置项"
    ),
)
class SystemConfigViewSet(viewsets.ReadOnlyModelViewSet):
    """系统配置视图集（只读）"""
    
    serializer_class = SystemConfigSerializer
    filter_backends = [filters.SearchFilter]
    search_fields = ['key', 'description']
    
    def get_queryset(self):
        """获取查询集"""
        return SystemConfig.objects.filter(is_active=True).order_by('key')
