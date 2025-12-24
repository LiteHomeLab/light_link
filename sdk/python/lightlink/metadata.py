"""LightLink 元数据定义"""
from dataclasses import dataclass, field
from typing import Dict, Any, List, Optional
from datetime import datetime


@dataclass
class ParameterMetadata:
    """参数元数据"""
    name: str
    type: str  # string, number, boolean, array, object
    required: bool
    description: str
    default: Optional[Any] = None


@dataclass
class ReturnMetadata:
    """返回值元数据"""
    name: str
    type: str
    description: str


@dataclass
class ExampleMetadata:
    """示例元数据"""
    input: Dict[str, Any]
    output: Dict[str, Any]
    description: str


@dataclass
class MethodMetadata:
    """方法元数据"""
    name: str
    description: str
    params: List[ParameterMetadata]
    returns: List[ReturnMetadata]
    example: Optional[ExampleMetadata] = None
    tags: List[str] = field(default_factory=list)
    deprecated: bool = False


@dataclass
class ServiceMetadata:
    """服务元数据"""
    name: str
    version: str
    description: str
    author: str
    tags: List[str]
    methods: List[MethodMetadata]
    registered_at: datetime
    last_seen: datetime
