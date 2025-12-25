"""Python Math Service - A complete math service with 4 methods"""
import asyncio
import logging
from lightlink.service import Service
from lightlink.metadata import (
    MethodMetadata, ParameterMetadata,
    ReturnMetadata, ExampleMetadata
)

logging.basicConfig(level=logging.INFO)


async def add_handler(args: dict) -> dict:
    """Add two numbers"""
    a = args["a"]
    b = args["b"]
    return {"sum": a + b}


async def multiply_handler(args: dict) -> dict:
    """Multiply two numbers"""
    a = args["a"]
    b = args["b"]
    return {"product": a * b}


async def power_handler(args: dict) -> dict:
    """Calculate base to the power of exp"""
    base = args["base"]
    exp = args["exp"]
    result = base ** exp
    return {"result": result}


async def divide_handler(args: dict) -> dict:
    """Divide numerator by denominator"""
    numerator = args["numerator"]
    denominator = args["denominator"]
    if denominator == 0:
        raise ValueError("division by zero")
    return {"quotient": numerator / denominator}


async def main():
    svc = Service("math-service", "nats://localhost:4222")

    # Method metadata for add
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
                description="The sum of a and b"
            )
        ],
        example=ExampleMetadata(
            input={"a": 10, "b": 20},
            output={"sum": 30},
            description="10 + 20 = 30"
        )
    )

    # Method metadata for multiply
    multiply_meta = MethodMetadata(
        name="multiply",
        description="Multiply two numbers",
        params=[
            ParameterMetadata(
                name="a",
                type="number",
                required=True,
                description="First factor"
            ),
            ParameterMetadata(
                name="b",
                type="number",
                required=True,
                description="Second factor"
            )
        ],
        returns=[
            ReturnMetadata(
                name="product",
                type="number",
                description="The product of a and b"
            )
        ],
        example=ExampleMetadata(
            input={"a": 5, "b": 6},
            output={"product": 30},
            description="5 * 6 = 30"
        )
    )

    # Method metadata for power
    power_meta = MethodMetadata(
        name="power",
        description="Calculate a to the power of b",
        params=[
            ParameterMetadata(
                name="base",
                type="number",
                required=True,
                description="The base number"
            ),
            ParameterMetadata(
                name="exp",
                type="number",
                required=True,
                description="The exponent"
            )
        ],
        returns=[
            ReturnMetadata(
                name="result",
                type="number",
                description="base raised to the power of exp"
            )
        ],
        example=ExampleMetadata(
            input={"base": 2, "exp": 10},
            output={"result": 1024},
            description="2^10 = 1024"
        )
    )

    # Method metadata for divide
    divide_meta = MethodMetadata(
        name="divide",
        description="Divide two numbers",
        params=[
            ParameterMetadata(
                name="numerator",
                type="number",
                required=True,
                description="The number to be divided"
            ),
            ParameterMetadata(
                name="denominator",
                type="number",
                required=True,
                description="The number to divide by"
            )
        ],
        returns=[
            ReturnMetadata(
                name="quotient",
                type="number",
                description="The result of division"
            )
        ],
        example=ExampleMetadata(
            input={"numerator": 100, "denominator": 4},
            output={"quotient": 25},
            description="100 / 4 = 25"
        )
    )

    # Register methods with metadata
    await svc.register_method_with_metadata("add", add_handler, add_meta)
    await svc.register_method_with_metadata("multiply", multiply_handler, multiply_meta)
    await svc.register_method_with_metadata("power", power_handler, power_meta)
    await svc.register_method_with_metadata("divide", divide_handler, divide_meta)

    await svc.start()

    logging.info("Math service registered with 4 methods. Press Ctrl+C to stop...")
    try:
        await asyncio.Future()
    except KeyboardInterrupt:
        await svc.stop()


if __name__ == "__main__":
    asyncio.run(main())
