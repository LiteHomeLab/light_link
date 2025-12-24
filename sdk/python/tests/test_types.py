import pytest
from lightlink.types import RPCRequest, RPCResponse


def test_rpc_request_creation():
    req = RPCRequest(id="123", method="test", args={"a": 1})
    assert req.id == "123"
    assert req.method == "test"


def test_rpc_response_creation():
    resp = RPCResponse(id="123", success=True, result={"value": 42})
    assert resp.success is True
    assert resp.result["value"] == 42
