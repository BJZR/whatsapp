package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "modernc.org/sqlite"
)

type Server struct {
	waClient *whatsmeow.Client
}

type OrderRequest struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

func main() {
	dbPath := "whatsapp.db"
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath), dbLog)
	if err != nil {
		log.Fatalf("Error al conectar base de datos: %v", err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		log.Fatalf("Error al obtener dispositivo: %v", err)
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	srv := &Server{waClient: client}

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.QR:
			fmt.Println("Escanea este código QR para vincular WhatsApp (Logs de Render):")
			q, _ := qrcode.New(v.Code, qrcode.Medium)
			fmt.Println(q.ToString(true))
		case *events.Connected:
			fmt.Println("WhatsApp conectado correctamente.")
		}
	})

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			log.Fatalf("Error al conectar: %v", err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				q, _ := qrcode.New(evt.Code, qrcode.Medium)
				fmt.Println(q.ToString(true))
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			log.Fatalf("Error al conectar sesión existente: %v", err)
		}
	}

	// Keep-alive routine para evitar inactividad en Render
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			fmt.Println("[Keep-Alive] Servicio activo:", time.Now().Format(time.RFC3339))
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/send", srv.handleSendOrder)

	fmt.Printf("Servidor corriendo en puerto %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

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

	_, err = s.waClient.SendMessage(context.Background(), jid, &waE2E.Message{
		Conversation: proto.String(req.Message),
	})

	if err != nil {
		http.Error(w, "Error al enviar mensaje de WhatsApp", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Enviado"}`))
}
