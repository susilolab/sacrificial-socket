/*
A complex web app example that implements ssgrpc for synchronizing multiple Sacrificial Socket instances
*/
package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/raz-varren/log"
	"github.com/susilolab/sacrificial-socket"
	"github.com/susilolab/sacrificial-socket/backend/ssgrpc"
	//"time"
)

type socketcast struct {
	SocketID string `json:"socketId"`
	Data     string `json:"data"`
}

type roomcast struct {
	Room string `json:"room"`
	Data string `json:"data"`
}

type message struct {
	Message string `json:"message"`
}

var (
	webPort      = flag.String("webport", ":8081", "host:port number used for webpage and socket connections")
	key          = flag.String("key", "./keys/snakeoil.key", "tls key used for grpc peer connections")
	cert         = flag.String("cert", "./keys/snakeoil.crt", "tls cert used for grpc peer connections")
	peerList     = flag.String("peers", "", "comma separated list of peerCN@host:port peers for grpc connections. if peerCN is not provided host will be used as the peer Common Name")
	sharedKey    = flag.String("sharedkey", "insecuresharedkeystring", "the shared key used to sign HMAC-SHA256 JWTs for authenticating grpc peers. must be the same on all peers")
	grpcHostPort = flag.String("grpchostport", ":30001", "listen host:port for grpc peer connections")
	insecure     = flag.Bool("insecure", false, "if the insecure flag is set, no tls or shared key authentication will be used in the grpc peer connections. insecure should not be used on production instances")
)

func main() {
	flag.Parse()

	log.SetDefaultLogger(log.NewLogger(os.Stdout, log.LogLevelDbg))

	s := ss.NewServer()

	s.On("echo", Echo)
	s.On("echobin", EchoBin)
	s.On("echojson", EchoJSON)
	s.On("join", Join)
	s.On("leave", Leave)
	s.On("socketcast", Socketcast)
	s.On("socketcastbin", SocketcastBin)
	s.On("socketcastjson", SocketcastJSON)
	s.On("roomcast", Roomcast)
	s.On("roomcastbin", RoomcastBin)
	s.On("roomcastjson", RoomcastJSON)
	s.On("broadcast", Broadcast)
	s.On("broadcastbin", BroadcastBin)
	s.On("broadcastjson", BroadcastJSON)

	if *peerList == "" {
		log.Err.Println("must provide peers to connect to")
		flag.PrintDefaults()
		return
	}

	peers := strings.Split(*peerList, ",")

	var b ss.MultihomeBackend

	if *insecure {
		b = ssgrpc.NewInsecureBackend(*grpcHostPort, peers)
	} else {
		b = ssgrpc.NewBackend(*key, *cert, *grpcHostPort, []byte(*sharedKey), peers)
	}

	s.SetMultihomeBackend(b)

	c := make(chan bool)
	s.EnableSignalShutdown(c)

	go func() {
		<-c
		os.Exit(0)
	}()

	http.Handle("/socket", s)
	http.Handle("/", http.FileServer(http.Dir("webroot")))

	var err error
	if *insecure {
		err = http.ListenAndServe(*webPort, nil)
	} else {
		err = http.ListenAndServeTLS(*webPort, *cert, *key, nil)
	}

	if err != nil {
		log.Err.Fatalln(err)
	}
}

func Echo(s *ss.Socket, data []byte) {
	s.Emit("echo", string(data))
}

func EchoBin(s *ss.Socket, data []byte) {
	s.Emit("echobin", data)
}

func EchoJSON(s *ss.Socket, data []byte) {
	var m message
	err := json.Unmarshal(data, &m)
	check(err)

	s.Emit("echojson", m)
}

func Join(s *ss.Socket, data []byte) {
	d := string(data)
	s.Join(d)
	s.Emit("echo", "joined room:"+d)
}

func Leave(s *ss.Socket, data []byte) {
	d := string(data)
	s.Leave(d)
	s.Emit("echo", "left room:"+d)
}

func Socketcast(s *ss.Socket, data []byte) {
	var sc socketcast
	err := json.Unmarshal(data, &sc)
	check(err)

	s.Socketcast(sc.SocketID, "socketcast", sc.Data)
}

func SocketcastBin(s *ss.Socket, data []byte) {
	var sc socketcast
	err := json.Unmarshal(data, &sc)
	check(err)

	s.Socketcast(sc.SocketID, "socketcastbin", []byte(sc.Data))
}

func SocketcastJSON(s *ss.Socket, data []byte) {
	var sc socketcast
	err := json.Unmarshal(data, &sc)
	check(err)

	s.Socketcast(sc.SocketID, "socketcastjson", sc)
}

func Roomcast(s *ss.Socket, data []byte) {
	var r roomcast
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, "roomcast", r.Data)
}

func RoomcastBin(s *ss.Socket, data []byte) {
	var r roomcast
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, "roomcastbin", []byte(r.Data))
}

func RoomcastJSON(s *ss.Socket, data []byte) {
	var r roomcast
	err := json.Unmarshal(data, &r)
	check(err)

	s.Roomcast(r.Room, "roomcastjson", r)
}

func Broadcast(s *ss.Socket, data []byte) {
	s.Broadcast("broadcast", string(data))
}

func BroadcastBin(s *ss.Socket, data []byte) {
	s.Broadcast("broadcastbin", data)
}

func BroadcastJSON(s *ss.Socket, data []byte) {
	var m message
	err := json.Unmarshal(data, &m)
	check(err)

	s.Broadcast("broadcastjson", m)
}

func check(err error) {
	if err != nil {
		log.Err.Println(err)
	}
}
