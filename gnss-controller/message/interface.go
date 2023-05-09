package message

type UbxMessageHandler interface {
	HandleUbxMessage(interface{}) error
}
