# Flash
A lightning-fast, simple, green, feeless, and secure cryptocurrency designed to meet real-world needs with next-generation technology.

## Consensus 
Each peer has a private key with zero or more coins associated with it. In order to send a transaction, a peer must get a verification signature from peers whose total coin value is greater than 50% of the total coins in existence. This is no mining in Flash so all coins are allocated to their initial keys in the genesis file (genesis.yaml).

**Example**

There are only three peers in the network: Alice, Bob and Eve. Alice wants to send Bob 20 coins. Each peer has 1000 coins. That means there were 3000 total coins minted in the genesis file. 

Alice needs signatures from peers that manage over 1500 coins in total. She sends her transaction to Bob and Eve who reply with a signature. She now has signatures worth 2000 coins: 2/3 of the total coins and enough to make her transaction valid. She sends her transaction along with the signatures to all the peers in the network and asks them to commit them to their database. Bob and Alices balances will then be updated and the transaction will be complete. 

## How to setup a network
First build the project
```
go mod tidy
go build
```

Then generate three private keys

```
mkdir key
./flash gen ./key/alice && ./flash gen ./key/bob &&./flash gen ./key/eve
```

Next run a node for alice

```
./flash start ./key/alice -p 2000
```

Go to the *My Node* page then press *c* to copy the Multiaddress.

Create a file called *boostrap.txt* and pate in the address.

The part of the Muliaddress after the last slash is the PeerID. We need to copy that and put it in a file called genesis.yaml which will then look something like this. The 1000 is the starting number of coins.
```
QmUHRt5oVsRvzUSKDCdqQ7vjKEpgVTGhhbAwn8FAtW1Yu5: 1000
```
Repeat this for bob and eve putting their Multiaddress and PeerID on new lines in the corresponding file. You need to specify a different port for each node which is done via the *-p* flag. Your config files should look something like this

genesis.yaml
```
QmUHRt5oVsRvzUSKDCdqQ7vjKEpgVTGhhbAwn8FAtW1Yu5: 1000
QmcKY3aNFhLjksuMT65rDuq3C3JZQcoEaB8JXPiJR5sAkP: 1000
QmaM4yng1KjjbRsadFkyaZRYX4FscTj61A98rThyLsepFi: 1000
```

bootstrp.txt
```
/ip4/127.0.0.1/tcp/2000/p2p/QmUHRt5oVsRvzUSKDCdqQ7vjKEpgVTGhhbAwn8FAtW1Yu5
/ip4/127.0.0.1/tcp/2001/p2p/QmcKY3aNFhLjksuMT65rDuq3C3JZQcoEaB8JXPiJR5sAkP
/ip4/127.0.0.1/tcp/2002/p2p/QmaM4yng1KjjbRsadFkyaZRYX4FscTj61A98rThyLsepFi
```

Next start three terminals and start one node in each
```
./flash start ./keys/alice -p 2000
```
```
./flash start ./keys/bob -p 2001
```
```
./flash start ./keys/eve -p 2002
```
Now go to Alice’s console window, select Transfer, paste in Bob’s peer ID, enter an amount then hit enter to execute the transaction. The transaction should complete in milliseconds using a lightning bug’s sneeze worth of electricity⚡
