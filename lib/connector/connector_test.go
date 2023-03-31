/*
 * Copyright (c) 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package connector

import (
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/model"
	"testing"
)

func Test_encodeLocalId(t *testing.T) {
	result := model.EncodeLocalId("aaaaaaaaa%bbbbbbb+cccccccccc#ddddddd/eeeeeeeeeeee%2Bffffffffffffff")
	expected := "aaaaaaaaa%25bbbbbbb%2Bcccccccccc%23ddddddd%2Feeeeeeeeeeee%252Bffffffffffffff"
	if result != expected {
		t.Error("\n", result, "\n", expected)
	}
}

func Test_decodeLocalId(t *testing.T) {
	result := model.DecodeLocalId("aaaaaaaaa%25bbbbbbb%2Bcccccccccc%23ddddddd%2Feeeeeeeeeeee%252Bffffffffffffff")
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
		if result := model.DecodeLocalId(model.EncodeLocalId(str)); result != str {
			t.Error("\n", result, "\n", str)
		}
	})
}
