package listener

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/libs/common"
	httpClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/zenanetwork/iris/bridge/setu/queue"
	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/helper"
)

const (
	ListenerServiceStr = "listener"

	RootChainListenerStr  = "rootchain"
	IrisListenerStr       = "iris"
	MaticChainListenerStr = "maticchain"
)

// var logger = util.Logger().With("service", ListenerServiceStr)

// ListenerService starts and stops all chain event listeners
type ListenerService struct {
	// Base service
	common.BaseService
	listeners []Listener
}

// NewListenerService returns new service object for listneing to events
func NewListenerService(cdc *codec.Codec, queueConnector *queue.QueueConnector, httpClient *httpClient.HTTP) *ListenerService {
	var logger = util.Logger().With("service", ListenerServiceStr)

	// creating listener object
	listenerService := &ListenerService{}

	listenerService.BaseService = *common.NewBaseService(logger, ListenerServiceStr, listenerService)

	rootchainListener := NewRootChainListener()
	rootchainListener.BaseListener = *NewBaseListener(cdc, queueConnector, httpClient, helper.GetMainClient(), RootChainListenerStr, rootchainListener)
	listenerService.listeners = append(listenerService.listeners, rootchainListener)

	maticchainListener := &MaticChainListener{}
	maticchainListener.BaseListener = *NewBaseListener(cdc, queueConnector, httpClient, helper.GetMaticClient(), MaticChainListenerStr, maticchainListener)
	listenerService.listeners = append(listenerService.listeners, maticchainListener)

	irisListener := &IrisListener{}
	irisListener.BaseListener = *NewBaseListener(cdc, queueConnector, httpClient, nil, IrisListenerStr, irisListener)
	listenerService.listeners = append(listenerService.listeners, irisListener)

	return listenerService
}

// OnStart starts new block subscription
func (listenerService *ListenerService) OnStart() error {
	if err := listenerService.BaseService.OnStart(); err != nil {
		listenerService.Logger.Error("OnStart | OnStart", "Error", err)
	} // Always call the overridden method.

	// start chain listeners
	for _, listener := range listenerService.listeners {
		if err := listener.Start(); err != nil {
			listenerService.Logger.Error("OnStart | Start", "Error", err)
		}
	}

	listenerService.Logger.Info("all listeners Started")

	return nil
}

// OnStop stops all necessary go routines
func (listenerService *ListenerService) OnStop() {
	listenerService.BaseService.OnStop() // Always call the overridden method.

	// start chain listeners
	for _, listener := range listenerService.listeners {
		listener.Stop()
	}

	listenerService.Logger.Info("all listeners stopped")
}
