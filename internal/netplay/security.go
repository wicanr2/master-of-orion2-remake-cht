package netplay

// security.go：網路大廳的可選傳輸保護。
//
// 原版只假設在 IPX／數據機網路上，沒有可直接搬用的現代身份層。remake 的
// 公開網路模式因此採「每次連線一次性 challenge + HMAC proof」，並可再套
// TLS 1.3。TLS 憑證預設是記憶體內自簽，真正的對局身份由共享密碼驗證；若
// 未設定密碼，TLS 仍提供加密，但不宣稱能辨認真正的主機。

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// ProtocolVersion 是目前 remake 網路協定版本。它與原版 IPX 封包不是同一格式。
const ProtocolVersion = 2

// LobbyOptions 是主機與客戶端共用的傳輸保護設定。
type LobbyOptions struct {
	// AuthToken 為空時不啟用身份驗證；非空時只接受相同密碼的 challenge proof。
	AuthToken string
	// EnableTLS 開啟 TLS 1.3。未提供 TLSConfig 時由主機產生短期記憶體憑證。
	EnableTLS bool
	// TLSConfig 可供整合測試或正式部署注入憑證／信任根；不會寫入磁碟。
	TLSConfig *tls.Config
}

// JoinOptions 是客戶端加入或重連時的設定。
type JoinOptions struct {
	LobbyOptions
	// ResumeToken 由主機在第一次加入後發給該玩家，之後只用它恢復原玩家編號。
	ResumeToken string
}

const authLabel = "moo2-remake-netplay-v2"

func authProof(secret, challenge string) string {
	if secret == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(authLabel))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(challenge))
	return hex.EncodeToString(h.Sum(nil))
}

func verifyAuth(secret, challenge, proof string) bool {
	want := authProof(secret, challenge)
	if want == "" {
		return proof == ""
	}
	got, err := hex.DecodeString(proof)
	if err != nil {
		return false
	}
	wantBytes, err := hex.DecodeString(want)
	if err != nil {
		return false
	}
	return hmac.Equal(got, wantBytes)
}

func randomHex(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("netplay: 隨機值長度必須為正")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("netplay: 產生隨機值：%w", err)
	}
	return hex.EncodeToString(b), nil
}

func newChallenge() (string, error) { return randomHex(24) }

func newResumeToken() (string, error) { return randomHex(32) }

func serverTLSConfig(opts LobbyOptions) (*tls.Config, error) {
	cfg := &tls.Config{MinVersion: tls.VersionTLS13}
	if opts.TLSConfig != nil {
		cfg = opts.TLSConfig.Clone()
		if cfg.MinVersion == 0 {
			cfg.MinVersion = tls.VersionTLS13
		}
	}
	if len(cfg.Certificates) == 0 {
		cert, err := ephemeralCertificate()
		if err != nil {
			return nil, err
		}
		cfg.Certificates = []tls.Certificate{cert}
	}
	return cfg, nil
}

func clientTLSConfig(opts LobbyOptions) *tls.Config {
	cfg := &tls.Config{MinVersion: tls.VersionTLS13}
	if opts.TLSConfig != nil {
		cfg = opts.TLSConfig.Clone()
		if cfg.MinVersion == 0 {
			cfg.MinVersion = tls.VersionTLS13
		}
	}
	// 預設主機憑證是每次啟動新產生的自簽憑證；沒有 RootCAs 時只能以
	// challenge/HMAC 驗證身份。使用者若提供 RootCAs，則保留正式憑證驗證。
	if cfg.RootCAs == nil && !cfg.InsecureSkipVerify {
		cfg.InsecureSkipVerify = true // #nosec G402 -- 預設憑證是記憶體內自簽，身份靠 HMAC
	}
	return cfg
}

func ephemeralCertificate() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("netplay: 產生 TLS 金鑰：%w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("netplay: 產生 TLS 序號：%w", err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "moo2-remake-session"},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"moo2-remake-session"},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("netplay: 建立 TLS 憑證：%w", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("netplay: 編碼 TLS 金鑰：%w", err)
	}
	return tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
	)
}
