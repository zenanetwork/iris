package checkpoint

/*
Checkpoint module is responsible for validating checkpoint in iris.

Sending checkpoint is a 2 phase process.
1. Send `MsgCheckpoint`: Here the transaction sender proposes the new checkpoint by sending the start block, end block and the roothash of the new checkpoint
						which is basically the Merkle Root of all the blocks from start to end.
2. Validate this by `handleMsgCheckpoint`: Here the transaction is validated by fetching all the headers and validating if the incoming checkpoint proposal is valid.
3. Once this `MsgCheckpoint` is deemed valid, the bridge collects all the votes and sends the checkpoint to ethereum chain smart contract.
4. As soon as this transaction on ethereum chain goes through we start with phase 2 of checkpoint submission process on iris
5. We send another transaction called `MsgCheckpointAck`: Here the transaction basically claims that the checkpoint earlier submitted has been processed on the ethereum chain
*/
