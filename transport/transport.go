package transport

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/libp2p/go-libp2p/core/network"
)

func SendBytes(payload []byte, s network.Stream) error {

	err := binary.Write(s, binary.LittleEndian, int64(len(payload)))

	if err != nil {
		return err
	}

	s.Write(payload)

	return nil
}

func ReceiveBytes(s network.Stream) ([]byte, error) {
	buf := bufio.NewReader(s)

	var len int64
	err := binary.Read(s, binary.LittleEndian, &len)

	if err != nil {
		return nil, err
	}

	payload := make([]byte, len)
	_, err = io.ReadFull(buf, payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
