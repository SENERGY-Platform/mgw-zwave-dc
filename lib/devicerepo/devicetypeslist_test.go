/*
 * Copyright (c) 2026 InfAI (CC SES)
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

package devicerepo

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/devicerepo/fallback"
	"github.com/SENERGY-Platform/models/go/models"
)

// the fallback file is read back as generic json and cast into []models.DeviceType;
// aspect lists and the deprecated single aspect id must both survive this round trip
func TestDeviceTypeFallbackKeepsAspectIds(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "fallback.json")

	deviceTypes := []models.DeviceType{{
		Id:   "dt1",
		Name: "dt1",
		Services: []models.Service{{
			Id:      "s1",
			LocalId: "s1",
			Outputs: []models.Content{{
				Id: "c1",
				ContentVariable: models.ContentVariable{
					Id:   "v1",
					Name: "value",
					Type: models.Structure,
					SubContentVariables: []models.ContentVariable{
						{Id: "v2", Name: "list", Type: models.Float, AspectIds: []string{"a1", "a2"}},
						{Id: "v3", Name: "single", Type: models.Float, AspectId: "a3"},
						{Id: "v4", Name: "alias", Type: models.Float, AspectId: "a4", AspectIds: []string{"a4"}},
					},
				},
			}},
		}},
	}}

	writer, err := fallback.NewFallback(filename)
	if err != nil {
		t.Fatal(err)
	}
	err = writer.Set(DtFallbackKey, deviceTypes)
	if err != nil {
		t.Fatal(err)
	}

	reader, err := fallback.NewFallback(filename)
	if err != nil {
		t.Fatal(err)
	}
	repo := &DeviceRepo{fallback: reader}
	actual, err := repo.getDeviceTypeListFromFallback()
	if err != nil {
		t.Fatal(err)
	}
	if len(actual) != 1 || len(actual[0].Services) != 1 || len(actual[0].Services[0].Outputs) != 1 {
		t.Fatalf("unexpected structure: %#v", actual)
	}
	sub := actual[0].Services[0].Outputs[0].ContentVariable.SubContentVariables
	if len(sub) != 3 {
		t.Fatalf("unexpected sub variables: %#v", sub)
	}

	t.Run("aspect list", func(t *testing.T) {
		if !reflect.DeepEqual(sub[0].AspectIds, []string{"a1", "a2"}) || sub[0].AspectId != "" {
			t.Errorf("%#v %#v", sub[0].AspectId, sub[0].AspectIds)
		}
	})
	t.Run("deprecated single aspect id", func(t *testing.T) {
		if sub[1].AspectId != "a3" || sub[1].AspectIds != nil {
			t.Errorf("%#v %#v", sub[1].AspectId, sub[1].AspectIds)
		}
	})
	t.Run("single aspect id as alias of one element list", func(t *testing.T) {
		if sub[2].AspectId != "a4" || !reflect.DeepEqual(sub[2].AspectIds, []string{"a4"}) {
			t.Errorf("%#v %#v", sub[2].AspectId, sub[2].AspectIds)
		}
	})
	t.Run("round trip is lossless", func(t *testing.T) {
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), `"aspect_ids"`) {
			t.Errorf("fallback file does not contain aspect_ids:\n%s", content)
		}
		if !reflect.DeepEqual(actual, deviceTypes) {
			t.Errorf("round trip changed device types:\n%s", content)
		}
	})
}
