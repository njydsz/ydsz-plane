# 国密算法 FIPS 合规文档（P2-12）

> 适用角色：运维工程师 / 等保测评人员 / 安全审计  
> 目标：说明 Plane 系统中国密算法的覆盖范围、构建方式及密钥管理，支撑等保三级合规  
> 最后更新：2025-07-10

---

## 1. 支持的国密算法清单

| 算法 | GM/T 标准编号 | 用途 | 密钥长度 |
|------|---------------|------|----------|
| SM2 | GM/T 0003.1-2012 | 非对称加密、数字签名 | 256 bit |
| SM3 | GM/T 0004-2012 | 哈希 / 消息摘要 | 256 bit 输出 |
| SM4 | GM/T 0002-2012 | 对称加密（CBC/GCM 模式） | 128 bit |

> 引用标准均符合国家密码管理局（OSCCA）发布规范。

---

## 2. Build Tag 切换说明

Plane 支持通过 Go build tag 切换密码算法实现，按实际部署环境选择：

### 2.1 Build Tag

| Build Tag | 默认算法 | 国密算法 | 说明 |
|-----------|----------|----------|------|
| 无 tag | AES-256-GCM / SHA-256 / ECDSA P-256 | — | 标准 FIPS 140-2 实现 |
| `cngm` | — | SM4 / SM2 / SM3 | 国密算法实现 |

### 2.2 构建示例

```bash
# 国密版本构建
CGO_ENABLED=0 go build -tags cngm -o plane-server ./cmd/server

# 国密版本镜像构建
docker build --build-arg GO_TAGS=cngm -t plane-server:cngm .
```

### 2.3 运行时代码路径

```
internal/infrastructure/crypto/cngm_init.go      -- 国密初始化（-tags cngm 时编译）
internal/infrastructure/crypto/standard_init.go  -- 标准算法初始化（默认编译）
```

编译时通过 build tag 自动选择实现，业务层代码无需修改。

---

## 3. 密钥管理

### 3.1 环境变量注入方式

密钥通过环境变量注入，不支持配置文件明文存储。部署时通过 Secret 管理工具（K8s Secret / Vault）下发：

| 环境变量名 | 用途 | 算法 |
|-----------|------|------|
| `PLANE_SM2_PRIVATE_KEY_HEX` | SM2 签名/解密私钥 | SM2 |
| `PLANE_SM2_PUBLIC_KEY_HEX` | SM2 验签/加密公钥 | SM2 |
| `PLANE_SM4_KEY_HEX` | SM4 加解密密钥 | SM4 |

### 3.2 密钥长度要求

| 算法 | 密钥长度 | 格式要求 |
|------|----------|----------|
| SM2 | 256 bit | Hex 编码，64 字符（不含 0x 前缀） |
| SM3 | N/A（哈希算法无密钥） | — |
| SM4 | 128 bit | Hex 编码，32 字符（不含 0x 前缀） |

### 3.3 密钥轮换策略

- client_secret 加密密钥：建议每 90 天轮换一次
- 轮换时新旧密钥共存窗口期 = 当前 JWT 最大有效期（默认 7 天），确保存量 token 仍可解密验证
- 轮换后无需重新登录所有用户

---

## 4. 密码应用场景

### 4.1 client_secret 加密存储

- **场景**：OAuth2 客户端密钥、第三方集成 API Key 持久化
- **算法**：SM4-CBC（默认）/ SM4-GCM
- **字段**：`clients.encrypted_secret`、`integrations.api_key_encrypted`

### 4.2 TLS 传输加密

- **场景**：服务端 HTTPS、gRPC 内部通信
- **证书**：SM2 证书链（需 OSCCA 认证的 SSL 证书）
- **Go 配置**：

```go
// -tags cngm 时自动启用 SM2 支持的 TLS cipher suites
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
    CipherSuites: []uint16{
        // GM/T 0024 定义的国密 cipher suite
        tls.TLS_SM4_GCM_SM3,
        tls.TLS_SM4_CCM_SM3,
    },
}
```

### 4.3 文件完整性校验

- **场景**：上传附件、导出数据的完整性摘要
- **算法**：SM3
- **字段**：`attachments.content_hash`（以 `sm3:` 前缀标识）

---

## 5. 等保三级对应条款

| 等保条款 | 条款描述 | Plane 对应措施 |
|----------|----------|----------------|
| 8.1.4.2 | 数据传输保密性：通信过程中使用密码技术保证数据不被泄露 | 服务端及内部通信启用 SM2 证书 + SM4 加密通道 |
| 8.1.4.3 | 数据存储保密性：采用密码技术保证涉及敏感信息的存储保密性 | client_secret 等敏感字段使用 SM4 加密后入库；密钥与数据分离管理（K8s Secret / Vault） |

### 合规自检清单

- [ ] 生产镜像使用 `-tags cngm` 构建
- [ ] TLS 配置包含 GM/T 0024 国密 cipher suite
- [ ] 环境变量中不包含明文密钥（通过 K8s Secret 或 Vault 注入）
- [ ] 敏感数据字段入库前有密文前缀标识（如 `enc_sm4:`）
- [ ] 密钥有明确的轮换记录和负责人
- [ ] TLS 证书为 OSCCA 认证机构签发

---

> 审批记录：本文件经安全团队评审，如有变更需重新走 review 流程。  
> 联系方式：安全组 #security-plane
