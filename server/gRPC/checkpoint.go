package gRPC

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"

	proto "github.com/zenanetwork/zenaproto/iris"
	protoutils "github.com/zenanetwork/zenaproto/utils"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Checkpoint struct {
	Proposer    hmTypes.IrisAddress `json:"proposer"`
	StartBlock  uint64              `json:"start_block"`
	EndBlock    uint64              `json:"end_block"`
	RootHash    hmTypes.IrisHash    `json:"root_hash"`
	ZenaChainID string              `json:"zena_chain_id"`
	TimeStamp   uint64              `json:"timestamp"`
}

func (h *IrisGRPCServer) FetchCheckpointCount(_ context.Context, _ *emptypb.Empty) (*proto.FetchCheckpointCountResponse, error) {
	cliCtx := cliContext.NewCLIContext().WithCodec(h.cdc)

	result, err := helper.FetchFromAPI(cliCtx, helper.GetIrisServerEndpoint(fetchCheckpointCount))
	if err != nil {
		logger.Error("Error while fetching checkpoint count")
		return nil, err
	}

	resp := &proto.FetchCheckpointCountResponse{}
	resp.Height = fmt.Sprint(result.Height)

	if err := json.Unmarshal(result.Result, &resp.Result); err != nil {
		logger.Error("Error unmarshalling checkpoint count", "error", err)
		return nil, err
	}

	return resp, nil
}

func (h *IrisGRPCServer) FetchCheckpoint(_ context.Context, in *proto.FetchCheckpointRequest) (*proto.FetchCheckpointResponse, error) {
	cliCtx := cliContext.NewCLIContext().WithCodec(h.cdc)

	url := ""
	if in.ID == -1 {
		url = fmt.Sprintf(fetchCheckpoint, "latest")
	} else {
		url = fmt.Sprintf(fetchCheckpoint, fmt.Sprint(in.ID))
	}

	result, err := helper.FetchFromAPI(cliCtx, helper.GetIrisServerEndpoint(url))

	if err != nil {
		logger.Error("Error while fetching checkpoint")
		return nil, err
	}

	checkPoint := &Checkpoint{}
	if err := json.Unmarshal(result.Result, &checkPoint); err != nil {
		logger.Error("Error unmarshalling checkpoint", "error", err)
		return nil, err
	}

	var hash [32]byte

	copy(hash[:], checkPoint.RootHash.Bytes())

	var address [20]byte

	copy(address[:], checkPoint.Proposer.Bytes())

	if checkPoint.TimeStamp > math.MaxInt64 {
		return nil, fmt.Errorf("timestamp value out of range for int64: %d", checkPoint.TimeStamp)
	}

	resp := &proto.FetchCheckpointResponse{}

	resp.Height = fmt.Sprint(result.Height)
	resp.Result = &proto.Checkpoint{
		StartBlock:  checkPoint.StartBlock,
		EndBlock:    checkPoint.EndBlock,
		RootHash:    protoutils.ConvertHashToH256(hash),
		Proposer:    protoutils.ConvertAddressToH160(address),
		Timestamp:   timestamppb.New(time.Unix(big.NewInt(0).SetUint64(checkPoint.TimeStamp).Int64(), 0)),
		ZenaChainID: checkPoint.ZenaChainID,
	}

	return resp, nil
}
