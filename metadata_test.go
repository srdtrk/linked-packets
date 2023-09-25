package linkedpackets_test

import (
	"encoding/json"
	"testing"

	ibcmock "github.com/cosmos/ibc-go/v8/testing/mock"

	"github.com/stretchr/testify/require"

	"github.com/srdtrk/linkedpackets"
)

func TestMetadataFromVersion(t *testing.T) {
	testMetadata := linkedpackets.Metadata{
		AppVersion:           ibcmock.Version,
		LinkedPacketsVersion: linkedpackets.Version,
	}

	versionBz, err := json.Marshal(&testMetadata)
	require.NoError(t, err)

	expectedMetadata := `{"linked_packets_version":"` + linkedpackets.Version + `","app_version":"` + ibcmock.Version + `"}`
	require.Equal(t, expectedMetadata, string(versionBz))

	metadata, err := linkedpackets.MetadataFromVersion(string(versionBz))
	require.NoError(t, err)
	require.Equal(t, ibcmock.Version, metadata.AppVersion)
	require.Equal(t, linkedpackets.Version, metadata.LinkedPacketsVersion)

	metadata, err = linkedpackets.MetadataFromVersion("")
	require.Error(t, err)
	require.ErrorIs(t, err, linkedpackets.ErrInvalidVersion)
	require.Empty(t, metadata)
}
