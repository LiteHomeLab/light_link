"""LightLink 服务端模块"""
import asyncio
import json
import uuid
from typing import Dict, Callable, Any, Optional
from datetime import datetime
import logging

from nats.aio.client import Client as NATSClient
from lightlink.types import RPCRequest, RPCResponse
from lightlink.metadata import ServiceMetadata, MethodMetadata

logger = logging.getLogger(__name__)

RPCHandler = Callable[[Dict[str, Any]], Any]


class Service:
    """LightLink 服务端"""

    DEFAULT_HEARTBEAT_INTERVAL = 30

    def __init__(
        self,
        name: str,
        nats_url: str = "nats://localhost:4222"
    ):
        self.name = name
        self.nats_url = nats_url
        self.nc: Optional[NATSClient] = None
        self._rpc_handlers: Dict[str, RPCHandler] = {}
        self._method_metadata: Dict[str, MethodMetadata] = {}
        self._heartbeat_task: Optional[asyncio.Task] = None
        self._heartbeat_stop = asyncio.Event()
        self._running = False

    async def register_rpc(self, method: str, handler: RPCHandler) -> None:
        """注册 RPC 方法"""
        self._rpc_handlers[method] = handler
        logger.info(f"Registered RPC method: {method}")

    async def register_method_with_metadata(
        self,
        method: str,
        handler: RPCHandler,
        metadata: MethodMetadata
    ) -> None:
        """注册带元数据的 RPC 方法"""
        self._method_metadata[method] = metadata
        await self.register_rpc(method, handler)

    async def has_rpc(self, method: str) -> bool:
        """检查方法是否已注册"""
        return method in self._rpc_handlers

    async def start(self) -> None:
        """启动服务"""
        if self._running:
            raise RuntimeError("Service already running")

        self.nc = NATSClient()
        await self.nc.connect(self.nats_url)

        subject = f"$SRV.{self.name}.>"
        await self.nc.subscribe(subject, cb=self._handle_rpc)

        await self._start_heartbeat()
        self._running = True
        logger.info(f"Service '{self.name}' started")

    async def _handle_rpc(self, msg) -> None:
        """处理 RPC 请求"""
        try:
            request_data = json.loads(msg.data.decode())
            request = RPCRequest(**request_data)

            handler = self._rpc_handlers.get(request.method)
            if handler is None:
                await self._send_error(msg, request.id, f"Method not found: {request.method}")
                return

            result = await handler(request.args)
            await self._send_success(msg, request.id, result)

        except Exception as e:
            logger.error(f"Error handling RPC: {e}")
            await self._send_error(msg, "", str(e))

    async def _send_success(self, msg, request_id: str, result: Dict[str, Any]) -> None:
        """发送成功响应"""
        response = RPCResponse(id=request_id, success=True, result=result)
        msg.respond(json.dumps(response.__dict__).encode())

    async def _send_error(self, msg, request_id: str, error: str) -> None:
        """发送错误响应"""
        response = RPCResponse(id=request_id, success=False, error=error)
        msg.respond(json.dumps(response.__dict__).encode())

    async def _start_heartbeat(self) -> None:
        """启动心跳"""
        async def heartbeat_loop():
            while not self._heartbeat_stop.is_set():
                await self._send_heartbeat()
                try:
                    await asyncio.wait_for(
                        self._heartbeat_stop.wait(),
                        timeout=self.DEFAULT_HEARTBEAT_INTERVAL
                    )
                except asyncio.TimeoutError:
                    continue

        self._heartbeat_task = asyncio.create_task(heartbeat_loop())

    async def _send_heartbeat(self) -> None:
        """发送心跳"""
        heartbeat = {
            "service": self.name,
            "version": "1.0.0",
            "timestamp": datetime.utcnow().isoformat()
        }
        subject = f"$LL.heartbeat.{self.name}"
        await self.nc.publish(subject, json.dumps(heartbeat).encode())

    async def stop(self) -> None:
        """停止服务"""
        if not self._running:
            return

        self._heartbeat_stop.set()
        if self._heartbeat_task:
            await self._heartbeat_task

        if self.nc:
            await self.nc.close()

        self._running = False
        logger.info(f"Service '{self.name}' stopped")
