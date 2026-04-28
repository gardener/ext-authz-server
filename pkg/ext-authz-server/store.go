// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package extauthzserver

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- needed to verify legacy SHA1 htpasswd credentials
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type store struct {
	secrets map[string]auth
}
type auth struct {
	username       string
	hashedPassword []byte
}

func newStore(dir fs.FS) (*store, error) {
	secrets, err := readSecrets(dir)
	if err != nil {
		return nil, fmt.Errorf("reading secrets: %w", err)
	}
	return &store{
		secrets: secrets,
	}, nil
}

// IsValid checks whether the provided authorization header is valid for the given host.
func (s *store) IsValid(host string, authorization string) error {
	username, password, ok := parseBasicAuthHeader(authorization)
	if !ok {
		return fmt.Errorf("invalid authorization header")
	}
	auth, ok := s.secrets[host]
	if !ok {
		return fmt.Errorf("no password for host %s", host)
	}
	if auth.username != username {
		return errors.New("username mismatch")
	}
	return comparePassword(auth.hashedPassword, password)
}

// comparePassword verifies the presented password against the stored hash.
// It supports bcrypt ($2a$/$2y$) and legacy SHA1 ({SHA}) htpasswd formats.
func comparePassword(storedHash, password []byte) error {
	if bytes.HasPrefix(storedHash, []byte("{SHA}")) {
		return compareSHA1(storedHash[len("{SHA}"):], password)
	}
	return bcrypt.CompareHashAndPassword(storedHash, password)
}

// compareSHA1 verifies a password against a base64-encoded SHA1 hash.
// #nosec G401 -- needed to verify legacy SHA1 htpasswd credentials generated before June 2024
func compareSHA1(storedBase64Hash, password []byte) error {
	expectedHash, err := base64.StdEncoding.DecodeString(string(storedBase64Hash))
	if err != nil {
		return fmt.Errorf("failed to decode SHA1 hash: %w", err)
	}
	actualHash := sha1.Sum(password)
	if !bytes.Equal(expectedHash, actualHash[:]) {
		return errors.New("password mismatch")
	}
	return nil
}

func readSecrets(dir fs.FS) (map[string]auth, error) {
	secrets := map[string]auth{}
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, fmt.Errorf("failed reading directory listing: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		data, err := fs.ReadFile(dir, name)
		if err != nil {
			return nil, fmt.Errorf("failed reading file %s: %w", name, err)
		}
		username, password, ok := parseBasicAuth(data)
		if !ok {
			return nil, fmt.Errorf("file %s does not contain valid basic auth", name)
		}
		secrets[name] = auth{
			username:       username,
			hashedPassword: password,
		}
	}
	return secrets, nil
}

func parseBasicAuthHeader(authHeader string) (username string, password []byte, ok bool) {
	const prefix = "Basic "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", nil, false
	}
	striped := authHeader[len(prefix):]
	c, err := base64.StdEncoding.DecodeString(striped)
	if err != nil {
		return "", nil, false
	}
	return parseBasicAuth(c)
}

func parseBasicAuth(c []byte) (string, []byte, bool) {
	username, password, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", nil, false
	}
	return username, []byte(password), true
}
