package qqmusic

import (
	"encoding/json"
	"strconv"
	"strings"
)

type loginResponseError struct {
	Kind                         string
	Gateway, Business, LoginType int
}

func (e *loginResponseError) Error() string { return e.message() }
func (e *loginResponseError) message() string {
	switch e.Kind {
	case "qr-signature":
		return "QQ did not return a QR signature (Q201). Refresh and try again"
	case "qr-image":
		return "The sign-in service returned an invalid QR image (Q202). Refresh and try again"
	case "qr-key":
		return "The QR code has expired (Q203). Refresh and try again"
	case "rejected":
		return "QQ Music rejected the sign-in (service code " + strconv.Itoa(e.Gateway) + "/" + strconv.Itoa(e.Business) + ")"
	case "json":
		return "QQ Music's sign-in response could not be parsed (L201)"
	case "status":
		return "QQ Music did not return a sign-in status (L202)"
	case "data":
		return "QQ Music did not return sign-in data (L203)"
	case "id":
		return "QQ Music returned an invalid account ID (L204)"
	case "key":
		return "QQ Music did not return sign-in credentials (L205)"
	case "type":
		return "QQ Music returned an unsupported sign-in type (" + strconv.Itoa(e.LoginType) + ")"
	case "expired":
		return "QQ Music returned expired credentials. Scan the code again"
	}
	return "QQ Music's sign-in response was invalid (L200)"
}

// The reference accepts numeric strings for optional expiry metadata; optional
// fields of an unavailable type do not invalidate the actual login credential.
func loginOptionalInt(raw json.RawMessage) int64 {
	var value json.Number
	if json.Unmarshal(raw, &value) == nil {
		if n, err := value.Int64(); err == nil {
			return n
		}
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		n, _ := strconv.ParseInt(text, 10, 64)
		return n
	}
	return 0
}

func decodeQQLoginResponse(body []byte) (Session, error) {
	return decodeLoginResponse(body, 2)
}

func decodeLoginResponse(body []byte, expectedType int) (Session, error) {
	var result struct {
		Code *int `json:"code"`
		Req  *struct {
			Code *int            `json:"code"`
			Data json.RawMessage `json:"data"`
		} `json:"req"`
	}
	if json.Unmarshal(body, &result) != nil {
		return Session{}, &loginResponseError{Kind: "json"}
	}
	if result.Code == nil || result.Req == nil || result.Req.Code == nil {
		return Session{}, &loginResponseError{Kind: "status"}
	}
	if *result.Code != 0 || *result.Req.Code != 0 {
		return Session{}, &loginResponseError{Kind: "rejected", Gateway: *result.Code, Business: *result.Req.Code}
	}
	var data object
	if json.Unmarshal(result.Req.Data, &data) != nil || data == nil {
		return Session{}, &loginResponseError{Kind: "data"}
	}
	id := ""
	if expectedType == 1 {
		id = text(data, "str_musicid")
	}
	if id == "" {
		id = text(data, "musicid")
	}
	if id == "" {
		id = strings.TrimSpace(string(data["musicid"]))
	}
	idNumber, idErr := strconv.ParseInt(id, 10, 64)
	if idErr != nil || idNumber <= 0 {
		return Session{}, &loginResponseError{Kind: "id"}
	}
	key := text(data, "musickey")
	if key == "" {
		return Session{}, &loginResponseError{Kind: "key"}
	}
	loginType := int(loginOptionalInt(data["loginType"]))
	if loginType == 0 && expectedType == 1 {
		loginType = 1
	}
	if loginType != expectedType {
		return Session{}, &loginResponseError{Kind: "type", LoginType: loginType}
	}
	s := Session{MusicID: id, MusicKey: key, LoginType: loginType, CreatedAt: loginOptionalInt(data["musickeyCreateTime"]), ExpiresIn: loginOptionalInt(data["keyExpiresIn"]), Nickname: text(data, "nick"), AvatarURL: text(data, "logo"), EncryptUin: text(data, "encryptUin")}
	s.OpenID, s.RefreshToken, s.UnionID, s.RefreshKey = text(data, "openid"), text(data, "refresh_token"), text(data, "unionid"), text(data, "refresh_key")
	if !s.valid() {
		return Session{}, &loginResponseError{Kind: "expired"}
	}
	return s, nil
}
