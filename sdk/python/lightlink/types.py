"""LightLink 类型定义"""
from dataclasses import dataclass, field
from typing import Dict, Any, Optional


@dataclass
class RPCRequest:
    """RPC 请求"""
    id: str = ""
    method: str = ""
    args: Dict[str, Any] = field(default_factory=dict)


@dataclass
class RPCResponse:
    """RPC 响应"""
    id: str = ""
    success: bool = False
    result: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
