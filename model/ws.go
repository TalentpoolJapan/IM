package model

func (m *Model) WsUnregister(s *MemInitUser) {
	m.MemRemoveConnByTouser(s.Touser, s.SessionId)
}
