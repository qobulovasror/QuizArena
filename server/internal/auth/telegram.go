package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"
)

var ErrInvalidInitData = errors.New("Telegram initData yaroqsiz")

// TelegramUser — initData ichidagi `user` maydoni (kerakli qism).
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// ValidateInitData — Telegram Web App initData'sini bot-token bilan tekshiradi.
//
// Algoritm (Telegram rasmiy):
//
//	secret      = HMAC-SHA256(key="WebAppData", msg=botToken)
//	data_check  = "key=value" juftlari (hash'siz), kalit bo'yicha saralangan, "\n" bilan
//	computed    = hex(HMAC-SHA256(key=secret, msg=data_check))
//	computed == hash bo'lishi kerak.
func ValidateInitData(initData, botToken string) (TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, ErrInvalidInitData
	}
	hash := values.Get("hash")
	if hash == "" || botToken == "" {
		return TelegramUser{}, ErrInvalidInitData
	}

	// data_check_string: hash'dan tashqari barcha juftlar, kalit bo'yicha saralangan.
	keys := make([]string, 0, len(values))
	for k := range values {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+values.Get(k))
	}
	dataCheck := strings.Join(pairs, "\n")

	secret := hmacSum([]byte("WebAppData"), []byte(botToken))
	computed := hex.EncodeToString(hmacSum(secret, []byte(dataCheck)))
	if !hmac.Equal([]byte(computed), []byte(hash)) {
		return TelegramUser{}, ErrInvalidInitData
	}

	var u TelegramUser
	if err := json.Unmarshal([]byte(values.Get("user")), &u); err != nil || u.ID == 0 {
		return TelegramUser{}, ErrInvalidInitData
	}
	// TODO(B2+): auth_date yangiligini tekshirish (replay himoyasi).
	return u, nil
}

func hmacSum(key, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	return h.Sum(nil)
}
