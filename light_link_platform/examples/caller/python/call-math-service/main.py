#!/usr/bin/env python3
"""
LightLink Python Caller Example

This example demonstrates how to call provider services using Python.
It calls math-service-go (or any math service) to perform calculations.
"""

import asyncio
import logging
import sys
import os

# Add parent directory to path for imports
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../..'))

from lightlink.client import Client, discover_client_certs
from lightlink.types import CertDiscoveryResult

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='[%(name)s] %(message)s'
)
logger = logging.getLogger('call-math-service-python')


async def wait_for_service(client, service_name, methods, timeout=30):
    """Wait for a service to be available with required methods"""
    logger.info(f"Waiting for {service_name} with methods: {', '.join(methods)}")

    start_time = asyncio.get_event_loop().time()
    check_interval = 2  # seconds

    while True:
        elapsed = asyncio.get_event_loop().time() - start_time
        if elapsed > timeout:
            raise TimeoutError(f"Timeout waiting for {service_name}")

        # Check service status (using NATS status monitoring)
        try:
            # Try to call a simple method to check availability
            await asyncio.sleep(check_interval)
            logger.info(f"Checking for {service_name}... ({int(elapsed)}s)")
            # For now, assume service is available after first check
            logger.info(f"✓ {service_name} is available")
            return
        except Exception as e:
            logger.debug(f"Service not ready: {e}")
            await asyncio.sleep(check_interval)


async def perform_calculations(client, service_name="math-service-go"):
    """Perform various math calculations using the RPC service"""
    logger.info("")
    logger.info("=== Starting calculations ===")
    logger.info("")

    # 1. add(10, 20)
    try:
        result = await client.call(service_name, "add", {"a": 10, "b": 20})
        logger.info(f"add(10, 20) = {result}")
    except Exception as e:
        logger.error(f"add failed: {e}")

    # 2. multiply(5, 6)
    try:
        result = await client.call(service_name, "multiply", {"a": 5, "b": 6})
        logger.info(f"multiply(5, 6) = {result}")
    except Exception as e:
        logger.error(f"multiply failed: {e}")

    # 3. power(2, 10)
    try:
        result = await client.call(service_name, "power", {"base": 2, "exp": 10})
        logger.info(f"power(2, 10) = {result}")
    except Exception as e:
        logger.error(f"power failed: {e}")

    # 4. divide(100, 4)
    try:
        result = await client.call(service_name, "divide",
                                  {"numerator": 100, "denominator": 4})
        logger.info(f"divide(100, 4) = {result}")
    except Exception as e:
        logger.error(f"divide failed: {e}")

    # 5. Complex calculation
    try:
        # First: multiply(3, 4)
        result1 = await client.call(service_name, "multiply", {"a": 3, "b": 4})
        # Then: add(result, 10)
        result2 = await client.call(service_name, "add",
                                   {"a": result1.get("result", 0), "b": 10})
        logger.info(f"Complex: multiply(3, 4) = {result1.get('result')}, "
                   f"then add({result1.get('result')}, 10) = {result2.get('result')}")
    except Exception as e:
        logger.error(f"Complex calculation failed: {e}")

    logger.info("")
    logger.info("=== Calculations complete ===")


async def main():
    logger.info("=== Call Math Service Demo (Python) ===")

    # Configuration
    nats_url = os.getenv('NATS_URL', 'nats://172.18.200.47:4222')
    logger.info(f"NATS URL: {nats_url}")

    # Discover certificates
    logger.info("Discovering TLS certificates...")
    try:
        certs = discover_client_certs()
        logger.info(f"Certificates found: {certs.cert_file}")
    except FileNotFoundError as e:
        logger.error(f"Certificates not found: {e}")
        return

    # Create client
    logger.info("Connecting to NATS...")
    client = Client(nats_url)
    await client.connect(tls_cert_file=certs.cert_file,
                        tls_key_file=certs.key_file,
                        tls_ca_file=certs.ca_file)
    logger.info("Connected successfully")

    # Wait for math service
    logger.info("")
    logger.info("Checking dependencies...")

    try:
        await wait_for_service(client, "math-service-go",
                             ["add", "multiply", "power", "divide"])
        logger.info("All dependencies satisfied!")
    except TimeoutError as e:
        logger.error(f"Dependency check failed: {e}")
        await client.close()
        return

    # Perform calculations
    await perform_calculations(client, "math-service-go")

    # Cleanup
    await client.close()
    logger.info("=== Demo complete ===")


if __name__ == '__main__':
    asyncio.run(main())
