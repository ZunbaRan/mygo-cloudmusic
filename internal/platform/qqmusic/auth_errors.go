package qqmusic

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
)

// LoginError retains the cause for classification without putting temporary
// authorization URLs or credentials in the error's printable representation.
type LoginError struct {
	Step string
	Err  error
}

type loginHTTPError int

func (e loginHTTPError) Error() string { return fmt.Sprintf("HTTP %d", int(e)) }

func (e *LoginError) Error() string { return LoginErrorMessage(e) }
func (e *LoginError) Unwrap() error { return e.Err }

func loginFailure(step string, err error) error {
	if err == nil {
		return nil
	}
	var failure *LoginError
	if errors.As(err, &failure) {
		return err
	}
	return &LoginError{Step: step, Err: err}
}

func loginStep(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "login"
	}
	switch u.Path {
	case "/ssl/ptqrshow":
		return "qr-show"
	case "/tme/trpc/proxy":
		return "device-bootstrap"
	case "/ssl/ptqrlogin":
		return "qr-poll"
	case "/check_sig":
		return "check-sig"
	case "/oauth2.0/login_jump":
		return "login-jump"
	case "/oauth2.0/authorize":
		return "authorize"
	case "/cgi-bin/musicu.fcg":
		return "qq-login"
	}
	return "login"
}

// LoginErrorMessage is safe to show in the fixed-width native login dialog.
func LoginErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var response *loginResponseError
	if errors.As(err, &response) {
		return response.message()
	}
	stage := "Sign In"
	var failure *LoginError
	if errors.As(err, &failure) {
		switch failure.Step {
		case "qr-show":
			stage = "QR code"
		case "device-bootstrap":
			stage = "Sign-in device setup"
		case "qr-poll":
			stage = "QR status check"
		case "qr-image":
			stage = "QR image decoding"
		case "check-sig":
			stage = "QQ authorization check"
		case "login-jump":
			stage = "QQ authorization redirect"
		case "authorize":
			stage = "QQ authorization"
		case "qq-login":
			stage = "QQ Music sign-in"
		}
	}
	var status loginHTTPError
	if errors.As(err, &status) {
		return fmt.Sprintf("The %s service returned HTTP %d. Refresh the QR code and try again", stage, int(status))
	}
	var dns *net.DNSError
	if errors.As(err, &dns) {
		return stage + " service is unreachable. Check your network and refresh the QR code"
	}
	var certificate x509.UnknownAuthorityError
	var hostname x509.HostnameError
	if errors.As(err, &certificate) || errors.As(err, &hostname) {
		return stage + " certificate check failed. Check your network and try again"
	}
	var network net.Error
	if errors.As(err, &network) {
		if network.Timeout() {
			return stage + " timed out. Refresh the QR code and try again"
		}
		if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			return stage + " connection failed (N201). Check your network and try again"
		}
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return stage + " connection was interrupted. Refresh the QR code and try again"
	}
	if failure != nil && failure.Step == "qq-login" {
		return "QQ Music sign-in response was abnormal (L200)"
	}
	return stage + " failed. Refresh the QR code and try again"
}
