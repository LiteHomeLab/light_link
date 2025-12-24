import asyncio
import json
import uuid
import os
import ssl
from nats.aio.client import Client as NATSClient
from nats.errors import TimeoutError, NotFoundError, BadRequestError


class TLSConfig:
    """TLS configuration"""
    def __init__(self, ca_file, cert_file, key_file, server_name=None):
        self.ca_file = ca_file
        self.cert_file = cert_file
        self.key_file = key_file
        self.server_name = server_name


class Client:
    """LightLink Python Client"""
    def __init__(self, url="nats://172.18.200.47:4222", tls_config=None):
        self.url = url
        self.tls_config = tls_config
        self.nc = None
        self.js = None
        self._subscriptions = []

    async def connect(self):
        """Connect to NATS server"""
        self.nc = NATSClient()

        options = {
            "name": "LightLink Python Client",
            "reconnect_time_wait": 2,
            "max_reconnect_attempts": 10,
        }

        # Add TLS configuration
        if self.tls_config:
            ssl_ctx = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
            ssl_ctx.load_verify_locations(self.tls_config.ca_file)
            ssl_ctx.load_cert_chain(
                certfile=self.tls_config.cert_file,
                keyfile=self.tls_config.key_file
            )
            ssl_ctx.minimum_version = ssl.TLSVersion.TLSv1_2
            if self.tls_config.server_name:
                ssl_ctx.server_hostname = self.tls_config.server_name
            options["tls"] = ssl_ctx

        await self.nc.connect(self.url, **options)

        # Get JetStream context
        self.js = self.nc.jetstream()

    async def close(self):
        """Close connection"""
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

        sub = await self.nc.subscribe(subject, cb=cb)
        self._subscriptions.append(sub)
        return sub

    async def unsubscribe(self, sub):
        """Unsubscribe"""
        await sub.unsubscribe()
        if sub in self._subscriptions:
            self._subscriptions.remove(sub)

    async def set_state(self, key, value):
        """Set state value"""
        kv = await self._get_or_create_kv("light_link_state")
        await kv.put(key, json.dumps(value).encode())

    async def get_state(self, key):
        """Get state value"""
        kv = await self._get_or_create_kv("light_link_state")
        entry = await kv.get(key)
        return json.loads(entry.value.decode())

    async def watch_state(self, key, handler):
        """Watch state changes"""
        kv = await self._get_or_create_kv("light_link_state")
        watcher = await kv.watch(key)

        async def _watch():
            async for entry in watcher:
                if entry is not None:
                    value = json.loads(entry.value.decode())
                    await handler(value)

        task = asyncio.create_task(_watch())
        return task

    async def _get_or_create_kv(self, bucket_name):
        """Get or create KV bucket"""
        try:
            kv = await self.js.key_value(bucket_name)
        except (NotFoundError, BadRequestError):
            kv = await self.js.create_key_value(bucket=bucket_name)
        return kv

    async def upload_file(self, file_path, remote_name):
        """Upload file to Object Store"""
        obj_store = await self._get_or_create_object_store("light_link_files")

        # Generate file ID
        file_id = str(uuid.uuid4())

        # Upload file
        file_size = os.path.getsize(file_path)
        chunk_size = 1024 * 1024  # 1MB chunks

        with open(file_path, "rb") as f:
            chunk_num = 0
            while True:
                chunk = f.read(chunk_size)
                if not chunk:
                    break
                chunk_key = f"{file_id}_{chunk_num}"
                await obj_store.put(chunk_key, chunk)
                chunk_num += 1

        # Publish file metadata
        metadata = {
            "file_id": file_id,
            "file_name": remote_name,
            "file_size": file_size,
            "chunk_num": chunk_num
        }
        await self.publish("file.uploaded", metadata)

        return file_id

    async def download_file(self, file_id, local_path):
        """Download file from Object Store"""
        obj_store = await self._get_or_create_object_store("light_link_files")

        # Subscribe to get metadata first (or we could store in KV)
        # For simplicity, we'll download all chunks
        chunk_num = 0
        with open(local_path, "wb") as f:
            while True:
                chunk_key = f"{file_id}_{chunk_num}"
                try:
                    chunk = await obj_store.get(chunk_key)
                    f.write(chunk.data)
                    chunk_num += 1
                except NotFoundError:
                    # No more chunks
                    break
                except (TimeoutError, BadRequestError) as e:
                    raise RuntimeError(f"Failed to download chunk {chunk_num}: {e}")

    async def _get_or_create_object_store(self, bucket_name):
        """Get or create Object Store"""
        try:
            obj_store = await self.js.object_store(bucket_name)
        except (NotFoundError, BadRequestError):
            obj_store = await self.js.create_object_store(bucket_name)
        return obj_store
