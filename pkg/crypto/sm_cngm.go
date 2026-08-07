//go:build cngm
// +build cngm

// Package crypto — 国密算法实现 (SM2/SM3/SM4)。
//
// 启用方式: go build -tags cngm
//
// 依赖:
//   - github.com/tjfoc/gmsm (国密算法标准实现)
//   - SM2: 基于 SM2 P-256 曲线的椭圆曲线签名
//   - SM3: 256 位密码杂凑算法 (GB/T 32905-2016)
//   - SM4: 128 位分组密码算法 (GB/T 32907-2016)
//
// 信创合规:
//   - 满足 GB/T 39786-2021《信息安全技术 信息系统密码应用基本要求》
//   - 满足 GM/T 0054-2018《信息系统密码应用基本要求》
package crypto

// 此文件在 build tag "cngm" 启用时提供实际的国密算法实现。
// 默认情况下（无 cngm tag），使用标准库算法作为 fallback。
//
// 实际生产环境使用方式:
//   1. go get github.com/tjfoc/gmsm
//   2. 取消下面注释的 import 和实现
//   3. go build -tags cngm

/*
import (
	"crypto/rand"
	"fmt"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
)

// SM3Hasher 国密 SM3 哈希实现。
type SM3Hasher struct{}

func (h *SM3Hasher) Sum(data []byte) []byte {
	hash := sm3.Sm3Sum(data)
	return hash[:]
}

func (h *SM3Hasher) SumHex(data []byte) string {
	return hex.EncodeToString(h.Sum(data))
}

func (h *SM3Hasher) Size() int { return SM3Size }

func (h *SM3Hasher) Name() string { return "SM3" }

// SM4Cipher 国密 SM4 对称加密实现。
type SM4Cipher struct {
	key []byte
}

func NewSM4Cipher(key []byte) (*SM4Cipher, error) {
	if len(key) != SM4KeySize {
		return nil, fmt.Errorf("crypto: SM4 key must be %d bytes", SM4KeySize)
	}
	return &SM4Cipher{key: key}, nil
}

func (c *SM4Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	return sm4.Sm4Ecb(c.key, plaintext, true)
}

func (c *SM4Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	return sm4.Sm4Ecb(c.key, ciphertext, false)
}

func (c *SM4Cipher) Name() string { return "SM4-ECB" }

// SM2Signer 国密 SM2 签名实现。
type SM2Signer struct {
	privateKey *sm2.PrivateKey
	publicKey  *sm2.PublicKey
}

func NewSM2Signer(privateKeyHex string) (*SM2Signer, error) {
	priv, err := sm2.ReadPrivateKeyFromHex(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("crypto: parse SM2 private key: %w", err)
	}
	return &SM2Signer{
		privateKey: priv,
		publicKey:  &priv.PublicKey,
	}, nil
}

func (s *SM2Signer) Sign(data []byte) ([]byte, error) {
	return s.privateKey.Sign(rand.Reader, data, nil)
}

func (s *SM2Signer) Verify(data, signature []byte) bool {
	return s.publicKey.Verify(data, signature)
}

func (s *SM2Signer) Name() string { return "SM2" }
*/
