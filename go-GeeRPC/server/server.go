package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-GeeRPC/encode"
	"go/ast"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const MagicNumber = 0x3bef5c

// GeeRPC 客户端固定采用 JSON 编码 Option，后续的 header 和 body 的编码方式由 Option 中的 CodeType 指定，服务端首先使用 JSON 解码 Option，然后通过 Option 得 CodeType 解码剩余的内容。即报文将以这样的形式发送：
type Option struct {
	MagicNumber    int
	CodecType      encode.Type
	ConnectTimeout time.Duration // 0代表没有限制
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      encode.GobType,
	ConnectTimeout: time.Second * 9999999,
	HandleTimeout:  time.Second * 9999999,
}

type Server struct {
	ServiceMap sync.Map
}

func (server *Server) Register(rcvr interface{}) error {
	s := NewService(rcvr)
	if _, dup := server.ServiceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc :service already defined:" + s.name)
	}
	return nil
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		//没有找到.
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.ServiceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.Method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rcp server: accept error", err)
			return
		}
		go s.ServerConn(conn)
	}
}

//检查 MagicNumber 和 CodeType 的值是否正确。然后根据 CodeType 得到对应的消息编解码器，接下来的处理交给 serverCodec
func (server *Server) ServerConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := encode.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn), &opt)

}

var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc encode.Codec, opt *Option) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Err = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout)
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h            *encode.Header // header of request
	argv, replyv reflect.Value  // argv and replyv of request
	mtype        *methodType
	svc          *service
}

func (server *Server) readRequestHeader(cc encode.Codec) (*encode.Header, error) {
	var h encode.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc encode.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}

	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.NewArgv()
	req.replyv = req.mtype.NewReplyv()
	// 确保返回值是个指针，读取body的结果是个指针
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

func (s *Server) sendResponse(cc encode.Codec, h *encode.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server:weite response error", err)
	}
}
func (server *Server) handleRequest(cc encode.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	send := make(chan struct{})
	go func() {
		err := req.svc.Call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Err = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			return
		}
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		send <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-send
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Err = fmt.Sprintf("rpc server: 请求处理超时: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-send
	}

}

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	NumCall   uint64
}

func (receiver *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&receiver.NumCall) //原子性
}
func (receiver *methodType) NewArgv() reflect.Value {
	var argv reflect.Value
	if receiver.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(receiver.ArgType.Elem())
	} else {
		argv = reflect.New(receiver.ArgType).Elem()
	}
	return argv
}

func (receiver *methodType) NewReplyv() reflect.Value {
	replyv := reflect.New(receiver.ReplyType.Elem())
	switch receiver.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(receiver.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(receiver.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	Method map[string]*methodType
}

func (s *service) registerMethods() {
	s.Method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			// 过滤两个导出或内置类型的入参（反射时为 3 个，第 0 个是自身，类似于 python 的 self，java 中的 this）
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			// 过滤 返回值有且只有 1 个，类型为 error
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.Method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
func (s *service) Call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.NumCall, 1)
	f := m.method.Func
	returnValue := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValue[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func NewService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

const (
	connected        = "200 Connected to Gee RPC"
	defaultRPCPath   = "/_geeprc_"
	defaultDebugPath = "/debug/geerpc"
)

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServerConn(conn)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
func (server *Server) HandleHTTP() {
	http.Handle(defaultRPCPath, server)
	http.Handle(defaultDebugPath, debugHTTP{server})
	log.Println("rpc server debug path:", defaultDebugPath)
}
