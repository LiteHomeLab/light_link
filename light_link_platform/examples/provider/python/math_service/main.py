"""Python Math Service - A complete math service with 4 methods"""
import sys
import os

# Add SDK to path using absolute path
sdk_path = r"C:\WorkSpace\Go2Hell\src\github.com\LiteHomeLab\light_link\sdk\python"
if sdk_path not in sys.path:
    sys.path.insert(0, sdk_path)

import asyncio
import logging
from lightlink.service import Service
from lightlink.metadata import (
    MethodMetadata, ParameterMetadata,
    ReturnMetadata, ExampleMetadata
)
from lightlink.client import discover_client_certs, create_ssl_context_from_discovery

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
    print("=== Python Math Service ===")
    print("\n[1/5] Discovering TLS certificates...")
    try:
        cert_result = discover_client_certs()
        print(f"Certificates found:")
        print(f"  CA:   {cert_result.ca_file}")
        print(f"  Cert: {cert_result.cert_file}")
        print(f"  Key:  {cert_result.key_file}")

        # Create SSL context from discovered certificates (skip verify for self-signed certs)
        ssl_ctx = create_ssl_context_from_discovery(cert_result, verify=False)

        print("\n[2/5] Creating service with TLS...")
        svc = Service(
            "math-service-python",
            "nats://172.18.200.47:4222",
            tls_config=ssl_ctx,
            auto_tls=False
        )
    except FileNotFoundError as e:
        print(f"ERROR: {e}")
        print("Please copy the 'client/' folder to your service directory.")
        return

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
    print("\n[3/5] Registering methods with metadata...")
    await svc.register_method_with_metadata("add", add_handler, add_meta)
    print("  - add: registered")
    await svc.register_method_with_metadata("multiply", multiply_handler, multiply_meta)
    print("  - multiply: registered")
    await svc.register_method_with_metadata("power", power_handler, power_meta)
    print("  - power: registered")
    await svc.register_method_with_metadata("divide", divide_handler, divide_meta)
    print("  - divide: registered")

    # Start service first (creates NATS connection)
    print("\n[4/5] Starting service...")
    await svc.start()
    print("Service started successfully!")

    # Build and register service metadata (requires NATS connection)
    print("\n[5/5] Registering service metadata...")
    metadata = svc.build_current_metadata(
        version="v1.0.0",
        description="A mathematical operations service providing basic and advanced math functions",
        author="LiteHomeLab",
        tags=["demo", "math", "calculator"]
    )
    await svc.register_metadata(metadata)
    print(f"Service metadata registered to NATS!")
    print(f"  Service: {metadata.name}")
    print(f"  Version: {metadata.version}")
    print(f"  Methods: {len(metadata.methods)}")

    print("\n=== Service Information ===")
    print(f"Service Name: {svc.name}")
    print("Registered Methods:")
    for name, meta in svc._method_metadata.items():
        print(f"  - {name}: {meta.description}")

    print("\n=== Math Service Complete ===")
    print("\nThe service is now running and will send heartbeat every 30 seconds.")
    print("Press Ctrl+C to stop the service.")

    try:
        await asyncio.Future()
    except KeyboardInterrupt:
        await svc.stop()


if __name__ == "__main__":
    asyncio.run(main())
