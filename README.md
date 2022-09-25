# towers-of-pau

Proto-Danksharding is (after the merge) the next big upgrade for Ethereum. 
It allows us to provide cheaper data availability for Rollups, thus contributing directly to making L2s the web3 experience we all dream of (low transaction costs, high throughput and the same security as mainnet). 
The ceremony is secure if at least 1 participant behaves honestly, in order to make sure that this trust assumption is met, it is very important to have as many people as possible contributing to the ceremony.

## What it does
We created a server that coordinates the ceremony and clients to participate in it.
The coordinator assigns time slots to participants. When the time slot is reached, the participant asks the coordinator for the current state of the ceremony. The coordinator will send the current state. The participant then modifies the state (according to the rules of the ceremony) and sends the new state to the coordinator. The coordinator will verify the new state and - if valid - store it and write a copy to disk.
All submissions are published, s.th. participants can make sure that their submissions are included and everyone can verify after the fact that the ceremony was done correctly.

## How to run
```
cd cmd/participant
go build
./participant https://dknopik.de
```
You can see your results on https://dknopik.de
