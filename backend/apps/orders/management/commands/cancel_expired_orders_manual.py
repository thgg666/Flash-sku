"""
Django管理命令：手动取消过期订单
"""

import logging
from django.core.management.base import BaseCommand
from django.utils import timezone

from apps.orders.tasks import cancel_expired_orders, monitor_order_timeouts

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = '手动取消过期订单'
    
    def add_arguments(self, parser):
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='只显示统计信息，不实际取消订单'
        )
        parser.add_argument(
            '--monitor',
            action='store_true',
            help='显示订单超时监控统计'
        )
        
    def handle(self, *args, **options):
        """主处理函数"""
        dry_run = options['dry_run']
        show_monitor = options['monitor']
        
        self.stdout.write(
            self.style.SUCCESS(f'开始手动处理过期订单 - 时间: {timezone.now()}')
        )
        
        if show_monitor:
            self._show_monitor_stats()
        
        if dry_run:
            self._dry_run()
        else:
            self._execute_cancel()
            
    def _show_monitor_stats(self):
        """显示监控统计"""
        self.stdout.write(
            self.style.WARNING('获取订单超时监控统计...')
        )
        
        try:
            result = monitor_order_timeouts.delay()
            stats = result.get(timeout=30)
            
            if stats['success']:
                self.stdout.write(
                    self.style.SUCCESS('订单超时监控统计:')
                )
                for key, value in stats['stats'].items():
                    self.stdout.write(f'  {key}: {value}')
            else:
                self.stdout.write(
                    self.style.ERROR(f'获取监控统计失败: {stats.get("error")}')
                )
                
        except Exception as e:
            self.stdout.write(
                self.style.ERROR(f'获取监控统计时发生异常: {str(e)}')
            )
            
    def _dry_run(self):
        """模拟运行，只显示统计"""
        from apps.orders.models import Order
        
        self.stdout.write(
            self.style.WARNING('模拟运行模式 - 不会实际取消订单')
        )
        
        try:
            # 查找过期订单
            expired_orders = Order.objects.filter(
                status='pending_payment',
                payment_deadline__lt=timezone.now()
            ).select_related('activity', 'user')
            
            count = expired_orders.count()
            
            self.stdout.write(
                self.style.SUCCESS(f'找到 {count} 个过期订单')
            )
            
            if count > 0:
                self.stdout.write('过期订单详情:')
                for order in expired_orders[:10]:  # 只显示前10个
                    self.stdout.write(
                        f'  订单ID: {order.id}, '
                        f'用户: {order.user.username}, '
                        f'活动: {order.activity.id}, '
                        f'过期时间: {order.payment_deadline}, '
                        f'金额: {order.total_amount}'
                    )
                
                if count > 10:
                    self.stdout.write(f'  ... 还有 {count - 10} 个订单')
                    
        except Exception as e:
            self.stdout.write(
                self.style.ERROR(f'查询过期订单时发生异常: {str(e)}')
            )
            
    def _execute_cancel(self):
        """执行取消操作"""
        self.stdout.write(
            self.style.WARNING('执行过期订单取消...')
        )
        
        try:
            result = cancel_expired_orders.delay()
            cancel_result = result.get(timeout=300)  # 5分钟超时
            
            if cancel_result['success']:
                self.stdout.write(
                    self.style.SUCCESS(
                        f'过期订单取消完成:\n'
                        f'  成功取消: {cancel_result["cancelled_count"]}\n'
                        f'  取消失败: {cancel_result["failed_count"]}\n'
                        f'  跳过处理: {cancel_result["skipped_count"]}\n'
                        f'  总计发现: {cancel_result["total_found"]}\n'
                        f'  处理耗时: {cancel_result["duration_seconds"]:.2f}秒'
                    )
                )
            else:
                self.stdout.write(
                    self.style.ERROR(f'取消过期订单失败: {cancel_result.get("error")}')
                )
                
        except Exception as e:
            self.stdout.write(
                self.style.ERROR(f'执行取消任务时发生异常: {str(e)}')
            )
