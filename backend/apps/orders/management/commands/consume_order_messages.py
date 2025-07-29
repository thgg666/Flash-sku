"""
Django管理命令：消费Go秒杀服务发送的订单消息
"""

import json
import logging
import signal
import sys
from typing import Dict, Any

import pika
from django.core.management.base import BaseCommand
from django.conf import settings

from apps.orders.tasks import process_seckill_order_from_go

logger = logging.getLogger(__name__)


class Command(BaseCommand):
    help = '消费Go秒杀服务发送的订单消息'
    
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.connection = None
        self.channel = None
        self.should_stop = False
        
    def add_arguments(self, parser):
        parser.add_argument(
            '--queue',
            type=str,
            default='seckill_orders',
            help='RabbitMQ队列名称 (默认: seckill_orders)'
        )
        parser.add_argument(
            '--exchange',
            type=str,
            default='seckill',
            help='RabbitMQ交换机名称 (默认: seckill)'
        )
        parser.add_argument(
            '--routing-key',
            type=str,
            default='order.created',
            help='路由键 (默认: order.created)'
        )
        
    def handle(self, *args, **options):
        """主处理函数"""
        queue_name = options['queue']
        exchange_name = options['exchange']
        routing_key = options['routing_key']
        
        self.stdout.write(
            self.style.SUCCESS(f'开始消费订单消息 - 队列: {queue_name}, 交换机: {exchange_name}')
        )
        
        # 设置信号处理器
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
        
        try:
            # 建立RabbitMQ连接
            self._setup_rabbitmq_connection(queue_name, exchange_name, routing_key)
            
            # 开始消费消息
            self._start_consuming()
            
        except Exception as e:
            self.stdout.write(
                self.style.ERROR(f'消费者启动失败: {str(e)}')
            )
            logger.error(f"消费者启动失败: {str(e)}")
            sys.exit(1)
        finally:
            self._cleanup()
            
    def _setup_rabbitmq_connection(self, queue_name: str, exchange_name: str, routing_key: str):
        """设置RabbitMQ连接"""
        try:
            # 从Django设置获取RabbitMQ URL
            rabbitmq_url = getattr(settings, 'CELERY_BROKER_URL', 'amqp://guest:guest@localhost:5672/')
            
            # 建立连接
            parameters = pika.URLParameters(rabbitmq_url)
            self.connection = pika.BlockingConnection(parameters)
            self.channel = self.connection.channel()
            
            # 声明交换机
            self.channel.exchange_declare(
                exchange=exchange_name,
                exchange_type='topic',
                durable=True
            )
            
            # 声明队列
            self.channel.queue_declare(
                queue=queue_name,
                durable=True
            )
            
            # 绑定队列到交换机
            self.channel.queue_bind(
                exchange=exchange_name,
                queue=queue_name,
                routing_key=routing_key
            )
            
            # 设置QoS
            self.channel.basic_qos(prefetch_count=1)
            
            # 设置消费者
            self.channel.basic_consume(
                queue=queue_name,
                on_message_callback=self._on_message,
                auto_ack=False
            )
            
            logger.info(f"RabbitMQ连接建立成功 - 队列: {queue_name}")
            
        except Exception as e:
            logger.error(f"建立RabbitMQ连接失败: {str(e)}")
            raise
            
    def _on_message(self, channel, method, properties, body):
        """处理接收到的消息"""
        try:
            # 解析消息
            message = json.loads(body.decode('utf-8'))
            
            logger.info(f"接收到订单消息 - 订单ID: {message.get('order_id')}")
            
            # 异步处理订单消息
            task_result = process_seckill_order_from_go.delay(message)
            
            logger.info(f"订单消息已提交处理 - 任务ID: {task_result.id}")
            
            # 确认消息
            channel.basic_ack(delivery_tag=method.delivery_tag)
            
        except json.JSONDecodeError as e:
            logger.error(f"消息格式错误: {str(e)}, 消息内容: {body}")
            # 拒绝消息，不重新入队
            channel.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            
        except Exception as e:
            logger.error(f"处理消息时发生异常: {str(e)}")
            # 拒绝消息，重新入队
            channel.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
            
    def _start_consuming(self):
        """开始消费消息"""
        self.stdout.write(
            self.style.SUCCESS('开始监听订单消息...')
        )
        
        try:
            while not self.should_stop:
                self.connection.process_data_events(time_limit=1)
        except KeyboardInterrupt:
            self.stdout.write(
                self.style.WARNING('接收到中断信号，正在停止消费者...')
            )
            self.should_stop = True
            
    def _signal_handler(self, signum, frame):
        """信号处理器"""
        self.stdout.write(
            self.style.WARNING(f'接收到信号 {signum}，正在停止消费者...')
        )
        self.should_stop = True
        
    def _cleanup(self):
        """清理资源"""
        try:
            if self.channel and not self.channel.is_closed:
                self.channel.stop_consuming()
                self.channel.close()
                
            if self.connection and not self.connection.is_closed:
                self.connection.close()
                
            self.stdout.write(
                self.style.SUCCESS('消费者已停止，资源已清理')
            )
            
        except Exception as e:
            logger.error(f"清理资源时发生异常: {str(e)}")
