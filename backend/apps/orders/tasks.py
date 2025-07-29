"""
Flash Sku - Orders Celery Tasks
订单相关的异步任务处理
"""

import logging
import json
from datetime import timedelta
from decimal import Decimal
from typing import Dict, Any

from celery import shared_task
from django.utils import timezone
from django.db import transaction, models
from django.core.exceptions import ObjectDoesNotExist
from django.contrib.auth import get_user_model

from .models import Order
from apps.activities.models import SeckillActivity

User = get_user_model()

logger = logging.getLogger(__name__)


@shared_task(bind=True, max_retries=3, default_retry_delay=60)
def process_seckill_order_from_go(self, order_message: Dict[str, Any]) -> Dict[str, Any]:
    """
    处理从Go秒杀服务发送的订单消息

    Args:
        order_message: Go服务发送的订单消息
        {
            "order_id": "order_123456",
            "user_id": "user123",
            "activity_id": "activity456",
            "product_id": "product789",
            "quantity": 1,
            "price": 99.99,
            "status": "pending",
            "created_at": "2025-01-29T12:00:00Z"
        }

    Returns:
        Dict: 包含订单处理结果的字典
    """
    try:
        logger.info(f"开始处理Go秒杀订单消息 - 订单ID: {order_message.get('order_id')}")

        # 提取消息数据
        go_order_id = order_message.get('order_id')
        user_id = order_message.get('user_id')
        activity_id = order_message.get('activity_id')
        quantity = order_message.get('quantity', 1)
        price = order_message.get('price', 0.0)

        # 参数验证
        if not all([go_order_id, user_id, activity_id]):
            logger.error(f"订单消息参数不完整 - 消息: {order_message}")
            return {
                'success': False,
                'error': '订单消息参数不完整',
                'go_order_id': go_order_id
            }

        with transaction.atomic():
            # 获取用户和活动信息
            try:
                user = User.objects.get(id=user_id)
                activity = SeckillActivity.objects.select_for_update().get(id=activity_id)
            except ObjectDoesNotExist as e:
                logger.error(f"用户或活动不存在 - 用户ID: {user_id}, 活动ID: {activity_id}")
                return {
                    'success': False,
                    'error': '用户或活动不存在',
                    'go_order_id': go_order_id
                }

            # 检查是否已经创建过订单（防重复处理）
            existing_order = Order.objects.filter(
                user=user,
                activity=activity
            ).first()

            if existing_order:
                logger.warning(f"订单已存在，跳过创建 - 用户ID: {user_id}, 活动ID: {activity_id}, 订单ID: {existing_order.id}")
                return {
                    'success': True,
                    'error': None,
                    'order_id': existing_order.id,
                    'go_order_id': go_order_id,
                    'message': '订单已存在'
                }

            # 创建Django订单
            order = Order.objects.create(
                user=user,
                activity=activity,
                product_name=activity.product.name,
                seckill_price=activity.seckill_price,  # 使用活动价格而不是Go传递的价格
                quantity=quantity,
                status='pending_payment'
            )

            logger.info(f"Django订单创建成功 - Django订单ID: {order.id}, Go订单ID: {go_order_id}")

            # 设置订单超时取消任务
            cancel_order_task.apply_async(
                args=[order.id],
                eta=order.payment_deadline,
                queue='orders'
            )

            return {
                'success': True,
                'error': None,
                'order_id': order.id,
                'go_order_id': go_order_id,
                'total_amount': str(order.total_amount),
                'payment_deadline': order.payment_deadline.isoformat()
            }

    except Exception as exc:
        logger.error(f"处理Go订单消息时发生异常 - Go订单ID: {order_message.get('order_id')}, 错误: {str(exc)}")

        # 重试机制
        if self.request.retries < self.max_retries:
            logger.info(f"重试处理Go订单消息 - 第{self.request.retries + 1}次重试")
            raise self.retry(exc=exc, countdown=60 * (self.request.retries + 1))

        return {
            'success': False,
            'error': f'系统异常: {str(exc)}',
            'go_order_id': order_message.get('order_id')
        }


@shared_task(bind=True, max_retries=3, default_retry_delay=60)
def create_seckill_order(self, user_id: int, activity_id: int, quantity: int = 1) -> Dict[str, Any]:
    """
    创建秒杀订单的异步任务
    
    Args:
        user_id: 用户ID
        activity_id: 秒杀活动ID
        quantity: 购买数量
        
    Returns:
        Dict: 包含订单创建结果的字典
    """
    try:
        logger.info(f"开始创建秒杀订单 - 用户ID: {user_id}, 活动ID: {activity_id}, 数量: {quantity}")
        
        with transaction.atomic():
            # 获取用户和活动信息
            try:
                user = User.objects.get(id=user_id)
                activity = SeckillActivity.objects.select_for_update().get(id=activity_id)
            except ObjectDoesNotExist as e:
                logger.error(f"用户或活动不存在 - 用户ID: {user_id}, 活动ID: {activity_id}")
                return {
                    'success': False,
                    'error': '用户或活动不存在',
                    'order_id': None
                }
            
            # 检查活动状态
            if not activity.is_active():
                logger.warning(f"活动未激活 - 活动ID: {activity_id}, 状态: {activity.status}")
                return {
                    'success': False,
                    'error': '活动未激活',
                    'order_id': None
                }
            
            # 检查库存
            if activity.available_stock < quantity:
                logger.warning(f"库存不足 - 活动ID: {activity_id}, 可用库存: {activity.available_stock}, 需求: {quantity}")
                return {
                    'success': False,
                    'error': '库存不足',
                    'order_id': None
                }
            
            # 检查用户是否已经购买过该活动
            existing_order = Order.objects.filter(user=user, activity=activity).first()
            if existing_order:
                logger.warning(f"用户已购买过该活动 - 用户ID: {user_id}, 活动ID: {activity_id}")
                return {
                    'success': False,
                    'error': '您已经购买过该活动商品',
                    'order_id': existing_order.id
                }
            
            # 扣减库存
            activity.available_stock -= quantity
            activity.save()
            
            # 创建订单
            order = Order.objects.create(
                user=user,
                activity=activity,
                product_name=activity.product.name,
                seckill_price=activity.seckill_price,
                quantity=quantity,
                status='pending_payment'
            )
            
            logger.info(f"订单创建成功 - 订单ID: {order.id}, 用户ID: {user_id}, 活动ID: {activity_id}")
            
            # 设置订单超时取消任务
            cancel_order_task.apply_async(
                args=[order.id],
                eta=order.payment_deadline,
                queue='orders'
            )
            
            return {
                'success': True,
                'error': None,
                'order_id': order.id,
                'total_amount': str(order.total_amount),
                'payment_deadline': order.payment_deadline.isoformat()
            }
            
    except Exception as exc:
        logger.error(f"创建订单时发生异常 - 用户ID: {user_id}, 活动ID: {activity_id}, 错误: {str(exc)}")
        
        # 重试机制
        if self.request.retries < self.max_retries:
            logger.info(f"重试创建订单 - 第{self.request.retries + 1}次重试")
            raise self.retry(exc=exc, countdown=60 * (self.request.retries + 1))
        
        return {
            'success': False,
            'error': f'系统异常: {str(exc)}',
            'order_id': None
        }


@shared_task(bind=True)
def cancel_order_task(self, order_id: int) -> Dict[str, Any]:
    """
    取消订单的异步任务
    
    Args:
        order_id: 订单ID
        
    Returns:
        Dict: 包含取消结果的字典
    """
    try:
        logger.info(f"开始取消订单 - 订单ID: {order_id}")
        
        with transaction.atomic():
            try:
                order = Order.objects.select_for_update().get(id=order_id)
            except ObjectDoesNotExist:
                logger.warning(f"订单不存在 - 订单ID: {order_id}")
                return {
                    'success': False,
                    'error': '订单不存在',
                    'order_id': order_id
                }
            
            # 检查订单是否可以取消
            if order.status != 'pending_payment':
                logger.info(f"订单状态不允许取消 - 订单ID: {order_id}, 状态: {order.status}")
                return {
                    'success': False,
                    'error': '订单状态不允许取消',
                    'order_id': order_id
                }
            
            # 检查是否已过期
            if not order.is_expired():
                logger.info(f"订单未过期，不需要取消 - 订单ID: {order_id}")
                return {
                    'success': False,
                    'error': '订单未过期',
                    'order_id': order_id
                }
            
            # 取消订单并回滚库存
            success = order.cancel_order(reason='支付超时自动取消')
            
            if success:
                logger.info(f"订单取消成功 - 订单ID: {order_id}")
                return {
                    'success': True,
                    'error': None,
                    'order_id': order_id
                }
            else:
                logger.error(f"订单取消失败 - 订单ID: {order_id}")
                return {
                    'success': False,
                    'error': '订单取消失败',
                    'order_id': order_id
                }
                
    except Exception as exc:
        logger.error(f"取消订单时发生异常 - 订单ID: {order_id}, 错误: {str(exc)}")
        return {
            'success': False,
            'error': f'系统异常: {str(exc)}',
            'order_id': order_id
        }


@shared_task(bind=True)
def cancel_expired_orders(self) -> Dict[str, Any]:
    """
    定时任务：取消所有过期的订单

    Returns:
        Dict: 包含处理结果的字典
    """
    start_time = timezone.now()
    try:
        logger.info("开始执行过期订单取消任务")

        # 查找所有过期的待支付订单，限制批次大小避免内存问题
        batch_size = 100
        expired_orders = Order.objects.filter(
            status='pending_payment',
            payment_deadline__lt=timezone.now()
        ).select_related('activity')[:batch_size]

        total_found = expired_orders.count()
        cancelled_count = 0
        failed_count = 0
        skipped_count = 0

        logger.info(f"找到 {total_found} 个过期订单需要处理")

        for order in expired_orders:
            try:
                with transaction.atomic():
                    # 重新获取订单以确保数据最新
                    order_for_update = Order.objects.select_for_update().get(id=order.id)

                    # 再次检查订单状态，防止并发修改
                    if order_for_update.status != 'pending_payment':
                        skipped_count += 1
                        logger.info(f"订单状态已变更，跳过处理 - 订单ID: {order.id}, 当前状态: {order_for_update.status}")
                        continue

                    # 再次检查是否过期
                    if not order_for_update.is_expired():
                        skipped_count += 1
                        logger.info(f"订单未过期，跳过处理 - 订单ID: {order.id}")
                        continue

                    if order_for_update.cancel_order(reason='支付超时自动取消'):
                        cancelled_count += 1
                        logger.info(f"过期订单取消成功 - 订单ID: {order.id}, 用户ID: {order.user_id}, 活动ID: {order.activity_id}")
                    else:
                        failed_count += 1
                        logger.warning(f"过期订单取消失败 - 订单ID: {order.id}")

            except ObjectDoesNotExist:
                skipped_count += 1
                logger.warning(f"订单不存在，跳过处理 - 订单ID: {order.id}")
            except Exception as e:
                failed_count += 1
                logger.error(f"取消过期订单时发生异常 - 订单ID: {order.id}, 错误: {str(e)}")

        end_time = timezone.now()
        duration = (end_time - start_time).total_seconds()

        logger.info(f"过期订单取消任务完成 - 耗时: {duration:.2f}秒, 成功: {cancelled_count}, 失败: {failed_count}, 跳过: {skipped_count}")

        # 如果还有更多过期订单，安排下一批处理
        if total_found >= batch_size:
            logger.info("检测到更多过期订单，将在30秒后处理下一批")
            cancel_expired_orders.apply_async(countdown=30)

        return {
            'success': True,
            'cancelled_count': cancelled_count,
            'failed_count': failed_count,
            'skipped_count': skipped_count,
            'total_found': total_found,
            'duration_seconds': duration
        }

    except Exception as exc:
        end_time = timezone.now()
        duration = (end_time - start_time).total_seconds()
        logger.error(f"执行过期订单取消任务时发生异常: {str(exc)}, 耗时: {duration:.2f}秒")
        return {
            'success': False,
            'error': str(exc),
            'cancelled_count': 0,
            'failed_count': 0,
            'duration_seconds': duration
        }


@shared_task
def monitor_order_timeouts() -> Dict[str, Any]:
    """
    监控任务：统计订单超时情况

    Returns:
        Dict: 包含监控统计的字典
    """
    try:
        logger.info("开始执行订单超时监控任务")

        now = timezone.now()

        # 统计各种状态的订单
        stats = {
            'pending_payment_total': Order.objects.filter(status='pending_payment').count(),
            'pending_payment_expired': Order.objects.filter(
                status='pending_payment',
                payment_deadline__lt=now
            ).count(),
            'pending_payment_expiring_soon': Order.objects.filter(
                status='pending_payment',
                payment_deadline__gte=now,
                payment_deadline__lt=now + timedelta(minutes=5)
            ).count(),
            'cancelled_today': Order.objects.filter(
                status='cancelled',
                cancelled_at__gte=now.replace(hour=0, minute=0, second=0, microsecond=0)
            ).count(),
            'paid_today': Order.objects.filter(
                status='paid',
                paid_at__gte=now.replace(hour=0, minute=0, second=0, microsecond=0)
            ).count()
        }

        # 计算超时率
        total_orders_today = Order.objects.filter(
            created_at__gte=now.replace(hour=0, minute=0, second=0, microsecond=0)
        ).count()

        if total_orders_today > 0:
            timeout_rate = (stats['cancelled_today'] / total_orders_today) * 100
        else:
            timeout_rate = 0

        stats['timeout_rate_percent'] = round(timeout_rate, 2)
        stats['total_orders_today'] = total_orders_today

        logger.info(f"订单超时监控完成 - 统计: {stats}")

        # 如果超时率过高，记录警告
        if timeout_rate > 50:  # 超时率超过50%
            logger.warning(f"订单超时率过高: {timeout_rate:.2f}%, 需要关注")

        return {
            'success': True,
            'stats': stats,
            'timestamp': now.isoformat()
        }

    except Exception as exc:
        logger.error(f"执行订单超时监控任务时发生异常: {str(exc)}")
        return {
            'success': False,
            'error': str(exc),
            'timestamp': timezone.now().isoformat()
        }


@shared_task
def check_stock_consistency() -> Dict[str, Any]:
    """
    检查库存一致性的任务

    验证活动库存与实际订单数据的一致性

    Returns:
        Dict: 包含检查结果的字典
    """
    try:
        logger.info("开始执行库存一致性检查")

        from apps.activities.models import SeckillActivity

        inconsistent_activities = []
        total_checked = 0

        # 检查所有活跃的秒杀活动
        activities = SeckillActivity.objects.filter(status='active')

        for activity in activities:
            total_checked += 1

            # 计算已售出的数量（已支付的订单）
            sold_quantity = Order.objects.filter(
                activity=activity,
                status='paid'
            ).aggregate(
                total=models.Sum('quantity')
            )['total'] or 0

            # 计算待支付的数量
            pending_quantity = Order.objects.filter(
                activity=activity,
                status='pending_payment'
            ).aggregate(
                total=models.Sum('quantity')
            )['total'] or 0

            # 理论上的可用库存
            theoretical_stock = activity.total_stock - sold_quantity - pending_quantity

            # 实际的可用库存
            actual_stock = activity.available_stock

            # 检查是否一致
            if theoretical_stock != actual_stock:
                inconsistency = {
                    'activity_id': activity.id,
                    'activity_name': activity.product.name,
                    'total_stock': activity.total_stock,
                    'sold_quantity': sold_quantity,
                    'pending_quantity': pending_quantity,
                    'theoretical_stock': theoretical_stock,
                    'actual_stock': actual_stock,
                    'difference': actual_stock - theoretical_stock
                }
                inconsistent_activities.append(inconsistency)

                logger.warning(f"发现库存不一致 - 活动ID: {activity.id}, "
                             f"理论库存: {theoretical_stock}, 实际库存: {actual_stock}, "
                             f"差异: {actual_stock - theoretical_stock}")

        result = {
            'success': True,
            'total_checked': total_checked,
            'inconsistent_count': len(inconsistent_activities),
            'inconsistent_activities': inconsistent_activities,
            'timestamp': timezone.now().isoformat()
        }

        if inconsistent_activities:
            logger.error(f"库存一致性检查发现 {len(inconsistent_activities)} 个不一致的活动")
        else:
            logger.info(f"库存一致性检查完成，所有 {total_checked} 个活动库存一致")

        return result

    except Exception as exc:
        logger.error(f"执行库存一致性检查时发生异常: {str(exc)}")
        return {
            'success': False,
            'error': str(exc),
            'timestamp': timezone.now().isoformat()
        }


@shared_task(bind=True, max_retries=3, default_retry_delay=30)
def rollback_stock(self, activity_id: int, quantity: int, order_id: int = None) -> Dict[str, Any]:
    """
    回滚库存的异步任务

    Args:
        activity_id: 活动ID
        quantity: 回滚数量
        order_id: 订单ID（可选，用于日志记录）

    Returns:
        Dict: 包含回滚结果的字典
    """
    start_time = timezone.now()

    # 参数验证
    if not isinstance(activity_id, int) or activity_id <= 0:
        logger.error(f"无效的活动ID: {activity_id}")
        return {
            'success': False,
            'error': '无效的活动ID',
            'activity_id': activity_id
        }

    if not isinstance(quantity, int) or quantity <= 0:
        logger.error(f"无效的回滚数量: {quantity}")
        return {
            'success': False,
            'error': '无效的回滚数量',
            'activity_id': activity_id
        }

    try:
        logger.info(f"开始回滚库存 - 活动ID: {activity_id}, 数量: {quantity}, 订单ID: {order_id}")

        with transaction.atomic():
            try:
                activity = SeckillActivity.objects.select_for_update().get(id=activity_id)
            except ObjectDoesNotExist:
                logger.error(f"活动不存在 - 活动ID: {activity_id}")
                return {
                    'success': False,
                    'error': '活动不存在',
                    'activity_id': activity_id
                }

            # 记录原始库存
            original_stock = activity.available_stock

            # 检查活动状态
            if not activity.is_active():
                logger.warning(f"活动未激活，但仍执行库存回滚 - 活动ID: {activity_id}, 状态: {activity.status}")

            # 回滚库存
            new_stock = original_stock + quantity

            # 确保库存不超过总库存
            if new_stock > activity.total_stock:
                logger.warning(f"回滚后库存超过总库存，调整为总库存 - 活动ID: {activity_id}, "
                             f"计算库存: {new_stock}, 总库存: {activity.total_stock}")
                new_stock = activity.total_stock
                actual_rollback = new_stock - original_stock
            else:
                actual_rollback = quantity

            activity.available_stock = new_stock
            activity.save()

            end_time = timezone.now()
            duration = (end_time - start_time).total_seconds()

            logger.info(f"库存回滚成功 - 活动ID: {activity_id}, "
                       f"请求回滚: {quantity}, 实际回滚: {actual_rollback}, "
                       f"原库存: {original_stock}, 新库存: {new_stock}, "
                       f"耗时: {duration:.3f}秒")

            return {
                'success': True,
                'error': None,
                'activity_id': activity_id,
                'order_id': order_id,
                'requested_quantity': quantity,
                'actual_rollback': actual_rollback,
                'original_stock': original_stock,
                'current_stock': new_stock,
                'duration_seconds': duration
            }

    except Exception as exc:
        end_time = timezone.now()
        duration = (end_time - start_time).total_seconds()

        logger.error(f"回滚库存时发生异常 - 活动ID: {activity_id}, 订单ID: {order_id}, "
                    f"错误: {str(exc)}, 耗时: {duration:.3f}秒")

        # 重试机制
        if self.request.retries < self.max_retries:
            retry_countdown = 30 * (self.request.retries + 1)
            logger.info(f"重试回滚库存 - 第{self.request.retries + 1}次重试, "
                       f"{retry_countdown}秒后执行")
            raise self.retry(exc=exc, countdown=retry_countdown)

        return {
            'success': False,
            'error': f'系统异常: {str(exc)}',
            'activity_id': activity_id,
            'order_id': order_id,
            'duration_seconds': duration
        }
