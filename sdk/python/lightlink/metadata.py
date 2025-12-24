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

    def to_dict(self) -> Dict[str, Any]:
        return {
            "name": self.name,
            "type": self.type,
            "required": self.required,
            "description": self.description,
            "default": self.default
        }


@dataclass
class ReturnMetadata:
    """返回值元数据"""
    name: str
    type: str
    description: str

    def to_dict(self) -> Dict[str, Any]:
        return {
            "name": self.name,
            "type": self.type,
            "description": self.description
        }


@dataclass
class ExampleMetadata:
    """示例元数据"""
    input: Dict[str, Any]
    output: Dict[str, Any]
    description: str

    def to_dict(self) -> Dict[str, Any]:
        return {
            "input": self.input,
            "output": self.output,
            "description": self.description
        }


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

    def to_dict(self) -> Dict[str, Any]:
        return {
            "name": self.name,
            "description": self.description,
            "params": [p.to_dict() for p in self.params],
            "returns": [r.to_dict() for r in self.returns],
            "example": self.example.to_dict() if self.example else None,
            "tags": self.tags,
            "deprecated": self.deprecated
        }


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

    def to_dict(self) -> Dict[str, Any]:
        return {
            "name": self.name,
            "version": self.version,
            "description": self.description,
            "author": self.author,
            "tags": self.tags,
            "methods": [m.to_dict() for m in self.methods],
            "registered_at": self.registered_at.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "last_seen": self.last_seen.strftime("%Y-%m-%dT%H:%M:%SZ")
        }
