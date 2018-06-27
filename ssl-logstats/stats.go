package main

import (
	"github.com/RoboCup-SSL/ssl-go-tools/sslproto"
	"io"
)

type FrameProcessor interface {
	ProcessDetection(*sslproto.LogMessage, *sslproto.SSL_DetectionFrame)
	ProcessReferee(*sslproto.LogMessage, *sslproto.SSL_Referee)
	Init(logFile string) error
	io.Closer
}
