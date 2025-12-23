import asyncio
import json
import uuid
from nats.aio.client import Client as NATSClient
from nats.errors import TimeoutError


class Client:
    def __init__(self, url="nats://localhost:4222"):
        self.url = url
        self.nc = None

    async def connect(self):
        self.nc = NATSClient()
        await self.nc.connect(self.url)

    async def close(self):
        if self.nc:
            await self.nc.close()

    async def call(self, service, method, args, timeout=5.0):
        """RPC call"""
        subject = f"$SRV.{service}.{method}"
        request = {
            "id": str(uuid.uuid4()),
            "method": method,
            "args": args
        }

        try:
            msg = await self.nc.request(
                subject,
                json.dumps(request).encode(),
                timeout=timeout
            )
            response = json.loads(msg.data.decode())
            if not response.get("success"):
                raise Exception(response.get("error"))
            return response.get("result")
        except TimeoutError:
            raise Exception("RPC timeout")

    async def publish(self, subject, data):
        """Publish message"""
        await self.nc.publish(subject, json.dumps(data).encode())

    async def subscribe(self, subject, handler):
        """Subscribe to messages"""
        async def cb(msg):
            data = json.loads(msg.data.decode())
            await handler(data)

        await self.nc.subscribe(subject, cb)
