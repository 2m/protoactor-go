package shared

import "github.com/AsynkronIT/gam/cluster/grain"

//a Go struct implementing the Hello interface
type hello struct {
	grain.Grain
}

func (*hello) SayHello(r *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: "hello " + r.Name}, nil
}

func (*hello) Add(r *AddRequest) (*AddResponse, error) {
	return &AddResponse{Result: r.A + r.B}, nil
}

func init() {
	//apply DI and setup logic
	HelloFactory(func() Hello { return &hello{} })
}
