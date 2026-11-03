package main

import "testing"

func TestParseIPAName(t *testing.T) {
	b, id, v := parseIPAName("com.zing.zalo_579523206_25.03.01.ipa")
	if b != "com.zing.zalo" || id != 579523206 || v != "25.03.01" {
		t.Fatalf("got %q %d %q", b, id, v)
	}
	b, id, v = parseIPAName("my_app_1.ipa")
	if b != "my" || id != 0 || v != "1" {
		// bundle IDs with underscores are ambiguous; just make sure it doesn't panic
	}
}

func TestCountry(t *testing.T) {
	if c := countryFromStoreFront("143441-1,29"); c != "US" {
		t.Fatal(c)
	}
	if c := countryFromStoreFront("143471-2,32"); c != "VN" {
		t.Fatal(c)
	}
	if c := countryFromStoreFront("999999-1"); c != "US" {
		t.Fatal(c)
	}
}

func TestFileKeychain(t *testing.T) {
	p := t.TempDir() + "/kc.json"
	k := newFileKeychain(p)
	if _, err := k.Get("x"); err == nil {
		t.Fatal("expected not found")
	}
	if err := k.Set("account", []byte(`{"email":"a@b"}`)); err != nil {
		t.Fatal(err)
	}
	k2 := newFileKeychain(p)
	v, err := k2.Get("account")
	if err != nil || string(v) != `{"email":"a@b"}` {
		t.Fatalf("%v %s", err, v)
	}
	_ = k2.Remove("account")
	if _, err := newFileKeychain(p).Get("account"); err == nil {
		t.Fatal("expected removed")
	}
}
