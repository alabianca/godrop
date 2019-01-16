## Godrop

This package enables establishing peer to peer connections via mDNS or TCP punch through.


### Usage

Create a `godrop` via `godrop.NewGodrop(opt ...Option)` . Godrop follows the Variadic Configuration Function pattern to configure the instance.
`Option` is of type `func(drop *Godrop)` In this function you have the option to override default configurations. The following properties can be configured:

|Property      |Type  |Description
|--------------|:----:|:--------------------------------------------------------------------------------------------------------------------------------------------------------|
|Port          |string| The port the service with be listening on and answering mDNS queries for                                                                                |
|IP            |string| The IP of the service                                                                                                                                   |
|ServiceName   |string| The name of the service. Default _godrop._tcp.local                                                                                                     |
|Host          |string| The host of the service. Default godrop.local                                                                                                           |
|ServiceWeight |uint16| Service weight. mDNS purposes. Default 0                                                                                                                |
|TTL           |uint32| TTL of queries responses. Default 0                                                                                                                     |
|Priority      |uint16| Priority mDNS purposes. Default 0                                                                                                                       |
|RelayIP       |string| The IP of a relay server (Only required if sharing files via TCP holepunch)                                                                             |
|RelayPort     |string| The Port of the relay server (Only required if sharing files via TCP holepunch)                                                                         |
|ListenAddr    |string| The IP the service will be listening on. Same as `IP` (Only required if sharing files via TCP holepunch)                                                |
|LocalPort     |string| The Port of the service. Same as `Port` (Only required if sharing files via TCP holepunch)                                                              |
|LocalRelayIP  |string| The IP the service will be listening for P2P connections. Should always be the same as `ListenAddr` (Only required if sharing files via TCP holepunch)  |
|UID           |string| The unique user id. Other peers can find you by this id. (Only required if sharing files via TCP holepunch)                                             |

You generally only need to override the defaults when doing tcp holepunching. For example if you have a well known relay server listening at `159.89.152.222` on port `8080`
you can provide an `Option` function into `godrop.NewGodrop(opt ...Option)`

```go
func configRelayServer(drop *godrop.Godrop) {
    drop.RelayIP = "159.89.152.222"
    drop.RelayPort = "8080"
}
```

#### mDNS

To establish connections via mDNS follow this pattern:

```go
drop, err := godrop.NewGodrop()

	if err != nil {
		panic(err)
	}

	connStrategy := drop.NewP2PConn("mdns")

	p2pConn, err := connStrategy.Connect("") //pass an empty string to connect to the first peer available

	if err != nil {
		fmt.Println("Could Not establish P2P Connection")
		os.Exit(1)
    }
    
    //you now have access to the underlying tcp connection with p2pConn.Conn which is of type net.TCPConn
```


