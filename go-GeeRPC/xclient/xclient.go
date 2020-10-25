package xclient

import (
	"context"
	"go-GeeRPC/client"
	"go-GeeRPC/server"
	"io"
	"reflect"
	"sync"
)

type Xclient struct {
	d       Discovery
	mode    SelectMode
	opt     *server.Option
	mu      sync.Mutex
	clients map[string]*client.Client
}

func NewXClient(d Discovery, mode SelectMode, opt *server.Option) *Xclient {
	return &Xclient{d: d, mode: mode, opt: opt, clients: make(map[string]*client.Client)}
}
func (x Xclient) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	for key, client := range x.clients {
		client.Close()
		delete(x.clients, key)
	}
	return nil

}

var _ io.Closer = (*Xclient)(nil)

func (xc *Xclient) dial(rpcAddr string) (*client.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	c, ok := xc.clients[rpcAddr]
	if ok && !c.IsAvailable() {
		_ = c.Close()
		delete(xc.clients, rpcAddr)
		c = nil
	}
	if c == nil {
		var err error
		c, err = client.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = c
	}
	return c, nil
}

func (xc *Xclient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

func (xc *Xclient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

func (xc *Xclient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil

	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var cloneReply interface{}
			if reply != nil {
				cloneReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}

			err := xc.call(rpcAddr, ctx, serviceMethod, args, cloneReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(cloneReply).Elem())
				replyDone = true
			}
			mu.Unlock()

		}()
	}
	wg.Wait()
	return e

}
