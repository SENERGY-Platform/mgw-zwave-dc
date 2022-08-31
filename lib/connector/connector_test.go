package connector

import "testing"

func Test_encodeLocalId(t *testing.T) {
	result := encodeLocalId("aaaaaaaaa%bbbbbbb+cccccccccc#ddddddd/eeeeeeeeeeee%2Bffffffffffffff")
	expected := "aaaaaaaaa%25bbbbbbb%2Bcccccccccc%23ddddddd%2Feeeeeeeeeeee%252Bffffffffffffff"
	if result != expected {
		t.Error("\n", result, "\n", expected)
	}
}

func Test_decodeLocalId(t *testing.T) {
	result := decodeLocalId("aaaaaaaaa%25bbbbbbb%2Bcccccccccc%23ddddddd%2Feeeeeeeeeeee%252Bffffffffffffff")
	expected := "aaaaaaaaa%bbbbbbb+cccccccccc#ddddddd/eeeeeeeeeeee%2Bffffffffffffff"
	if result != expected {
		t.Error("\n", result, "\n", expected)
	}
}

func FuzzLocalIdEncode(f *testing.F) {
	f.Add("+")
	f.Add("/")
	f.Add("%")
	f.Add("#")
	f.Add("2B")
	f.Add("%+#/%2B")
	f.Add("aaaaaaaaa%bbbbbbb+cccccccccc#ddddddd/eeeeeeeeeeee%2Bffffffffffffff")
	f.Fuzz(func(t *testing.T, str string) {
		if result := decodeLocalId(encodeLocalId(str)); result != str {
			t.Error("\n", result, "\n", str)
		}
	})
}
