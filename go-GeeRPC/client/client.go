package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-GeeRPC/encode"
	"go-GeeRPC/server"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

//结构体 Call 来承载一次 RPC 调用所需要的信息
type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call //Done 的类型是 chan *Call，当调用结束时，会调用 call.done() 通知调用方
}

func (call *Call) done() {
	call.Done <- call
}

// 结构体CLient
type Client struct {
	cc       encode.Codec //消息的编解码器
	opt      *server.Option
	sending  sync.Mutex //保证请求的有序发送，即防止出现多个请求报文混淆
	header   encode.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call //存储未处理完的请求，键是编号，值是 Call 实例
	closing  bool             // 用户主动停止
	shutdown bool             // 服务器通知停止
}

var ErrShutdown = errors.New("connection is shut down")

func (c Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close()
}

func (receiver *Client) IsAvailable() bool {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	return !receiver.shutdown && !receiver.closing
}

var _ io.Closer = (*Client)(nil)

//将参数 call 添加到 client.pending 中，并更新 client.seq。
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

//根据 seq，从 client.pending 中移除对应的 call
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

//服务端或客户端发生错误时调用
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

// 接受响应
//call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了。
//call 存在，但服务端处理出错，即 h.Error 不为空。
//call 存在，服务端处理正常，那么需要从 body 中读取 Reply 的值
func (client *Client) receive() {
	var err error
	for err == nil {
		var h encode.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			err = client.cc.ReadBody(nil)
		case h.Err != "":
			call.Error = fmt.Errorf(h.Err)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body" + err.Error())
			}
			call.done()
		}
	}
	client.terminateCalls(err)

}
func NewClient(conn net.Conn, opt *server.Option) (*Client, error) {
	f := encode.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc encode.Codec, opt *server.Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}
func parseOptions(opts ...*server.Option) (*server.Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return server.DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = server.DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = server.DefaultOption.CodecType
	}
	return opt, nil
}

func Dial(network, address string, opts ...*server.Option) (*Client, error) {
	return dialTimeout(NewClient, network, address, opts...)
}

func (client *Client) send(call *Call) {
	// 防止发送请求错乱
	client.sending.Lock()
	defer client.sending.Unlock()

	// 注册
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// 处理header
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Err = ""

	// 序列化，发送请求

	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

//Go 是一个异步接口，返回 call 实例
func (client *Client) Go(serverMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client:done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serverMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))

	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client:调用错误：" + ctx.Err().Error())
	case v := <-call.Done:
		return v.Error
	}
}

type clientResult struct {
	client *Client
	err    error
}
type newClientFunc func(conn net.Conn, opt *server.Option) (client *Client, err error)

func dialTimeout(f newClientFunc, network, address string, opts ...*server.Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	// 关闭连接
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan clientResult)
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{client: client, err: err}

	}()
	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client:网络连接超时")
	case result := <-ch:
		return result.client, result.err
	}

}

const (
	connected        = "200 Connected to Gee RPC"
	defaultRPCPath   = "/_geeprc_"
	defaultDebugPath = "/debug/geerpc"
)

func NewHTTPClient(conn net.Conn, opt *server.Option) (*Client, error) {
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRPCPath))

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}

func DialHTTP(network, address string, opts ...*server.Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, address, opts...)
}

func XDial(rpcAddr string, opts ...*server.Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		// tcp, unix or other transport protocol
		return Dial(protocol, addr, opts...)
	}
}
