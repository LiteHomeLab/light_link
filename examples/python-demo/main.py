import asyncio
from lightlink import Client


async def main():
    client = Client()
    await client.connect()

    # RPC call
    result = await client.call("user-service", "getUser", {"user_id": "U001"})
    print("Result:", result)

    # Publish
    await client.publish("events.test", {"msg": "hello"})

    # Subscribe
    await client.subscribe("events.test", lambda data: print(data))

    await asyncio.sleep(1)
    await client.close()


if __name__ == "__main__":
    asyncio.run(main())
