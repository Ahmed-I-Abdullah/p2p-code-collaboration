# P2P Code Collaboration Application

This application is a peer-to-peer (P2P) networking implementation using Go and the libp2p framework.

## Installation

### Prerequisites

- Go (I am using version 1.20)

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
go build  -o p2p ./cmd/main.go
```

## Flags
- **rendezvous:** Specifies a unique identifier for a P2P network group. Nodes with the same rendezvous string can discover and connect to each other. (default: "meet me here")

- **peer:** Adds a peer's multiaddress to the list of bootstrap nodes. This enables a node to connect to existing peers for network bootstrapping.

- **listen:** Specifies the multiaddress for the node to listen on. Other nodes can connect to this address to establish P2P connections.

- **is_bootstrap:** Indicates whether the node serves as a bootstrap node. Bootstrap nodes facilitate the initial connection establishment for new nodes joining the network. (default: false)

- **grpcport:** Specifies the port for the gRPC server. gRPC is used for communication between nodes in the P2P network.

- **gitport:** Specifies the port for the Git daemon. This port enables Git operations over the P2P network.

- **repos_dir:** Specifies the directory from which Git repositories are served. This directory holds the shared code repositories accessible to nodes in the P2P network. (default: "./repos")

## Usage
### Running a Bootstrap Node
To start a bootstrap node, run the following command:
```bash
./p2p -listen /ip4/<your_ip>/tcp/<node_port> -rendezvous test -is_bootstrap true
```

### Connecting to the Bootstrap Node
Once the bootstrap node is running, you can connect to it using the same binary with additional flags:
```bash
./p2p -listen /ip4/<your_ip>/tcp/<node_port> -rendezvous test -grpcport <grpc_port> -gitport <git_daemon_port> -repos_dir <repos_dir_path>  -peer <bootstrap_address>
```
Replace placeholders <your_ip>, <node_port>, <grpc_port>, <git_daemon_port>, <repos_dir_path>, and <bootstrap_address> with appropriate values.

For example:
```bash
./p2p -listen /ip4/172.20.10.12/tcp/6667 -rendezvous test -grpcport 3000 -gitport 3001 -repos_dir repos1 -peer /ip4/172.20.10.12/tcp/6666/p2p/12D3KooWKV9yGUYG5KBwmj5hge332gYzKhwaJ9RjBJX2HE86zYVt
```


