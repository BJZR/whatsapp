func (s *Server) handleSendOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	jid, err := whatsmeow.ParseJID(req.Phone + "@s.whatsapp.net")
	if err != nil {
		http.Error(w, "Número de teléfono inválido", http.StatusBadRequest)
		return
	}

	// Usando la estructura nativa de whatsmeow para texto plano
	_, err = s.waClient.SendMessage(context.Background(), jid, &whatsmeow.Proto.Message{
		Conversation: proto.String(req.Message),
	})

	if err != nil {
		http.Error(w, "Error al enviar mensaje de WhatsApp", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Enviado"}`))
}
