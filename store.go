package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/majd/ipatool/v2/pkg/keychain"
	"github.com/majd/ipatool/v2/pkg/util/machine"
	"github.com/majd/ipatool/v2/pkg/util/operatingsystem"
)

// fileKeychain là keychain tối giản lưu JSON trong thư mục dữ liệu của app.
// ipatool chỉ cần Get/Set/Remove nên không cần keyring hệ thống.
type fileKeychain struct {
	mu   sync.Mutex
	path string
	data map[string][]byte
}

var errKeyNotFound = errors.New("key not found")

func newFileKeychain(path string) *fileKeychain {
	k := &fileKeychain{path: path, data: map[string][]byte{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &k.data)
	}
	return k
}

func (k *fileKeychain) flush() error {
	b, err := json.MarshalIndent(k.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(k.path, b, 0o600)
}

func (k *fileKeychain) Get(key string) ([]byte, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	v, ok := k.data[key]
	if !ok {
		return nil, errKeyNotFound
	}
	return v, nil
}

func (k *fileKeychain) Set(key string, data []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = data
	return k.flush()
}

func (k *fileKeychain) Remove(key string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	delete(k.data, key)
	return k.flush()
}

var _ keychain.Keychain = (*fileKeychain)(nil)

// stateDir trả về thư mục chứa cookie + tài khoản đã đăng nhập.
func stateDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "IPADownloader")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

type storeDeps struct {
	Store     appstore.AppStore
	CookieJar *cookiejar.Jar
	Keychain  *fileKeychain
	Dir       string
}

func newStore() (*storeDeps, error) {
	dir, err := stateDir()
	if err != nil {
		return nil, fmt.Errorf("không tạo được thư mục dữ liệu: %w", err)
	}

	jar, err := cookiejar.New(&cookiejar.Options{Filename: filepath.Join(dir, "cookies.json")})
	if err != nil {
		return nil, fmt.Errorf("không khởi tạo được cookie jar: %w", err)
	}

	kc := newFileKeychain(filepath.Join(dir, "keychain.json"))
	osys := operatingsystem.New()

	store := appstore.NewAppStore(appstore.Args{
		Keychain:        kc,
		CookieJar:       jar,
		OperatingSystem: osys,
		Machine:         machine.New(machine.Args{OS: osys}),
	})

	return &storeDeps{Store: store, CookieJar: jar, Keychain: kc, Dir: dir}, nil
}
