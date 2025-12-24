import pytest
import asyncio
from lightlink.service import Service


@pytest.mark.asyncio
async def test_register_rpc():
    svc = Service("test", "nats://localhost:4222")
    await svc.register_rpc("test", lambda args: {"result": 42})
    assert await svc.has_rpc("test")


@pytest.mark.asyncio
async def test_service_name():
    svc = Service("test-service", "nats://localhost:4222")
    assert svc.name == "test-service"
