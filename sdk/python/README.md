# LightLink Python SDK

Python 客户端和服务端 SDK 用于 LightLink 多语言 RPC 框架。

## 安装

```bash
pip install lightlink
```

## 快速开始

### 服务端

```python
import asyncio
from lightlink import Service

async def handler(args):
    return args

async def main():
    svc = Service("my-service", "nats://localhost:4222")
    await svc.register_rpc("echo", handler)
    await svc.start()

asyncio.run(main())
```

### 元数据注册

```python
from lightlink.metadata import MethodMetadata, ParameterMetadata

meta = MethodMetadata(
    name="add",
    params=[
        ParameterMetadata(name="a", type="number", required=True, description="First")
    ]
)

await svc.register_method_with_metadata("add", handler, meta)
```

## 协议兼容性

- JSON 格式与 Go SDK 完全一致
- 支持 asyncio
- 支持 NATS JetStream

## 示例

- `examples/rpc_service.py` - 基本 RPC 服务示例
- `examples/metadata_demo.py` - 元数据注册示例
