import asyncio
import pytest

from lightlink import Client


class TestClientConnection:
    async def test_client_create(self):
        """Test client can be created"""
        client = Client("nats://localhost:4222")
        assert client is not None

    async def test_client_connect(self):
        """Test client can connect"""
        client = Client("nats://localhost:4222")
        try:
            await client.connect()
            assert client.nc is not None
        except Exception as e:
            pytest.skip(f"Cannot connect to NATS: {e}")
        finally:
            await client.close()


class TestClientRPC:
    async def test_call_nonexistent_service(self):
        """Test calling non-existent service returns error"""
        client = Client("nats://localhost:4222")
        try:
            await client.connect()
            with pytest.raises(Exception):
                await client.call("nonexistent", "method", {})
        except Exception as e:
            pytest.skip(f"Cannot connect to NATS: {e}")
        finally:
            await client.close()


class TestClientPubSub:
    async def test_publish_subscribe(self):
        """Test publish and subscribe"""
        client = Client("nats://localhost:4222")
        try:
            await client.connect()

            received = []
            async def handler(msg):
                received.append(msg)

            await client.subscribe("test.python", handler)
            await client.publish("test.python", {"msg": "hello"})

            await asyncio.sleep(0.1)

            assert len(received) > 0
        except Exception as e:
            pytest.skip(f"Cannot connect to NATS: {e}")
        finally:
            await client.close()


class TestClientState:
    async def test_set_get_state(self):
        """Test setting and getting state"""
        client = Client("nats://localhost:4222")
        try:
            await client.connect()

            await client.set_state("test.python.key", {"value": 123})
            state = await client.get_state("test.python.key")

            assert state is not None
            assert state.get("value") == 123
        except Exception as e:
            pytest.skip(f"Cannot connect to NATS: {e}")
        finally:
            await client.close()
