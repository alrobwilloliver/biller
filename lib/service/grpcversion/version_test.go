package grpcversion

import (
	"context"
	"testing"

	"biller/lib/version"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func Test_Service_GetVersion(t *testing.T) {
	version.Version = "123"
	version.BuildTime = "Prehistoric Era"

	res, err := service{}.GetVersion(context.Background(), &GetVersionRequest{})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(res, &GetVersionResponse{BuildTime: "Prehistoric Era", Version: "123"}))
}
