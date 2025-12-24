"""RPC 服务端示例 (with TLS)"""
import asyncio
import ssl
import logging

from lightlink.service import Service

logging.basicConfig(level=logging.INFO)

# Create SSL context with client certificates (Windows paths)
ssl_ctx = ssl.create_default_context(ssl.Purpose.CLIENT_AUTH)
ssl_ctx.load_cert_chain(
    r"C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\light_link\deploy\nats\tls\client-app.crt",
    r"C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\light_link\deploy\nats\tls\client-app.key"
)
ssl_ctx.load_verify_locations(r"C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\light_link\deploy\nats\tls\ca.crt")
ssl_ctx.verify_mode = ssl.CERT_REQUIRED


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
    svc = Service(
        "python-math-service",
        "tls://localhost:4222",
        tls_config=ssl_ctx
    )

    await svc.register_rpc("add", add_handler)
    await svc.register_rpc("multiply", multiply_handler)

    await svc.start()
    logging.info("Python Service with TLS started. Press Ctrl+C to stop...")

    try:
        await asyncio.Future()
    except KeyboardInterrupt:
        await svc.stop()


if __name__ == "__main__":
    asyncio.run(main())
