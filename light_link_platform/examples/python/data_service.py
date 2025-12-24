"""Data Processing Service Example"""
import asyncio
import logging
from lightlink.service import Service
from lightlink.metadata import (
    ServiceMetadata, MethodMetadata, ParameterMetadata,
    ReturnMetadata, ExampleMetadata
)

logging.basicConfig(level=logging.INFO)


async def filter_handler(args: dict) -> dict:
    """Filter data by condition"""
    data = args.get("data", [])
    min_value = args.get("min", 0)

    filtered = [x for x in data if x >= min_value]
    return {"result": filtered}


async def transform_handler(args: dict) -> dict:
    """Transform data by multiplying"""
    data = args.get("data", [])
    multiplier = args.get("multiplier", 1)

    transformed = [x * multiplier for x in data]
    return {"result": transformed}


async def aggregate_handler(args: dict) -> dict:
    """Aggregate data with sum, avg, min, max"""
    data = args.get("data", [])

    if not data:
        return {"sum": 0, "avg": 0, "min": 0, "max": 0, "count": 0}

    return {
        "sum": sum(data),
        "avg": sum(data) / len(data),
        "min": min(data),
        "max": max(data),
        "count": len(data)
    }


async def main():
    # Try with just URL, no TLS
    import os
    nats_url = os.getenv("NATS_URL", "nats://172.18.200.47:4222")
    svc = Service("python-data-service", nats_url)

    # Register filter method with metadata
    filter_meta = MethodMetadata(
        name="filter",
        description="Filter numeric data by minimum value",
        params=[
            ParameterMetadata(
                name="data",
                type="array",
                required=True,
                description="Array of numbers to filter"
            ),
            ParameterMetadata(
                name="min",
                type="number",
                required=False,
                description="Minimum value threshold",
                default=0
            )
        ],
        returns=[
            ReturnMetadata(
                name="result",
                type="array",
                description="Filtered array"
            )
        ],
        example=ExampleMetadata(
            input={"data": [1, 5, 3, 7, 2], "min": 3},
            output={"result": [5, 3, 7]},
            description="Filter values >= 3"
        ),
        tags=["data", "filter"]
    )
    await svc.register_method_with_metadata("filter", filter_handler, filter_meta)

    # Register transform method with metadata
    transform_meta = MethodMetadata(
        name="transform",
        description="Transform numeric data by multiplying",
        params=[
            ParameterMetadata(
                name="data",
                type="array",
                required=True,
                description="Array of numbers to transform"
            ),
            ParameterMetadata(
                name="multiplier",
                type="number",
                required=False,
                description="Multiplication factor",
                default=1
            )
        ],
        returns=[
            ReturnMetadata(
                name="result",
                type="array",
                description="Transformed array"
            )
        ],
        example=ExampleMetadata(
            input={"data": [1, 2, 3], "multiplier": 2},
            output={"result": [2, 4, 6]},
            description="Multiply each value by 2"
        ),
        tags=["data", "transform"]
    )
    await svc.register_method_with_metadata("transform", transform_handler, transform_meta)

    # Register aggregate method with metadata
    aggregate_meta = MethodMetadata(
        name="aggregate",
        description="Calculate statistics (sum, avg, min, max, count)",
        params=[
            ParameterMetadata(
                name="data",
                type="array",
                required=True,
                description="Array of numbers to analyze"
            )
        ],
        returns=[
            ReturnMetadata(name="sum", type="number", description="Sum of all values"),
            ReturnMetadata(name="avg", type="number", description="Average value"),
            ReturnMetadata(name="min", type="number", description="Minimum value"),
            ReturnMetadata(name="max", type="number", description="Maximum value"),
            ReturnMetadata(name="count", type="number", description="Number of items")
        ],
        example=ExampleMetadata(
            input={"data": [1, 2, 3, 4, 5]},
            output={"sum": 15, "avg": 3.0, "min": 1, "max": 5, "count": 5},
            description="Calculate statistics"
        ),
        tags=["data", "aggregate", "statistics"]
    )
    await svc.register_method_with_metadata("aggregate", aggregate_handler, aggregate_meta)

    await svc.start()

    # Build and register metadata
    metadata = svc.build_current_metadata(
        version="1.0.0",
        description="Python Data Processing Service - Filter, transform and aggregate numeric data",
        author="LightLink Team",
        tags=["python", "data", "processing"]
    )
    await svc.register_metadata(metadata)

    logging.info("Python Data Service started and registered.")
    logging.info("Service: python-data-service")
    logging.info("Methods: filter, transform, aggregate")
    logging.info("Press Ctrl+C to stop...")

    try:
        await asyncio.Future()
    except KeyboardInterrupt:
        await svc.stop()


if __name__ == "__main__":
    asyncio.run(main())
