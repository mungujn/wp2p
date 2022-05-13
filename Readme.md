# Distributed Local Web Server

A p2p web server application. This serves web pages from a local directory and allows other peers connected to the same p2p network to be able to view them.

Consider two users, `A` and `B`. `A` starts up an instance of the program setting, at minimum, a `username` for themselves and the directory on their machine that they want to share. User `B` does the same.

Now, on `A`'s machine, accessing `http://localhost:{port}/` will display a list of usernames currently connected to the same p2p network. Clicking on any of the usernames will render `http://localhost:{port}/{username}/index.html` on `A`'s machine.

## Environment Configuration

| Name                  | Required | Default     | Description                                                                                                    |
| --------------------- | -------- | ----------- | -------------------------------------------------------------------------------------------------------------- |
| USERNAME              | Yes      | me          | The username that will be used to represent this node on the network                                           |
| LOCAL_ROOT_FOLDER     | Yes      | test_folder | The local folder to serve web pages from                                                                       |
| LOCAL_WEB_SERVER_PORT | Yes      | 8080        | The port to use for the local web server                                                                       |
| LOCAL_NODE_PORT       | Yes      | 4040        | The port to use for the p2p subsytem                                                                           |
| LOCAL_NODE_HOST       | No       | 0.0.0.0     | The host address that the p2p subsytem will bind to                                                            |
| NETWORK_NAME          | No       | local       | A unique string that identifies the p2p network you want to connect to                                         |
| PROTOCOL_ID           | No       | localfiles  | A unique string that identifies the p2p network version                                                        |
| PROTOCOL_VERSION      | No       | 0.1         | A further refining of the current p2p networks version                                                         |
| RUN_GLOBAL            | No       | false       | Whether or not to use the bridge peer                                                                           |
| CUSTOM_BOOTSTRAP_PEER | No       |             | This can be a peer with a public IP address that can be used to expand the network by serving as a bridge peer |
| DEBUG                 | No       | false       | If set to true, this generates a fixed host ID for a given node port                                           |
| LOG_LEVEL             | No       | DEBUG       | Application log level, possible values: ERROR, INFO, TRACE                                                     |

## The System

This system has 3 main subsytems:

- Web server
- Application
- Local files
- p2p network

### Web Server

This serves up web pages. It is configured to run on `localhost` on the port specified in the environment configuration. It can serve web pages from the configured `LOCAL_ROOT_FOLDER` and from other peers on the p2p network. 

The web server defines a `System` interface that specifies two main methods; one for retrieving a file and one for retrieving a list of online users. 

This subsytem is implemented in the `server` module.

### Application

This implements the systems main logic as defined by the `System` interface. The web server sends over the full web request `GET` path to the file subsystem. The main application logic here is seperation of the username from the actual file path being requested. The application then makes a request for the specified file (from the appropriate user) and returns the file contents. 

The application defines a `FileProvider` interface that specifies three main methods; one for setting up the provider, one for retrieving a file from a specific user and one for retrieving a list of online users. 

This subsytem is implemented in the `app` module.

### Local Files

This is a local files implementation of the `FileProvider` interface specified by the application subsystem. It is primarily used in the unit tests for the `app` module to test application logic without needing a full p2p network. 

It is implemented in the `local` module.

### p2p network

This is arguably the heart of the application. It is the primary p2p network implementation, implemented in the `remote` module. It is used to connect to other peers on the network and retrieve files from these peers. It's also the primary utilizer of most of the environment variables specified above. 

The module implements the `FileProvider` interface defined by the application subsytem (located in the `app` module) and provides p2p-networking integrated implementations of the 3 methods that the interface specifies, `StartHost`, `GetFile` and `GetOnlineNodes`.

- `StartHost` creates a `libp2p` node/host and sets up a handler for incoming connections. It also initializes `mDNS` which is used to discover other peers on the network.

- `GetFile` is used to retrieve files from other peers on the network.

- `GetOnlineNodes` returns a list of online users in the network. Nodes in the network are identified using ID's that are not exactly human readable. These ID's are mapped to usernames in a custom handshake protocol that is implemented in this `remote` module.

This modules tests are at a relatively high level. Two full nodes are spun up and file exchange between them is tested.

## Running

- Run `make deps` to install the required dependencies
- Run `make compile` to compile the app for your environment
- In one shell instance, configure at least `USERNAME`, `LOCAL_ROOT_FOLDER`, `LOCAL_WEB_SERVER_PORT` and `LOCAL_NODE_PORT` and then start up the application using `make run`
- In a second shell instance, configure the four variables above as well, pointing them to different values and then start up the second app instance

## Notes

- The bridge capability is as yet untested. For now, best to leave `RUN_GLOBAL` set to false.

## Credit

The p2p implementation was built with a lot of guidance from the excellent [examples](https://github.com/libp2p/go-libp2p/tree/master/examples) in the go-libp2p library
