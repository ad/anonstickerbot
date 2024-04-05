package sender

import "fmt"

func (sender *Sender) SendResult(s SendResult) error {
	if s.Error != nil {
		sender.lgr.Error(fmt.Sprintf("message id %d sent to %d error %q: %s", s.MessageID, s.ChatID, s.Error, s.Msg))
		return s.Error
	}

	sender.lgr.Info(fmt.Sprintf("message id %d sent to %d: %s", s.MessageID, s.ChatID, s.Msg))

	return nil
}
