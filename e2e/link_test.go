package main

import (
	"context"
	"testing"

	"github.com/srdtrk/linkedpackets/interchaintest/v2/testsuite"
	"github.com/stretchr/testify/suite"
)

type LinkTestSuite struct {
	testsuite.TestSuite
}

func (s *LinkTestSuite) SetupLinkSuite(ctx context.Context) {
	s.SetupSuite(ctx, chainSpecs)
}

func (s *LinkTestSuite) TestLink() {
	ctx := context.Background()
	s.SetupLinkSuite(ctx)
}

func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(LinkTestSuite))
}
