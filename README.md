# registry-etcd (*This is a community driven project*)

Some application runtime use [etcd](https://github.com/etcd-io/etcd) for service discovery.

## How to use with Kitex server?

```go
import (
    ...
    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    etcd "github.com/kitex-contrib/registry-etcd"
    ...
)

func main() {
    ...
    r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}) // r should not be reused.
    if err != nil {
        log.Fatal(err)
    }
    // https://www.cloudwego.io/docs/tutorials/framework-exten/registry/#integrate-into-kitex
    server, err := echo.NewServer(new(EchoImpl), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "echo"}), server.WithRegistry(r))
    if err != nil {
        log.Fatal(err)
    }
    err = server.Run()
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```


## How to use with Kitex client?

```go
import (
    ...
    "github.com/cloudwego/kitex/client"
    etcd "github.com/kitex-contrib/registry-etcd"
    ...
)

func main() {
    ...
    r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
    if err != nil {
        log.Fatal(err)
    }
    client, err := echo.NewClient("echo", client.WithResolver(r))
    if err != nil {
        log.Fatal(err)
    }
    ...
}
```

## Authentication

### server
```go
package main

import (
    ...
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

type HelloImpl struct{}

func (h *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{
		Message: req.Message,
	}
	return
}

func main() {
	// creates a etcd based registry with given username and password
	r, err := etcd.NewEtcdRegistryWithAuth([]string{"127.0.0.1:2379"}, "username", "password")
	if err != nil {
		log.Fatal(err)
	}
	server := hello.NewServer(new(HelloImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "Hello",
	}))
	err = server.Run()
	if err != nil {
		log.Fatal(err)
	}
}

```

### client
```go
package main

import (
    ...
	"github.com/cloudwego/kitex/client"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() { 
	// creates a etcd based resolver with given username and password
	r, err := etcd.NewEtcdResolverWithAuth([]string{"127.0.0.1:2379"}, "username", "password")
	if err != nil {
		log.Fatal(err)
	}
	client := hello.MustNewClient("Hello", client.WithResolver(r))
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		resp, err := client.Echo(ctx, &api.Request{Message: "Hello"})
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(resp)
		time.Sleep(time.Second)
	}
}
```

## More info

See [example](/example).


## Compatibility

Compatible with etcd server v3.




maintained by: [lizheming](https://github.com/duduainankai)
