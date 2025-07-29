"""
Flash Sku - Orders Views
订单管理相关视图
"""

import logging
from rest_framework import status
from rest_framework.decorators import api_view, permission_classes
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response
from django.core.exceptions import ObjectDoesNotExist

from .models import Order
from .tasks import create_seckill_order
from apps.activities.models import SeckillActivity

logger = logging.getLogger(__name__)


@api_view(['POST'])
@permission_classes([IsAuthenticated])
def create_seckill_order_view(request):
    """
    创建秒杀订单API

    POST /api/orders/seckill/
    {
        "activity_id": 1,
        "quantity": 1
    }
    """
    try:
        activity_id = request.data.get('activity_id')
        quantity = request.data.get('quantity', 1)

        # 参数验证
        if not activity_id:
            return Response({
                'success': False,
                'error': '活动ID不能为空',
                'code': 'MISSING_ACTIVITY_ID'
            }, status=status.HTTP_400_BAD_REQUEST)

        if not isinstance(quantity, int) or quantity <= 0:
            return Response({
                'success': False,
                'error': '购买数量必须是正整数',
                'code': 'INVALID_QUANTITY'
            }, status=status.HTTP_400_BAD_REQUEST)

        # 检查活动是否存在
        try:
            activity = SeckillActivity.objects.get(id=activity_id)
        except ObjectDoesNotExist:
            return Response({
                'success': False,
                'error': '活动不存在',
                'code': 'ACTIVITY_NOT_FOUND'
            }, status=status.HTTP_404_NOT_FOUND)

        # 检查用户是否已经有该活动的订单
        existing_order = Order.objects.filter(
            user=request.user,
            activity=activity
        ).first()

        if existing_order:
            return Response({
                'success': False,
                'error': '您已经购买过该活动商品',
                'code': 'ALREADY_PURCHASED',
                'order_id': existing_order.id
            }, status=status.HTTP_400_BAD_REQUEST)

        # 异步创建订单
        task_result = create_seckill_order.delay(
            user_id=request.user.id,
            activity_id=activity_id,
            quantity=quantity
        )

        logger.info(f"秒杀订单创建任务已提交 - 用户ID: {request.user.id}, 活动ID: {activity_id}, 任务ID: {task_result.id}")

        return Response({
            'success': True,
            'message': '订单创建请求已提交，正在处理中...',
            'task_id': task_result.id,
            'code': 'ORDER_SUBMITTED'
        }, status=status.HTTP_202_ACCEPTED)

    except Exception as e:
        logger.error(f"创建秒杀订单时发生异常 - 用户ID: {request.user.id}, 错误: {str(e)}")
        return Response({
            'success': False,
            'error': '系统异常，请稍后重试',
            'code': 'SYSTEM_ERROR'
        }, status=status.HTTP_500_INTERNAL_SERVER_ERROR)


@api_view(['GET'])
@permission_classes([IsAuthenticated])
def check_order_status(request, task_id):
    """
    检查订单创建状态API

    GET /api/orders/status/{task_id}/
    """
    try:
        from celery.result import AsyncResult

        task_result = AsyncResult(task_id)

        if task_result.ready():
            result = task_result.result
            if isinstance(result, dict):
                return Response({
                    'task_id': task_id,
                    'status': 'completed',
                    'result': result
                }, status=status.HTTP_200_OK)
            else:
                return Response({
                    'task_id': task_id,
                    'status': 'failed',
                    'error': str(result)
                }, status=status.HTTP_200_OK)
        else:
            return Response({
                'task_id': task_id,
                'status': 'pending',
                'message': '订单正在创建中...'
            }, status=status.HTTP_200_OK)

    except Exception as e:
        logger.error(f"检查订单状态时发生异常 - 任务ID: {task_id}, 错误: {str(e)}")
        return Response({
            'task_id': task_id,
            'status': 'error',
            'error': '系统异常'
        }, status=status.HTTP_500_INTERNAL_SERVER_ERROR)


@api_view(['GET'])
@permission_classes([IsAuthenticated])
def user_orders(request):
    """
    获取用户订单列表API

    GET /api/orders/
    """
    try:
        orders = Order.objects.filter(user=request.user).order_by('-created_at')

        orders_data = []
        for order in orders:
            orders_data.append({
                'id': order.id,
                'activity_id': order.activity.id,
                'product_name': order.product_name,
                'seckill_price': str(order.seckill_price),
                'quantity': order.quantity,
                'total_amount': str(order.total_amount),
                'status': order.status,
                'payment_deadline': order.payment_deadline.isoformat() if order.payment_deadline else None,
                'remaining_time': order.get_remaining_time(),
                'created_at': order.created_at.isoformat(),
                'can_pay': order.can_pay(),
                'can_cancel': order.can_cancel()
            })

        return Response({
            'success': True,
            'orders': orders_data,
            'total': len(orders_data)
        }, status=status.HTTP_200_OK)

    except Exception as e:
        logger.error(f"获取用户订单时发生异常 - 用户ID: {request.user.id}, 错误: {str(e)}")
        return Response({
            'success': False,
            'error': '系统异常'
        }, status=status.HTTP_500_INTERNAL_SERVER_ERROR)


@api_view(['POST'])
@permission_classes([IsAuthenticated])
def cancel_order(request, order_id):
    """
    取消订单API

    POST /api/orders/{order_id}/cancel/
    """
    try:
        try:
            order = Order.objects.get(id=order_id, user=request.user)
        except ObjectDoesNotExist:
            return Response({
                'success': False,
                'error': '订单不存在',
                'code': 'ORDER_NOT_FOUND'
            }, status=status.HTTP_404_NOT_FOUND)

        if not order.can_cancel():
            return Response({
                'success': False,
                'error': '订单状态不允许取消',
                'code': 'CANNOT_CANCEL'
            }, status=status.HTTP_400_BAD_REQUEST)

        success = order.cancel_order(reason='用户主动取消')

        if success:
            logger.info(f"用户取消订单成功 - 订单ID: {order_id}, 用户ID: {request.user.id}")
            return Response({
                'success': True,
                'message': '订单已取消',
                'order_id': order_id
            }, status=status.HTTP_200_OK)
        else:
            return Response({
                'success': False,
                'error': '订单取消失败',
                'code': 'CANCEL_FAILED'
            }, status=status.HTTP_400_BAD_REQUEST)

    except Exception as e:
        logger.error(f"取消订单时发生异常 - 订单ID: {order_id}, 用户ID: {request.user.id}, 错误: {str(e)}")
        return Response({
            'success': False,
            'error': '系统异常'
        }, status=status.HTTP_500_INTERNAL_SERVER_ERROR)
