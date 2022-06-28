package zoox

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-zoox/crypto/aes"
	"github.com/go-zoox/crypto/hmac"
)

var sessionKey = "gsession"
var sessionSignature = "gsession.sig"

// Session is the http session based on cookie.
type Session struct {
	ctx       *Context
	data      map[string]string
	crypto    *aes.CFB
	secretKey []byte
	isParsed  bool
}

func newSession(ctx *Context) *Session {
	crypto, err := aes.NewCFB(256, &aes.HexEncoding{}, nil)
	if err != nil {
		panic(err)
	}

	secretKey := []byte("go-zoox")
	if ctx.App.SecretKey != "" && len(ctx.App.SecretKey) < 32 {
		rest := 32 - len(secretKey)
		secretKey = []byte(ctx.App.SecretKey + strings.Repeat("0", rest))
	} else {
		secretKey = []byte(ctx.App.SecretKey[:32])
	}

	return &Session{
		ctx:       ctx,
		secretKey: secretKey,
		crypto:    crypto,
		data: map[string]string{
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		},
	}
}

func (s *Session) parse() {
	if s.isParsed {
		return
	}

	s.isParsed = true

	sessionEncrypted := s.ctx.Cookie.Get(sessionKey)
	sessionSignature := s.ctx.Cookie.Get(sessionSignature)

	if signatureX := hmac.Sha256(string(s.secretKey), sessionEncrypted); signatureX != sessionSignature {
		return
	}

	// sessionEncrypted, err := base64.RawStdEncoding.DecodeString(sessionRaw)
	// if err != nil {
	// 	s.data = make(map[string]string)
	// 	return
	// }

	session, err := s.crypto.Decrypt([]byte(sessionEncrypted), s.secretKey)
	if err != nil {
		return
	}

	if session == nil {
		return
	}

	var data map[string]string
	if err := json.Unmarshal(session, &data); err != nil {
		return
	}

	for key, value := range data {
		s.data[key] = value
	}
}

func (s *Session) flush() {
	if s.data == nil {
		return
	}

	d, err := json.Marshal(s.data)
	if err != nil {
		return
	}

	dEncrypted, err := s.crypto.Encrypt(d, s.secretKey)
	if err != nil {
		return
	}

	// dRaw := base64.RawStdEncoding.EncodeToString(dEncrypted)
	data := string(dEncrypted)
	signature := hmac.Sha256(string(s.secretKey), data)
	s.ctx.Cookie.Set(sessionKey, data, 7*24*time.Hour)
	s.ctx.Cookie.Set(sessionSignature, signature, 7*24*time.Hour)
}

// Get gets the value by key.
func (s *Session) Get(key string) string {
	s.parse()

	if value, ok := s.data[key]; ok {
		return value
	}

	return ""
}

// Set sets the value by key.
func (s *Session) Set(key string, value string) {
	s.parse()

	s.data[key] = value
	s.data["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	s.flush()
}

// Del deletes the value by key.
func (s *Session) Del(key string) {
	s.parse()

	delete(s.data, key)

	s.flush()
}
