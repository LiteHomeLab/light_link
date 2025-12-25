import asyncio
import json
import uuid
import os
import ssl
from pathlib import Path
from typing import Optional
from nats.aio.client import Client as NATSClient
from nats.errors import TimeoutError, NotFoundError, BadRequestError


# Certificate discovery constants
DEFAULT_CLIENT_CERT_DIR = "./client"
DEFAULT_SERVER_CERT_DIR = "./nats-server"
DEFAULT_SERVER_NAME = "nats-server"


class CertDiscoveryResult:
    """Certificate discovery result"""
    def __init__(self, ca_file: str, cert_file: str, key_file: str, server_name: str, found: bool):
        self.ca_file = ca_file
        self.cert_file = cert_file
        self.key_file = key_file
        self.server_name = server_name
        self.found = found


def discover_client_certs() -> CertDiscoveryResult:
    """
    Automatically discover client certificates.

    Search order:
    1. ./client
    2. ../client
    3. ../../client
    4. ../../../client
    5. ../../../../client
    6. ../../../../../client

    Returns:
        CertDiscoveryResult: Discovery result

    Raises:
        FileNotFoundError: When certificates are not found
    """
    search_paths = [
        DEFAULT_CLIENT_CERT_DIR,
        "../client",
        "../../client",
        "../../../client",
        "../../../../client",
        "../../../../../client"
    ]

    for base_path in search_paths:
        result = _check_cert_directory(base_path, "client")
        if result.found:
            return result

    raise FileNotFoundError(
        f"Client certificates not found in search paths: {search_paths}. "
        f"Please copy the 'client/' folder from generated certificates to your project."
    )


def discover_server_certs() -> CertDiscoveryResult:
    """
    Automatically discover server certificates.

    Search order:
    1. ./nats-server
    2. ../nats-server
    3. ../../nats-server
    4. ../../../nats-server
    5. ../../../../nats-server
    6. ../../../../../nats-server

    Returns:
        CertDiscoveryResult: Discovery result

    Raises:
        FileNotFoundError: When certificates are not found
    """
    search_paths = [
        DEFAULT_SERVER_CERT_DIR,
        "../nats-server",
        "../../nats-server",
        "../../../nats-server",
        "../../../../nats-server",
        "../../../../../nats-server"
    ]

    for base_path in search_paths:
        result = _check_cert_directory(base_path, "server")
        if result.found:
            return result

    raise FileNotFoundError(
        f"Server certificates not found in search paths: {search_paths}. "
        f"Please copy the 'nats-server/' folder from generated certificates to your project."
    )


def _check_cert_directory(dir_path: str, cert_type: str) -> CertDiscoveryResult:
    """Check if certificate files exist in directory"""
    cert_file = os.path.join(dir_path, f"{cert_type}.crt")
    key_file = os.path.join(dir_path, f"{cert_type}.key")
    ca_file = os.path.join(dir_path, "ca.crt")

    if os.path.isfile(ca_file) and os.path.isfile(cert_file) and os.path.isfile(key_file):
        return CertDiscoveryResult(
            ca_file=ca_file,
            cert_file=cert_file,
            key_file=key_file,
            server_name=DEFAULT_SERVER_NAME,
            found=True
        )

    return CertDiscoveryResult("", "", "", "", False)


def create_ssl_context_from_discovery(result: CertDiscoveryResult) -> ssl.SSLContext:
    """Create SSL context from discovery result"""
    ssl_ctx = ssl.create_default_context(ssl.Purpose.SERVER_AUTH)
    ssl_ctx.load_verify_locations(result.ca_file)
    ssl_ctx.load_cert_chain(
        certfile=result.cert_file,
        keyfile=result.key_file
    )
    ssl_ctx.minimum_version = ssl.TLSVersion.TLSv1_2
    if result.server_name:
        ssl_ctx.server_hostname = result.server_name
    return ssl_ctx


class TLSConfig:
    """TLS configuration"""
    def __init__(self, ca_file, cert_file, key_file, server_name=None):
        self.ca_file = ca_file
        self.cert_file = cert_file
        self.key_file = key_file
        self.server_name = server_name or DEFAULT_SERVER_NAME

    @classmethod
    def from_auto_discovery(cls) -> 'TLSConfig':
        """
        Create TLS configuration from auto-discovery.

        Returns:
            TLSConfig: TLS configuration object

        Raises:
            FileNotFoundError: When certificates are not found
        """
        result = discover_client_certs()
        return cls(
            ca_file=result.ca_file,
            cert_file=result.cert_file,
            key_file=result.key_file,
            server_name=result.server_name
        )


class Client:
    """LightLink Python Client"""
    def __init__(self, url="nats://172.18.200.47:4222", tls_config=None, auto_tls=False):
        """
        Initialize client.

        Args:
            url: NATS server URL
            tls_config: TLS configuration (TLSConfig object)
            auto_tls: Whether to auto-discover TLS certificates (mutually exclusive with tls_config)
        """
        if auto_tls and tls_config:
            raise ValueError("Cannot specify both auto_tls and tls_config")

        self.url = url
        self.tls_config = tls_config
        self.auto_tls = auto_tls
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

        # Handle TLS configuration
        if self.auto_tls:
            # Auto-discover certificates
            discovery_result = discover_client_certs()
            ssl_ctx = create_ssl_context_from_discovery(discovery_result)
            options["tls"] = ssl_ctx
        elif self.tls_config:
            # Use provided TLS configuration
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
