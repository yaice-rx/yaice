package network

type IPacket interface {
	GetHeadLen() int                                 //获取包头长度方法
	Pack(msg IMessage) []byte                        //封包方法
	Unpack(buff []byte) ([]byte, []byte, int, error) //拆包方法
}
