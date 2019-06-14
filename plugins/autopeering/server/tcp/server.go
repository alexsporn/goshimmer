package tcp

import (
	"math"
	"net"
	"strconv"

	"github.com/iotaledger/goshimmer/packages/daemon"
	"github.com/iotaledger/goshimmer/packages/events"
	"github.com/iotaledger/goshimmer/packages/network"
	"github.com/iotaledger/goshimmer/packages/network/tcp"
	"github.com/iotaledger/goshimmer/packages/node"
	"github.com/iotaledger/goshimmer/plugins/autopeering/parameters"
	"github.com/iotaledger/goshimmer/plugins/autopeering/types/ping"
	"github.com/iotaledger/goshimmer/plugins/autopeering/types/request"
	"github.com/iotaledger/goshimmer/plugins/autopeering/types/response"
	"github.com/pkg/errors"
)

var server = tcp.NewServer()

func ConfigureServer(plugin *node.Plugin) {
	server.Events.Connect.Attach(events.NewClosure(HandleConnection))
	server.Events.Error.Attach(events.NewClosure(func(err error) {
		plugin.LogFailure("error in tcp server: " + err.Error())
	}))
	server.Events.Start.Attach(events.NewClosure(func() {
		if *parameters.ADDRESS.Value == "0.0.0.0" {
			plugin.LogSuccess("Starting TCP Server (port " + strconv.Itoa(*parameters.PORT.Value) + ") ... done")
		} else {
			plugin.LogSuccess("Starting TCP Server (" + *parameters.ADDRESS.Value + ":" + strconv.Itoa(*parameters.PORT.Value) + ") ... done")
		}
	}))
	server.Events.Shutdown.Attach(events.NewClosure(func() {
		plugin.LogSuccess("Stopping TCP Server ... done")
	}))
}

func RunServer(plugin *node.Plugin) {
	daemon.BackgroundWorker(func() {
		if *parameters.ADDRESS.Value == "0.0.0.0" {
			plugin.LogInfo("Starting TCP Server (port " + strconv.Itoa(*parameters.PORT.Value) + ") ...")
		} else {
			plugin.LogInfo("Starting TCP Server (" + *parameters.ADDRESS.Value + ":" + strconv.Itoa(*parameters.PORT.Value) + ") ...")
		}

		server.Listen(*parameters.PORT.Value)
	})
}

func ShutdownServer(plugin *node.Plugin) {
	plugin.LogInfo("Stopping TCP Server ...")

	server.Shutdown()
}

func HandleConnection(conn *network.ManagedConnection) {
	conn.SetTimeout(IDLE_TIMEOUT)

	var connectionState = STATE_INITIAL
	var receiveBuffer []byte
	var offset int

	conn.Events.ReceiveData.Attach(events.NewClosure(func(data []byte) {
		ProcessIncomingPacket(&connectionState, &receiveBuffer, conn, data, &offset)
	}))

	go conn.Read(make([]byte, int(math.Max(ping.MARSHALLED_TOTAL_SIZE, math.Max(request.MARSHALLED_TOTAL_SIZE, response.MARSHALLED_TOTAL_SIZE)))))
}

func ProcessIncomingPacket(connectionState *byte, receiveBuffer *[]byte, conn *network.ManagedConnection, data []byte, offset *int) {
	if *connectionState == STATE_INITIAL {
		var err error
		if *connectionState, *receiveBuffer, err = parsePackageHeader(data); err != nil {
			Events.Error.Trigger(conn.RemoteAddr().(*net.TCPAddr).IP, err)

			conn.Close()

			return
		}

		*offset = 0

		switch *connectionState {
		case STATE_REQUEST:
			*receiveBuffer = make([]byte, request.MARSHALLED_TOTAL_SIZE)
		case STATE_RESPONSE:
			*receiveBuffer = make([]byte, response.MARSHALLED_TOTAL_SIZE)
		case STATE_PING:
			*receiveBuffer = make([]byte, ping.MARSHALLED_TOTAL_SIZE)
		}
	}

	switch *connectionState {
	case STATE_REQUEST:
		processIncomingRequestPacket(connectionState, receiveBuffer, conn, data, offset)
	case STATE_RESPONSE:
		processIncomingResponsePacket(connectionState, receiveBuffer, conn, data, offset)
	case STATE_PING:
		processIncomingPingPacket(connectionState, receiveBuffer, conn, data, offset)
	}
}

func parsePackageHeader(data []byte) (byte, []byte, error) {
	var connectionState byte
	var receiveBuffer []byte

	switch data[0] {
	case request.MARSHALLED_PACKET_HEADER:
		receiveBuffer = make([]byte, request.MARSHALLED_TOTAL_SIZE)

		connectionState = STATE_REQUEST
	case response.MARHSALLED_PACKET_HEADER:
		receiveBuffer = make([]byte, response.MARSHALLED_TOTAL_SIZE)

		connectionState = STATE_RESPONSE
	case ping.MARSHALLED_PACKET_HEADER:
		receiveBuffer = make([]byte, ping.MARSHALLED_TOTAL_SIZE)

		connectionState = STATE_PING
	default:
		return 0, nil, errors.New("invalid package header")
	}

	return connectionState, receiveBuffer, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func processIncomingRequestPacket(connectionState *byte, receiveBuffer *[]byte, conn *network.ManagedConnection, data []byte, offset *int) {
	remainingCapacity := min(request.MARSHALLED_TOTAL_SIZE-*offset, len(data))

	copy((*receiveBuffer)[*offset:], data[:remainingCapacity])

	if *offset+len(data) < request.MARSHALLED_TOTAL_SIZE {
		*offset += len(data)
	} else {
		if req, err := request.Unmarshal(*receiveBuffer); err != nil {
			Events.Error.Trigger(conn.RemoteAddr().(*net.TCPAddr).IP, err)

			conn.Close()

			return
		} else {
			req.Issuer.Conn = conn
			req.Issuer.Address = conn.RemoteAddr().(*net.TCPAddr).IP

			req.Issuer.Conn.Events.Close.Attach(events.NewClosure(func() {
				req.Issuer.Conn = nil
			}))

			Events.ReceiveRequest.Trigger(req)
		}

		*connectionState = STATE_INITIAL

		if *offset+len(data) > request.MARSHALLED_TOTAL_SIZE {
			ProcessIncomingPacket(connectionState, receiveBuffer, conn, data[remainingCapacity:], offset)
		}
	}
}

func processIncomingResponsePacket(connectionState *byte, receiveBuffer *[]byte, conn *network.ManagedConnection, data []byte, offset *int) {
	remainingCapacity := min(response.MARSHALLED_TOTAL_SIZE-*offset, len(data))

	copy((*receiveBuffer)[*offset:], data[:remainingCapacity])

	if *offset+len(data) < response.MARSHALLED_TOTAL_SIZE {
		*offset += len(data)
	} else {
		if res, err := response.Unmarshal(*receiveBuffer); err != nil {
			Events.Error.Trigger(conn.RemoteAddr().(*net.TCPAddr).IP, err)

			conn.Close()

			return
		} else {
			res.Issuer.Conn = conn
			res.Issuer.Address = conn.RemoteAddr().(*net.TCPAddr).IP

			res.Issuer.Conn.Events.Close.Attach(events.NewClosure(func() {
				res.Issuer.Conn = nil
			}))

			Events.ReceiveResponse.Trigger(res)
		}

		*connectionState = STATE_INITIAL

		if *offset+len(data) > response.MARSHALLED_TOTAL_SIZE {
			ProcessIncomingPacket(connectionState, receiveBuffer, conn, data[remainingCapacity:], offset)
		}
	}
}

func processIncomingPingPacket(connectionState *byte, receiveBuffer *[]byte, conn *network.ManagedConnection, data []byte, offset *int) {
	remainingCapacity := min(ping.MARSHALLED_TOTAL_SIZE-*offset, len(data))

	copy((*receiveBuffer)[*offset:], data[:remainingCapacity])

	if *offset+len(data) < ping.MARSHALLED_TOTAL_SIZE {
		*offset += len(data)
	} else {
		if ping, err := ping.Unmarshal(*receiveBuffer); err != nil {
			Events.Error.Trigger(conn.RemoteAddr().(*net.TCPAddr).IP, err)

			conn.Close()

			return
		} else {
			ping.Issuer.Conn = conn
			ping.Issuer.Address = conn.RemoteAddr().(*net.TCPAddr).IP

			ping.Issuer.Conn.Events.Close.Attach(events.NewClosure(func() {
				ping.Issuer.Conn = nil
			}))

			Events.ReceivePing.Trigger(ping)
		}

		*connectionState = STATE_INITIAL

		if *offset+len(data) > ping.MARSHALLED_TOTAL_SIZE {
			ProcessIncomingPacket(connectionState, receiveBuffer, conn, data[remainingCapacity:], offset)
		}
	}
}
