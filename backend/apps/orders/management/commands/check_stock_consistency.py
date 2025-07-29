"""
Django管理命令：检查和修复库存一致性
"""

import logging
from django.core.management.base import BaseCommand
from django.utils import timezone
from django.db import transaction, models

from apps.orders.tasks import check_stock_consistency
from apps.activities.models import SeckillActivity
from apps.orders.models import Order

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = '检查和修复库存一致性'
    
    def add_arguments(self, parser):
        parser.add_argument(
            '--fix',
            action='store_true',
            help='自动修复发现的库存不一致问题'
        )
        parser.add_argument(
            '--activity-id',
            type=int,
            help='只检查指定活动的库存一致性'
        )
        parser.add_argument(
            '--dry-run',
            action='store_true',
            help='只显示检查结果，不执行修复'
        )
        
    def handle(self, *args, **options):
        """主处理函数"""
        fix_issues = options['fix']
        activity_id = options['activity_id']
        dry_run = options['dry_run']
        
        self.stdout.write(
            self.style.SUCCESS(f'开始库存一致性检查 - 时间: {timezone.now()}')
        )
        
        if activity_id:
            self._check_single_activity(activity_id, fix_issues, dry_run)
        else:
            self._check_all_activities(fix_issues, dry_run)
            
    def _check_single_activity(self, activity_id: int, fix_issues: bool, dry_run: bool):
        """检查单个活动的库存一致性"""
        try:
            activity = SeckillActivity.objects.get(id=activity_id)
            self.stdout.write(f'检查活动: {activity.product.name} (ID: {activity_id})')
            
            inconsistency = self._check_activity_consistency(activity)
            
            if inconsistency:
                self._display_inconsistency(inconsistency)
                
                if fix_issues and not dry_run:
                    self._fix_activity_stock(activity, inconsistency)
                elif dry_run:
                    self.stdout.write(
                        self.style.WARNING('模拟运行模式 - 不会实际修复库存')
                    )
            else:
                self.stdout.write(
                    self.style.SUCCESS(f'活动 {activity_id} 库存一致')
                )
                
        except SeckillActivity.DoesNotExist:
            self.stdout.write(
                self.style.ERROR(f'活动 {activity_id} 不存在')
            )
            
    def _check_all_activities(self, fix_issues: bool, dry_run: bool):
        """检查所有活动的库存一致性"""
        self.stdout.write('执行全面库存一致性检查...')
        
        try:
            result = check_stock_consistency.delay()
            check_result = result.get(timeout=60)
            
            if check_result['success']:
                self.stdout.write(
                    self.style.SUCCESS(
                        f'检查完成 - 总计: {check_result["total_checked"]}, '
                        f'不一致: {check_result["inconsistent_count"]}'
                    )
                )
                
                if check_result['inconsistent_activities']:
                    for inconsistency in check_result['inconsistent_activities']:
                        self._display_inconsistency(inconsistency)
                        
                        if fix_issues and not dry_run:
                            try:
                                activity = SeckillActivity.objects.get(id=inconsistency['activity_id'])
                                self._fix_activity_stock(activity, inconsistency)
                            except SeckillActivity.DoesNotExist:
                                self.stdout.write(
                                    self.style.ERROR(f'活动 {inconsistency["activity_id"]} 不存在')
                                )
                        elif dry_run:
                            self.stdout.write(
                                self.style.WARNING('模拟运行模式 - 不会实际修复库存')
                            )
                else:
                    self.stdout.write(
                        self.style.SUCCESS('所有活动库存一致！')
                    )
            else:
                self.stdout.write(
                    self.style.ERROR(f'检查失败: {check_result.get("error")}')
                )
                
        except Exception as e:
            self.stdout.write(
                self.style.ERROR(f'执行检查时发生异常: {str(e)}')
            )
            
    def _check_activity_consistency(self, activity):
        """检查单个活动的一致性"""
        # 计算已售出的数量
        sold_quantity = Order.objects.filter(
            activity=activity,
            status='paid'
        ).aggregate(total=models.Sum('quantity'))['total'] or 0
        
        # 计算待支付的数量
        pending_quantity = Order.objects.filter(
            activity=activity,
            status='pending_payment'
        ).aggregate(total=models.Sum('quantity'))['total'] or 0
        
        # 理论库存
        theoretical_stock = activity.total_stock - sold_quantity - pending_quantity
        actual_stock = activity.available_stock
        
        if theoretical_stock != actual_stock:
            return {
                'activity_id': activity.id,
                'activity_name': activity.product.name,
                'total_stock': activity.total_stock,
                'sold_quantity': sold_quantity,
                'pending_quantity': pending_quantity,
                'theoretical_stock': theoretical_stock,
                'actual_stock': actual_stock,
                'difference': actual_stock - theoretical_stock
            }
        return None
        
    def _display_inconsistency(self, inconsistency):
        """显示不一致信息"""
        self.stdout.write(
            self.style.ERROR(
                f'\n库存不一致 - 活动: {inconsistency["activity_name"]} (ID: {inconsistency["activity_id"]})\n'
                f'  总库存: {inconsistency["total_stock"]}\n'
                f'  已售出: {inconsistency["sold_quantity"]}\n'
                f'  待支付: {inconsistency["pending_quantity"]}\n'
                f'  理论库存: {inconsistency["theoretical_stock"]}\n'
                f'  实际库存: {inconsistency["actual_stock"]}\n'
                f'  差异: {inconsistency["difference"]}'
            )
        )
        
    def _fix_activity_stock(self, activity, inconsistency):
        """修复活动库存"""
        try:
            with transaction.atomic():
                activity.available_stock = inconsistency['theoretical_stock']
                activity.save()
                
                self.stdout.write(
                    self.style.SUCCESS(
                        f'已修复活动 {activity.id} 的库存: '
                        f'{inconsistency["actual_stock"]} -> {inconsistency["theoretical_stock"]}'
                    )
                )
                
        except Exception as e:
            self.stdout.write(
                self.style.ERROR(f'修复活动 {activity.id} 库存时发生异常: {str(e)}')
            )
