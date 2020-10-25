package server

import (
	"fmt"
	"html/template"
	"net/http"
)

const debugText = `<html>
	<body>
	<title>GeeRPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
			<th align=center>Method</th><th align=center>Calls</th>
			{{range  $k, $v := .Method}}
				<tr>
				<td align=left font=fixed>{{$k}}({{$v.ArgType}}, {{$v.ReplyType}}) error</td>
				<td align=center>{{$v.NumCalls}}</td>
				</tr>
			{{end}}
		</table>
		
	{{end}}
	</body>
	</html>`

type debugHTTP struct {
	*Server
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}

var debug = template.Must(template.New("RPC debug").Parse(debugText))

// Runs at /debug/geerpc
func (server debugHTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Build a sorted version of the data.

	var services []debugService
	server.ServiceMap.Range(func(namei, svci interface{}) bool {
		svc := svci.(*service)
		services = append(services, debugService{
			Name:   namei.(string),
			Method: svc.Method,
		})
		return true
	})
	err := debug.Execute(w, services)
	if err != nil {
		_, _ = fmt.Fprintln(w, "rpc: error executing template:", err.Error())
	}
}
