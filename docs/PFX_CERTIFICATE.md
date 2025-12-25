# PFX 证书生成说明

C# SDK 的 NATS.Client 需要使用 PFX 格式的证书文件（包含证书和私钥在同一文件中）。

## 生成 PFX 文件

使用 OpenSSL 将现有的 `.crt` 和 `.key` 文件转换为 `.pfx` 格式：

```bash
cd light_link_platform/client

openssl pkcs12 -export -out client.pfx \
    -inkey client.key \
    -in client.crt \
    -passout pass:lightlink
```

## 密码说明

默认密码为 `lightlink`。如需更改密码，请同步修改：
- `sdk/csharp/LightLink/TLSConfig.cs` 中的 `PfxPassword` 默认值

## 安全注意事项

- PFX 文件包含私钥，请勿提交到版本控制系统
- `client/` 目录已在 `.gitignore` 中排除
- 生产环境应使用更强的密码保护 PFX 文件
