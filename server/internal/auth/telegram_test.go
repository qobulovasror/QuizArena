package auth

import (
	"encoding/hex"
	"net/url"
	"testing"
)

// signInitData — test uchun yaroqli initData yasaydi (Telegram algoritmi).
func signInitData(botToken, userJSON, authDate string) string {
	dataCheck := "auth_date=" + authDate + "\nuser=" + userJSON
	secret := hmacSum([]byte("WebAppData"), []byte(botToken))
	hash := hex.EncodeToString(hmacSum(secret, []byte(dataCheck)))
	v := url.Values{}
	v.Set("auth_date", authDate)
	v.Set("user", userJSON)
	v.Set("hash", hash)
	return v.Encode()
}

func TestValidateInitData(t *testing.T) {
	const botToken = "123456:TESTTOKEN"
	userJSON := `{"id":42,"first_name":"Ali","username":"ali"}`
	initData := signInitData(botToken, userJSON, "1700000000")

	u, err := ValidateInitData(initData, botToken)
	if err != nil {
		t.Fatalf("yaroqli initData o'tishi kerak: %v", err)
	}
	if u.ID != 42 || u.FirstName != "Ali" || u.Username != "ali" {
		t.Fatalf("user noto'g'ri: %+v", u)
	}
}

func TestValidateInitDataRejects(t *testing.T) {
	const botToken = "123456:TESTTOKEN"
	userJSON := `{"id":42,"first_name":"Ali"}`
	initData := signInitData(botToken, userJSON, "1700000000")

	// Boshqa token bilan — rad etilishi kerak.
	if _, err := ValidateInitData(initData, "boshqa:token"); err == nil {
		t.Fatal("noto'g'ri token bilan rad etilishi kerak")
	}
	// Buzilgan hash.
	if _, err := ValidateInitData(initData+"x", botToken); err == nil {
		t.Fatal("buzilgan initData rad etilishi kerak")
	}
	// hash umuman yo'q.
	if _, err := ValidateInitData("user="+url.QueryEscape(userJSON), botToken); err == nil {
		t.Fatal("hash'siz rad etilishi kerak")
	}
}
