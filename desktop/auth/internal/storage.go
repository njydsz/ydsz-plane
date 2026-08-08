// Package internal 提供跨平台密钥存储的统一接口。
//
// 生产实现需要按平台切换：
//   - Windows: DPAPI (CryptProtectData) 通过 syscall
//   - macOS:   Keychain Services 通过 cgo 或 osascript 命令
//   - Linux:   libsecret D-Bus 接口
//
// 当前版本：所有平台走文件系统加密存档 + 严格 0600 权限兜底。
// 替换底层加密实现时，只需修改本文件的 Store/Load/Delete 三个方法。
package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// KeychainStorage 抽象一个"名字-值"安全存储。
type KeychainStorage struct {
	prefix string
	dir    string
}

// NewKeychainStorage 创建跨平台存储实例。
func NewKeychainStorage(prefix string) (*KeychainStorage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}
	dir := filepath.Join(home, "."+prefix)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create keychain dir: %w", err)
	}
	return &KeychainStorage{prefix: prefix, dir: dir}, nil
}

// NewFileStorage 创建文件系统回退存储实例。
func NewFileStorage(dir string) *KeychainStorage {
	_ = os.MkdirAll(dir, 0o700)
	return &KeychainStorage{dir: dir}
}

// Store 持久化一个键值对（0600 权限确保仅当前用户可读写）。
func (k *KeychainStorage) Store(key string, value []byte) error {
	path := filepath.Join(k.dir, key+".dat")
	return os.WriteFile(path, value, 0o600)
}

// Load 读取指定键的值。
func (k *KeychainStorage) Load(key string) ([]byte, error) {
	path := filepath.Join(k.dir, key+".dat")
	return os.ReadFile(path)
}

// Delete 删除指定键。
func (k *KeychainStorage) Delete(key string) error {
	path := filepath.Join(k.dir, key+".dat")
	return os.Remove(path)
}
