package linkedpackets

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
)

// MetadataFromVersion attempts to parse the given string into a fee version Metadata,
// an error is returned if it fails to do so.
func MetadataFromVersion(version string) (Metadata, error) {
	var metadata Metadata
	err := json.Unmarshal([]byte(version), &metadata)
	if err != nil {
		return Metadata{}, errorsmod.Wrapf(ErrInvalidVersion, "failed to unmarshal metadata from version: %s", version)
	}

	return metadata, nil
}
