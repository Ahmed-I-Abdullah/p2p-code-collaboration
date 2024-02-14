# P2P Code Collaboration Application

This application is a peer-to-peer (P2P) networking implementation using Go and the libp2p framework.

## Installation

### Prerequisites

- Go (I am using version 1.19)

### Clone the Repository
```bash
git clone git@github.com:Ahmed-I-Abdullah/p2p-code-collaboration.git
cd p2p-code-collaboration
```

## Build the Project

### Install Dependencies

You can install the project's dependecies using the following command:

```bash
go mod tidy
```

### Build the Project
You can build the project using the following command:
```bash
go build -o p2p
```

## Usage
### Running a Bootstrap Node
To start a bootstrap node, run the following command:
```bash
./p2p -listen /ip4/<your_ip>/tcp/<node_port> -rendezvous test -is_bootstrap true
```

### Connecting to the Bootstrap Node
Once the bootstrap node is running, you can connect to it using the same binary:
```bash
./p2p -listen /ip4/<your_ip>/tcp/<node_port> -rendezvous test -peer <bootstrap_address>
```

For example:
```bash
./p2p -listen /ip4/<your_ip>/tcp/<node_port> -rendezvous test -peer <bootstrap_address>
```


