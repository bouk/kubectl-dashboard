package main

import (
	"fmt"
	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
)

type authManager struct{}

func (a *authManager) Login(s *authApi.LoginSpec) (*authApi.AuthResponse, error) {
	return nil, fmt.Errorf("nope")
}
func (a *authManager) Refresh(s string) (string, error) {
	return s, nil
}
func (a *authManager) AuthenticationModes() []authApi.AuthenticationMode {
	return nil
}
func (a *authManager) AuthenticationSkippable() bool {
	return true
}

var _ authApi.AuthManager = &authManager{}
