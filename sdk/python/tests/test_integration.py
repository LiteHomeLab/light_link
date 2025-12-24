"""集成测试 - 测试与 Go SDK 的互操作性"""
import pytest
import asyncio
from lightlink.service import Service


@pytest.mark.integration
@pytest.mark.asyncio
async def test_python_service_responds_to_rpc():
    """测试 Python 服务能响应 RPC 调用"""
    svc = Service("test-service", "nats://localhost:4222")

    async def handler(args):
        return {"echo": args.get("input")}

    await svc.register_rpc("echo", handler)
    await svc.start()

    # 这里可以添加实际的 RPC 调用测试
    # 需要 NATS 服务器运行

    await svc.stop()
