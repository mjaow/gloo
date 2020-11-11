package protoutils

import (
	"reflect"
	"strings"

	proto2 "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	"github.com/rotisserie/eris"
)

type MultiAnyResolver struct{}

func (m *MultiAnyResolver) Resolve(typeUrl string) (proto.Message, error) {
	messageType := typeUrl
	if slash := strings.LastIndex(typeUrl, "/"); slash >= 0 {
		messageType = messageType[slash+1:]
	}
	var mt reflect.Type
	mt = proto.MessageType(messageType)
	if mt == nil {
		mt = proto2.MessageType(messageType)
		if mt == nil {
			return nil, eris.Errorf("unknown message type %q", messageType)
		}
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil

}

