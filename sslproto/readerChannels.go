package sslproto

func (l *LogReader) CreateVisionWrapperChannel(channel chan *SSL_WrapperPacket) {
	logMessageChannel := make(chan *LogMessage, 100)
	go l.CreateLogMessageChannel(logMessageChannel)

	for logMessage := range logMessageChannel {
		if logMessage.MessageType == MESSAGE_SSL_VISION_2014 {
			visionMsg := logMessage.ParseVisionWrapper()
			channel <- visionMsg
		}
	}
	close(channel)
	return
}

func (l *LogReader) CreateVisionDetectionChannel(channel chan *SSL_DetectionFrame) {
	logMessageChannel := make(chan *LogMessage, 100)
	go l.CreateLogMessageChannel(logMessageChannel)

	for logMessage := range logMessageChannel {
		if logMessage.MessageType == MESSAGE_SSL_VISION_2014 {
			visionMsg := logMessage.ParseVisionWrapper()
			if visionMsg.Detection != nil {
				channel <- visionMsg.Detection
			}
		}
	}
	close(channel)
	return
}

func (l *LogReader) CreateRefereeChannel(channel chan *SSL_Referee) {
	logMessageChannel := make(chan *LogMessage, 100)
	go l.CreateLogMessageChannel(logMessageChannel)

	for logMessage := range logMessageChannel {
		if logMessage.MessageType == MESSAGE_SSL_REFBOX_2013 {
			refereeMsg := logMessage.ParseReferee()
			channel <- refereeMsg
		}
	}
	close(channel)
	return
}

func (l *LogReader) CreateLogMessageChannel(channel chan *LogMessage) (err error) {
	for l.HasMessage() {
		msg, err := l.ReadMessage()
		if err != nil {
			break
		}

		channel <- msg
	}
	close(channel)
	return
}
