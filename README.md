# P2P Storage and Computing For Organizations
This project aims to provide a foundation for decentralized organization governence system.

## Usage
To test:
    go test ./... -covermode=count

To build:
    go build -o themachine cmd/themachine/main.go

To run:
    ./themachine




## Notes:
- Payment (or burning) should be required for Genesis transactions to avoid spam.
- Payment should be required for Computation requests.
- A node would have its own keys plus keys provided by organizations of which it is a member
- With a Genesis transaction an organization is created.
- Organization assigns keys to already existing nodes.
- Transactions requires "targets" (addressed keys) to sign it as a prerequisite before saving.
- Transactions are of many types, which basically hold data like a file system. 


## TODOs:
Current state is just a draft. Many TODO candidates include:
- Complete syncronization process
- Switching to libp2p or devp2p for networking
- Switching to BLS for signatures to keep multi-signed transactions small in size
- Key distribution infrastructure for keys derived by organizations
- Consensus mechanism for determining valid transactions
- Forming optimal genesis and rule structures
- Introduce DHT and Sharding
- Computation infrastructure (sandbox, language, remote calls)
- Complete Web UIs
- Payment token for computation. Consider ERC-20