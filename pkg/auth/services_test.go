/*
 * SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"testing"
)

const service Services = `^outbound\|1194\|\|vpn-seed-server(-[0-4])?\..*\.svc\.cluster\.local$`

func TestCheck(t *testing.T) {
	type testdata struct {
		svc           string
		expectedMatch bool
	}
	tests := []testdata{
		{"outbound|1194||vpn-seed-server.foo.svc.cluster.local", true},
		{"outbound|1194||vpn-seed-server-0.foo.svc.cluster.local", true},
		{"outbound|1194||vpn-seed-server-2.foo.svc.cluster.local", true},
		{"outbound|1194||vpn-seed-server-11.foo.svc.cluster.local", false},
		{"outbound|1194||vpn-seed-server.foo.svc.cluster.local.bar", false},
	}

	for _, test := range tests {
		match, err := service.Check(test.svc)
		if err != nil {
			t.Errorf("%s failed: %s", test.svc, err)
			continue
		}
		if match != test.expectedMatch {
			t.Errorf("%s match mismatch: expected %t", test.svc, test.expectedMatch)
		}
	}
}
