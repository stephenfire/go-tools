package tools

import "testing"

func TestNodeVersion(t *testing.T) {
	tests := []struct {
		major, minor, patch uint64
		alpha               bool
		err                 bool
		v                   Version
		s                   string
	}{
		{0, 0, 0, false, false, 0, "0.0.0"},
		{0, 0, 0, true, false, 0, "0.0.0"},
		{0, 10, 0, true, false, -10_000_000, "0.10.0.a"},
		{0, 10, 999_999, false, false, 10_999_999, "0.10.999999"},
		{0, 10, 1_000_000, false, true, 0, ""},
		{1_000, 100, 36, false, false, 1_000_000_100_000_036, "1000.100.36"},
		{1_000_000, 100, 36, false, true, 0, ""},
		{3, 2_000_000, 36, false, true, 0, ""},
		{3, 100, 6000, true, false, -3_000_100_006_000, "3.100.6000.a"},
	}

	for _, test := range tests {
		ver, err := NewVersion(test.major, test.minor, test.patch, test.alpha)
		if err != nil {
			if test.err {
				t.Logf("major:%d minor:%d patch:%d error:%v check", test.major, test.minor, test.patch, err)
			} else {
				t.Fatalf("major:%d minor:%d patch:%d failed:%v", test.major, test.minor, test.patch, err)
			}
		} else {
			if test.err {
				t.Fatalf("major:%d minor:%d patch:%d should failed, but: %s", test.major, test.minor, test.patch, ver)
			} else {
				if ver != test.v {
					t.Fatalf("major:%d minor:%d patch:%d should be %d but %d", test.major, test.minor, test.patch, test.v, ver)
				} else {
					if ver.Major() == test.major && ver.Minor() == test.minor && ver.Patch() == test.patch {
						t.Logf("major:%d minor:%d patch:%d -> %s check", test.major, test.minor, test.patch, ver)

						verstr := ver.String()
						if verstr != test.s {
							t.Fatalf("want %s, but %s", test.s, verstr)
						}

						ver1, err := ParseVersion(verstr)
						if err != nil {
							t.Fatalf("parse %s failed: %v", verstr, err)
						}
						if ver1 != test.v {
							t.Fatalf("want %d, but %d", test.v, ver1)
						}

					} else {
						t.Fatalf("major:%d minor:%d patch:%d %s check versions failed", test.major, test.minor, test.patch, ver)
					}
				}
			}
		}
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		input string
		err   bool
		v     Version
	}{
		{"7.2.6", false, 7_000_002_000_006},
		{"7.2.6.a", false, -7_000_002_000_006},
		{"7x.2.6", true, 0},
		{"7009.999999.6", false, 7_009_999_999_000_006},
		{"7009.999999.6.a", false, -7_009_999_999_000_006},
		{"7.5.0", false, 7_000_005_000_000},
		{"7.0.0.a", false, -7_000_000_000_000},
		{"0.0.0", false, 0},
		{"0.0.0.a", true, 0},
	}

	for _, test := range tests {
		ver, err := ParseVersion(test.input)
		if err != nil {
			if test.err {
				t.Logf("input:%s error:%v check", test.input, err)
			} else {
				t.Fatalf("input:%s failed: %v", test.input, err)
			}
		} else {
			if test.err {
				t.Fatalf("input:%s should error, but didn't, got: %s", test.input, ver)
			} else {
				if ver != test.v {
					t.Fatalf("input:%s should be %d but got %d", test.input, test.v, ver)
				} else {
					t.Logf("input:%s is %d check", test.input, ver)
				}
			}
		}
	}
}
