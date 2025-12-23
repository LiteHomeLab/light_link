import asyncio
import os
import sys

# Add parent directory to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../..", "sdk", "python"))

from lightlink import Client


async def main():
    # Get NATS URL from environment or use default
    nats_url = os.getenv("NATS_URL", "nats://172.18.200.47:4222")

    print("=== Python SDK Demo ===")
    print(f"NATS URL: {nats_url}")

    client = Client(url=nats_url)
    await client.connect()

    # Test 1: Pub/Sub
    print("\n[1/3] Testing pub/sub...")
    received = []

    async def handler(data):
        received.append(data)
        print(f"  Received: {data}")

    sub = await client.subscribe("test.python", handler)

    # Publish messages
    for i in range(3):
        await client.publish("test.python", {"id": i, "msg": f"Hello #{i}"})

    await asyncio.sleep(1)
    await client.unsubscribe(sub)
    print(f"  Total received: {len(received)} messages")

    # Test 2: State management
    print("\n[2/3] Testing state management...")
    await client.set_state("device.python01", {"temperature": 25.5, "humidity": 60})
    state = await client.get_state("device.python01")
    print(f"  State: {state}")

    # Test 3: File transfer
    print("\n[3/3] Testing file transfer...")
    test_file = "test_python.txt"
    download_file = "downloaded_python.txt"

    # Create test file
    with open(test_file, "w") as f:
        f.write("Hello from Python SDK!")

    try:
        # Upload
        file_id = await client.upload_file(test_file, test_file)
        print(f"  Uploaded: {file_id}")

        # Download
        await client.download_file(file_id, download_file)
        print(f"  Downloaded: {download_file}")

        # Verify
        with open(download_file, "r") as f:
            content = f.read()
        print(f"  Content: {content}")

        # Cleanup
        os.remove(test_file)
        os.remove(download_file)
        print("  File transfer verified!")
    except Exception as e:
        print(f"  File transfer error: {e}")

    await client.close()
    print("\n=== Python SDK Demo Complete ===")


if __name__ == "__main__":
    asyncio.run(main())
