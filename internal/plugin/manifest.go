package plugin

import "encoding/json"

// unmarshalManifest accepts the community's original metadata block when a
// package does not declare a blackpie block of its own.
func unmarshalManifest(b []byte, dst any) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	if _, ok := fields["blackpie"]; !ok {
		if legacy, ok := fields["magpie"]; ok {
			fields["blackpie"] = legacy
			var err error
			b, err = json.Marshal(fields)
			if err != nil {
				return err
			}
		}
	}
	return json.Unmarshal(b, dst)
}
