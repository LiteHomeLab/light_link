"""RPC 服务端示例"""
import asyncio
import logging
from lightlink.service import Service

logging.basicConfig(level=logging.INFO)


async def add_handler(args: dict) -> dict:
    """加法处理器"""
    a = args["a"]
    b = args["b"]
    return {"sum": a + b}


async def multiply_handler(args: dict) -> dict:
    """乘法处理器"""
    a = args["a"]
    b = args["b"]
    return {"product": a * b}


async def main():
    svc = Service("math-service", "nats://localhost:4222")

    await svc.register_rpc("add", add_handler)
    await svc.register_rpc("multiply", multiply_handler)

    await svc.start()
    logging.info("Service started. Press Ctrl+C to stop...")

    try:
        await asyncio.Future()
    except KeyboardInterrupt:
        await svc.stop()


if __name__ == "__main__":
    asyncio.run(main())
