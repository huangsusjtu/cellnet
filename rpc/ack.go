package rpc

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proto/coredef"
	"github.com/davyxu/cellnet/socket"
)

type Response interface {
	Feedback(interface{})
}

type response struct {
	ses cellnet.Session
	req *coredef.RemoteCallREQ
}

func (self *response) Feedback(msg interface{}) {

	pkt, _ := cellnet.BuildPacket(msg)

	self.ses.Send(&coredef.RemoteCallACK{
		MsgID:  pkt.MsgID,
		Data:   pkt.Data,
		CallID: self.req.CallID,
	})
}

func (self *response) ContextID() int {
	return int(self.req.MsgID)
}

func InstallServer(p cellnet.Peer) {

	// 服务端
	socket.RegisterSessionMessage(p, coredef.RemoteCallREQ{}, func(content interface{}, ses cellnet.Session) {
		msg := content.(*coredef.RemoteCallREQ)

		// 客户端发过来的请求消息注入到回调中
		p.CallData(&response{
			ses: ses,
			req: msg,
		})

	})

}

// 注册连接消息
func RegisterMessage(eq cellnet.EventQueue, msgIns interface{}, userHandler func(Response, interface{})) {

	msgMeta := cellnet.NewMessageMeta(msgIns)

	// 将消息注册到mapper中, 提供反射用
	socket.MapNameID(msgMeta.Name, msgMeta.ID)

	eq.RegisterCallback(msgMeta.ID, func(data interface{}) {

		if ev, ok := data.(*response); ok {

			rawMsg, err := cellnet.ParsePacket(&cellnet.Packet{
				MsgID: ev.req.MsgID,
				Data:  ev.req.Data,
			}, msgMeta.Type)

			if err != nil {
				log.Errorln("unmarshaling error:", err)
				return
			}

			userHandler(ev, rawMsg)

		}

	})

}
