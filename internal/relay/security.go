package relay

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

func randomSecret(size int) string {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
func digest(value string) []byte { sum := sha256.Sum256([]byte(value)); return sum[:] }
func seal(key []byte, value, aad string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(value), []byte(aad)), nil
}
func unseal(key, data []byte, aad string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted material")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], []byte(aad))
	return string(plain), err
}

type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}
type Policy struct {
	TestAuthority      string
	Resolver           Resolver
	PrivateAuthorities []string
}

func (policy Policy) privateAllowed(authority string) bool {
	for _, allowed := range policy.PrivateAuthorities {
		if allowed == authority {
			return true
		}
	}
	return false
}

var errDestinationForbidden = errors.New("destination forbidden")

var blocked = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("::/96"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("100::/64"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}

func publicIP(address netip.Addr) bool {
	address = address.Unmap()
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() {
		return false
	}
	for _, prefix := range blocked {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}
func (policy Policy) resolve(ctx context.Context, host string) ([]netip.Addr, error) {
	if address, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{address}, nil
	}
	resolver := policy.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	return resolver.LookupNetIP(ctx, "ip", host)
}
func (policy Policy) validateIPs(ctx context.Context, host, authority string) (netip.Addr, error) {
	addresses, err := policy.resolve(ctx, host)
	if err != nil || len(addresses) == 0 {
		return netip.Addr{}, errors.New("destination unavailable")
	}
	for _, address := range addresses {
		if authority == policy.TestAuthority && address.IsLoopback() {
			continue
		}
		if policy.privateAllowed(authority) && address.IsPrivate() && !address.IsLoopback() && !address.IsLinkLocalUnicast() {
			continue
		}
		if !publicIP(address) {
			return netip.Addr{}, errDestinationForbidden
		}
	}
	return addresses[0], nil
}
func (policy Policy) URL(ctx context.Context, raw string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || strings.ContainsAny(raw, "\\\r\n") || parsed.Opaque != "" {
		return "", errors.New("invalid destination")
	}
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && (parsed.Host == policy.TestAuthority || policy.privateAllowed(parsed.Host))) {
		return "", errors.New("HTTPS required")
	}
	if parsed.RawPath != "" || strings.ContainsAny(parsed.Path, "\r\n\x00") || strings.Contains(parsed.Path, "..") || strings.Contains(parsed.Path, "//") {
		return "", errors.New("invalid path")
	}
	if port := parsed.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return "", errors.New("invalid port")
		}
	}
	parsed.Path = strings.TrimSuffix(path.Clean("/"+strings.TrimPrefix(parsed.Path, "/")), "/")
	if _, err := policy.validateIPs(ctx, parsed.Hostname(), parsed.Host); err != nil {
		return "", err
	}
	return parsed.String(), nil
}
func (policy Policy) Dial(ctx context.Context, network, authority string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	host, port, err := net.SplitHostPort(authority)
	if err != nil {
		return nil, err
	}
	testAuthority := authority
	if strings.HasSuffix(authority, ":443") && !policy.privateAllowed(authority) {
		testAuthority = host
	}
	address, err := policy.validateIPs(ctx, host, testAuthority)
	if err != nil {
		return nil, err
	}
	return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(address.String(), port))
}
