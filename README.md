## Godrop

Godrop uses zeroconf to discover other Godrop peers in the local network

### Configure Godrop

Godrop's `NewGodrop(...Option)` function uses variadic configuration functions to configure itslef

#### Example
```go
func configGodrop(drop *Godrop) {
    drop.Port = 4000
    drop.IP = "127.0.0.1"
    Host = "godrop.local"
    UID = "My user name" //instance name
}

drop, err := NewGodrop(configGodrop)
```

### Make Godrop Discoverable in the local network

To make your godrop service discoverable, you first need to register your godrop service by using
`Godrop.NewMDNSService("someDir/someCatPicture.png")`.
This will make your service discoverable in the local network and allow connecting peers to clone `someCatPicture.png`.

#### Example
```go
server,err := drop.NewMDNSService("someDir/someCatPicture.png") 
```
This will return a server that is listening at the address/port specified when the drop instance was created. It will return an error if the `someDir/someCatPicture.png` does not exist.
At this point you can already find the service using `dns-sd`. In your terminal run:
`dns-sd -B _godrop._tcp`. You will see it show up.

### Discover Godrop services. 

Just like `dns-sd` you can discover godrop services in the local network using `Godrop.Discover`

#### Example
```go
godropServices,err := drop.Discover(15) // browses the network for 15 seconds
```

`godropServices` is a slice of `zeroconf.ServiceEntry`'s 

### Lookup specific instances

Rather than discovering all services you can also lookup an services by it's instance name.

#### Example
```go
service,err := drop.Lookup("alex") // browses the network for 15 seconds
```

### Connect to a service

You can directly connect to another godrop service if you know their instance name. Usually you will do a `Discover` call to browse for services in the network. Then do a `Godrop.Connect(instance string)` to connect to an active service. 
Connect will internall do a `Lookup` for the provided instance name.

#### Example
```go
session, err := drop.Connect("alex")
```