package useragent

import "testing"

func TestUserAgent(t *testing.T) {
	ua := New("Mozilla/5.0 (Linux; Android 5.1.1; Nexus 6 Build/LYZ28E) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.83 Mobile Safari/537.36")
	if !ua.IsMobile() {
		t.Error("IsMobile() should return true")
	}

	if ua.IsTablet() {
		t.Error("IsTablet() should return false")
	}

	if ua.IsBot() {
		t.Error("IsBot() should return false")
	}

	if ua.IsWeChat() {
		t.Error("IsWechat() should return false")
	}
}
