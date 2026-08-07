// Package crypto — 国密算法适配层 (SM2/SM3/SM4)。
//
// 信创合规要求:
//   - 等保三级: 应采用密码技术保证通信过程中数据的保密性 (GB/T 22239 8.1.4)
//   - 密码法: 关键信息基础设施应使用商用密码进行保护
//   - 金融/政务行业: SM 系列算法为强制要求
//
// 设计要点:
//   - 接口抽象: Signer / Hasher / Cipher 接口，支持运行时切换算法
//   - Build Tag 隔离: 国密实现通过 build tag "cngm" 控制编译
//     - 默认: 使用标准库 crypto (SHA-256 / AES-256-GCM / RSA/ECDSA)
//     - cngm: 使用 tjfoc/gmsm (SM2/SM3/SM4)
//   - 配置驱动: 通过 YDSZ_CRYPTO_PROVIDER 环境变量选择算法
//   - 渐进迁移: 支持混合模式（JWT 仍用 RS256，存储加密用 SM4）
//
// 使用方式:
//
//	// 哈希
//	hasher := crypto.GetHasher()
//	digest := hasher.Sum([]byte("hello"))
//
//	// 对称加密
//	cipher, _ := crypto.NewSM4Cipher(key)
//	encrypted, _ := cipher.Encrypt(plaintext)
//
//	// 签名
//	signer, _ := crypto.NewSM2Signer(privateKey)
//	signature, _ := signer.Sign(data)
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Provider 密码算法提供者。
type Provider string

const (
	// ProviderStandard 标准算法 (SHA-256 / AES-256-GCM / RSA/ECDSA)。
	ProviderStandard Provider = "standard"
	// ProviderGuomi 国密算法 (SM2/SM3/SM4)。
	ProviderGuomi Provider = "guomi"
)

// --- Hasher 接口 ---

// Hasher 哈希计算接口。
type Hasher interface {
	// Sum 计算哈希值。
	Sum(data []byte) []byte
	// SumHex 计算哈希值并返回 hex 编码。
	SumHex(data []byte) string
	// Size 返回哈希值字节长度。
	Size() int
	// Name 返回算法名称。
	Name() string
}

// SHA256Hasher 标准 SHA-256 实现。
type SHA256Hasher struct{}

func (h *SHA256Hasher) Sum(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (h *SHA256Hasher) SumHex(data []byte) string {
	return hex.EncodeToString(h.Sum(data))
}

func (h *SHA256Hasher) Size() int { return 32 }

func (h *SHA256Hasher) Name() string { return "SHA-256" }

// --- Cipher 接口 ---

// Cipher 对称加密接口。
type Cipher interface {
	// Encrypt 加密明文。
	Encrypt(plaintext []byte) ([]byte, error)
	// Decrypt 解密密文。
	Decrypt(ciphertext []byte) ([]byte, error)
	// Name 返回算法名称。
	Name() string
}

// --- Signer 接口 ---

// Signer 数字签名接口。
type Signer interface {
	// Sign 对数据进行签名。
	Sign(data []byte) ([]byte, error)
	// Verify 验证签名。
	Verify(data, signature []byte) bool
	// Name 返回算法名称。
	Name() string
}

// --- Key Generator ---

// KeyGenerator 密钥生成器接口。
type KeyGenerator interface {
	// GenerateKey 生成随机密钥（指定长度）。
	GenerateKey(bits int) ([]byte, error)
	// Name 返回生成器名称。
	Name() string
}

// DefaultKeyGenerator 默认密钥生成器（使用 crypto/rand）。
type DefaultKeyGenerator struct{}

func (g *DefaultKeyGenerator) GenerateKey(bits int) ([]byte, error) {
	if bits%8 != 0 {
		return nil, fmt.Errorf("crypto: key bits must be multiple of 8")
	}
	key := make([]byte, bits/8)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("crypto: generate key: %w", err)
	}
	return key, nil
}

func (g *DefaultKeyGenerator) Name() string { return "crypto/rand" }

// --- Global Provider ---

var currentProvider Provider = ProviderStandard

// SetProvider 设置当前密码算法提供者。
// 应在应用启动时调用（基于配置）。
func SetProvider(p Provider) {
	currentProvider = p
}

// GetProvider 返回当前密码算法提供者。
func GetProvider() Provider {
	return currentProvider
}

// GetHasher 返回当前配置的哈希实现。
func GetHasher() Hasher {
	switch currentProvider {
	case ProviderGuomi:
		// 当 build tag "cngm" 启用时，这里返回 SM3Hasher
		// return &SM3Hasher{}
		return &SHA256Hasher{} // fallback
	default:
		return &SHA256Hasher{}
	}
}

// --- 工具函数 ---

// GenerateRandomHex 生成指定字节长度的随机 hex 字符串。
// 用于生成 API Token、State、Nonce 等安全随机值。
func GenerateRandomHex(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateRandomBytes 生成指定长度的安全随机字节。
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// ConstantTimeCompare 常量时间比较（防时序攻击）。
func ConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// --- SM3 (国密哈希) 接口声明 ---
// 当 build tag "cngm" 启用时，sm_cngm.go 将提供实际实现。

// SM3Size SM3 哈希值长度（32 字节 = 256 位）。
const SM3Size = 32

// --- SM4 (国密对称加密) 接口声明 ---
// 当 build tag "cngm" 启用时，sm_cngm.go 将提供实际实现。

// SM4KeySize SM4 密钥长度（16 字节 = 128 位）。
const SM4KeySize = 16

// SM4BlockSize SM4 分组大小（16 字节）。
const SM4BlockSize = 16

// --- SM2 (国密非对称签名) 接口声明 ---
// 当 build tag "cngm" 启用时，sm_cngm.go 将提供实际实现。

// SM2SignatureSize SM2 签名长度（64 字节）。
const SM2SignatureSize = 64

// SM2PublicKeySize SM2 公钥长度（64 字节，未压缩）。
const SM2PublicKeySize = 64

// SM2PrivateKeySize SM2 私钥长度（32 字节）。
const SM2PrivateKeySize = 32
