"""元数据注册示例"""
import asyncio
import logging
from lightlink.service import Service
from lightlink.metadata import (
    MethodMetadata, ParameterMetadata,
    ReturnMetadata, ExampleMetadata
)

logging.basicConfig(level=logging.INFO)


async def add_handler(args: dict) -> dict:
    """加法处理器"""
    a = args["a"]
    b = args["b"]
    return {"sum": a + b}


async def main():
    svc = Service("math-service", "nats://localhost:4222")

    add_meta = MethodMetadata(
        name="add",
        description="Add two numbers together",
        params=[
            ParameterMetadata(
                name="a",
                type="number",
                required=True,
                description="First number"
            ),
            ParameterMetadata(
                name="b",
                type="number",
                required=True,
                description="Second number"
            )
        ],
        returns=[
            ReturnMetadata(
                name="sum",
                type="number",
                description="The sum"
            )
        ],
        example=ExampleMetadata(
            input={"a": 10, "b": 20},
            output={"sum": 30},
            description="10 + 20 = 30"
        )
    )

    await svc.register_method_with_metadata("add", add_handler, add_meta)
    await svc.start()

    logging.info("Service with metadata registered. Press Ctrl+C to stop...")
    try:
        await asyncio.Future()
    except KeyboardInterrupt:
        await svc.stop()


if __name__ == "__main__":
    asyncio.run(main())
